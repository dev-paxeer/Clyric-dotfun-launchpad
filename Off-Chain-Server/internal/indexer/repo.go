package indexer

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct{
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

func (r *Repo) UpsertPool(ctx context.Context, poolAddr, tokenAddr, oracleAddr string, createdBlock int64, createdTx string, createdTime *time.Time) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO pools(pool_address, token_address, oracle_address, created_block, created_tx, created_time)
		VALUES($1,$2,$3,$4,$5,$6)
		ON CONFLICT(pool_address) DO UPDATE SET token_address = EXCLUDED.token_address, oracle_address = EXCLUDED.oracle_address
	`, poolAddr, tokenAddr, oracleAddr, createdBlock, createdTx, createdTime)
	return err
}

func (r *Repo) UpdatePoolSnapshot(ctx context.Context, poolAddr string, reserveUSDC, reserveToken, spotX18, floorX18 *string) error {
    _, err := r.pool.Exec(ctx, `
        UPDATE pools SET
            reserve_usdc = COALESCE($2::numeric, reserve_usdc),
            reserve_token = COALESCE($3::numeric, reserve_token),
            spot_x18     = COALESCE($4::numeric, spot_x18),
            floor_x18    = COALESCE($5::numeric, floor_x18)
        WHERE pool_address = $1
    `, poolAddr, reserveUSDC, reserveToken, spotX18, floorX18)
    return err
}

func (r *Repo) InsertPriceUpdate(ctx context.Context, poolAddr, priceX18, floorX18, txHash string, blockNumber int64, logIndex int, blockTime *time.Time, confirmed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO price_updates(pool_address, price_x18, floor_x18, block_number, tx_hash, log_index, block_time, confirmed)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
	`, poolAddr, priceX18, floorX18, blockNumber, txHash, logIndex, blockTime, confirmed)
	return err
}

func (r *Repo) InsertReserves(ctx context.Context, poolAddr, reserveUSDC, reserveToken, txHash string, blockNumber int64, logIndex int, blockTime *time.Time, confirmed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO reserves(pool_address, reserve_usdc, reserve_token, block_number, tx_hash, log_index, block_time, confirmed)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
	`, poolAddr, reserveUSDC, reserveToken, blockNumber, txHash, logIndex, blockTime, confirmed)
	return err
}

func (r *Repo) InsertSwap(ctx context.Context, poolAddr, sender string, usdcToToken bool, amountIn, amountOut, recipient, txHash string, blockNumber int64, logIndex int, blockTime *time.Time, confirmed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO swaps(pool_address, sender, usdc_to_token, amount_in, amount_out, recipient, block_number, tx_hash, log_index, block_time, confirmed)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, poolAddr, sender, usdcToToken, amountIn, amountOut, recipient, blockNumber, txHash, logIndex, blockTime, confirmed)
	return err
}

func (r *Repo) InsertLiquidity(ctx context.Context, poolAddr, eventType, provider, amountUSDC, amountToken, lpAmount, txHash string, blockNumber int64, logIndex int, blockTime *time.Time, confirmed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO liquidity_events(pool_address, event_type, provider, amount_usdc, amount_token, lp_amount, block_number, tx_hash, log_index, block_time, confirmed)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, poolAddr, eventType, provider, amountUSDC, amountToken, lpAmount, blockNumber, txHash, logIndex, blockTime, confirmed)
	return err
}

func (r *Repo) InsertOracleUpdate(ctx context.Context, poolAddr, priceCumulative, txHash string, oracleTs int64, blockNumber int64, logIndex int, blockTime *time.Time, confirmed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO oracle_updates(pool_address, price_cumulative, oracle_timestamp, block_number, tx_hash, log_index, block_time, confirmed)
		VALUES($1,$2,$3,$4,$5,$6,$7,$8)
	`, poolAddr, priceCumulative, oracleTs, blockNumber, txHash, logIndex, blockTime, confirmed)
	return err
}

func (r *Repo) InsertCreatorFees(ctx context.Context, poolAddr, amountUSDC, txHash string, blockNumber int64, logIndex int, blockTime *time.Time, confirmed bool) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO creator_fees(pool_address, amount_usdc, block_number, tx_hash, log_index, block_time, confirmed)
		VALUES($1,$2,$3,$4,$5,$6,$7)
	`, poolAddr, amountUSDC, blockNumber, txHash, logIndex, blockTime, confirmed)
	return err
}

func (r *Repo) LookupPoolByOracle(ctx context.Context, oracleAddr string) (string, error) {
	var poolAddr string
	err := r.pool.QueryRow(ctx, `SELECT pool_address FROM pools WHERE oracle_address = $1`, oracleAddr).Scan(&poolAddr)
	if err == pgx.ErrNoRows {
		return "", nil
	}
	return poolAddr, err
}
