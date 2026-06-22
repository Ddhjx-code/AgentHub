package knowledge

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/Ddhjx-code/AgentHub/internal/chunker"
	"github.com/Ddhjx-code/AgentHub/internal/embedding"
	"github.com/Ddhjx-code/AgentHub/internal/model"
	kbRepo "github.com/Ddhjx-code/AgentHub/internal/repository/knowledge"
	"github.com/Ddhjx-code/AgentHub/internal/vectorstore"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
)

type Service interface {
	CreateKB(ctx context.Context, kb *model.KnowledgeBase) error
	GetKB(ctx context.Context, id int64) (*model.KnowledgeBase, error)
	ListKBs(ctx context.Context) ([]*model.KnowledgeBase, error)
	UpdateKB(ctx context.Context, kb *model.KnowledgeBase) error
	DeleteKB(ctx context.Context, id int64) error

	UploadDocument(ctx context.Context, kbID int64, name, content string) (*model.Document, error)
	ListDocuments(ctx context.Context, kbID int64) ([]*model.Document, error)
	DeleteDocument(ctx context.Context, kbID, docID int64) error

	BindAgentKB(ctx context.Context, agentID, kbID int64) error
	UnbindAgentKB(ctx context.Context, agentID, kbID int64) error
	ListAgentKBs(ctx context.Context, agentID int64) ([]*model.KnowledgeBase, error)

	Search(ctx context.Context, agentID int64, query string, topK int) ([]vectorstore.SearchResult, error)
}

type service struct {
	repo        kbRepo.Repository
	embClient   embedding.Client
	vectorStore vectorstore.Store
	logger      *slog.Logger
}

func NewService(
	repo kbRepo.Repository,
	embClient embedding.Client,
	vs vectorstore.Store,
	logger *slog.Logger,
) Service {
	return &service{
		repo:        repo,
		embClient:   embClient,
		vectorStore: vs,
		logger:      logger,
	}
}

func (s *service) CreateKB(ctx context.Context, kb *model.KnowledgeBase) error {
	if kb.Status == "" {
		kb.Status = model.KBStatusActive
	}
	if kb.ChunkSize <= 0 {
		kb.ChunkSize = 512
	}
	if kb.ChunkOverlap < 0 {
		kb.ChunkOverlap = 0
	}
	return s.repo.CreateKB(ctx, kb)
}

func (s *service) GetKB(ctx context.Context, id int64) (*model.KnowledgeBase, error) {
	kb, err := s.repo.GetKBByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if kb == nil {
		return nil, errcode.ErrKBNotFound
	}
	return kb, nil
}

func (s *service) ListKBs(ctx context.Context) ([]*model.KnowledgeBase, error) {
	return s.repo.ListKBs(ctx)
}

func (s *service) UpdateKB(ctx context.Context, kb *model.KnowledgeBase) error {
	existing, err := s.repo.GetKBByID(ctx, kb.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errcode.ErrKBNotFound
	}
	return s.repo.UpdateKB(ctx, kb)
}

func (s *service) DeleteKB(ctx context.Context, id int64) error {
	existing, err := s.repo.GetKBByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errcode.ErrKBNotFound
	}

	if err := s.vectorStore.DeleteByKnowledgeBase(ctx, id); err != nil {
		s.logger.Error("delete vectors by kb failed", "kb_id", id, "error", err)
	}

	return s.repo.DeleteKB(ctx, id)
}

