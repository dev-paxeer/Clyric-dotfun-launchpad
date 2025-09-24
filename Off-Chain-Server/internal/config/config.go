package config

import (
    "fmt"
    "os"

    "gopkg.in/yaml.v3"
)

// Config holds application configuration loaded from YAML and env overrides.
type Config struct {
	RPC struct {
		WS   string `yaml:"ws"`
		HTTP string `yaml:"http"`
	} `yaml:"rpc"`
	Contracts struct {
		Factory string `yaml:"factory"`
		USDC    string `yaml:"usdc"`
	} `yaml:"contracts"`
	Indexer struct {
		StartBlock    uint64 `yaml:"startBlock"`
		Confirmations uint64 `yaml:"confirmations"`
		BatchSize     uint64 `yaml:"batchSize"`
	} `yaml:"indexer"`
	Postgres struct {
		DSN string `yaml:"dsn"`
	} `yaml:"postgres"`
}

func Load(path string) (*Config, error) {
    b, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var c Config
    if err := yaml.Unmarshal(b, &c); err != nil {
        return nil, err
    }
	// Env overrides
	if v := os.Getenv("PAXEER_RPC_WS"); v != "" {
		c.RPC.WS = v
	}
	if v := os.Getenv("PAXEER_RPC_HTTP"); v != "" {
		c.RPC.HTTP = v
	}
	if v := os.Getenv("PAXEER_FACTORY"); v != "" {
		c.Contracts.Factory = v
	}
	if v := os.Getenv("PAXEER_USDC"); v != "" {
		c.Contracts.USDC = v
	}
	if v := os.Getenv("PAXEER_DB_DSN"); v != "" {
		c.Postgres.DSN = v
	}
	if v := os.Getenv("PAXEER_START_BLOCK"); v != "" {
		if parsed, perr := parseUint(v); perr == nil {
			c.Indexer.StartBlock = parsed
		}
	}
	if v := os.Getenv("PAXEER_CONFIRMATIONS"); v != "" {
		if parsed, perr := parseUint(v); perr == nil {
			c.Indexer.Confirmations = parsed
		}
	}
	if v := os.Getenv("PAXEER_BATCH_SIZE"); v != "" {
		if parsed, perr := parseUint(v); perr == nil {
			c.Indexer.BatchSize = parsed
		}
	}
	// defaults
	if c.Indexer.Confirmations == 0 {
		c.Indexer.Confirmations = 2
	}
	if c.Indexer.BatchSize == 0 {
		c.Indexer.BatchSize = 5000
	}
    return &c, nil
}

func parseUint(s string) (uint64, error) {
    var (
        v   uint64
        err error
    )
    // allow hex or dec
    if len(s) > 2 && (s[:2] == "0x" || s[:2] == "0X") {
        _, err = fmt.Sscanf(s, "%x", &v)
    } else {
        _, err = fmt.Sscanf(s, "%d", &v)
    }
    return v, err
}
