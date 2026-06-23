package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/model"
	agentRepo "github.com/Ddhjx-code/AgentHub/internal/repository/agent"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
)

type Service interface {
	Create(ctx context.Context, req CreateRequest) (*model.Agent, error)
	Update(ctx context.Context, id int64, req UpdateRequest) (*model.Agent, error)
	Delete(ctx context.Context, id int64) error
	ToggleStatus(ctx context.Context, id int64) (*model.Agent, error)
	AdminList(ctx context.Context, page, limit int) ([]*model.Agent, int, error)
	AdminGetByID(ctx context.Context, id int64) (*AgentDetail, error)

	ListActive(ctx context.Context, page, limit int, category, tag string) ([]*model.Agent, int, error)
	GetByID(ctx context.Context, id int64) (*AgentDetail, error)
}

type CreateRequest struct {
	Name        string
	Icon        string
	Color       string
	Category    string
	ShortDesc   string
	FullDesc    string
	Tags        []string
	Cost        int
	Prompt      string
	Temperature float64
	MaxTokens   int
	ModelName   string
	BaseURL     string
	APIKey      string
	Featured    bool
	Speed       string
	Precision   string
	Tools       []ToolRequest
}

type ToolRequest struct {
	Name        string
	Description string
	Type        string
	InputSchema string
	Config      string
}

type UpdateRequest struct {
	Name        *string
	Icon        *string
	Color       *string
	Category    *string
	ShortDesc   *string
	FullDesc    *string
	Tags        []string
	Cost        *int
	Prompt      *string
	Temperature *float64
	MaxTokens   *int
	ModelName   *string
	BaseURL     *string
	APIKey      *string
	Featured    *bool
	Speed       *string
	Precision   *string
	Tools       []ToolRequest
}

