package indexer

import (
    "encoding/json"
    "fmt"
    "strings"

    "github.com/ethereum/go-ethereum/accounts/abi"
    abiassets "github.com/paxeer/offchain-server/abiassets"
)

type ABIs struct {
	Factory abi.ABI
	Pool    abi.ABI
	Oracle  abi.ABI

	// Event IDs cache
	SigPoolCreated    string
	SigPriceUpdate    string
	SigSync           string
	SigSwap           string
	SigAddLiquidity   string
	SigRemoveLiquidity string
	SigCollectCreatorFees string
	SigInitialTokenSeeded string
	SigOracleUpdate   string
}

func loadABI(name string) (abi.ABI, error) {
    b, err := abiassets.Files.ReadFile("abis/" + name + ".json")
    if err != nil {
        return abi.ABI{}, err
    }
    // The JSON is event-only or partial; abi.JSON supports plain arrays.
    content := strings.TrimSpace(string(b))
    if strings.HasPrefix(content, "[") {
        var tmp []any
        if err := json.Unmarshal(b, &tmp); err == nil {
            return abi.JSON(strings.NewReader(content))
        }
    }
    return abi.JSON(strings.NewReader(content))
}

func LoadABIs() (*ABIs, error) {
	f, err := loadABI("LaunchpadFactory")
	if err != nil {
		return nil, fmt.Errorf("load factory abi: %w", err)
	}
	p, err := loadABI("LaunchPool")
	if err != nil {
		return nil, fmt.Errorf("load pool abi: %w", err)
	}
	o, err := loadABI("LaunchPoolOracle")
	if err != nil {
		return nil, fmt.Errorf("load oracle abi: %w", err)
	}
	out := &ABIs{Factory: f, Pool: p, Oracle: o}
	out.SigPoolCreated = f.Events["PoolCreated"].ID.String()
	out.SigPriceUpdate = p.Events["PriceUpdate"].ID.String()
	out.SigSync = p.Events["Sync"].ID.String()
	out.SigSwap = p.Events["Swap"].ID.String()
	out.SigAddLiquidity = p.Events["AddLiquidity"].ID.String()
	out.SigRemoveLiquidity = p.Events["RemoveLiquidity"].ID.String()
	out.SigCollectCreatorFees = p.Events["CollectCreatorFees"].ID.String()
	out.SigInitialTokenSeeded = p.Events["InitialTokenSeeded"].ID.String()
	out.SigOracleUpdate = o.Events["OracleUpdate"].ID.String()
	return out, nil
}
