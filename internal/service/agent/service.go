package agent

import (
	"context"
	"fmt"
	"log/slog"

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
	Engine      string
	Prompt      string
	Temperature float64
	MaxTokens   int
	Featured    bool
	Speed       string
	Precision   string

	CozeWorkflowID  string
	CozeAPIKey      string
	CozeRegion      string
	CozeInputField  string
	CozeOutputField string

	N8NWebhookURL  string
	N8NAuthType    string
	N8NAuthToken   string
	N8NTimeout     int
	N8NPayloadTmpl string
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
	Engine      *string
	Prompt      *string
	Temperature *float64
	MaxTokens   *int
	Featured    *bool
	Speed       *string
	Precision   *string

	CozeWorkflowID  *string
	CozeAPIKey      *string
	CozeRegion      *string
	CozeInputField  *string
	CozeOutputField *string

	N8NWebhookURL  *string
	N8NAuthType    *string
	N8NAuthToken   *string
	N8NTimeout     *int
	N8NPayloadTmpl *string
}

type AgentDetail struct {
	*model.Agent
	CozeWorkflow *model.CozeWorkflow `json:"coze_workflow,omitempty"`
	N8NWorkflow  *model.N8NWorkflow  `json:"n8n_workflow,omitempty"`
}

type service struct {
	agentRepo agentRepo.Repository
	logger    *slog.Logger
}

