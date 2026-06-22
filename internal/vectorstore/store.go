package vectorstore

import "context"

type SearchResult struct {
	ChunkID   int64
	Content   string
	Score     float32
	DocName   string
}

const DefaultMinScore float32 = 0.5

type Store interface {
	Store(ctx context.Context, knowledgeBaseID, documentID int64, chunks []ChunkData) error
	Search(ctx context.Context, knowledgeBaseID int64, queryEmbedding []float32, topK int, minScore float32) ([]SearchResult, error)
	SearchBM25(ctx context.Context, knowledgeBaseID int64, query string, topK int) ([]SearchResult, error)
	DeleteByDocument(ctx context.Context, documentID int64) error
	DeleteByKnowledgeBase(ctx context.Context, knowledgeBaseID int64) error
}

type ChunkData struct {
	Index     int
	Content   string
	Embedding []float32
}
