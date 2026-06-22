package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/database"
	"github.com/Ddhjx-code/AgentHub/internal/llm"
	"github.com/Ddhjx-code/AgentHub/internal/model"
	agentRepo "github.com/Ddhjx-code/AgentHub/internal/repository/agent"
	convRepo "github.com/Ddhjx-code/AgentHub/internal/repository/conversation"
	msgRepo "github.com/Ddhjx-code/AgentHub/internal/repository/message"
	txRepo "github.com/Ddhjx-code/AgentHub/internal/repository/transaction"
	walletRepo "github.com/Ddhjx-code/AgentHub/internal/repository/wallet"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
)

type mockLLMClient struct {
	responses []llm.ChatResponse
	callIndex int
}

func (m *mockLLMClient) ChatCompletion(_ context.Context, _, _ string, _ llm.ChatRequest) (*llm.ChatResponse, error) {
	if m.callIndex >= len(m.responses) {
		return &m.responses[len(m.responses)-1], nil
	}
	resp := m.responses[m.callIndex]
	m.callIndex++
	return &resp, nil
}

type mockToolExecutor struct {
	results map[string]string
}

func (m *mockToolExecutor) Execute(_ context.Context, _ string, _ json.RawMessage, arguments string) (string, error) {
	if result, ok := m.results[arguments]; ok {
		return result, nil
	}
	return "mock result", nil
}

type testEnv struct {
	svc       Service
	agentRepo agentRepo.Repository
	walletRepo walletRepo.Repository
	userID    int64
}

func setupTest(t *testing.T, llmClient llm.Client, toolExec *mockToolExecutor) *testEnv {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	result, err := db.Exec(
		`INSERT INTO users (email, name, password, status) VALUES (?, ?, ?, ?)`,
		"chat@example.com", "ChatUser", "hashed", "active",
	)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	userID, _ := result.LastInsertId()

	ar := agentRepo.NewRepository(db)
	cr := convRepo.NewRepository(db)
	mr := msgRepo.NewRepository(db)
	wr := walletRepo.NewRepository(db)
	tr := txRepo.NewRepository(db)
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	svc := NewService(ar, cr, mr, wr, tr, llmClient, toolExec, nil, lg)

	return &testEnv{
		svc:        svc,
		agentRepo:  ar,
		walletRepo: wr,
		userID:     userID,
	}
}

func createActiveAgent(t *testing.T, env *testEnv, cost int) *model.Agent {
	t.Helper()
	agent := &model.Agent{
		Name:        "TestAgent",
		Status:      model.AgentStatusActive,
		Cost:        cost,
		ModelName:   "test-model",
		BaseURL:     "http://localhost",
		APIKey:      "test-key",
		Prompt:      "You are a helpful assistant.",
		Temperature: 0.7,
		MaxTokens:   1024,
		Tags:        []string{},
	}
	if err := env.agentRepo.Create(context.Background(), agent); err != nil {
		t.Fatalf("create agent: %v", err)
	}
	return agent
}

func TestSendMessageBasic(t *testing.T) {
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"}}},
		},
	}
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	agent := createActiveAgent(t, env, 0)

	resp, err := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "Hi there")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if resp.Reply != "Hello!" {
		t.Fatalf("expected Hello!, got %s", resp.Reply)
	}
	if resp.ConversationID == 0 {
		t.Fatal("expected non-zero conversation ID")
	}
}

func TestSendMessageWithToolCalling(t *testing.T) {
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{
				Message: llm.Message{
					Role: "assistant",
					ToolCalls: []llm.ToolCall{
						{ID: "call_1", Type: "function", Function: llm.ToolCallFunction{Name: "search", Arguments: `{"query":"test"}`}},
					},
				},
				FinishReason: "tool_calls",
			}}},
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Found: test results"}, FinishReason: "stop"}}},
		},
	}
	mockTool := &mockToolExecutor{results: map[string]string{`{"query":"test"}`: "search results for test"}}
	env := setupTest(t, mockLLM, mockTool)
	ctx := context.Background()

	agent := createActiveAgent(t, env, 0)
	env.agentRepo.CreateTool(ctx, &model.AgentTool{
		AgentID:     agent.ID,
		Name:        "search",
		Description: "Search the web",
		Type:        model.ToolTypeCoze,
		InputSchema: json.RawMessage(`{"type":"object"}`),
		Config:      json.RawMessage(`{}`),
	})

	resp, err := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "search for test")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if resp.Reply != "Found: test results" {
		t.Fatalf("expected Found: test results, got %s", resp.Reply)
	}
}

