package agent

import (
	"context"
	"testing"

	"github.com/Ddhjx-code/AgentHub/internal/database"
	"github.com/Ddhjx-code/AgentHub/internal/model"
)

func setupTestDB(t *testing.T) Repository {
	t.Helper()
	db, err := database.New(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewRepository(db)
}

func newTestAgent(name, engine string) *model.Agent {
	return &model.Agent{
		Name:        name,
		Engine:      engine,
		Status:      model.AgentStatusInactive,
		Tags:        []string{"test", "demo"},
		Cost:        10,
		Temperature: 0.7,
		MaxTokens:   2048,
	}
}

func TestCreateAndGetByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("TestBot", model.EngineCoze)
	if err := repo.Create(ctx, agent); err != nil {
		t.Fatalf("create agent: %v", err)
	}
	if agent.ID == 0 {
		t.Fatal("expected non-zero agent ID")
	}

	found, err := repo.GetByID(ctx, agent.ID)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if found == nil {
		t.Fatal("expected agent, got nil")
	}
	if found.Name != "TestBot" {
		t.Fatalf("expected name TestBot, got %s", found.Name)
	}
	if found.Engine != model.EngineCoze {
		t.Fatalf("expected engine coze, got %s", found.Engine)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	found, err := repo.GetByID(ctx, 9999)
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if found != nil {
		t.Fatal("expected nil for non-existent id")
	}
}

func TestUpdateAgent(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("Original", model.EngineCoze)
	repo.Create(ctx, agent)

	agent.Name = "Updated"
	agent.Cost = 50
	if err := repo.Update(ctx, agent); err != nil {
		t.Fatalf("update agent: %v", err)
	}

	found, _ := repo.GetByID(ctx, agent.ID)
	if found.Name != "Updated" {
		t.Fatalf("expected name Updated, got %s", found.Name)
	}
	if found.Cost != 50 {
		t.Fatalf("expected cost 50, got %d", found.Cost)
	}
}

func TestDeleteAgent(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("ToDelete", model.EngineCoze)
	repo.Create(ctx, agent)

	if err := repo.Delete(ctx, agent.ID); err != nil {
		t.Fatalf("delete agent: %v", err)
	}

	found, _ := repo.GetByID(ctx, agent.ID)
	if found != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestListWithPagination(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		repo.Create(ctx, newTestAgent("Bot", model.EngineCoze))
	}

	agents, total, err := repo.List(ctx, ListFilter{Page: 1, Limit: 2})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 5 {
		t.Fatalf("expected total 5, got %d", total)
	}
	if len(agents) != 2 {
		t.Fatalf("expected 2 agents, got %d", len(agents))
	}
}

func TestListFilterByStatus(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	active := newTestAgent("Active", model.EngineCoze)
	active.Status = model.AgentStatusActive
	repo.Create(ctx, active)

	inactive := newTestAgent("Inactive", model.EngineCoze)
	repo.Create(ctx, inactive)

	agents, total, _ := repo.List(ctx, ListFilter{Status: model.AgentStatusActive, Page: 1, Limit: 20})
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if agents[0].Name != "Active" {
		t.Fatalf("expected Active, got %s", agents[0].Name)
	}
}

func TestListFilterByCategory(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	a1 := newTestAgent("A1", model.EngineCoze)
	a1.Category = "writing"
	repo.Create(ctx, a1)

	a2 := newTestAgent("A2", model.EngineCoze)
	a2.Category = "coding"
	repo.Create(ctx, a2)

	agents, total, _ := repo.List(ctx, ListFilter{Category: "coding", Page: 1, Limit: 20})
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if agents[0].Name != "A2" {
		t.Fatalf("expected A2, got %s", agents[0].Name)
	}
}

func TestListFilterByTag(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	a1 := newTestAgent("A1", model.EngineCoze)
	a1.Tags = []string{"ai", "writing"}
	repo.Create(ctx, a1)

	a2 := newTestAgent("A2", model.EngineCoze)
	a2.Tags = []string{"ai", "coding"}
	repo.Create(ctx, a2)

	agents, total, _ := repo.List(ctx, ListFilter{Tag: "coding", Page: 1, Limit: 20})
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if agents[0].Name != "A2" {
		t.Fatalf("expected A2, got %s", agents[0].Name)
	}
}

func TestCozeWorkflowCRUD(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("CozeBot", model.EngineCoze)
	repo.Create(ctx, agent)

	wf := &model.CozeWorkflow{
		AgentID:     agent.ID,
		WorkflowID:  "wf_123",
		APIKey:      "key_secret",
		Region:      "cn",
		InputField:  "input",
		OutputField: "output",
	}
	if err := repo.CreateCozeWorkflow(ctx, wf); err != nil {
		t.Fatalf("create coze workflow: %v", err)
	}

	found, err := repo.GetCozeWorkflow(ctx, agent.ID)
	if err != nil {
		t.Fatalf("get coze workflow: %v", err)
	}
	if found.WorkflowID != "wf_123" {
		t.Fatalf("expected wf_123, got %s", found.WorkflowID)
	}
	if found.APIKey != "key_secret" {
		t.Fatalf("expected key_secret, got %s", found.APIKey)
	}

	found.WorkflowID = "wf_456"
	if err := repo.UpdateCozeWorkflow(ctx, found); err != nil {
		t.Fatalf("update coze workflow: %v", err)
	}
	updated, _ := repo.GetCozeWorkflow(ctx, agent.ID)
	if updated.WorkflowID != "wf_456" {
		t.Fatalf("expected wf_456, got %s", updated.WorkflowID)
	}

	if err := repo.DeleteCozeWorkflow(ctx, agent.ID); err != nil {
		t.Fatalf("delete coze workflow: %v", err)
	}
	deleted, _ := repo.GetCozeWorkflow(ctx, agent.ID)
	if deleted != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestN8NWorkflowCRUD(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("N8NBot", model.EngineN8N)
	repo.Create(ctx, agent)

	wf := &model.N8NWorkflow{
		AgentID:     agent.ID,
		WebhookURL:  "/webhook/abc",
		AuthType:    "bearer",
		AuthToken:   "token_secret",
		Timeout:     60,
		PayloadTmpl: `{"input": "{{.input}}"}`,
	}
	if err := repo.CreateN8NWorkflow(ctx, wf); err != nil {
		t.Fatalf("create n8n workflow: %v", err)
	}

	found, err := repo.GetN8NWorkflow(ctx, agent.ID)
	if err != nil {
		t.Fatalf("get n8n workflow: %v", err)
	}
	if found.WebhookURL != "/webhook/abc" {
		t.Fatalf("expected /webhook/abc, got %s", found.WebhookURL)
	}
	if found.Timeout != 60 {
		t.Fatalf("expected timeout 60, got %d", found.Timeout)
	}

	found.WebhookURL = "/webhook/xyz"
	if err := repo.UpdateN8NWorkflow(ctx, found); err != nil {
		t.Fatalf("update n8n workflow: %v", err)
	}
	updated, _ := repo.GetN8NWorkflow(ctx, agent.ID)
	if updated.WebhookURL != "/webhook/xyz" {
		t.Fatalf("expected /webhook/xyz, got %s", updated.WebhookURL)
	}

	if err := repo.DeleteN8NWorkflow(ctx, agent.ID); err != nil {
		t.Fatalf("delete n8n workflow: %v", err)
	}
	deleted, _ := repo.GetN8NWorkflow(ctx, agent.ID)
	if deleted != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestCascadeDelete(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("CascadeBot", model.EngineCoze)
	repo.Create(ctx, agent)
	repo.CreateCozeWorkflow(ctx, &model.CozeWorkflow{
		AgentID:    agent.ID,
		WorkflowID: "cascade_wf",
	})

	repo.Delete(ctx, agent.ID)

	wf, _ := repo.GetCozeWorkflow(ctx, agent.ID)
	if wf != nil {
		t.Fatal("expected workflow to be cascade deleted")
	}
}

func TestTagsJSONRoundTrip(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("TagBot", model.EngineCoze)
	agent.Tags = []string{"alpha", "beta", "gamma"}
	repo.Create(ctx, agent)

	found, _ := repo.GetByID(ctx, agent.ID)
	if len(found.Tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(found.Tags))
	}
	if found.Tags[0] != "alpha" || found.Tags[1] != "beta" || found.Tags[2] != "gamma" {
		t.Fatalf("unexpected tags: %v", found.Tags)
	}
}

func TestFeaturedBoolRoundTrip(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("FeatBot", model.EngineCoze)
	agent.Featured = true
	repo.Create(ctx, agent)

	found, _ := repo.GetByID(ctx, agent.ID)
	if !found.Featured {
		t.Fatal("expected featured to be true")
	}

	agent.Featured = false
	repo.Update(ctx, agent)
	found, _ = repo.GetByID(ctx, agent.ID)
	if found.Featured {
		t.Fatal("expected featured to be false")
	}
}
