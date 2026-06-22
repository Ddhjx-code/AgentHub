package knowledge

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/database"
	"github.com/Ddhjx-code/AgentHub/internal/model"
	kbRepo "github.com/Ddhjx-code/AgentHub/internal/repository/knowledge"
	"github.com/Ddhjx-code/AgentHub/internal/vectorstore"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
)

type mockEmbeddingClient struct {
	dimension int
}

func (m *mockEmbeddingClient) Embed(_ context.Context, _, _, _ string, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		vec := make([]float32, m.dimension)
		for j := range vec {
			vec[j] = float32(i+1) * 0.1
		}
		result[i] = vec
	}
	return result, nil
}

type testEnv struct {
	svc  Service
	repo kbRepo.Repository
	db   *sql.DB
}

func setupTest(t *testing.T) *testEnv {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	repo := kbRepo.NewRepository(db)
	vs := vectorstore.NewSQLiteStore(db)
	embClient := &mockEmbeddingClient{dimension: 3}
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	svc := NewService(repo, embClient, vs, lg)
	return &testEnv{svc: svc, repo: repo, db: db}
}

func createTestAgent(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	result, err := db.Exec(
		`INSERT INTO agents (name, status, tags) VALUES (?, ?, ?)`,
		"TestAgent", "active", "[]")
	if err != nil {
		t.Fatalf("create test agent: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func TestCreateAndGetKB(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	kb := &model.KnowledgeBase{
		Name:             "Test KB",
		Description:      "A test knowledge base",
		EmbeddingBaseURL: "http://localhost:11434",
		EmbeddingAPIKey:  "test-key",
		EmbeddingModel:   "nomic-embed-text",
	}
	if err := env.svc.CreateKB(ctx, kb); err != nil {
		t.Fatalf("create kb: %v", err)
	}
	if kb.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if kb.Status != model.KBStatusActive {
		t.Fatalf("expected active status, got %s", kb.Status)
	}
	if kb.ChunkSize != 512 {
		t.Fatalf("expected default chunk size 512, got %d", kb.ChunkSize)
	}

	got, err := env.svc.GetKB(ctx, kb.ID)
	if err != nil {
		t.Fatalf("get kb: %v", err)
	}
	if got.Name != "Test KB" {
		t.Fatalf("expected Test KB, got %s", got.Name)
	}
}

func TestListKBs(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	env.svc.CreateKB(ctx, &model.KnowledgeBase{
		Name: "KB1", EmbeddingBaseURL: "http://localhost", EmbeddingAPIKey: "k", EmbeddingModel: "m",
	})
	env.svc.CreateKB(ctx, &model.KnowledgeBase{
		Name: "KB2", EmbeddingBaseURL: "http://localhost", EmbeddingAPIKey: "k", EmbeddingModel: "m",
	})

	kbs, err := env.svc.ListKBs(ctx)
	if err != nil {
		t.Fatalf("list kbs: %v", err)
	}
	if len(kbs) != 2 {
		t.Fatalf("expected 2, got %d", len(kbs))
	}
}

func TestUpdateKB(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	kb := &model.KnowledgeBase{
		Name: "Original", EmbeddingBaseURL: "http://localhost", EmbeddingAPIKey: "k", EmbeddingModel: "m",
	}
	env.svc.CreateKB(ctx, kb)

	kb.Name = "Updated"
	if err := env.svc.UpdateKB(ctx, kb); err != nil {
		t.Fatalf("update kb: %v", err)
	}

	got, _ := env.svc.GetKB(ctx, kb.ID)
	if got.Name != "Updated" {
		t.Fatalf("expected Updated, got %s", got.Name)
	}
}

func TestDeleteKB(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	kb := &model.KnowledgeBase{
		Name: "ToDelete", EmbeddingBaseURL: "http://localhost", EmbeddingAPIKey: "k", EmbeddingModel: "m",
	}
	env.svc.CreateKB(ctx, kb)

	if err := env.svc.DeleteKB(ctx, kb.ID); err != nil {
		t.Fatalf("delete kb: %v", err)
	}

	_, err := env.svc.GetKB(ctx, kb.ID)
	if err != errcode.ErrKBNotFound {
		t.Fatalf("expected ErrKBNotFound, got %v", err)
	}
}

func TestUploadDocumentAndSearch(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	kb := &model.KnowledgeBase{
		Name:             "Search KB",
		EmbeddingBaseURL: "http://localhost",
		EmbeddingAPIKey:  "k",
		EmbeddingModel:   "m",
		ChunkSize:        50,
		ChunkOverlap:     10,
	}
	env.svc.CreateKB(ctx, kb)

	content := "Go is a statically typed, compiled programming language. It was designed at Google by Robert Griesemer, Rob Pike, and Ken Thompson."
	doc, err := env.svc.UploadDocument(ctx, kb.ID, "go_intro.txt", content)
	if err != nil {
		t.Fatalf("upload document: %v", err)
	}
	if doc.Status != model.DocStatusCompleted {
		t.Fatalf("expected completed, got %s", doc.Status)
	}
	if doc.ChunkCount == 0 {
		t.Fatal("expected non-zero chunk count")
	}

	agentID := createTestAgent(t, env.db)
	env.repo.BindAgentKB(ctx, agentID, kb.ID)

	results, err := env.svc.Search(ctx, agentID, "Go programming", 3)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected search results")
	}
}

