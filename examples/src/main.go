package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const pgURI = "postgres://user:password@postgresql:5432/db?sslmode=disable&connect_timeout=5"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(pgURI)
	if err != nil {
		log.Fatalf("bad DSN: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("dial failed: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping failed: %v", err)
	}

	var now time.Time
	if err := pool.QueryRow(ctx, "select now()").Scan(&now); err != nil {
		log.Fatalf("query failed: %v", err)
	}
	log.Printf("IT WORKEDD --------- ok: %s", now.UTC().Format(time.RFC3339))
}
