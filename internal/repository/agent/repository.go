package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ddhjx-code/AgentHub/internal/model"
)

type Repository interface {
	Create(ctx context.Context, agent *model.Agent) error
	GetByID(ctx context.Context, id int64) (*model.Agent, error)
	Update(ctx context.Context, agent *model.Agent) error
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context, filter ListFilter) ([]*model.Agent, int, error)
	IncrementCalls(ctx context.Context, id int64) error

	CreateTool(ctx context.Context, tool *model.AgentTool) error
	ListToolsByAgentID(ctx context.Context, agentID int64) ([]*model.AgentTool, error)
	GetToolByID(ctx context.Context, id int64) (*model.AgentTool, error)
	UpdateTool(ctx context.Context, tool *model.AgentTool) error
	DeleteTool(ctx context.Context, id int64) error
	DeleteToolsByAgentID(ctx context.Context, agentID int64) error
}

type ListFilter struct {
	Status   string
	Category string
	Tag      string
	Page     int
	Limit    int
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, agent *model.Agent) error {
	query := `INSERT INTO agents (name, icon, color, category, short_desc, full_desc, tags,
              cost, status, prompt, temperature, max_tokens, model_name, base_url, api_key,
              rating, calls, featured, speed, precision, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		agent.Name, agent.Icon, agent.Color, agent.Category, agent.ShortDesc, agent.FullDesc,
		tagsToJSON(agent.Tags), agent.Cost, agent.Status, agent.Prompt,
		agent.Temperature, agent.MaxTokens, agent.ModelName, agent.BaseURL, agent.APIKey,
		agent.Rating, agent.Calls, boolToInt(agent.Featured),
		agent.Speed, agent.Precision, now, now)
	if err != nil {
		return fmt.Errorf("insert agent: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	agent.ID = id
	agent.CreatedAt = now
	agent.UpdatedAt = now
	return nil
}

func (r *repository) GetByID(ctx context.Context, id int64) (*model.Agent, error) {
	query := `SELECT id, name, icon, color, category, short_desc, full_desc, tags,
              cost, status, prompt, temperature, max_tokens, model_name, base_url, api_key,
              rating, calls, featured, speed, precision, created_at, updated_at
              FROM agents WHERE id = ?`
	return r.scanAgent(r.db.QueryRowContext(ctx, query, id))
}

func (r *repository) Update(ctx context.Context, agent *model.Agent) error {
	query := `UPDATE agents SET name=?, icon=?, color=?, category=?, short_desc=?, full_desc=?,
              tags=?, cost=?, status=?, prompt=?, temperature=?, max_tokens=?,
              model_name=?, base_url=?, api_key=?,
              rating=?, calls=?, featured=?, speed=?, precision=?, updated_at=?
              WHERE id=?`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		agent.Name, agent.Icon, agent.Color, agent.Category, agent.ShortDesc, agent.FullDesc,
		tagsToJSON(agent.Tags), agent.Cost, agent.Status, agent.Prompt,
		agent.Temperature, agent.MaxTokens, agent.ModelName, agent.BaseURL, agent.APIKey,
		agent.Rating, agent.Calls, boolToInt(agent.Featured),
		agent.Speed, agent.Precision, now, agent.ID)
	if err != nil {
		return fmt.Errorf("update agent: %w", err)
	}
	agent.UpdatedAt = now
	return nil
}

func (r *repository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM agents WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete agent: %w", err)
	}
	return nil
}

func (r *repository) List(ctx context.Context, filter ListFilter) ([]*model.Agent, int, error) {
	where, args := buildWhere(filter)

	var total int
	countQuery := "SELECT COUNT(*) FROM agents" + where
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count agents: %w", err)
	}

	selectQuery := `SELECT id, name, icon, color, category, short_desc, full_desc, tags,
                    cost, status, prompt, temperature, max_tokens, model_name, base_url, api_key,
                    rating, calls, featured, speed, precision, created_at, updated_at
                    FROM agents` + where + ` ORDER BY created_at DESC LIMIT ? OFFSET ?`

	offset := (filter.Page - 1) * filter.Limit
	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list agents: %w", err)
	}
	defer rows.Close()

	var agents []*model.Agent
	for rows.Next() {
		agent, err := r.scanAgentRow(rows)
		if err != nil {
			return nil, 0, err
		}
		agents = append(agents, agent)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate agents: %w", err)
	}
	return agents, total, nil
}

func (r *repository) IncrementCalls(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE agents SET calls = calls + 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("increment calls: %w", err)
	}
	return nil
}

// Tool methods

func (r *repository) CreateTool(ctx context.Context, tool *model.AgentTool) error {
	query := `INSERT INTO agent_tools (agent_id, name, description, type, input_schema, config, created_at)
              VALUES (?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		tool.AgentID, tool.Name, tool.Description, tool.Type,
		string(tool.InputSchema), string(tool.Config), now)
	if err != nil {
		return fmt.Errorf("insert tool: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("last insert id: %w", err)
	}
	tool.ID = id
	tool.CreatedAt = now
	return nil
}

func (r *repository) ListToolsByAgentID(ctx context.Context, agentID int64) ([]*model.AgentTool, error) {
	query := `SELECT id, agent_id, name, description, type, input_schema, config, created_at
              FROM agent_tools WHERE agent_id = ? ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query, agentID)
	if err != nil {
		return nil, fmt.Errorf("list tools: %w", err)
	}
	defer rows.Close()

	var tools []*model.AgentTool
	for rows.Next() {
		tool := &model.AgentTool{}
		var inputSchema, config string
		err := rows.Scan(&tool.ID, &tool.AgentID, &tool.Name, &tool.Description,
			&tool.Type, &inputSchema, &config, &tool.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan tool: %w", err)
		}
		tool.InputSchema = json.RawMessage(inputSchema)
		tool.Config = json.RawMessage(config)
		tools = append(tools, tool)
	}
	return tools, rows.Err()
}

func (r *repository) GetToolByID(ctx context.Context, id int64) (*model.AgentTool, error) {
	query := `SELECT id, agent_id, name, description, type, input_schema, config, created_at
              FROM agent_tools WHERE id = ?`
	tool := &model.AgentTool{}
	var inputSchema, config string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tool.ID, &tool.AgentID, &tool.Name, &tool.Description,
		&tool.Type, &inputSchema, &config, &tool.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query tool: %w", err)
	}
	tool.InputSchema = json.RawMessage(inputSchema)
	tool.Config = json.RawMessage(config)
	return tool, nil
}

func (r *repository) UpdateTool(ctx context.Context, tool *model.AgentTool) error {
	query := `UPDATE agent_tools SET name=?, description=?, type=?, input_schema=?, config=? WHERE id=?`
	_, err := r.db.ExecContext(ctx, query,
		tool.Name, tool.Description, tool.Type,
		string(tool.InputSchema), string(tool.Config), tool.ID)
	if err != nil {
		return fmt.Errorf("update tool: %w", err)
	}
	return nil
}

func (r *repository) DeleteTool(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM agent_tools WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete tool: %w", err)
	}
	return nil
}

func (r *repository) DeleteToolsByAgentID(ctx context.Context, agentID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM agent_tools WHERE agent_id = ?`, agentID)
	if err != nil {
		return fmt.Errorf("delete tools by agent: %w", err)
	}
	return nil
}