func (s *service) UploadDocument(ctx context.Context, kbID int64, name, content string) (*model.Document, error) {
	kb, err := s.repo.GetKBByID(ctx, kbID)
	if err != nil {
		return nil, err
	}
	if kb == nil {
		return nil, errcode.ErrKBNotFound
	}

	doc := &model.Document{
		KnowledgeBaseID: kbID,
		Name:            name,
		Content:         content,
		CharCount:       utf8.RuneCountInString(content),
		Status:          model.DocStatusPending,
	}
	if err := s.repo.CreateDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("create document: %w", err)
	}

	if err := s.repo.UpdateDocumentStatus(ctx, doc.ID, model.DocStatusProcessing, 0); err != nil {
		return nil, fmt.Errorf("update status to processing: %w", err)
	}

	chunks := chunker.Split(content, kb.ChunkSize, kb.ChunkOverlap)
	if len(chunks) == 0 {
		s.repo.UpdateDocumentStatus(ctx, doc.ID, model.DocStatusCompleted, 0)
		doc.Status = model.DocStatusCompleted
		return doc, nil
	}

	embeddings, err := s.embClient.Embed(ctx, kb.EmbeddingBaseURL, kb.EmbeddingAPIKey, kb.EmbeddingModel, chunks)
	if err != nil {
		s.repo.UpdateDocumentStatus(ctx, doc.ID, model.DocStatusFailed, 0)
		return nil, errcode.ErrEmbeddingError
	}

	if kb.Dimension == 0 && len(embeddings) > 0 && len(embeddings[0]) > 0 {
		s.repo.UpdateKBDimension(ctx, kbID, len(embeddings[0]))
	}

	chunkData := make([]vectorstore.ChunkData, len(chunks))
	for i, text := range chunks {
		var emb []float32
		if i < len(embeddings) {
			emb = embeddings[i]
		}
		chunkData[i] = vectorstore.ChunkData{
			Index:     i,
			Content:   text,
			Embedding: emb,
		}
	}

	if err := s.vectorStore.Store(ctx, kbID, doc.ID, chunkData); err != nil {
		s.repo.UpdateDocumentStatus(ctx, doc.ID, model.DocStatusFailed, 0)
		return nil, fmt.Errorf("store vectors: %w", err)
	}

	s.repo.UpdateDocumentStatus(ctx, doc.ID, model.DocStatusCompleted, len(chunks))
	doc.Status = model.DocStatusCompleted
	doc.ChunkCount = len(chunks)
	return doc, nil
}

func (s *service) ListDocuments(ctx context.Context, kbID int64) ([]*model.Document, error) {
	kb, err := s.repo.GetKBByID(ctx, kbID)
	if err != nil {
		return nil, err
	}
	if kb == nil {
		return nil, errcode.ErrKBNotFound
	}
	return s.repo.ListDocumentsByKBID(ctx, kbID)
}

func (s *service) DeleteDocument(ctx context.Context, kbID, docID int64) error {
	doc, err := s.repo.GetDocumentByID(ctx, docID)
	if err != nil {
		return err
	}
	if doc == nil || doc.KnowledgeBaseID != kbID {
		return errcode.ErrDocNotFound
	}

	if err := s.vectorStore.DeleteByDocument(ctx, docID); err != nil {
		s.logger.Error("delete vectors by doc failed", "doc_id", docID, "error", err)
	}

	return s.repo.DeleteDocument(ctx, docID)
}

func (s *service) BindAgentKB(ctx context.Context, agentID, kbID int64) error {
	kb, err := s.repo.GetKBByID(ctx, kbID)
	if err != nil {
		return err
	}
	if kb == nil {
		return errcode.ErrKBNotFound
	}
	return s.repo.BindAgentKB(ctx, agentID, kbID)
}

func (s *service) UnbindAgentKB(ctx context.Context, agentID, kbID int64) error {
	return s.repo.UnbindAgentKB(ctx, agentID, kbID)
}

func (s *service) ListAgentKBs(ctx context.Context, agentID int64) ([]*model.KnowledgeBase, error) {
	return s.repo.ListKBsByAgentID(ctx, agentID)
}

func (s *service) Search(ctx context.Context, agentID int64, query string, topK int) ([]vectorstore.SearchResult, error) {
	kbs, err := s.repo.ListKBsByAgentID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("list agent kbs: %w", err)
	}
	if len(kbs) == 0 {
		return nil, nil
	}

	if topK <= 0 {
		topK = 5
	}

	var allResults []vectorstore.SearchResult
	for _, kb := range kbs {
		if kb.Status != model.KBStatusActive {
			continue
		}

		embeddings, err := s.embClient.Embed(ctx, kb.EmbeddingBaseURL, kb.EmbeddingAPIKey, kb.EmbeddingModel, []string{query})
		if err != nil {
			s.logger.Error("embed query failed", "kb_id", kb.ID, "error", err)
			continue
		}
		if len(embeddings) == 0 || len(embeddings[0]) == 0 {
			continue
		}

		results, err := s.vectorStore.Search(ctx, kb.ID, embeddings[0], topK)
		if err != nil {
			s.logger.Error("vector search failed", "kb_id", kb.ID, "error", err)
			continue
		}
		allResults = append(allResults, results...)
	}

	if len(allResults) > topK {
		allResults = allResults[:topK]
	}

	return allResults, nil
}

func FormatSearchResults(results []vectorstore.SearchResult) string {
	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder
	for i, r := range results {
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		fmt.Fprintf(&sb, "[Source: %s]\n%s", r.DocName, r.Content)
	}
	return sb.String()
}
