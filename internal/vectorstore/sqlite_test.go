package vectorstore

import (
	"context"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/database"
)

func setupTestDB(t *testing.T) *sqliteStore {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	db.Exec(`INSERT INTO knowledge_bases (id, name) VALUES (1, 'test-kb')`)
	db.Exec(`INSERT INTO documents (id, knowledge_base_id, name, content) VALUES (1, 1, 'doc1.txt', 'test content')`)

	return &sqliteStore{db: db}
}

func TestStoreAndSearch(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	chunks := []ChunkData{
		{Index: 0, Content: "Go is a programming language", Embedding: []float32{1, 0, 0}},
		{Index: 1, Content: "Python is also a language", Embedding: []float32{0, 1, 0}},
		{Index: 2, Content: "Rust is fast and safe", Embedding: []float32{0, 0, 1}},
	}
	if err := store.Store(ctx, 1, 1, chunks); err != nil {
		t.Fatalf("store chunks: %v", err)
	}

	results, err := store.Search(ctx, 1, []float32{1, 0, 0}, 2, 0)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Content != "Go is a programming language" {
		t.Errorf("expected Go chunk first, got %q", results[0].Content)
	}
}

func TestSearchMinScore(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	chunks := []ChunkData{
		{Index: 0, Content: "highly relevant", Embedding: []float32{1, 0, 0}},
		{Index: 1, Content: "somewhat relevant", Embedding: []float32{0.7, 0.7, 0}},
		{Index: 2, Content: "not relevant", Embedding: []float32{0, 0, 1}},
	}
	store.Store(ctx, 1, 1, chunks)

	results, err := store.Search(ctx, 1, []float32{1, 0, 0}, 10, 0.8)
	if err != nil {
		t.Fatalf("search: %v", err)
	}

	for _, r := range results {
		if r.Score < 0.8 {
			t.Errorf("result %q has score %f below threshold", r.Content, r.Score)
		}
	}
	if len(results) == 3 {
		t.Error("expected some results to be filtered by minScore")
	}
}

func TestSearchBM25(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	chunks := []ChunkData{
		{Index: 0, Content: "Go language goroutine concurrency", Embedding: []float32{1, 0, 0}},
		{Index: 1, Content: "Python asyncio event loop", Embedding: []float32{0, 1, 0}},
		{Index: 2, Content: "Go channel communication pattern", Embedding: []float32{0, 0, 1}},
	}
	store.Store(ctx, 1, 1, chunks)

	results, err := store.SearchBM25(ctx, 1, "Go", 5)
	if err != nil {
		t.Fatalf("bm25 search: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 BM25 results for 'Go', got %d", len(results))
	}

	for _, r := range results {
		if r.Content != "Go language goroutine concurrency" && r.Content != "Go channel communication pattern" {
			t.Errorf("unexpected result: %q", r.Content)
		}
	}
}

func TestFTSSyncOnDelete(t *testing.T) {
	store := setupTestDB(t)
	ctx := context.Background()

	chunks := []ChunkData{
		{Index: 0, Content: "Go programming", Embedding: []float32{1, 0, 0}},
		{Index: 1, Content: "Python scripting", Embedding: []float32{0, 1, 0}},
	}
	store.Store(ctx, 1, 1, chunks)

	results, _ := store.SearchBM25(ctx, 1, "Go", 5)
	if len(results) != 1 {
		t.Fatalf("expected 1 BM25 result before delete, got %d", len(results))
	}

	store.DeleteByDocument(ctx, 1)

	results, _ = store.SearchBM25(ctx, 1, "Go", 5)
	if len(results) != 0 {
		t.Fatalf("expected 0 BM25 results after delete, got %d", len(results))
	}
}
