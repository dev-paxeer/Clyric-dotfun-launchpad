package main

import (
    "context"
    "flag"
    "log"
    "net/http"
    "os"

    "github.com/paxeer/offchain-server/internal/api"
    "github.com/paxeer/offchain-server/internal/config"
    "github.com/paxeer/offchain-server/internal/db"
)

func main() {
    cfgPath := flag.String("config", "configs/config.yaml", "path to config.yaml")
    flag.Parse()

    c, err := config.Load(*cfgPath)
    if err != nil { log.Fatalf("load config: %v", err) }

    database, err := db.Connect(context.Background(), c.Postgres.DSN)
    if err != nil { log.Fatalf("db connect: %v", err) }
    defer database.Close()

    if err := database.Migrate(context.Background(), "migrations"); err != nil {
        log.Fatalf("db migrate: %v", err)
    }

    addr := os.Getenv("PAXEER_API_ADDR")
    if addr == "" { addr = ":8080" }

    srv := &http.Server{ Addr: addr, Handler: api.New(database.Pool) }
    log.Printf("API listening on %s", addr)
    log.Fatal(srv.ListenAndServe())
}
