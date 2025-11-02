package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func createTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS documents (
			id SERIAL PRIMARY KEY,
			filename TEXT NOT NULL,
			content TEXT NOT NULL,
			embedding vector(768),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_embedding ON documents USING ivfflat (embedding vector_cosine_ops);
	`
	_, err := pool.Exec(ctx, query)
	return err
}

func emptyTable(ctx context.Context, pool *pgxpool.Pool) error {
	query := "DELETE FROM documents"
	_, err := pool.Exec(ctx, query)
	return err
}

func insertDocument(ctx context.Context, pool *pgxpool.Pool, filename, content string, embedding []float64) error {
	// Convert embedding to pgvector format string
	embedStr := "["
	for i, val := range embedding {
		if i > 0 {
			embedStr += ","
		}
		embedStr += fmt.Sprintf("%f", val)
	}
	embedStr += "]"

	query := `
		INSERT INTO documents (filename, content, embedding)
		VALUES ($1, $2, $3::vector)
	`
	_, err := pool.Exec(ctx, query, filename, content, embedStr)
	return err
}

func Seed() {
	time.Sleep(2 * time.Second)

	ctx := context.Background()

	// Connect to PostgreSQL
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

	// Enable pgvector extension
	if _, err := pool.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		log.Fatalf("failed to enable vector extension: %v", err)
	}

	// Empty table
	// We need to empty the table, if you run this example more than once you'll be left with multiple copies of the same document
	// This project is just an example, you probably want to create some sort of logic to either update records or not import existing ones
	if err := emptyTable(ctx, pool); err != nil {
		log.Fatalf("failed to empty table: %v", err)
	}
	fmt.Println("Table ready")

	// Create table
	if err := createTable(ctx, pool); err != nil {
		log.Fatalf("failed to create table: %v", err)
	}
	fmt.Println("Table ready")

	// Read all .md files from data directory
	files, err := filepath.Glob("data/*.md")
	if err != nil {
		log.Fatalf("failed to read data directory: %v", err)
	}

	if len(files) == 0 {
		fmt.Println("No .md files found in data directory")
		return
	}

	// Process each file
	for _, file := range files {
		filename := filepath.Base(file)
		fmt.Printf("Processing %s...", filename)

		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("failed to read %s: %v", filename, err)
			continue
		}

		// Get embedding
		embedding, err := getEmbedding(string(content))
		if err != nil {
			log.Printf("failed to get embedding for %s: %v", filename, err)
			continue
		}

		// Insert into database
		if err := insertDocument(ctx, pool, filename, string(content), embedding); err != nil {
			log.Printf("failed to insert %s: %v", filename, err)
			continue
		}

		fmt.Println(" ✓")
	}

	fmt.Println("Seeding complete!")
}
