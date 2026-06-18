package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

func New(dsn string) (*sql.DB, error) {
	dir := filepath.Dir(dsn)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create data dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return nil, fmt.Errorf("exec pragma %q: %w", p, err)
		}
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createWalletsTable,
		createAgentsTable,
		createCozeWorkflowsTable,
		createN8NWorkflowsTable,
		createAgentToolsTable,
		createAgentToolsIndex,
		createConversationsTable,
		createMessagesTable,
		createTransactionsTable,
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	migrateAddColumn(db, "users", "role", "TEXT NOT NULL DEFAULT 'user'")
	migrateAddColumn(db, "agents", "model_name", "TEXT NOT NULL DEFAULT ''")
	migrateAddColumn(db, "agents", "base_url", "TEXT NOT NULL DEFAULT ''")
	migrateAddColumn(db, "agents", "api_key", "TEXT NOT NULL DEFAULT ''")

	migrateWorkflowsToTools(db)

	return nil
}

func migrateAddColumn(db *sql.DB, table, column, colDef string) {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, colDef)
	_, err := db.Exec(query)
	if err != nil && strings.Contains(err.Error(), "duplicate column") {
		return
	}
}

func migrateWorkflowsToTools(db *sql.DB) {
	rows, err := db.Query(`SELECT agent_id, workflow_id, api_key, region, input_field, output_field FROM coze_workflows`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var agentID int64
		var workflowID, apiKey, region, inputField, outputField string
		rows.Scan(&agentID, &workflowID, &apiKey, &region, &inputField, &outputField)
		config, _ := json.Marshal(map[string]string{
			"workflow_id":  workflowID,
			"api_key":      apiKey,
			"region":       region,
			"input_field":  inputField,
			"output_field": outputField,
		})
		db.Exec(`INSERT OR IGNORE INTO agent_tools (agent_id, name, description, type, config) VALUES (?, ?, ?, ?, ?)`,
			agentID, "coze_workflow", "Coze workflow", "coze", string(config))
	}

	rows2, err := db.Query(`SELECT agent_id, webhook_url, auth_type, auth_token, timeout, payload_tmpl FROM n8n_workflows`)
	if err != nil {
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var agentID int64
		var webhookURL, authType, authToken, payloadTmpl string
		var timeout int
		rows2.Scan(&agentID, &webhookURL, &authType, &authToken, &timeout, &payloadTmpl)
		config, _ := json.Marshal(map[string]interface{}{
			"webhook_url":  webhookURL,
			"auth_type":    authType,
			"auth_token":   authToken,
			"timeout":      timeout,
			"payload_tmpl": payloadTmpl,
		})
		db.Exec(`INSERT OR IGNORE INTO agent_tools (agent_id, name, description, type, config) VALUES (?, ?, ?, ?, ?)`,
			agentID, "n8n_workflow", "N8N workflow", "n8n", string(config))
	}
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    email       TEXT    NOT NULL UNIQUE,
    name        TEXT    NOT NULL,
    password    TEXT    NOT NULL,
    avatar      TEXT    NOT NULL DEFAULT '',
    role        TEXT    NOT NULL DEFAULT 'user',
    status      TEXT    NOT NULL DEFAULT 'active',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
`

const createWalletsTable = `
CREATE TABLE IF NOT EXISTS wallets (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL UNIQUE,
    balance    INTEGER NOT NULL DEFAULT 0,
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
`

const createAgentsTable = `
CREATE TABLE IF NOT EXISTS agents (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT    NOT NULL,
    icon        TEXT    NOT NULL DEFAULT '',
    color       TEXT    NOT NULL DEFAULT '',
    category    TEXT    NOT NULL DEFAULT '',
    short_desc  TEXT    NOT NULL DEFAULT '',
    full_desc   TEXT    NOT NULL DEFAULT '',
    tags        TEXT    NOT NULL DEFAULT '[]',
    cost        INTEGER NOT NULL DEFAULT 0,
    engine      TEXT    NOT NULL DEFAULT '',
    status      TEXT    NOT NULL DEFAULT 'inactive',
    prompt      TEXT    NOT NULL DEFAULT '',
    temperature REAL    NOT NULL DEFAULT 0.7,
    max_tokens  INTEGER NOT NULL DEFAULT 2048,
    model_name  TEXT    NOT NULL DEFAULT '',
    base_url    TEXT    NOT NULL DEFAULT '',
    api_key     TEXT    NOT NULL DEFAULT '',
    rating      REAL    NOT NULL DEFAULT 0.0,
    calls       INTEGER NOT NULL DEFAULT 0,
    featured    INTEGER NOT NULL DEFAULT 0,
    speed       TEXT    NOT NULL DEFAULT '',
    precision   TEXT    NOT NULL DEFAULT '',
    created_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at  DATETIME NOT NULL DEFAULT (datetime('now'))
);
`

const createCozeWorkflowsTable = `
CREATE TABLE IF NOT EXISTS coze_workflows (
    agent_id     INTEGER PRIMARY KEY,
    workflow_id  TEXT    NOT NULL DEFAULT '',
    api_key      TEXT    NOT NULL DEFAULT '',
    region       TEXT    NOT NULL DEFAULT '',
    input_field  TEXT    NOT NULL DEFAULT '',
    output_field TEXT    NOT NULL DEFAULT '',
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
`

const createN8NWorkflowsTable = `
CREATE TABLE IF NOT EXISTS n8n_workflows (
    agent_id      INTEGER PRIMARY KEY,
    webhook_url   TEXT    NOT NULL DEFAULT '',
    auth_type     TEXT    NOT NULL DEFAULT '',
    auth_token    TEXT    NOT NULL DEFAULT '',
    timeout       INTEGER NOT NULL DEFAULT 30,
    payload_tmpl  TEXT    NOT NULL DEFAULT '',
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
`

const createAgentToolsTable = `
CREATE TABLE IF NOT EXISTS agent_tools (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id     INTEGER NOT NULL,
    name         TEXT    NOT NULL,
    description  TEXT    NOT NULL DEFAULT '',
    type         TEXT    NOT NULL,
    input_schema TEXT    NOT NULL DEFAULT '{}',
    config       TEXT    NOT NULL DEFAULT '{}',
    created_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (agent_id) REFERENCES agents(id) ON DELETE CASCADE
);
`

const createAgentToolsIndex = `
CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_tools_agent_name ON agent_tools(agent_id, name);
`

const createConversationsTable = `
CREATE TABLE IF NOT EXISTS conversations (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    agent_id   INTEGER NOT NULL,
    title      TEXT    NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (agent_id) REFERENCES agents(id)
);
`

const createMessagesTable = `
CREATE TABLE IF NOT EXISTS messages (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    role            TEXT    NOT NULL,
    content         TEXT    NOT NULL DEFAULT '',
    tool_calls      TEXT    NOT NULL DEFAULT '',
    tool_call_id    TEXT    NOT NULL DEFAULT '',
    created_at      DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);
`

const createTransactionsTable = `
CREATE TABLE IF NOT EXISTS transactions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id    INTEGER NOT NULL,
    type       TEXT    NOT NULL,
    agent_id   INTEGER,
    agent_name TEXT    NOT NULL DEFAULT '',
    amount     INTEGER NOT NULL DEFAULT 0,
    status     TEXT    NOT NULL DEFAULT 'success',
    note       TEXT    NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT (datetime('now')),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
`
