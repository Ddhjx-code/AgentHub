package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/Ddhjx-code/AgentHub/internal/llm"
	"github.com/Ddhjx-code/AgentHub/internal/model"
	agentRepo "github.com/Ddhjx-code/AgentHub/internal/repository/agent"
	convRepo "github.com/Ddhjx-code/AgentHub/internal/repository/conversation"
	msgRepo "github.com/Ddhjx-code/AgentHub/internal/repository/message"
	txRepo "github.com/Ddhjx-code/AgentHub/internal/repository/transaction"
	walletRepo "github.com/Ddhjx-code/AgentHub/internal/repository/wallet"
	kbSvc "github.com/Ddhjx-code/AgentHub/internal/service/knowledge"
	"github.com/Ddhjx-code/AgentHub/internal/tool"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
)

const (
	maxToolLoops   = 5
	historyLimit   = 20
	titleMaxLength = 50
)

type Service interface {
	SendMessage(ctx context.Context, userID, agentID int64, conversationID *int64, content string) (*SendMessageResponse, error)
	ListConversations(ctx context.Context, userID int64) ([]*model.Conversation, error)
	GetMessages(ctx context.Context, userID, conversationID int64) ([]*model.Message, error)
	DeleteConversation(ctx context.Context, userID, conversationID int64) error
}

type SendMessageResponse struct {
	ConversationID int64  `json:"conversation_id"`
	Reply          string `json:"reply"`
}

type service struct {
	agentRepo  agentRepo.Repository
	convRepo   convRepo.Repository
	msgRepo    msgRepo.Repository
	walletRepo walletRepo.Repository
	txRepo     txRepo.Repository
	llmClient  llm.Client
	toolExec   tool.Executor
	kbSvc      kbSvc.Service
	logger     *slog.Logger
}

func NewService(
	ar agentRepo.Repository,
	cr convRepo.Repository,
	mr msgRepo.Repository,
	wr walletRepo.Repository,
	tr txRepo.Repository,
	lc llm.Client,
	te tool.Executor,
	ks kbSvc.Service,
	logger *slog.Logger,
) Service {
	return &service{
		agentRepo:  ar,
		convRepo:   cr,
		msgRepo:    mr,
		walletRepo: wr,
		txRepo:     tr,
		llmClient:  lc,
		toolExec:   te,
		kbSvc:      ks,
		logger:     logger,
	}
}

