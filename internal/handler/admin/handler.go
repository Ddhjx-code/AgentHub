package admin

import (
	"encoding/json"
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

type toolReq struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Type        string          `json:"type" binding:"required"`
	InputSchema json.RawMessage `json:"input_schema"`
	Config      json.RawMessage `json:"config"`
}

type createAgentReq struct {
	Name        string    `json:"name" binding:"required"`
	Icon        string    `json:"icon"`
	Color       string    `json:"color"`
	Category    string    `json:"category"`
	ShortDesc   string    `json:"short_desc"`
	FullDesc    string    `json:"full_desc"`
	Tags        []string  `json:"tags"`
	Cost        int       `json:"cost"`
	Prompt      string    `json:"prompt"`
	Temperature float64   `json:"temperature"`
	MaxTokens   int       `json:"max_tokens"`
	ModelName   string    `json:"model_name"`
	BaseURL     string    `json:"base_url"`
	APIKey      string    `json:"api_key"`
	Featured    bool      `json:"featured"`
	Speed       string    `json:"speed"`
	Precision   string    `json:"precision"`
	Tools       []toolReq `json:"tools"`
}

type updateAgentReq struct {
	Name        *string   `json:"name"`
	Icon        *string   `json:"icon"`
	Color       *string   `json:"color"`
	Category    *string   `json:"category"`
	ShortDesc   *string   `json:"short_desc"`
	FullDesc    *string   `json:"full_desc"`
	Tags        []string  `json:"tags"`
	Cost        *int      `json:"cost"`
	Prompt      *string   `json:"prompt"`
	Temperature *float64  `json:"temperature"`
	MaxTokens   *int      `json:"max_tokens"`
	ModelName   *string   `json:"model_name"`
	BaseURL     *string   `json:"base_url"`
	APIKey      *string   `json:"api_key"`
	Featured    *bool     `json:"featured"`
	Speed       *string   `json:"speed"`
	Precision   *string   `json:"precision"`
	Tools       []toolReq `json:"tools"`
}

func (h *Handler) CreateAgent(c *gin.Context) {
	var req createAgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	svcReq := agentSvc.CreateRequest{
		Name:        req.Name,
		Icon:        req.Icon,
		Color:       req.Color,
		Category:    req.Category,
		ShortDesc:   req.ShortDesc,
		FullDesc:    req.FullDesc,
		Tags:        req.Tags,
		Cost:        req.Cost,
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		ModelName:   req.ModelName,
		BaseURL:     req.BaseURL,
		APIKey:      req.APIKey,
		Featured:    req.Featured,
		Speed:       req.Speed,
		Precision:   req.Precision,
	}
	for _, t := range req.Tools {
		svcReq.Tools = append(svcReq.Tools, agentSvc.ToolRequest{
			Name:        t.Name,
			Description: t.Description,
			Type:        t.Type,
			InputSchema: rawOrDefault(t.InputSchema, `{"type":"object","properties":{}}`),
			Config:      rawOrDefault(t.Config, "{}"),
		})
	}

	agent, err := h.agentSvc.Create(c.Request.Context(), svcReq)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Created(c, agent)
}

func (h *Handler) ListAgents(c *gin.Context) {
	page, limit := parsePagination(c)

	agents, total, err := h.agentSvc.AdminList(c.Request.Context(), page, limit)
	if err != nil {
		handleError(c, err)
		return
	}
	response.SuccessWithMeta(c, agents, &response.Meta{
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *Handler) GetAgent(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	detail, err := h.agentSvc.AdminGetByID(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, detail)
}

func (h *Handler) UpdateAgent(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	var req updateAgentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	svcReq := agentSvc.UpdateRequest{
		Name:        req.Name,
		Icon:        req.Icon,
		Color:       req.Color,
		Category:    req.Category,
		ShortDesc:   req.ShortDesc,
		FullDesc:    req.FullDesc,
		Tags:        req.Tags,
		Cost:        req.Cost,
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		ModelName:   req.ModelName,
		BaseURL:     req.BaseURL,
		APIKey:      req.APIKey,
		Featured:    req.Featured,
		Speed:       req.Speed,
		Precision:   req.Precision,
	}
	if req.Tools != nil {
		var tools []agentSvc.ToolRequest
		for _, t := range req.Tools {
			tools = append(tools, agentSvc.ToolRequest{
				Name:        t.Name,
				Description: t.Description,
				Type:        t.Type,
				InputSchema: rawOrDefault(t.InputSchema, `{"type":"object","properties":{}}`),
				Config:      rawOrDefault(t.Config, "{}"),
			})
		}
		svcReq.Tools = tools
	}

	agent, err := h.agentSvc.Update(c.Request.Context(), id, svcReq)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, agent)
}

func (h *Handler) DeleteAgent(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	if err := h.agentSvc.Delete(c.Request.Context(), id); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ToggleAgentStatus(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParam)
		return
	}

	agent, err := h.agentSvc.ToggleStatus(c.Request.Context(), id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, agent)
}

func parseID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
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

func handleError(c *gin.Context, err error) {
	if ec, ok := err.(*errcode.ErrCode); ok {
		response.Error(c, ec)
		return
	}
	response.Error(c, errcode.ErrInternalServer)
}

func rawOrDefault(raw json.RawMessage, def string) string {
	if len(raw) == 0 {
		return def
	}
	return string(raw)
}
