package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const pgURI = "postgres://user:password@postgresql:5432/db?sslmode=disable&connect_timeout=5"
const ollamaEmbedURL = "http://ollama:11434/api/embeddings"
const ollamaURL = "http://ollama:11434/api/chat"

type EmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbedResponse struct {
	Embedding []float64 `json:"embedding"`
}

func getEmbedding(content string) ([]float64, error) {
	reqBody := EmbedRequest{
		Model:  "nomic-embed-text:latest",
		Prompt: content,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal failed: %w", err)
	}

	resp, err := http.Post(ollamaEmbedURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}

	var embedResp EmbedResponse
	if err := json.Unmarshal(body, &embedResp); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	return embedResp.Embedding, nil
}