func TestSendMessageInsufficientBalance(t *testing.T) {
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Hello!"}, FinishReason: "stop"}}},
		},
	}
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	agent := createActiveAgent(t, env, 10)

	_, err := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "Hi")
	if err != errcode.ErrInsufficientFund {
		t.Fatalf("expected ErrInsufficientFund, got %v", err)
	}
}

func TestSendMessageWithBilling(t *testing.T) {
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Paid reply"}, FinishReason: "stop"}}},
		},
	}
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	env.walletRepo.Create(ctx, &model.Wallet{UserID: env.userID, Balance: 100})
	agent := createActiveAgent(t, env, 10)

	resp, err := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "Hi")
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	if resp.Reply != "Paid reply" {
		t.Fatalf("expected Paid reply, got %s", resp.Reply)
	}

	wallet, _ := env.walletRepo.GetByUserID(ctx, env.userID)
	if wallet.Balance != 90 {
		t.Fatalf("expected balance 90, got %d", wallet.Balance)
	}
}

func TestSendMessageAgentOffline(t *testing.T) {
	mockLLM := &mockLLMClient{}
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	agent := &model.Agent{
		Name:   "OfflineAgent",
		Status: model.AgentStatusInactive,
		Tags:   []string{},
	}
	env.agentRepo.Create(ctx, agent)

	_, err := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "Hi")
	if err != errcode.ErrAgentOffline {
		t.Fatalf("expected ErrAgentOffline, got %v", err)
	}
}

func TestSendMessageContinueConversation(t *testing.T) {
	callCount := 0
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Reply 1"}, FinishReason: "stop"}}},
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Reply 2"}, FinishReason: "stop"}}},
		},
	}
	_ = callCount
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	agent := createActiveAgent(t, env, 0)

	resp1, err := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "First message")
	if err != nil {
		t.Fatalf("first message: %v", err)
	}

	convID := resp1.ConversationID
	resp2, err := env.svc.SendMessage(ctx, env.userID, agent.ID, &convID, "Second message")
	if err != nil {
		t.Fatalf("second message: %v", err)
	}
	if resp2.ConversationID != convID {
		t.Fatalf("expected same conversation ID %d, got %d", convID, resp2.ConversationID)
	}
	if resp2.Reply != "Reply 2" {
		t.Fatalf("expected Reply 2, got %s", resp2.Reply)
	}
}

func TestListConversations(t *testing.T) {
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "R1"}, FinishReason: "stop"}}},
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "R2"}, FinishReason: "stop"}}},
		},
	}
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	agent := createActiveAgent(t, env, 0)

	env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "Hello")
	env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "World")

	convs, err := env.svc.ListConversations(ctx, env.userID)
	if err != nil {
		t.Fatalf("list conversations: %v", err)
	}
	if len(convs) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(convs))
	}
}

func TestGetMessages(t *testing.T) {
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Reply"}, FinishReason: "stop"}}},
		},
	}
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	agent := createActiveAgent(t, env, 0)
	resp, _ := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "Hello")

	msgs, err := env.svc.GetMessages(ctx, env.userID, resp.ConversationID)
	if err != nil {
		t.Fatalf("get messages: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages (user+assistant), got %d", len(msgs))
	}
	if msgs[0].Role != model.RoleUser {
		t.Fatalf("expected first message role user, got %s", msgs[0].Role)
	}
	if msgs[1].Role != model.RoleAssistant {
		t.Fatalf("expected second message role assistant, got %s", msgs[1].Role)
	}
}

func TestDeleteConversation(t *testing.T) {
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Reply"}, FinishReason: "stop"}}},
		},
	}
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	agent := createActiveAgent(t, env, 0)
	resp, _ := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "Hello")

	if err := env.svc.DeleteConversation(ctx, env.userID, resp.ConversationID); err != nil {
		t.Fatalf("delete conversation: %v", err)
	}

	_, err := env.svc.GetMessages(ctx, env.userID, resp.ConversationID)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteConversationNotOwned(t *testing.T) {
	mockLLM := &mockLLMClient{
		responses: []llm.ChatResponse{
			{Choices: []llm.Choice{{Message: llm.Message{Role: "assistant", Content: "Reply"}, FinishReason: "stop"}}},
		},
	}
	env := setupTest(t, mockLLM, &mockToolExecutor{})
	ctx := context.Background()

	agent := createActiveAgent(t, env, 0)
	resp, _ := env.svc.SendMessage(ctx, env.userID, agent.ID, nil, "Hello")

	err := env.svc.DeleteConversation(ctx, 9999, resp.ConversationID)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound for other user, got %v", err)
	}
}
