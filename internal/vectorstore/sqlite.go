package vectorstore

import (
	"context"
	"database/sql"
	"fmt"
	"sort"

	"github.com/Ddhjx-code/AgentHub/internal/embedding"
)

type sqliteStore struct {
	db *sql.DB
}

func NewSQLiteStore(db *sql.DB) Store {
	return &sqliteStore{db: db}
}

func (s *sqliteStore) Store(ctx context.Context, knowledgeBaseID, documentID int64, chunks []ChunkData) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx,
		`INSERT INTO chunks (document_id, knowledge_base_id, chunk_index, content, embedding) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer stmt.Close()

	for _, c := range chunks {
		blob := embedding.EncodeFloat32s(c.Embedding)
		if _, err := stmt.ExecContext(ctx, documentID, knowledgeBaseID, c.Index, c.Content, blob); err != nil {
			return fmt.Errorf("insert chunk %d: %w", c.Index, err)
		}
	}

	return tx.Commit()
}

func (s *sqliteStore) Search(ctx context.Context, knowledgeBaseID int64, queryEmbedding []float32, topK int) ([]SearchResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.content, c.embedding, d.name
		 FROM chunks c
		 JOIN documents d ON d.id = c.document_id
		 WHERE c.knowledge_base_id = ? AND c.embedding IS NOT NULL`,
		knowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("query chunks: %w", err)
	}
	defer rows.Close()

	type scored struct {
		result SearchResult
		score  float32
	}

	var candidates []scored
	for rows.Next() {
		var id int64
		var content string
		var blob []byte
		var docName string
		if err := rows.Scan(&id, &content, &blob, &docName); err != nil {
			return nil, fmt.Errorf("scan chunk: %w", err)
		}

		chunkVec := embedding.DecodeFloat32s(blob)
		score := embedding.CosineSimilarity(queryEmbedding, chunkVec)
		candidates = append(candidates, scored{
			result: SearchResult{ChunkID: id, Content: content, Score: score, DocName: docName},
			score:  score,
		})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	if topK > len(candidates) {
		topK = len(candidates)
	}

	results := make([]SearchResult, topK)
	for i := 0; i < topK; i++ {
		results[i] = candidates[i].result
	}

	return results, nil
}

func (s *sqliteStore) DeleteByDocument(ctx context.Context, documentID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM chunks WHERE document_id = ?`, documentID)
	return err
}

func (s *sqliteStore) DeleteByKnowledgeBase(ctx context.Context, knowledgeBaseID int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM chunks WHERE knowledge_base_id = ?`, knowledgeBaseID)
	return err
}
