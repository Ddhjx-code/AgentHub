package chat

import (
	"strconv"

	"github.com/Ddhjx-code/AgentHub/internal/middleware"
	chatSvc "github.com/Ddhjx-code/AgentHub/internal/service/chat"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	"github.com/Ddhjx-code/AgentHub/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	chatSvc chatSvc.Service
}

func NewHandler(cs chatSvc.Service) *Handler {
	return &Handler{chatSvc: cs}
}

type sendMessageReq struct {
	AgentID        int64  `json:"agent_id" binding:"required"`
	ConversationID *int64 `json:"conversation_id"`
	Content        string `json:"content" binding:"required"`
}

func (h *Handler) SendMessage(c *gin.Context) {
	var req sendMessageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	userID := c.GetInt64(middleware.UserIDKey)
	result, err := h.chatSvc.SendMessage(c.Request.Context(), userID, req.AgentID, req.ConversationID, req.Content)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListConversations(c *gin.Context) {
	userID := c.GetInt64(middleware.UserIDKey)
	convs, err := h.chatSvc.ListConversations(c.Request.Context(), userID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, convs)
}

func (h *Handler) GetMessages(c *gin.Context) {
	userID := c.GetInt64(middleware.UserIDKey)
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	msgs, err := h.chatSvc.GetMessages(c.Request.Context(), userID, convID)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, msgs)
}

func (h *Handler) DeleteConversation(c *gin.Context) {
	userID := c.GetInt64(middleware.UserIDKey)
	convID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	if err := h.chatSvc.DeleteConversation(c.Request.Context(), userID, convID); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil)
}

func handleError(c *gin.Context, err error) {
	if ec, ok := err.(*errcode.ErrCode); ok {
		response.Error(c, ec)
		return
	}
	response.Error(c, errcode.ErrInternalServer)
}
