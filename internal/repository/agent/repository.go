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

	CreateCozeWorkflow(ctx context.Context, wf *model.CozeWorkflow) error
	GetCozeWorkflow(ctx context.Context, agentID int64) (*model.CozeWorkflow, error)
	UpdateCozeWorkflow(ctx context.Context, wf *model.CozeWorkflow) error
	DeleteCozeWorkflow(ctx context.Context, agentID int64) error

	CreateN8NWorkflow(ctx context.Context, wf *model.N8NWorkflow) error
	GetN8NWorkflow(ctx context.Context, agentID int64) (*model.N8NWorkflow, error)
	UpdateN8NWorkflow(ctx context.Context, wf *model.N8NWorkflow) error
	DeleteN8NWorkflow(ctx context.Context, agentID int64) error
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
              cost, engine, status, prompt, temperature, max_tokens, rating, calls, featured,
              speed, precision, created_at, updated_at)
              VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		agent.Name, agent.Icon, agent.Color, agent.Category, agent.ShortDesc, agent.FullDesc,
		tagsToJSON(agent.Tags), agent.Cost, agent.Engine, agent.Status, agent.Prompt,
		agent.Temperature, agent.MaxTokens, agent.Rating, agent.Calls, boolToInt(agent.Featured),
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
              cost, engine, status, prompt, temperature, max_tokens, rating, calls, featured,
              speed, precision, created_at, updated_at
              FROM agents WHERE id = ?`
	return r.scanAgent(r.db.QueryRowContext(ctx, query, id))
}

func (r *repository) Update(ctx context.Context, agent *model.Agent) error {
	query := `UPDATE agents SET name=?, icon=?, color=?, category=?, short_desc=?, full_desc=?,
              tags=?, cost=?, engine=?, status=?, prompt=?, temperature=?, max_tokens=?,
              rating=?, calls=?, featured=?, speed=?, precision=?, updated_at=?
              WHERE id=?`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		agent.Name, agent.Icon, agent.Color, agent.Category, agent.ShortDesc, agent.FullDesc,
		tagsToJSON(agent.Tags), agent.Cost, agent.Engine, agent.Status, agent.Prompt,
		agent.Temperature, agent.MaxTokens, agent.Rating, agent.Calls, boolToInt(agent.Featured),
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
                    cost, engine, status, prompt, temperature, max_tokens, rating, calls, featured,
                    speed, precision, created_at, updated_at
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

// Coze workflow

func (r *repository) CreateCozeWorkflow(ctx context.Context, wf *model.CozeWorkflow) error {
	query := `INSERT INTO coze_workflows (agent_id, workflow_id, api_key, region, input_field, output_field)
              VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		wf.AgentID, wf.WorkflowID, wf.APIKey, wf.Region, wf.InputField, wf.OutputField)
	if err != nil {
		return fmt.Errorf("insert coze workflow: %w", err)
	}
	return nil
}

func (r *repository) GetCozeWorkflow(ctx context.Context, agentID int64) (*model.CozeWorkflow, error) {
	query := `SELECT agent_id, workflow_id, api_key, region, input_field, output_field
              FROM coze_workflows WHERE agent_id = ?`
	wf := &model.CozeWorkflow{}
	err := r.db.QueryRowContext(ctx, query, agentID).Scan(
		&wf.AgentID, &wf.WorkflowID, &wf.APIKey, &wf.Region, &wf.InputField, &wf.OutputField)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query coze workflow: %w", err)
	}
	return wf, nil
}

func (r *repository) UpdateCozeWorkflow(ctx context.Context, wf *model.CozeWorkflow) error {
	query := `UPDATE coze_workflows SET workflow_id=?, api_key=?, region=?, input_field=?, output_field=?
              WHERE agent_id=?`
	_, err := r.db.ExecContext(ctx, query,
		wf.WorkflowID, wf.APIKey, wf.Region, wf.InputField, wf.OutputField, wf.AgentID)
	if err != nil {
		return fmt.Errorf("update coze workflow: %w", err)
	}
	return nil
}

func (r *repository) DeleteCozeWorkflow(ctx context.Context, agentID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM coze_workflows WHERE agent_id = ?`, agentID)
	if err != nil {
		return fmt.Errorf("delete coze workflow: %w", err)
	}
	return nil
}

// N8N workflow

func (r *repository) CreateN8NWorkflow(ctx context.Context, wf *model.N8NWorkflow) error {
	query := `INSERT INTO n8n_workflows (agent_id, webhook_url, auth_type, auth_token, timeout, payload_tmpl)
              VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		wf.AgentID, wf.WebhookURL, wf.AuthType, wf.AuthToken, wf.Timeout, wf.PayloadTmpl)
	if err != nil {
		return fmt.Errorf("insert n8n workflow: %w", err)
	}
	return nil
}

func (r *repository) GetN8NWorkflow(ctx context.Context, agentID int64) (*model.N8NWorkflow, error) {
	query := `SELECT agent_id, webhook_url, auth_type, auth_token, timeout, payload_tmpl
              FROM n8n_workflows WHERE agent_id = ?`
	wf := &model.N8NWorkflow{}
	err := r.db.QueryRowContext(ctx, query, agentID).Scan(
		&wf.AgentID, &wf.WebhookURL, &wf.AuthType, &wf.AuthToken, &wf.Timeout, &wf.PayloadTmpl)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query n8n workflow: %w", err)
	}
	return wf, nil
}

func (r *repository) UpdateN8NWorkflow(ctx context.Context, wf *model.N8NWorkflow) error {
	query := `UPDATE n8n_workflows SET webhook_url=?, auth_type=?, auth_token=?, timeout=?, payload_tmpl=?
              WHERE agent_id=?`
	_, err := r.db.ExecContext(ctx, query,
		wf.WebhookURL, wf.AuthType, wf.AuthToken, wf.Timeout, wf.PayloadTmpl, wf.AgentID)
	if err != nil {
		return fmt.Errorf("update n8n workflow: %w", err)
	}
	return nil
}

func (r *repository) DeleteN8NWorkflow(ctx context.Context, agentID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM n8n_workflows WHERE agent_id = ?`, agentID)
	if err != nil {
		return fmt.Errorf("delete n8n workflow: %w", err)
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
		&agent.Cost, &agent.Engine, &agent.Status, &agent.Prompt,
		&agent.Temperature, &agent.MaxTokens, &agent.Rating, &agent.Calls, &featuredInt,
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
		&agent.Cost, &agent.Engine, &agent.Status, &agent.Prompt,
		&agent.Temperature, &agent.MaxTokens, &agent.Rating, &agent.Calls, &featuredInt,
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
