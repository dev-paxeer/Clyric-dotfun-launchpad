package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*DB, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &DB{Pool: pool}, nil
}

func (d *DB) Close() {
	if d.Pool != nil {
		d.Pool.Close()
	}
}

func (d *DB) Migrate(ctx context.Context, dir string) error {
	if d.Pool == nil {
		return errors.New("nil pool")
	}
	// ensure migrations table
	_, err := d.Pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		id SERIAL PRIMARY KEY,
		filename TEXT UNIQUE NOT NULL,
		applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	if err != nil {
		return err
	}
	// collect applied filenames
	applied := map[string]bool{}
	rows, err := d.Pool.Query(ctx, `SELECT filename FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return err
		}
		applied[name] = true
	}
	rows.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	type mig struct{ name string; path string }
	var mgs []mig
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".sql" {
			continue
		}
		mgs = append(mgs, mig{name: e.Name(), path: filepath.Join(dir, e.Name())})
	}
	sort.Slice(mgs, func(i, j int) bool { return mgs[i].name < mgs[j].name })

	for _, m := range mgs {
		if applied[m.name] {
			continue
		}
		b, err := os.ReadFile(m.path)
		if err != nil {
			return err
		}
		tx, err := d.Pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer func() {
			if tx != nil {
				_ = tx.Rollback(ctx)
			}
		}()
		if _, err := tx.Exec(ctx, string(b)); err != nil {
			return fmt.Errorf("apply migration %s: %w", m.name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations(filename) VALUES($1)`, m.name); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		tx = nil
	}
	return nil
}