func (s *service) SendMessage(ctx context.Context, userID, agentID int64, conversationID *int64, content string) (*SendMessageResponse, error) {
	agent, err := s.agentRepo.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("get agent: %w", err)
	}
	if agent == nil || agent.Status != model.AgentStatusActive {
		return nil, errcode.ErrAgentOffline
	}

	if agent.Cost > 0 {
		wallet, err := s.walletRepo.GetByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("get wallet: %w", err)
		}
		if wallet == nil || wallet.Balance < int64(agent.Cost) {
			return nil, errcode.ErrInsufficientFund
		}
	}

	var convID int64
	if conversationID != nil && *conversationID > 0 {
		conv, err := s.convRepo.GetByID(ctx, *conversationID)
		if err != nil {
			return nil, fmt.Errorf("get conversation: %w", err)
		}
		if conv == nil || conv.UserID != userID {
			return nil, errcode.ErrNotFound
		}
		convID = conv.ID
	} else {
		title := content
		if len(title) > titleMaxLength {
			title = title[:titleMaxLength]
		}
		conv := &model.Conversation{
			UserID:  userID,
			AgentID: agentID,
			Title:   title,
		}
		if err := s.convRepo.Create(ctx, conv); err != nil {
			return nil, fmt.Errorf("create conversation: %w", err)
		}
		convID = conv.ID
	}

	userMsg := &model.Message{
		ConversationID: convID,
		Role:           model.RoleUser,
		Content:        content,
	}
	if err := s.msgRepo.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("save user message: %w", err)
	}

	history, err := s.msgRepo.ListByConversationID(ctx, convID, historyLimit)
	if err != nil {
		return nil, fmt.Errorf("load history: %w", err)
	}

	tools, err := s.agentRepo.ListToolsByAgentID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}

	systemPrompt := agent.Prompt
	hasKB := false

	if s.kbSvc != nil {
		results, err := s.kbSvc.Search(ctx, agentID, content, 5)
		if err != nil {
			s.logger.Error("RAG search failed", "agent_id", agentID, "error", err)
		}
		if len(results) > 0 {
			ragContext := kbSvc.FormatSearchResults(results)
			systemPrompt = systemPrompt +
				"\n\n## Reference Knowledge\n" +
				"The following are retrieved reference materials. " +
				"If they are not relevant to the user's question, ignore them and answer based on your own knowledge. " +
				"Do not fabricate information that is not present in the references.\n\n" +
				ragContext
			hasKB = true
		}
	}

	messages := buildLLMMessages(systemPrompt, history)
	llmTools := buildLLMTools(tools)

	if hasKB {
		llmTools = append(llmTools, llm.Tool{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        model.ToolTypeKnowledgeSearch,
				Description: "Search the knowledge base for relevant information. Use when you need more details about a topic.",
				Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"search query"}},"required":["query"]}`),
			},
		})
	}

	reply, err := s.llmLoop(ctx, agent, tools, messages, llmTools, convID)
	if err != nil {
		return nil, err
	}

	if agent.Cost > 0 {
		if err := s.walletRepo.Deduct(ctx, userID, agent.Cost); err != nil {
			s.logger.Error("deduct wallet failed", "user_id", userID, "error", err)
		} else {
			agentID := agent.ID
			s.txRepo.Create(ctx, &model.Transaction{
				UserID:    userID,
				Type:      model.TxTypeUse,
				AgentID:   &agentID,
				AgentName: agent.Name,
				Amount:    agent.Cost,
				Status:    model.TxStatusSuccess,
			})
		}
	}

	s.agentRepo.IncrementCalls(ctx, agent.ID)

	return &SendMessageResponse{
		ConversationID: convID,
		Reply:          reply,
	}, nil
}

func (s *service) llmLoop(ctx context.Context, agent *model.Agent, tools []*model.AgentTool, messages []llm.Message, llmTools []llm.Tool, convID int64) (string, error) {
	for i := 0; i < maxToolLoops; i++ {
		req := llm.ChatRequest{
			Model:       agent.ModelName,
			Messages:    messages,
			Temperature: agent.Temperature,
			MaxTokens:   agent.MaxTokens,
		}
		if len(llmTools) > 0 {
			req.Tools = llmTools
		}

		resp, err := s.llmClient.ChatCompletion(ctx, agent.BaseURL, agent.APIKey, req)
		if err != nil {
			return "", errcode.ErrLLMError
		}
		if len(resp.Choices) == 0 {
			return "", errcode.ErrLLMError
		}

		choice := resp.Choices[0]

		if len(choice.Message.ToolCalls) == 0 {
			assistantMsg := &model.Message{
				ConversationID: convID,
				Role:           model.RoleAssistant,
				Content:        choice.Message.Content,
			}
			s.msgRepo.Create(ctx, assistantMsg)
			return choice.Message.Content, nil
		}

		toolCallsJSON, _ := json.Marshal(choice.Message.ToolCalls)
		assistantMsg := &model.Message{
			ConversationID: convID,
			Role:           model.RoleAssistant,
			ToolCalls:      string(toolCallsJSON),
		}
		s.msgRepo.Create(ctx, assistantMsg)
		messages = append(messages, choice.Message)

		for _, tc := range choice.Message.ToolCalls {
			var result string
			if tc.Function.Name == model.ToolTypeKnowledgeSearch && s.kbSvc != nil {
				result = s.handleKnowledgeSearch(ctx, agent.ID, tc.Function.Arguments)
			} else {
				agentTool := findTool(tools, tc.Function.Name)
				if agentTool != nil {
					var execErr error
					result, execErr = s.toolExec.Execute(ctx, agentTool.Type, agentTool.Config, tc.Function.Arguments)
					if execErr != nil {
						result = fmt.Sprintf("Tool execution error: %s", execErr.Error())
						s.logger.Error("tool execution failed", "tool", tc.Function.Name, "error", execErr)
					}
				} else {
					result = fmt.Sprintf("Tool %q not found", tc.Function.Name)
				}
			}

			toolMsg := &model.Message{
				ConversationID: convID,
				Role:           model.RoleTool,
				Content:        result,
				ToolCallID:     tc.ID,
			}
			s.msgRepo.Create(ctx, toolMsg)
			messages = append(messages, llm.Message{
				Role:       model.RoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	return "", errcode.ErrToolLoopLimit
}

func (s *service) ListConversations(ctx context.Context, userID int64) ([]*model.Conversation, error) {
	convs, err := s.convRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	if convs == nil {
		return []*model.Conversation{}, nil
	}
	return convs, nil
}

func (s *service) GetMessages(ctx context.Context, userID, conversationID int64) ([]*model.Message, error) {
	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	if conv == nil || conv.UserID != userID {
		return nil, errcode.ErrNotFound
	}

	msgs, err := s.msgRepo.ListByConversationID(ctx, conversationID, 100)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	if msgs == nil {
		return []*model.Message{}, nil
	}
	return msgs, nil
}

func (s *service) DeleteConversation(ctx context.Context, userID, conversationID int64) error {
	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("get conversation: %w", err)
	}
	if conv == nil || conv.UserID != userID {
		return errcode.ErrNotFound
	}
	if err := s.convRepo.Delete(ctx, conversationID); err != nil {
		return fmt.Errorf("delete conversation: %w", err)
	}
	return nil
}

func buildLLMMessages(systemPrompt string, history []*model.Message) []llm.Message {
	var messages []llm.Message
	if systemPrompt != "" {
		messages = append(messages, llm.Message{Role: model.RoleSystem, Content: systemPrompt})
	}
	for _, m := range history {
		msg := llm.Message{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		if m.ToolCalls != "" {
			var toolCalls []llm.ToolCall
			json.Unmarshal([]byte(m.ToolCalls), &toolCalls)
			msg.ToolCalls = toolCalls
		}
		messages = append(messages, msg)
	}
	return messages
}

func buildLLMTools(tools []*model.AgentTool) []llm.Tool {
	var result []llm.Tool
	for _, t := range tools {
		result = append(result, llm.Tool{
			Type: "function",
			Function: llm.ToolFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.InputSchema,
			},
		})
	}
	return result
}

func findTool(tools []*model.AgentTool, name string) *model.AgentTool {
	for _, t := range tools {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (s *service) handleKnowledgeSearch(ctx context.Context, agentID int64, arguments string) string {
	var args struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(arguments), &args); err != nil {
		return "invalid search query"
	}
	if args.Query == "" {
		return "empty search query"
	}

	results, err := s.kbSvc.Search(ctx, agentID, args.Query, 5)
	if err != nil {
		s.logger.Error("knowledge search failed", "agent_id", agentID, "error", err)
		return "knowledge search failed"
	}
	if len(results) == 0 {
		return "no relevant information found"
	}

	return kbSvc.FormatSearchResults(results)
}
