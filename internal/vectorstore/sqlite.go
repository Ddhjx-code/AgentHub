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

	chunkStmt, err := tx.PrepareContext(ctx,
		`INSERT INTO chunks (document_id, knowledge_base_id, chunk_index, content, embedding) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare insert: %w", err)
	}
	defer chunkStmt.Close()

	ftsStmt, err := tx.PrepareContext(ctx,
		`INSERT INTO chunks_fts (rowid, content) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare fts insert: %w", err)
	}
	defer ftsStmt.Close()

	for _, c := range chunks {
		blob := embedding.EncodeFloat32s(c.Embedding)
		result, err := chunkStmt.ExecContext(ctx, documentID, knowledgeBaseID, c.Index, c.Content, blob)
		if err != nil {
			return fmt.Errorf("insert chunk %d: %w", c.Index, err)
		}
		chunkID, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("last insert id: %w", err)
		}
		if _, err := ftsStmt.ExecContext(ctx, chunkID, c.Content); err != nil {
			return fmt.Errorf("insert fts chunk %d: %w", c.Index, err)
		}
	}

	return tx.Commit()
}

func (s *sqliteStore) Search(ctx context.Context, knowledgeBaseID int64, queryEmbedding []float32, topK int, minScore float32) ([]SearchResult, error) {
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

	var results []SearchResult
	for i := 0; i < len(candidates) && len(results) < topK; i++ {
		if candidates[i].score < minScore {
			break
		}
		results = append(results, candidates[i].result)
	}

	return results, nil
}

func (s *sqliteStore) SearchBM25(ctx context.Context, knowledgeBaseID int64, query string, topK int) ([]SearchResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.content, d.name, f.rank
		 FROM chunks_fts f
		 JOIN chunks c ON c.id = f.rowid
		 JOIN documents d ON d.id = c.document_id
		 WHERE c.knowledge_base_id = ? AND chunks_fts MATCH ?
		 ORDER BY f.rank
		 LIMIT ?`,
		knowledgeBaseID, query, topK)
	if err != nil {
		return nil, fmt.Errorf("bm25 search: %w", err)
	}
	defer rows.Close()

	var results []SearchResult
	for rows.Next() {
		var r SearchResult
		var rank float64
		if err := rows.Scan(&r.ChunkID, &r.Content, &r.DocName, &rank); err != nil {
			return nil, fmt.Errorf("scan bm25 result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *sqliteStore) DeleteByDocument(ctx context.Context, documentID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.ExecContext(ctx,
		`DELETE FROM chunks_fts WHERE rowid IN (SELECT id FROM chunks WHERE document_id = ?)`,
		documentID)
	tx.ExecContext(ctx, `DELETE FROM chunks WHERE document_id = ?`, documentID)

	return tx.Commit()
}

func (s *sqliteStore) DeleteByKnowledgeBase(ctx context.Context, knowledgeBaseID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tx.ExecContext(ctx,
		`DELETE FROM chunks_fts WHERE rowid IN (SELECT id FROM chunks WHERE knowledge_base_id = ?)`,
		knowledgeBaseID)
	tx.ExecContext(ctx, `DELETE FROM chunks WHERE knowledge_base_id = ?`, knowledgeBaseID)

	return tx.Commit()
}
