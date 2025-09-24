package indexer

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Indexer struct {
	HTTP *ethclient.Client
	WS   *ethclient.Client

	Factory common.Address

	ABIs *ABIs
	DB   *pgxpool.Pool
	Repo *Repo

	Confirmations uint64
	BatchSize     uint64

	mu       sync.RWMutex
	pools    map[common.Address]struct{}
	oracles  map[common.Address]common.Address // oracle -> pool
	lastHead uint64
}

func NewIndexer(httpURL, wsURL string, factory string, db *pgxpool.Pool, abis *ABIs, confirmations, batch uint64) (*Indexer, error) {
	cliHTTP, err := ethclient.Dial(httpURL)
	if err != nil {
		return nil, fmt.Errorf("dial http: %w", err)
	}
	var cliWS *ethclient.Client
	if wsURL != "" {
		if c, err := ethclient.Dial(wsURL); err == nil {
			cliWS = c
		} else {
			// WS optional: continue without it; we will use HTTP polling
			log.Printf("[warn] ws dial failed (%v), falling back to HTTP polling", err)
		}
	}
	ix := &Indexer{
		HTTP:         cliHTTP,
		WS:           cliWS,
		Factory:      common.HexToAddress(factory),
		ABIs:         abis,
		DB:           db,
		Repo:         NewRepo(db),
		Confirmations: confirmations,
		BatchSize:     batch,
		pools:        make(map[common.Address]struct{}),
		oracles:      make(map[common.Address]common.Address),
	}
	return ix, nil
}

func (ix *Indexer) Close() {
	if ix.HTTP != nil {
		ix.HTTP.Close()
	}
	if ix.WS != nil {
		ix.WS.Close()
	}
}

func (ix *Indexer) safeHead(ctx context.Context) (uint64, error) {
	h, err := ix.HTTP.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, err
	}
	if h.Number == nil {
		return 0, fmt.Errorf("nil head number")
	}
	head := h.Number.Uint64()
	if head < ix.Confirmations {
		return 0, nil
	}
	return head - ix.Confirmations, nil
}

func (ix *Indexer) ensurePool(pool, token, oracle common.Address, createdBlock uint64, txHash common.Hash, blockTime *time.Time) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	if _, ok := ix.pools[pool]; !ok {
		ix.pools[pool] = struct{}{}
	}
	ix.oracles[oracle] = pool
	_ = ix.Repo.UpsertPool(context.Background(), pool.Hex(), token.Hex(), oracle.Hex(), int64(createdBlock), txHash.Hex(), blockTime)
}

func (ix *Indexer) Backfill(ctx context.Context, start uint64) error {
	log.Printf("[backfill] starting from block %d", start)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		safe, err := ix.safeHead(ctx)
		if err != nil {
			return err
		}
		if start == 0 || start > safe {
			time.Sleep(2 * time.Second)
			continue
		}
		end := start + ix.BatchSize - 1
		if end > safe {
			end = safe
		}
		if err := ix.scanRange(ctx, start, end); err != nil {
			return err
		}
		start = end + 1
	}
}

func (ix *Indexer) scanRange(ctx context.Context, from, to uint64) error {
	log.Printf("[scan] %d -> %d", from, to)

	// 1) Factory: PoolCreated
	q := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Addresses: []common.Address{ix.Factory},
		Topics:    [][]common.Hash{{ix.ABIs.Factory.Events["PoolCreated"].ID}},
	}
	logs, err := ix.HTTP.FilterLogs(ctx, q)
	if err != nil {
		return fmt.Errorf("filter factory: %w", err)
	}
	for _, lg := range logs {
		if err := ix.handleFactoryLog(ctx, lg); err != nil {
			log.Printf("handle factory log err: %v", err)
		}
	}

	// gather known pools and oracles
	ix.mu.RLock()
	var poolAddrs []common.Address
	for a := range ix.pools { poolAddrs = append(poolAddrs, a) }
	var oracleAddrs []common.Address
	for a := range ix.oracles { oracleAddrs = append(oracleAddrs, a) }
	ix.mu.RUnlock()

	// 2) Pool events
	if len(poolAddrs) > 0 {
		pq := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(from)),
			ToBlock:   big.NewInt(int64(to)),
			Addresses: poolAddrs,
		}
		plogs, err := ix.HTTP.FilterLogs(ctx, pq)
		if err != nil {
			return fmt.Errorf("filter pools: %w", err)
		}
		for _, lg := range plogs {
			if err := ix.handlePoolLog(ctx, lg); err != nil {
				log.Printf("handle pool log err: %v", err)
			}
		}
	}

	// 3) Oracle updates
	if len(oracleAddrs) > 0 {
		oq := ethereum.FilterQuery{
			FromBlock: big.NewInt(int64(from)),
			ToBlock:   big.NewInt(int64(to)),
			Addresses: oracleAddrs,
		}
		ologs, err := ix.HTTP.FilterLogs(ctx, oq)
		if err != nil {
			return fmt.Errorf("filter oracles: %w", err)
		}
		for _, lg := range ologs {
			if err := ix.handleOracleLog(ctx, lg); err != nil {
				log.Printf("handle oracle log err: %v", err)
			}
		}
	}
    // track last processed block
    ix.mu.Lock()
    if to > ix.lastHead { ix.lastHead = to }
    ix.mu.Unlock()
	return nil
}

