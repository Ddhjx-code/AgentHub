# AgentHub

AI Agent Service Platform — AI 智能体服务平台

## Project Structure

```
AgentHub/
├── cmd/server/              # Application entrypoint
├── internal/
│   ├── config/              # Configuration loading
│   ├── model/               # Domain models
│   ├── handler/             # HTTP handlers (controller layer)
│   │   ├── user/            # User registration, login, profile
│   │   ├── agent/           # Agent CRUD, browsing, invocation
│   │   ├── wallet/          # Balance query, recharge
│   │   ├── transaction/     # Transaction records
│   │   └── admin/           # Admin dashboard, management
│   ├── service/             # Business logic layer
│   │   ├── user/
│   │   ├── agent/
│   │   ├── wallet/
│   │   ├── transaction/
│   │   └── admin/
│   ├── repository/          # Data access layer
│   │   ├── user/
│   │   ├── agent/
│   │   ├── wallet/
│   │   └── transaction/
│   ├── middleware/           # Auth, CORS, rate limiting
│   ├── router/              # Route registration
│   └── workflow/            # Workflow engine adapters
│       ├── coze/            # Coze workflow API client
│       └── n8n/             # n8n webhook client
├── pkg/
│   ├── response/            # Unified API response
│   ├── errcode/             # Error codes
│   └── logger/              # Structured logging
├── configs/                 # Configuration files
├── scripts/                 # Build & deploy scripts
├── docs/                    # Documentation
├── demo/                    # UI prototype (excluded from git)
├── Makefile
└── go.mod
```

## Modules

| Module | Description |
|--------|-------------|
| user | User registration, login, JWT auth, profile management |
| agent | Agent CRUD, categories, configuration (prompts, parameters) |
| wallet | Credit (灵石) balance, recharge plans |
| transaction | Usage & recharge records, audit trail |
| workflow | Workflow engine adapters: Coze API, n8n Webhook |
| admin | Admin dashboard, user & agent management, stats |

## Quick Start

```bash
# Build
make build

# Run
make run

# Test
make test
```