func NewService(ar agentRepo.Repository, logger *slog.Logger) Service {
	return &service{agentRepo: ar, logger: logger}
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*model.Agent, error) {
	if req.Engine != model.EngineCoze && req.Engine != model.EngineN8N {
		return nil, errcode.ErrInvalidEngine
	}

	agent := &model.Agent{
		Name:        req.Name,
		Icon:        req.Icon,
		Color:       req.Color,
		Category:    req.Category,
		ShortDesc:   req.ShortDesc,
		FullDesc:    req.FullDesc,
		Tags:        req.Tags,
		Cost:        req.Cost,
		Engine:      req.Engine,
		Status:      model.AgentStatusInactive,
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Featured:    req.Featured,
		Speed:       req.Speed,
		Precision:   req.Precision,
	}

	if err := s.agentRepo.Create(ctx, agent); err != nil {
		return nil, fmt.Errorf("create agent: %w", err)
	}

	if err := s.createWorkflow(ctx, agent.ID, req); err != nil {
		return nil, fmt.Errorf("create workflow: %w", err)
	}

	s.logger.Info("agent created", "agent_id", agent.ID, "engine", agent.Engine)
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

	oldEngine := agent.Engine
	applyUpdate(agent, req)

	if req.Engine != nil && *req.Engine != oldEngine {
		if *req.Engine != model.EngineCoze && *req.Engine != model.EngineN8N {
			return nil, errcode.ErrInvalidEngine
		}
		if oldEngine == model.EngineCoze {
			s.agentRepo.DeleteCozeWorkflow(ctx, id)
		} else {
			s.agentRepo.DeleteN8NWorkflow(ctx, id)
		}
		createReq := workflowFromUpdate(id, req)
		if err := s.createWorkflowFromParts(ctx, *req.Engine, createReq); err != nil {
			return nil, fmt.Errorf("create new workflow: %w", err)
		}
	} else {
		if err := s.updateWorkflow(ctx, agent.ID, agent.Engine, req); err != nil {
			return nil, fmt.Errorf("update workflow: %w", err)
		}
	}

	if err := s.agentRepo.Update(ctx, agent); err != nil {
		return nil, fmt.Errorf("update agent: %w", err)
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
	return s.buildDetail(ctx, agent)
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
	detail := &AgentDetail{Agent: agent}
	switch agent.Engine {
	case model.EngineCoze:
		wf, err := s.agentRepo.GetCozeWorkflow(ctx, agent.ID)
		if err != nil {
			return nil, fmt.Errorf("get coze workflow: %w", err)
		}
		detail.CozeWorkflow = wf
	case model.EngineN8N:
		wf, err := s.agentRepo.GetN8NWorkflow(ctx, agent.ID)
		if err != nil {
			return nil, fmt.Errorf("get n8n workflow: %w", err)
		}
		detail.N8NWorkflow = wf
	}
	return detail, nil
}

func (s *service) createWorkflow(ctx context.Context, agentID int64, req CreateRequest) error {
	switch req.Engine {
	case model.EngineCoze:
		return s.agentRepo.CreateCozeWorkflow(ctx, &model.CozeWorkflow{
			AgentID:     agentID,
			WorkflowID:  req.CozeWorkflowID,
			APIKey:      req.CozeAPIKey,
			Region:      req.CozeRegion,
			InputField:  req.CozeInputField,
			OutputField: req.CozeOutputField,
		})
	case model.EngineN8N:
		timeout := req.N8NTimeout
		if timeout == 0 {
			timeout = 30
		}
		return s.agentRepo.CreateN8NWorkflow(ctx, &model.N8NWorkflow{
			AgentID:     agentID,
			WebhookURL:  req.N8NWebhookURL,
			AuthType:    req.N8NAuthType,
			AuthToken:   req.N8NAuthToken,
			Timeout:     timeout,
			PayloadTmpl: req.N8NPayloadTmpl,
		})
	}
	return nil
}

type workflowParts struct {
	CozeWorkflowID  string
	CozeAPIKey      string
	CozeRegion      string
	CozeInputField  string
	CozeOutputField string
	N8NWebhookURL   string
	N8NAuthType     string
	N8NAuthToken    string
	N8NTimeout      int
	N8NPayloadTmpl  string
}

func workflowFromUpdate(agentID int64, req UpdateRequest) workflowParts {
	var p workflowParts
	if req.CozeWorkflowID != nil {
		p.CozeWorkflowID = *req.CozeWorkflowID
	}
	if req.CozeAPIKey != nil {
		p.CozeAPIKey = *req.CozeAPIKey
	}
	if req.CozeRegion != nil {
		p.CozeRegion = *req.CozeRegion
	}
	if req.CozeInputField != nil {
		p.CozeInputField = *req.CozeInputField
	}
	if req.CozeOutputField != nil {
		p.CozeOutputField = *req.CozeOutputField
	}
	if req.N8NWebhookURL != nil {
		p.N8NWebhookURL = *req.N8NWebhookURL
	}
	if req.N8NAuthType != nil {
		p.N8NAuthType = *req.N8NAuthType
	}
	if req.N8NAuthToken != nil {
		p.N8NAuthToken = *req.N8NAuthToken
	}
	if req.N8NTimeout != nil {
		p.N8NTimeout = *req.N8NTimeout
	}
	if req.N8NPayloadTmpl != nil {
		p.N8NPayloadTmpl = *req.N8NPayloadTmpl
	}
	return p
}

func (s *service) createWorkflowFromParts(ctx context.Context, engine string, p workflowParts) error {
	switch engine {
	case model.EngineCoze:
		return s.agentRepo.CreateCozeWorkflow(ctx, &model.CozeWorkflow{
			WorkflowID:  p.CozeWorkflowID,
			APIKey:      p.CozeAPIKey,
			Region:      p.CozeRegion,
			InputField:  p.CozeInputField,
			OutputField: p.CozeOutputField,
		})
	case model.EngineN8N:
		timeout := p.N8NTimeout
		if timeout == 0 {
			timeout = 30
		}
		return s.agentRepo.CreateN8NWorkflow(ctx, &model.N8NWorkflow{
			WebhookURL:  p.N8NWebhookURL,
			AuthType:    p.N8NAuthType,
			AuthToken:   p.N8NAuthToken,
			Timeout:     timeout,
			PayloadTmpl: p.N8NPayloadTmpl,
		})
	}
	return nil
}

func (s *service) updateWorkflow(ctx context.Context, agentID int64, engine string, req UpdateRequest) error {
	hasCozeUpdate := req.CozeWorkflowID != nil || req.CozeAPIKey != nil || req.CozeRegion != nil ||
		req.CozeInputField != nil || req.CozeOutputField != nil
	hasN8NUpdate := req.N8NWebhookURL != nil || req.N8NAuthType != nil || req.N8NAuthToken != nil ||
		req.N8NTimeout != nil || req.N8NPayloadTmpl != nil

	switch engine {
	case model.EngineCoze:
		if !hasCozeUpdate {
			return nil
		}
		existing, err := s.agentRepo.GetCozeWorkflow(ctx, agentID)
		if err != nil {
			return err
		}
		if existing == nil {
			return nil
		}
		if req.CozeWorkflowID != nil {
			existing.WorkflowID = *req.CozeWorkflowID
		}
		if req.CozeAPIKey != nil {
			existing.APIKey = *req.CozeAPIKey
		}
		if req.CozeRegion != nil {
			existing.Region = *req.CozeRegion
		}
		if req.CozeInputField != nil {
			existing.InputField = *req.CozeInputField
		}
		if req.CozeOutputField != nil {
			existing.OutputField = *req.CozeOutputField
		}
		return s.agentRepo.UpdateCozeWorkflow(ctx, existing)
	case model.EngineN8N:
		if !hasN8NUpdate {
			return nil
		}
		existing, err := s.agentRepo.GetN8NWorkflow(ctx, agentID)
		if err != nil {
			return err
		}
		if existing == nil {
			return nil
		}
		if req.N8NWebhookURL != nil {
			existing.WebhookURL = *req.N8NWebhookURL
		}
		if req.N8NAuthType != nil {
			existing.AuthType = *req.N8NAuthType
		}
		if req.N8NAuthToken != nil {
			existing.AuthToken = *req.N8NAuthToken
		}
		if req.N8NTimeout != nil {
			existing.Timeout = *req.N8NTimeout
		}
		if req.N8NPayloadTmpl != nil {
			existing.PayloadTmpl = *req.N8NPayloadTmpl
		}
		return s.agentRepo.UpdateN8NWorkflow(ctx, existing)
	}
	return nil
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
	if req.Engine != nil {
		agent.Engine = *req.Engine
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