func TestListDocuments(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	kb := &model.KnowledgeBase{
		Name: "Docs KB", EmbeddingBaseURL: "http://localhost", EmbeddingAPIKey: "k", EmbeddingModel: "m",
	}
	env.svc.CreateKB(ctx, kb)

	env.svc.UploadDocument(ctx, kb.ID, "doc1.txt", "Hello World")
	env.svc.UploadDocument(ctx, kb.ID, "doc2.txt", "Goodbye World")

	docs, err := env.svc.ListDocuments(ctx, kb.ID)
	if err != nil {
		t.Fatalf("list documents: %v", err)
	}
	if len(docs) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(docs))
	}
}

func TestDeleteDocument(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	kb := &model.KnowledgeBase{
		Name: "Del Doc KB", EmbeddingBaseURL: "http://localhost", EmbeddingAPIKey: "k", EmbeddingModel: "m",
	}
	env.svc.CreateKB(ctx, kb)

	doc, _ := env.svc.UploadDocument(ctx, kb.ID, "temp.txt", "temporary content")

	if err := env.svc.DeleteDocument(ctx, kb.ID, doc.ID); err != nil {
		t.Fatalf("delete document: %v", err)
	}

	docs, _ := env.svc.ListDocuments(ctx, kb.ID)
	if len(docs) != 0 {
		t.Fatalf("expected 0 documents after delete, got %d", len(docs))
	}
}

func TestBindUnbindAgentKB(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	kb := &model.KnowledgeBase{
		Name: "Bind KB", EmbeddingBaseURL: "http://localhost", EmbeddingAPIKey: "k", EmbeddingModel: "m",
	}
	env.svc.CreateKB(ctx, kb)

	agentID := createTestAgent(t, env.db)
	if err := env.svc.BindAgentKB(ctx, agentID, kb.ID); err != nil {
		t.Fatalf("bind: %v", err)
	}

	kbs, err := env.svc.ListAgentKBs(ctx, agentID)
	if err != nil {
		t.Fatalf("list agent kbs: %v", err)
	}
	if len(kbs) != 1 || kbs[0].ID != kb.ID {
		t.Fatalf("expected 1 bound kb, got %d", len(kbs))
	}

	if err := env.svc.UnbindAgentKB(ctx, agentID, kb.ID); err != nil {
		t.Fatalf("unbind: %v", err)
	}

	kbs, _ = env.svc.ListAgentKBs(ctx, agentID)
	if len(kbs) != 0 {
		t.Fatalf("expected 0 after unbind, got %d", len(kbs))
	}
}

func TestGetKBNotFound(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	_, err := env.svc.GetKB(ctx, 999)
	if err != errcode.ErrKBNotFound {
		t.Fatalf("expected ErrKBNotFound, got %v", err)
	}
}

func TestDeleteDocumentNotFound(t *testing.T) {
	env := setupTest(t)
	ctx := context.Background()

	kb := &model.KnowledgeBase{
		Name: "NF KB", EmbeddingBaseURL: "http://localhost", EmbeddingAPIKey: "k", EmbeddingModel: "m",
	}
	env.svc.CreateKB(ctx, kb)

	err := env.svc.DeleteDocument(ctx, kb.ID, 999)
	if err != errcode.ErrDocNotFound {
		t.Fatalf("expected ErrDocNotFound, got %v", err)
	}
}

func TestFormatSearchResults(t *testing.T) {
	results := []vectorstore.SearchResult{
		{ChunkID: 1, Content: "chunk 1 content", Score: 0.9, DocName: "doc1.txt"},
		{ChunkID: 2, Content: "chunk 2 content", Score: 0.8, DocName: "doc2.txt"},
	}

	output := FormatSearchResults(results)
	if output == "" {
		t.Fatal("expected non-empty output")
	}

	empty := FormatSearchResults(nil)
	if empty != "" {
		t.Fatalf("expected empty for nil results, got %s", empty)
	}
}

func TestRRFMerge(t *testing.T) {
	vectorResults := []vectorstore.SearchResult{
		{ChunkID: 1, Content: "chunk A", Score: 0.9},
		{ChunkID: 2, Content: "chunk B", Score: 0.8},
		{ChunkID: 3, Content: "chunk C", Score: 0.7},
	}
	bm25Results := []vectorstore.SearchResult{
		{ChunkID: 3, Content: "chunk C", Score: 0},
		{ChunkID: 4, Content: "chunk D", Score: 0},
		{ChunkID: 1, Content: "chunk A", Score: 0},
	}

	merged := rrfMerge(vectorResults, bm25Results, 3)

	if len(merged) != 3 {
		t.Fatalf("expected 3 results, got %d", len(merged))
	}

	// chunk A and C appear in both lists, should rank higher than B and D
	topIDs := map[int64]bool{}
	for _, r := range merged[:2] {
		topIDs[r.ChunkID] = true
	}
	if !topIDs[1] || !topIDs[3] {
		t.Errorf("expected chunk 1 and 3 in top 2, got %v", merged[:2])
	}
}

func TestRRFMergeEmpty(t *testing.T) {
	merged := rrfMerge(nil, nil, 5)
	if len(merged) != 0 {
		t.Fatalf("expected 0 results, got %d", len(merged))
	}

	vectorOnly := rrfMerge(
		[]vectorstore.SearchResult{{ChunkID: 1, Content: "A"}},
		nil,
		5,
	)
	if len(vectorOnly) != 1 {
		t.Fatalf("expected 1 result, got %d", len(vectorOnly))
	}
}