type AdminToolView struct {
	ID          int64           `json:"id"`
	AgentID     int64           `json:"agent_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	InputSchema json.RawMessage `json:"input_schema"`
	Config      json.RawMessage `json:"config"`
	CreatedAt   time.Time       `json:"created_at"`
}

type AgentDetail struct {
	*model.Agent
	Tools      []*model.AgentTool `json:"tools"`
	AdminTools []AdminToolView    `json:"admin_tools,omitempty"`
	MaskedKey  string             `json:"api_key,omitempty"`
}

type service struct {
	agentRepo agentRepo.Repository
	logger    *slog.Logger
}

func NewService(ar agentRepo.Repository, logger *slog.Logger) Service {
	return &service{agentRepo: ar, logger: logger}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*model.Agent, error) {
	agent := &model.Agent{
		Name:        req.Name,
		Icon:        req.Icon,
		Color:       req.Color,
		Category:    req.Category,
		ShortDesc:   req.ShortDesc,
		FullDesc:    req.FullDesc,
		Tags:        req.Tags,
		Cost:        req.Cost,
		Status:      model.AgentStatusInactive,
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

	if err := s.agentRepo.Create(ctx, agent); err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	for _, t := range req.Tools {
		tool := &model.AgentTool{
			AgentID:     agent.ID,
			Name:        t.Name,
			Description: t.Description,
			Type:        t.Type,
			InputSchema: []byte(t.InputSchema),
			Config:      []byte(t.Config),
		}
		if err := s.agentRepo.CreateTool(ctx, tool); err != nil {
			return nil, fmt.Errorf("create tool %q: %w", t.Name, err)
		}
	}

	s.logger.Info("agent created", "agent_id", agent.ID)
	return agent, nil
}

func (s *service) Update(ctx context.Context, id int64, req UpdateRequest) (*model.Agent, error) {
	agent, err := s.agentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	if agent == nil {
		return nil, errcode.ErrNotFound
	}

	applyUpdate(agent, req)

	if err := s.agentRepo.Update(ctx, agent); err != nil {
		return nil, fmt.Errorf("update agent: %w", err)
	}

	if req.Tools != nil {
		if err := s.agentRepo.DeleteToolsByAgentID(ctx, id); err != nil {
			return nil, fmt.Errorf("delete old tools: %w", err)
		}
		for _, t := range req.Tools {
			tool := &model.AgentTool{
				AgentID:     id,
				Name:        t.Name,
				Description: t.Description,
				Type:        t.Type,
				InputSchema: []byte(t.InputSchema),
				Config:      []byte(t.Config),
			}
			if err := s.agentRepo.CreateTool(ctx, tool); err != nil {
				return nil, fmt.Errorf("create tool %q: %w", t.Name, err)
			}
		}
	}

	return agent, nil
}

func (s *service) Delete(ctx context.Context, id int64) error {
	agent, err := s.agentRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("get agent: %w", err)
	}
	if agent == nil {
		return errcode.ErrNotFound
	}
	if err := s.agentRepo.DeleteToolsByAgentID(ctx, id); err != nil {
		return fmt.Errorf("delete tools: %w", err)
	}
	if err := s.agentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete agent: %w", err)
	}
	s.logger.Info("agent deleted", "agent_id", id)
	return nil
}

func (s *service) ToggleStatus(ctx context.Context, id int64) (*model.Agent, error) {
	agent, err := s.agentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	if agent == nil {
		return nil, errcode.ErrNotFound
	}

	if agent.Status == model.AgentStatusActive {
		agent.Status = model.AgentStatusInactive
	} else {
		agent.Status = model.AgentStatusActive
	}

	if err := s.agentRepo.Update(ctx, agent); err != nil {
		return nil, fmt.Errorf("toggle status: %w", err)
	}
	return agent, nil
}

func (s *service) AdminList(ctx context.Context, page, limit int) ([]*model.Agent, int, error) {
	filter := agentRepo.ListFilter{Page: page, Limit: limit}
	agents, total, err := s.agentRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list agents: %w", err)
	}
	return agents, total, nil
}

func (s *service) AdminGetByID(ctx context.Context, id int64) (*AgentDetail, error) {
	agent, err := s.agentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	if agent == nil {
		return nil, errcode.ErrNotFound
	}
	return s.buildAdminDetail(ctx, agent)
}

func (s *service) ListActive(ctx context.Context, page, limit int, category, tag string) ([]*model.Agent, int, error) {
	filter := agentRepo.ListFilter{
		Status:   model.AgentStatusActive,
		Category: category,
		Tag:      tag,
		Page:     page,
		Limit:    limit,
	}
	agents, total, err := s.agentRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list active agents: %w", err)
	}
	return agents, total, nil
}

func (s *service) GetByID(ctx context.Context, id int64) (*AgentDetail, error) {
	agent, err := s.agentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	if agent == nil || agent.Status != model.AgentStatusActive {
		return nil, errcode.ErrNotFound
	}
	return s.buildDetail(ctx, agent)
}

func (s *service) buildDetail(ctx context.Context, agent *model.Agent) (*AgentDetail, error) {
	tools, err := s.agentRepo.ListToolsByAgentID(ctx, agent.ID)
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	if tools == nil {
		tools = []*model.AgentTool{}
	}
	detail := &AgentDetail{Agent: agent, Tools: tools}
	if agent.APIKey != "" {
		detail.MaskedKey = maskKey(agent.APIKey)
	}
	return detail, nil
}

func (s *service) buildAdminDetail(ctx context.Context, agent *model.Agent) (*AgentDetail, error) {
	detail, err := s.buildDetail(ctx, agent)
	if err != nil {
		return nil, err
	}
	adminTools := make([]AdminToolView, len(detail.Tools))
	for i, t := range detail.Tools {
		adminTools[i] = AdminToolView{
			ID:          t.ID,
			AgentID:     t.AgentID,
			Name:        t.Name,
			Description: t.Description,
			Type:        t.Type,
			InputSchema: t.InputSchema,
			Config:      t.Config,
			CreatedAt:   t.CreatedAt,
		}
	}
	detail.AdminTools = adminTools
	return detail, nil
}

func maskKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func applyUpdate(agent *model.Agent, req UpdateRequest) {
	if req.Name != nil {
		agent.Name = *req.Name
	}
	if req.Icon != nil {
		agent.Icon = *req.Icon
	}
	if req.Color != nil {
		agent.Color = *req.Color
	}
	if req.Category != nil {
		agent.Category = *req.Category
	}
	if req.ShortDesc != nil {
		agent.ShortDesc = *req.ShortDesc
	}
	if req.FullDesc != nil {
		agent.FullDesc = *req.FullDesc
	}
	if req.Tags != nil {
		agent.Tags = req.Tags
	}
	if req.Cost != nil {
		agent.Cost = *req.Cost
	}
	if req.Prompt != nil {
		agent.Prompt = *req.Prompt
	}
	if req.Temperature != nil {
		agent.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		agent.MaxTokens = *req.MaxTokens
	}
	if req.ModelName != nil {
		agent.ModelName = *req.ModelName
	}
	if req.BaseURL != nil {
		agent.BaseURL = *req.BaseURL
	}
	if req.APIKey != nil && *req.APIKey != "" {
		agent.APIKey = *req.APIKey
	}
	if req.Featured != nil {
		agent.Featured = *req.Featured
	}
	if req.Speed != nil {
		agent.Speed = *req.Speed
	}
	if req.Precision != nil {
		agent.Precision = *req.Precision
	}
}
