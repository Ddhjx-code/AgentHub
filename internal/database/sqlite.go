package database

import (
	"database/sql"
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
	}
	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}

	migrateAddRoleColumn(db)

	return nil
}

func migrateAddRoleColumn(db *sql.DB) {
	_, err := db.Exec(`ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user'`)
	if err != nil && strings.Contains(err.Error(), "duplicate column") {
		return
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
    engine      TEXT    NOT NULL DEFAULT 'coze',
    status      TEXT    NOT NULL DEFAULT 'inactive',
    prompt      TEXT    NOT NULL DEFAULT '',
    temperature REAL    NOT NULL DEFAULT 0.7,
    max_tokens  INTEGER NOT NULL DEFAULT 2048,
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
