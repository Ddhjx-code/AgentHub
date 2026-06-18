package agent

import (
	"context"
	"encoding/json"
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

func newTestAgent(name string) *model.Agent {
	return &model.Agent{
		Name:        name,
		Status:      model.AgentStatusInactive,
		Tags:        []string{"test", "demo"},
		Cost:        10,
		Temperature: 0.7,
		MaxTokens:   2048,
		ModelName:   "deepseek-chat",
		BaseURL:     "https://api.deepseek.com/v1",
		APIKey:      "sk-test",
	}
}

func TestCreateAndGetByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("TestBot")
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
	if found.ModelName != "deepseek-chat" {
		t.Fatalf("expected model deepseek-chat, got %s", found.ModelName)
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

	agent := newTestAgent("Original")
	repo.Create(ctx, agent)

	agent.Name = "Updated"
	agent.Cost = 50
	agent.ModelName = "gpt-4o"
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
	if found.ModelName != "gpt-4o" {
		t.Fatalf("expected model gpt-4o, got %s", found.ModelName)
	}
}

func TestDeleteAgent(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("ToDelete")
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
		repo.Create(ctx, newTestAgent("Bot"))
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

	active := newTestAgent("Active")
	active.Status = model.AgentStatusActive
	repo.Create(ctx, active)

	repo.Create(ctx, newTestAgent("Inactive"))

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

	a1 := newTestAgent("A1")
	a1.Category = "writing"
	repo.Create(ctx, a1)

	a2 := newTestAgent("A2")
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

	a1 := newTestAgent("A1")
	a1.Tags = []string{"ai", "writing"}
	repo.Create(ctx, a1)

	a2 := newTestAgent("A2")
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

func TestToolCRUD(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("ToolBot")
	repo.Create(ctx, agent)

	tool := &model.AgentTool{
		AgentID:     agent.ID,
		Name:        "search",
		Description: "Search the web",
		Type:        model.ToolTypeCoze,
		InputSchema: json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}}}`),
		Config:      json.RawMessage(`{"workflow_id":"wf_123","api_key":"key_abc"}`),
	}
	if err := repo.CreateTool(ctx, tool); err != nil {
		t.Fatalf("create tool: %v", err)
	}
	if tool.ID == 0 {
		t.Fatal("expected non-zero tool ID")
	}

	found, err := repo.GetToolByID(ctx, tool.ID)
	if err != nil {
		t.Fatalf("get tool: %v", err)
	}
	if found.Name != "search" {
		t.Fatalf("expected search, got %s", found.Name)
	}

	tools, err := repo.ListToolsByAgentID(ctx, agent.ID)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}

	tool.Description = "Updated description"
	if err := repo.UpdateTool(ctx, tool); err != nil {
		t.Fatalf("update tool: %v", err)
	}
	updated, _ := repo.GetToolByID(ctx, tool.ID)
	if updated.Description != "Updated description" {
		t.Fatalf("expected updated description, got %s", updated.Description)
	}

	if err := repo.DeleteTool(ctx, tool.ID); err != nil {
		t.Fatalf("delete tool: %v", err)
	}
	deleted, _ := repo.GetToolByID(ctx, tool.ID)
	if deleted != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestDeleteToolsByAgentID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("MultiToolBot")
	repo.Create(ctx, agent)

	for _, name := range []string{"tool1", "tool2", "tool3"} {
		repo.CreateTool(ctx, &model.AgentTool{
			AgentID:     agent.ID,
			Name:        name,
			Type:        model.ToolTypeN8N,
			InputSchema: json.RawMessage("{}"),
			Config:      json.RawMessage("{}"),
		})
	}

	tools, _ := repo.ListToolsByAgentID(ctx, agent.ID)
	if len(tools) != 3 {
		t.Fatalf("expected 3 tools, got %d", len(tools))
	}

	if err := repo.DeleteToolsByAgentID(ctx, agent.ID); err != nil {
		t.Fatalf("delete tools by agent: %v", err)
	}
	tools, _ = repo.ListToolsByAgentID(ctx, agent.ID)
	if len(tools) != 0 {
		t.Fatalf("expected 0 tools, got %d", len(tools))
	}
}

func TestCascadeDeleteTools(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("CascadeBot")
	repo.Create(ctx, agent)
	repo.CreateTool(ctx, &model.AgentTool{
		AgentID:     agent.ID,
		Name:        "cascade_tool",
		Type:        model.ToolTypeCoze,
		InputSchema: json.RawMessage("{}"),
		Config:      json.RawMessage("{}"),
	})

	repo.Delete(ctx, agent.ID)

	tools, _ := repo.ListToolsByAgentID(ctx, agent.ID)
	if len(tools) != 0 {
		t.Fatal("expected tools to be cascade deleted")
	}
}

func TestIncrementCalls(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("CallBot")
	repo.Create(ctx, agent)

	repo.IncrementCalls(ctx, agent.ID)
	repo.IncrementCalls(ctx, agent.ID)

	found, _ := repo.GetByID(ctx, agent.ID)
	if found.Calls != 2 {
		t.Fatalf("expected 2 calls, got %d", found.Calls)
	}
}

func TestTagsJSONRoundTrip(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	agent := newTestAgent("TagBot")
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

	agent := newTestAgent("FeatBot")
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
