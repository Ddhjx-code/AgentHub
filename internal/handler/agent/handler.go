package agent

import (
	"strconv"

	agentSvc "github.com/Ddhjx-code/AgentHub/internal/service/agent"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	"github.com/Ddhjx-code/AgentHub/pkg/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	agentSvc agentSvc.Service
}

func NewHandler(as agentSvc.Service) *Handler {
	return &Handler{agentSvc: as}
}

func (h *Handler) List(c *gin.Context) {
	page, limit := parsePagination(c)
	category := c.Query("category")
	tag := c.Query("tag")

	agents, total, err := h.agentSvc.ListActive(c.Request.Context(), page, limit, category, tag)
	if err != nil {
		if ec, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, ec)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.SuccessWithMeta(c, agents, &response.Meta{
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *Handler) Detail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	detail, err := h.agentSvc.GetByID(c.Request.Context(), id)
	if err != nil {
		if ec, ok := err.(*errcode.ErrCode); ok {
			response.Error(c, ec)
			return
		}
		response.Error(c, errcode.ErrInternalServer)
		return
	}
	response.Success(c, detail)
}

func parsePagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return page, limit
}