func (ix *Indexer) Subscribe(ctx context.Context) error {
	if ix.WS == nil {
		return fmt.Errorf("no ws client")
	}
	heads := make(chan *types.Header, 32)
	sub, err := ix.WS.SubscribeNewHead(ctx, heads)
	if err != nil {
		return fmt.Errorf("subscribe heads: %w", err)
	}
	defer sub.Unsubscribe()
	for {
		select {
		case err := <-sub.Err():
			return err
		case h := <-heads:
			if h == nil || h.Number == nil { continue }
			safe, err := ix.safeHead(ctx)
			if err != nil { return err }
			if safe == 0 { continue }
			ix.mu.Lock()
			if safe <= ix.lastHead {
				ix.mu.Unlock()
				continue
			}
			from := ix.lastHead + 1
			if from == 1 { from = safe }
			ix.lastHead = safe
			ix.mu.Unlock()
			if from <= safe {
				if err := ix.scanRange(ctx, from, safe); err != nil {
					log.Printf("scanRange live err: %v", err)
				}
			}
		}
	}
}

func (ix *Indexer) handleFactoryLog(ctx context.Context, lg types.Log) error {
	// Decode PoolCreated(token indexed, pool, oracle)
	e := ix.ABIs.Factory.Events["PoolCreated"]
	if lg.Topics[0] != e.ID { return nil }
	out := make(map[string]any)
	if err := ix.ABIs.Factory.UnpackIntoMap(out, e.Name, lg.Data); err != nil {
		return err
	}
	token := common.BytesToAddress(lg.Topics[1].Bytes())
	pool := out["pool"].(common.Address)
	oracle := out["oracle"].(common.Address)
	// block time
	var blkTime *time.Time
	h, err := ix.HTTP.HeaderByHash(ctx, lg.BlockHash)
	if err == nil && h != nil {
		t := time.Unix(int64(h.Time), 0).UTC()
		blkTime = &t
	}
	ix.ensurePool(pool, token, oracle, lg.BlockNumber, lg.TxHash, blkTime)
	return nil
}

