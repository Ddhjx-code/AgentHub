package knowledge

import (
	"strconv"

	"github.com/Ddhjx-code/AgentHub/internal/model"
	kbSvc "github.com/Ddhjx-code/AgentHub/internal/service/knowledge"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	"github.com/Ddhjx-code/AgentHub/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	kbSvc kbSvc.Service
}

func NewHandler(ks kbSvc.Service) *Handler {
	return &Handler{kbSvc: ks}
}

type createKBReq struct {
	Name             string `json:"name" binding:"required"`
	Description      string `json:"description"`
	EmbeddingBaseURL string `json:"embedding_base_url" binding:"required"`
	EmbeddingAPIKey  string `json:"embedding_api_key" binding:"required"`
	EmbeddingModel   string `json:"embedding_model" binding:"required"`
	ChunkSize        int    `json:"chunk_size"`
	ChunkOverlap     int    `json:"chunk_overlap"`
}

type updateKBReq struct {
	Name             *string `json:"name"`
	Description      *string `json:"description"`
	EmbeddingBaseURL *string `json:"embedding_base_url"`
	EmbeddingAPIKey  *string `json:"embedding_api_key"`
	EmbeddingModel   *string `json:"embedding_model"`
	ChunkSize        *int    `json:"chunk_size"`
	ChunkOverlap     *int    `json:"chunk_overlap"`
	Status           *string `json:"status"`
}

type uploadDocReq struct {
	Name    string `json:"name" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type bindKBReq struct {
	KnowledgeBaseID int64 `json:"knowledge_base_id" binding:"required"`
}

func (h *Handler) CreateKB(c *gin.Context) {
	var req createKBReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	kb := &model.KnowledgeBase{
		Name:             req.Name,
		Description:      req.Description,
		EmbeddingBaseURL: req.EmbeddingBaseURL,
		EmbeddingAPIKey:  req.EmbeddingAPIKey,
		EmbeddingModel:   req.EmbeddingModel,
		ChunkSize:        req.ChunkSize,
		ChunkOverlap:     req.ChunkOverlap,
	}
	if err := h.kbSvc.CreateKB(c.Request.Context(), kb); err != nil {
		handleError(c, err)
		return
	}
	response.Created(c, kb)
}

func (h *Handler) ListKBs(c *gin.Context) {
	kbs, err := h.kbSvc.ListKBs(c.Request.Context())
	if err != nil {
		handleError(c, err)
		return
	}
	if kbs == nil {
		kbs = []*model.KnowledgeBase{}
	}
	response.Success(c, kbs)
}

func (h *Handler) GetKB(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	kb, err := h.kbSvc.GetKB(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, kb)
}

func (h *Handler) UpdateKB(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	var req updateKBReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	kb, err := h.kbSvc.GetKB(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}

	if req.Name != nil {
		kb.Name = *req.Name
	}
	if req.Description != nil {
		kb.Description = *req.Description
	}
	if req.EmbeddingBaseURL != nil {
		kb.EmbeddingBaseURL = *req.EmbeddingBaseURL
	}
	if req.EmbeddingAPIKey != nil {
		kb.EmbeddingAPIKey = *req.EmbeddingAPIKey
	}
	if req.EmbeddingModel != nil {
		kb.EmbeddingModel = *req.EmbeddingModel
	}
	if req.ChunkSize != nil {
		kb.ChunkSize = *req.ChunkSize
	}
	if req.ChunkOverlap != nil {
		kb.ChunkOverlap = *req.ChunkOverlap
	}
	if req.Status != nil {
		kb.Status = *req.Status
	}

	if err := h.kbSvc.UpdateKB(c.Request.Context(), kb); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, kb)
}

func (h *Handler) DeleteKB(c *gin.Context) {
	id, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	if err := h.kbSvc.DeleteKB(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) UploadDocument(c *gin.Context) {
	kbID, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	var req uploadDocReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	doc, err := h.kbSvc.UploadDocument(c.Request.Context(), kbID, req.Name, req.Content)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Created(c, doc)
}

func (h *Handler) ListDocuments(c *gin.Context) {
	kbID, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	docs, err := h.kbSvc.ListDocuments(c.Request.Context(), kbID)
	if err != nil {
		handleError(c, err)
		return
	}
	if docs == nil {
		docs = []*model.Document{}
	}
	response.Success(c, docs)
}

func (h *Handler) DeleteDocument(c *gin.Context) {
	kbID, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}
	docID, err := parseID(c, "doc_id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	if err := h.kbSvc.DeleteDocument(c.Request.Context(), kbID, docID); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) BindAgentKB(c *gin.Context) {
	agentID, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	var req bindKBReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	if err := h.kbSvc.BindAgentKB(c.Request.Context(), agentID, req.KnowledgeBaseID); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) UnbindAgentKB(c *gin.Context) {
	agentID, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}
	kbID, err := parseID(c, "kb_id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	if err := h.kbSvc.UnbindAgentKB(c.Request.Context(), agentID, kbID); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListAgentKBs(c *gin.Context) {
	agentID, err := parseID(c, "id")
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	kbs, err := h.kbSvc.ListAgentKBs(c.Request.Context(), agentID)
	if err != nil {
		handleError(c, err)
		return
	}
	if kbs == nil {
		kbs = []*model.KnowledgeBase{}
	}
	response.Success(c, kbs)
}

func parseID(c *gin.Context, param string) (int64, error) {
	return strconv.ParseInt(c.Param(param), 10, 64)
}

func handleError(c *gin.Context, err error) {
	if ec, ok := err.(*errcode.ErrCode); ok {
		response.Error(c, ec)
		return
	}
	response.Error(c, errcode.ErrInternalServer)
}
