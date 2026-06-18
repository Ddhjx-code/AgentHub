package agent

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/database"
	"github.com/Ddhjx-code/AgentHub/internal/model"
	"github.com/Ddhjx-code/AgentHub/pkg/errcode"
	agentRepo "github.com/Ddhjx-code/AgentHub/internal/repository/agent"
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

func TestCreateCozeAgent(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, err := svc.Create(ctx, CreateRequest{
		Name:           "CozeBot",
		Engine:         model.EngineCoze,
		Tags:           []string{"ai"},
		Cost:           10,
		CozeWorkflowID: "wf_123",
		CozeAPIKey:     "key_abc",
	})
	if err != nil {
		t.Fatalf("create coze agent: %v", err)
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
	if detail.CozeWorkflow == nil {
		t.Fatal("expected coze workflow to be set")
	}
	if detail.CozeWorkflow.WorkflowID != "wf_123" {
		t.Fatalf("expected wf_123, got %s", detail.CozeWorkflow.WorkflowID)
	}
}

func TestCreateN8NAgent(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, err := svc.Create(ctx, CreateRequest{
		Name:          "N8NBot",
		Engine:        model.EngineN8N,
		N8NWebhookURL: "/webhook/test",
		N8NTimeout:    60,
	})
	if err != nil {
		t.Fatalf("create n8n agent: %v", err)
	}

	detail, _ := svc.AdminGetByID(ctx, agent.ID)
	if detail.N8NWorkflow == nil {
		t.Fatal("expected n8n workflow to be set")
	}
	if detail.N8NWorkflow.WebhookURL != "/webhook/test" {
		t.Fatalf("expected /webhook/test, got %s", detail.N8NWorkflow.WebhookURL)
	}
}

func TestCreateAgentInvalidEngine(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, CreateRequest{
		Name:   "BadBot",
		Engine: "unknown",
	})
	if err != errcode.ErrInvalidEngine {
		t.Fatalf("expected ErrInvalidEngine, got %v", err)
	}
}

func TestUpdateAgent(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, _ := svc.Create(ctx, CreateRequest{
		Name:   "Original",
		Engine: model.EngineCoze,
	})

	newName := "Updated"
	updated, err := svc.Update(ctx, agent.ID, UpdateRequest{Name: &newName})
	if err != nil {
		t.Fatalf("update agent: %v", err)
	}
	if updated.Name != "Updated" {
		t.Fatalf("expected Updated, got %s", updated.Name)
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

	agent, _ := svc.Create(ctx, CreateRequest{
		Name:   "ToDelete",
		Engine: model.EngineCoze,
	})

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

	agent, _ := svc.Create(ctx, CreateRequest{
		Name:   "ToggleBot",
		Engine: model.EngineCoze,
	})
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

	svc.Create(ctx, CreateRequest{Name: "A1", Engine: model.EngineCoze})
	a2, _ := svc.Create(ctx, CreateRequest{Name: "A2", Engine: model.EngineN8N})
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

	svc.Create(ctx, CreateRequest{Name: "Inactive", Engine: model.EngineCoze})
	active, _ := svc.Create(ctx, CreateRequest{Name: "Active", Engine: model.EngineCoze})
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

func TestListActiveWithCategoryFilter(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	a1, _ := svc.Create(ctx, CreateRequest{Name: "Writer", Engine: model.EngineCoze})
	svc.ToggleStatus(ctx, a1.ID)
	cat := "writing"
	svc.Update(ctx, a1.ID, UpdateRequest{Category: &cat})

	a2, _ := svc.Create(ctx, CreateRequest{Name: "Coder", Engine: model.EngineCoze})
	svc.ToggleStatus(ctx, a2.ID)
	cat2 := "coding"
	svc.Update(ctx, a2.ID, UpdateRequest{Category: &cat2})

	agents, total, _ := svc.ListActive(ctx, 1, 20, "coding", "")
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if agents[0].Name != "Coder" {
		t.Fatalf("expected Coder, got %s", agents[0].Name)
	}
}

func TestGetByIDActive(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	agent, _ := svc.Create(ctx, CreateRequest{
		Name:           "ActiveBot",
		Engine:         model.EngineCoze,
		CozeWorkflowID: "wf_active",
	})
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

	agent, _ := svc.Create(ctx, CreateRequest{
		Name:   "InactiveBot",
		Engine: model.EngineCoze,
	})

	_, err := svc.GetByID(ctx, agent.ID)
	if err != errcode.ErrNotFound {
		t.Fatalf("expected ErrNotFound for inactive agent, got %v", err)
	}
}