func (ix *Indexer) handlePoolLog(ctx context.Context, lg types.Log) error {
	sig := lg.Topics[0].Hex()
	switch sig {
	case ix.ABIs.SigPriceUpdate:
		out := struct{ PriceX18, FloorX18 *big.Int }{}
		if err := unpack(ix.ABIs.Pool, "PriceUpdate", lg, &out); err != nil { return err }
		poolAddr := lg.Address.Hex()
		var blkTime *time.Time
		h, err := ix.HTTP.HeaderByHash(ctx, lg.BlockHash)
		if err == nil && h != nil { t := time.Unix(int64(h.Time), 0).UTC(); blkTime = &t }
		if err := ix.Repo.InsertPriceUpdate(ctx, poolAddr, out.PriceX18.String(), out.FloorX18.String(), lg.TxHash.Hex(), int64(lg.BlockNumber), int(lg.Index), blkTime, true); err != nil { return err }
		// also update snapshot (only spot/floor)
		spot := out.PriceX18.String()
		floor := out.FloorX18.String()
		_ = ix.Repo.UpdatePoolSnapshot(ctx, poolAddr, nil, nil, &spot, &floor)
	case ix.ABIs.SigSync:
		out := struct{ ReserveUSDC, ReserveToken *big.Int }{}
		if err := unpack(ix.ABIs.Pool, "Sync", lg, &out); err != nil { return err }
		poolAddr := lg.Address.Hex()
		var blkTime *time.Time
		h, err := ix.HTTP.HeaderByHash(ctx, lg.BlockHash)
		if err == nil && h != nil { t := time.Unix(int64(h.Time), 0).UTC(); blkTime = &t }
		if err := ix.Repo.InsertReserves(ctx, poolAddr, out.ReserveUSDC.String(), out.ReserveToken.String(), lg.TxHash.Hex(), int64(lg.BlockNumber), int(lg.Index), blkTime, true); err != nil { return err }
		// update snapshot (only reserves)
		rusdc := out.ReserveUSDC.String()
		rtok := out.ReserveToken.String()
		_ = ix.Repo.UpdatePoolSnapshot(ctx, poolAddr, &rusdc, &rtok, nil, nil)
	case ix.ABIs.SigSwap:
		out := struct{ AmountIn, AmountOut *big.Int; UsdcToToken bool }{}
		if err := unpack(ix.ABIs.Pool, "Swap", lg, &out); err != nil { return err }
		// indexed: sender (topics[1]), to (topics[2])
		sender := common.BytesToAddress(lg.Topics[1].Bytes()).Hex()
		to := common.BytesToAddress(lg.Topics[2].Bytes()).Hex()
		poolAddr := lg.Address.Hex()
		var blkTime *time.Time
		h, err := ix.HTTP.HeaderByHash(ctx, lg.BlockHash)
		if err == nil && h != nil { t := time.Unix(int64(h.Time), 0).UTC(); blkTime = &t }
		return ix.Repo.InsertSwap(ctx, poolAddr, sender, out.UsdcToToken, out.AmountIn.String(), out.AmountOut.String(), to, lg.TxHash.Hex(), int64(lg.BlockNumber), int(lg.Index), blkTime, true)
	case ix.ABIs.SigAddLiquidity:
		out := struct{ AmountUSDC, AmountToken, LpMinted *big.Int }{}
		if err := unpack(ix.ABIs.Pool, "AddLiquidity", lg, &out); err != nil { return err }
		provider := common.BytesToAddress(lg.Topics[1].Bytes()).Hex()
		var blkTime *time.Time
		h, err := ix.HTTP.HeaderByHash(ctx, lg.BlockHash)
		if err == nil && h != nil { t := time.Unix(int64(h.Time), 0).UTC(); blkTime = &t }
		return ix.Repo.InsertLiquidity(ctx, lg.Address.Hex(), "add", provider, out.AmountUSDC.String(), out.AmountToken.String(), out.LpMinted.String(), lg.TxHash.Hex(), int64(lg.BlockNumber), int(lg.Index), blkTime, true)
	case ix.ABIs.SigRemoveLiquidity:
		out := struct{ LpBurned, AmountUSDC, AmountToken *big.Int }{}
		if err := unpack(ix.ABIs.Pool, "RemoveLiquidity", lg, &out); err != nil { return err }
		provider := common.BytesToAddress(lg.Topics[1].Bytes()).Hex()
		var blkTime *time.Time
		h, err := ix.HTTP.HeaderByHash(ctx, lg.BlockHash)
		if err == nil && h != nil { t := time.Unix(int64(h.Time), 0).UTC(); blkTime = &t }
		return ix.Repo.InsertLiquidity(ctx, lg.Address.Hex(), "remove", provider, out.AmountUSDC.String(), out.AmountToken.String(), out.LpBurned.String(), lg.TxHash.Hex(), int64(lg.BlockNumber), int(lg.Index), blkTime, true)
	case ix.ABIs.SigCollectCreatorFees:
		out := struct{ AmountUSDC *big.Int }{}
		if err := unpack(ix.ABIs.Pool, "CollectCreatorFees", lg, &out); err != nil { return err }
		var blkTime *time.Time
		h, err := ix.HTTP.HeaderByHash(ctx, lg.BlockHash)
		if err == nil && h != nil { t := time.Unix(int64(h.Time), 0).UTC(); blkTime = &t }
		return ix.Repo.InsertCreatorFees(ctx, lg.Address.Hex(), out.AmountUSDC.String(), lg.TxHash.Hex(), int64(lg.BlockNumber), int(lg.Index), blkTime, true)
	}
	return nil
}

func (ix *Indexer) handleOracleLog(ctx context.Context, lg types.Log) error {
	if lg.Topics[0].Hex() != ix.ABIs.SigOracleUpdate { return nil }
	out := struct{ PriceCumulative *big.Int; Timestamp uint32 }{}
	if err := unpack(ix.ABIs.Oracle, "OracleUpdate", lg, &out); err != nil { return err }
	var blkTime *time.Time
	h, err := ix.HTTP.HeaderByHash(ctx, lg.BlockHash)
	if err == nil && h != nil { t := time.Unix(int64(h.Time), 0).UTC(); blkTime = &t }
	ix.mu.RLock()
	pool := ix.oracles[lg.Address]
	ix.mu.RUnlock()
	if (pool == common.Address{}) {
		// try db lookup
		if p, err := ix.Repo.LookupPoolByOracle(ctx, lg.Address.Hex()); err == nil && p != "" {
			pool = common.HexToAddress(p)
			ix.mu.Lock(); ix.oracles[lg.Address] = pool; ix.mu.Unlock()
		}
	}
	if (pool == common.Address{}) { return nil }
	return ix.Repo.InsertOracleUpdate(ctx, pool.Hex(), out.PriceCumulative.String(), lg.TxHash.Hex(), int64(out.Timestamp), int64(lg.BlockNumber), int(lg.Index), blkTime, true)
}

func unpack(contractABI abi.ABI, event string, lg types.Log, out any) error {
	// FilterLogs returns topics separated; data contains only non-indexed fields
	return contractABI.UnpackIntoInterface(out, event, lg.Data)
}

// PollForever periodically checks the safe head with HTTP and scans new ranges.
func (ix *Indexer) PollForever(ctx context.Context, interval time.Duration) error {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            safe, err := ix.safeHead(ctx)
            if err != nil {
                log.Printf("safeHead err: %v", err)
                continue
            }
            if safe == 0 {
                continue
            }
            ix.mu.Lock()
            from := ix.lastHead + 1
            if from == 1 {
                from = safe
            }
            ix.lastHead = safe
            ix.mu.Unlock()
            if from <= safe {
                if err := ix.scanRange(ctx, from, safe); err != nil {
                    log.Printf("poll scanRange err: %v", err)
                }
            }
        }
    }
}
