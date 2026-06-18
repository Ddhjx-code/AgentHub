package agent

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/database"
	"github.com/Ddhjx-code/AgentHub/internal/model"
	agentRepo "github.com/Ddhjx-code/AgentHub/internal/repository/agent"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
)

func setupTestService(t *testing.T) Service {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	ar := agentRepo.NewRepository(db)
	lg := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	return NewService(ar, lg)
}

func TestCreateAgent(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, err := svc.Create(ctx, CreateRequest{
		Name:      "TestBot",
		Tags:      []string{"ai"},
		Cost:      10,
		ModelName: "deepseek-chat",
		BaseURL:   "https://api.deepseek.com/v1",
		APIKey:    "sk-test",
		Tools: []ToolRequest{
			{
				Name:        "search",
				Description: "Search the web",
				Type:        model.ToolTypeCoze,
				InputSchema: `{"type":"object"}`,
				Config:      `{"workflow_id":"wf_123"}`,
			},
		},
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if agent.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if agent.Status != model.AgentStatusInactive {
		t.Fatalf("expected inactive status, got %s", agent.Status)
	}

	detail, err := svc.AdminGetByID(ctx, agent.ID)
	if err != nil {
		t.Fatalf("admin get by id: %v", err)
	}
	if len(detail.Tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(detail.Tools))
	}
	if detail.Tools[0].Name != "search" {
		t.Fatalf("expected search, got %s", detail.Tools[0].Name)
	}
}

func TestCreateAgentNoTools(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, err := svc.Create(ctx, CreateRequest{
		Name:      "PureLLM",
		ModelName: "gpt-4o",
		BaseURL:   "https://api.openai.com/v1",
		APIKey:    "sk-test",
	})
	if err != nil {
		t.Fatalf("create agent: %v", err)
	}

	detail, _ := svc.AdminGetByID(ctx, agent.ID)
	if len(detail.Tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(detail.Tools))
	}
}

func TestUpdateAgent(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, _ := svc.Create(ctx, CreateRequest{
		Name:      "Original",
		ModelName: "deepseek-chat",
	})

	newName := "Updated"
	newModel := "gpt-4o"
	updated, err := svc.Update(ctx, agent.ID, UpdateRequest{
		Name:      &newName,
		ModelName: &newModel,
	})
	if err != nil {
		t.Fatalf("update agent: %v", err)
	}
	if updated.Name != "Updated" {
		t.Fatalf("expected Updated, got %s", updated.Name)
	}
}

func TestUpdateAgentWithTools(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, _ := svc.Create(ctx, CreateRequest{
		Name: "ToolBot",
		Tools: []ToolRequest{
			{Name: "old_tool", Type: model.ToolTypeCoze, InputSchema: "{}", Config: "{}"},
		},
	})

	_, err := svc.Update(ctx, agent.ID, UpdateRequest{
		Tools: []ToolRequest{
			{Name: "new_tool1", Type: model.ToolTypeCoze, InputSchema: "{}", Config: "{}"},
			{Name: "new_tool2", Type: model.ToolTypeN8N, InputSchema: "{}", Config: "{}"},
		},
	})
	if err != nil {
		t.Fatalf("update with tools: %v", err)
	}

	detail, _ := svc.AdminGetByID(ctx, agent.ID)
	if len(detail.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(detail.Tools))
	}
}

func TestUpdateAgentNotFound(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	newName := "Ghost"
	_, err := svc.Update(ctx, 9999, UpdateRequest{Name: &newName})
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteAgent(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, _ := svc.Create(ctx, CreateRequest{Name: "ToDelete"})

	if err := svc.Delete(ctx, agent.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err := svc.AdminGetByID(ctx, agent.ID)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestDeleteAgentNotFound(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	err := svc.Delete(ctx, 9999)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestToggleStatus(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, _ := svc.Create(ctx, CreateRequest{Name: "ToggleBot"})
	if agent.Status != model.AgentStatusInactive {
		t.Fatalf("expected inactive, got %s", agent.Status)
	}

	toggled, err := svc.ToggleStatus(ctx, agent.ID)
	if err != nil {
		t.Fatalf("toggle: %v", err)
	}
	if toggled.Status != model.AgentStatusActive {
		t.Fatalf("expected active, got %s", toggled.Status)
	}

	toggled2, _ := svc.ToggleStatus(ctx, agent.ID)
	if toggled2.Status != model.AgentStatusInactive {
		t.Fatalf("expected inactive, got %s", toggled2.Status)
	}
}

func TestAdminList(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateRequest{Name: "A1"})
	a2, _ := svc.Create(ctx, CreateRequest{Name: "A2"})
	svc.ToggleStatus(ctx, a2.ID)

	agents, total, err := svc.AdminList(ctx, 1, 20)
	if err != nil {
		t.Fatalf("admin list: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
}

func TestListActive(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	svc.Create(ctx, CreateRequest{Name: "Inactive"})
	active, _ := svc.Create(ctx, CreateRequest{Name: "Active"})
	svc.ToggleStatus(ctx, active.ID)

	agents, total, err := svc.ListActive(ctx, 1, 20, "", "")
	if err != nil {
		t.Fatalf("list active: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if agents[0].Name != "Active" {
		t.Fatalf("expected Active, got %s", agents[0].Name)
	}
}

func TestGetByIDActive(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, _ := svc.Create(ctx, CreateRequest{Name: "ActiveBot"})
	svc.ToggleStatus(ctx, agent.ID)

	detail, err := svc.GetByID(ctx, agent.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if detail.Name != "ActiveBot" {
		t.Fatalf("expected ActiveBot, got %s", detail.Name)
	}
}

func TestGetByIDInactive(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, _ := svc.Create(ctx, CreateRequest{Name: "InactiveBot"})

	_, err := svc.GetByID(ctx, agent.ID)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound for inactive agent, got %v", err)
	}
}
