package embedding

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Model != "text-embedding-3-small" {
			t.Errorf("unexpected model: %s", req.Model)
		}
		if len(req.Input) != 2 {
			t.Errorf("expected 2 inputs, got %d", len(req.Input))
		}

		resp := embeddingResponse{
			Data: []embeddingData{
				{Index: 0, Embedding: []float32{0.1, 0.2, 0.3}},
				{Index: 1, Embedding: []float32{0.4, 0.5, 0.6}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient()
	result, err := c.Embed(context.Background(), server.URL, "test-key", "text-embedding-3-small", []string{"hello", "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0][0] != 0.1 || result[1][2] != 0.6 {
		t.Errorf("unexpected embedding values")
	}
}

func TestEmbedAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":"rate limited"}`))
	}))
	defer server.Close()

	c := NewClient()
	_, err := c.Embed(context.Background(), server.URL, "key", "model", []string{"test"})
	if err == nil {
		t.Fatal("expected error")
	}
}
