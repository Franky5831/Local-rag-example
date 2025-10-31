package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

var questions = []string{
	"what do I have in the fridge?",
	"when did gundam first came out?",
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

type Document struct {
	Filename string
	Content  string
}

func searchRelevantDocs(ctx context.Context, pool *pgxpool.Pool, question string) ([]Document, error) {
	// Get embedding for the question
	embedding, err := getEmbedding(question)
	if err != nil {
		return nil, fmt.Errorf("failed to get embedding: %w", err)
	}

	// Convert embedding to pgvector format
	embedStr := "["
	for i, val := range embedding {
		if i > 0 {
			embedStr += ","
		}
		embedStr += fmt.Sprintf("%f", val)
	}
	embedStr += "]"

	// Search for similar documents
	// Gets only documents that are at least 0.7 relevant. It goes from 0 to 1
	// Gets only top 3 results
	query := `
		SELECT filename, content
		FROM documents
		WHERE (embedding <=> $1::vector) < 0.5
		ORDER BY embedding <=> $1::vector
		LIMIT 3
	`

	rows, err := pool.Query(ctx, query, embedStr)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var doc Document
		if err := rows.Scan(&doc.Filename, &doc.Content); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

func askLLM(question string, context string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: context,
		},
		{
			Role:    "user",
			Content: question,
		},
	}

	reqBody := ChatRequest{
		Model:    "gemma3:270m",
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}

	resp, err := http.Post(ollamaURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read failed: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal failed: %w", err)
	}

	return chatResp.Message.Content, nil
}

func processQuestion(ctx context.Context, pool *pgxpool.Pool, question string) {
	// Search for relevant documents
	docs, err := searchRelevantDocs(ctx, pool, question)
	if err != nil {
		log.Printf("Failed to search docs: %v", err)
		return
	}

	if len(docs) == 0 {
		fmt.Printf("\nQuestion: %s\nNo relevant documents found.\n", question)
		return
	}

	// Build context from documents
	var contextParts []string
	var filenames []string
	for _, doc := range docs {
		contextParts = append(contextParts, doc.Content)
		filenames = append(filenames, doc.Filename)
	}
	context := "Use the following information to answer the question:\n\n" + strings.Join(contextParts, "\n\n")

	// Ask LLM
	answer, err := askLLM(question, context)
	if err != nil {
		log.Printf("Failed to ask LLM: %v", err)
		return
	}

	// Print results
	fmt.Printf("\n" + strings.Repeat("=", 60) + "\n")
	fmt.Printf("Question: %s\n", question)
	fmt.Printf("Answer: %s\n", answer)
	fmt.Printf("Sources: %s\n", strings.Join(filenames, ", "))
	fmt.Printf(strings.Repeat("=", 60) + "\n")
}

func askQuestions(ctx context.Context, pool *pgxpool.Pool) {
	for _, question := range questions {
		processQuestion(ctx, pool, question)
	}
}

func Retrive() {
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
	// Ask questions
	fmt.Println("\nAsking questions...")
	askQuestions(ctx, pool)
}
