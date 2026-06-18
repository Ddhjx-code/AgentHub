package vectorstore

import "context"

type SearchResult struct {
	ChunkID   int64
	Content   string
	Score     float32
	DocName   string
}

type Store interface {
	Store(ctx context.Context, knowledgeBaseID, documentID int64, chunks []ChunkData) error
	Search(ctx context.Context, knowledgeBaseID int64, queryEmbedding []float32, topK int) ([]SearchResult, error)
	DeleteByDocument(ctx context.Context, documentID int64) error
	DeleteByKnowledgeBase(ctx context.Context, knowledgeBaseID int64) error
}

type ChunkData struct {
	Index     int
	Content   string
	Embedding []float32
}
