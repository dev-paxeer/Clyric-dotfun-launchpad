package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/paxeer/offchain-server/internal/config"
	"github.com/paxeer/offchain-server/internal/db"
	"github.com/paxeer/offchain-server/internal/indexer"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "path to config.yaml")
	flag.Parse()

	c, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()
	database, err := db.Connect(ctx, c.Postgres.DSN)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer database.Close()

	if err := database.Migrate(ctx, "migrations"); err != nil {
		log.Fatalf("db migrate: %v", err)
	}

	abis, err := indexer.LoadABIs()
	if err != nil {
		log.Fatalf("load abis: %v", err)
	}

	ix, err := indexer.NewIndexer(c.RPC.HTTP, c.RPC.WS, c.Contracts.Factory, database.Pool, abis, c.Indexer.Confirmations, c.Indexer.BatchSize)
	if err != nil {
		log.Fatalf("new indexer: %v", err)
	}
	defer ix.Close()

	// context with cancel on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Start backfill and live subscription/polling
	go func() {
		if err := ix.Backfill(ctx, c.Indexer.StartBlock); err != nil {
			log.Printf("backfill stopped: %v", err)
			cancel()
		}
	}()

	if ix.WS != nil {
		if err := ix.Subscribe(ctx); err != nil {
			log.Printf("subscribe stopped: %v", err)
		}
	} else {
		log.Printf("no ws client; starting HTTP polling loop")
		if err := ix.PollForever(ctx, 3*time.Second); err != nil {
			log.Printf("polling stopped: %v", err)
		}
	}

	// Give goroutines a moment to shutdown gracefully
	time.Sleep(1 * time.Second)
	log.Println("Indexer stopped.")
}