// helpers

func buildWhere(f ListFilter) (string, []interface{}) {
	where := " WHERE 1=1"
	var args []interface{}
	if f.Status != "" {
		where += " AND status = ?"
		args = append(args, f.Status)
	}
	if f.Category != "" {
		where += " AND category = ?"
		args = append(args, f.Category)
	}
	if f.Tag != "" {
		where += ` AND tags LIKE ?`
		args = append(args, fmt.Sprintf(`%%"%s"%%`, f.Tag))
	}
	return where, args
}

func (r *repository) scanAgent(row *sql.Row) (*model.Agent, error) {
	agent := &model.Agent{}
	var tagsJSON string
	var featuredInt int
	err := row.Scan(
		&agent.ID, &agent.Name, &agent.Icon, &agent.Color, &agent.Category,
		&agent.ShortDesc, &agent.FullDesc, &tagsJSON,
		&agent.Cost, &agent.Status, &agent.Prompt,
		&agent.Temperature, &agent.MaxTokens, &agent.ModelName, &agent.BaseURL, &agent.APIKey,
		&agent.Rating, &agent.Calls, &featuredInt,
		&agent.Speed, &agent.Precision, &agent.CreatedAt, &agent.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scan agent: %w", err)
	}
	agent.Tags = tagsFromJSON(tagsJSON)
	agent.Featured = featuredInt != 0
	return agent, nil
}

func (r *repository) scanAgentRow(rows *sql.Rows) (*model.Agent, error) {
	agent := &model.Agent{}
	var tagsJSON string
	var featuredInt int
	err := rows.Scan(
		&agent.ID, &agent.Name, &agent.Icon, &agent.Color, &agent.Category,
		&agent.ShortDesc, &agent.FullDesc, &tagsJSON,
		&agent.Cost, &agent.Status, &agent.Prompt,
		&agent.Temperature, &agent.MaxTokens, &agent.ModelName, &agent.BaseURL, &agent.APIKey,
		&agent.Rating, &agent.Calls, &featuredInt,
		&agent.Speed, &agent.Precision, &agent.CreatedAt, &agent.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("scan agent row: %w", err)
	}
	agent.Tags = tagsFromJSON(tagsJSON)
	agent.Featured = featuredInt != 0
	return agent, nil
}

func tagsToJSON(tags []string) string {
	if tags == nil {
		tags = []string{}
	}
	b, _ := json.Marshal(tags)
	return string(b)
}

func tagsFromJSON(s string) []string {
	var tags []string
	json.Unmarshal([]byte(s), &tags)
	if tags == nil {
		return []string{}
	}
	return tags
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
