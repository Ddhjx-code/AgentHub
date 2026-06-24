# AgentHub

AI Agent Service Platform — AI 智能体服务平台

一站式 AI 智能体管理与调用平台，支持多 LLM 后端接入、工作流引擎集成（Coze/n8n）、RAG 知识库增强、按量计费，提供完整的前后端实现。

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.21+ / Gin / SQLite3 |
| Frontend | React 18 / TypeScript / Vite / TailwindCSS 4 |
| LLM | OpenAI-compatible API (DeepSeek, GPT, etc.) |
| Tools | Coze Workflow / n8n Webhook |
| RAG | Embedding + SQLite vector store |
| Auth | JWT + bcrypt |

## Features

- **User System** — Registration, login, JWT authentication, role-based access (user/admin)
- **Agent Marketplace** — Browse, search, filter agents by category/tag, agent detail pages
- **Agent Management** — Admin CRUD with prompt, LLM config, tool binding, knowledge base binding
- **Chat** — Multi-turn conversation with agents, message history, conversation management
- **Tool Calling** — LLM function calling with Coze workflow and n8n webhook executors
- **Tool Config UI** — Admin panel for adding/editing tool configurations per agent
- **Knowledge Base (RAG)** — Document upload, chunking, embedding, similarity search with hybrid scoring
- **Wallet & Billing** — Credit-based (灵石) pay-per-use, transaction history
- **i18n** — Chinese/English toggle, persisted to localStorage
- **Chat Tool Display** — Collapsible tool call details in chat messages

## Project Structure

```
AgentHub/
├── cmd/server/main.go           # Entrypoint
├── configs/config.yaml          # Server configuration
├── internal/
│   ├── config/                  # Config loading (YAML)
│   ├── database/                # SQLite init & migrations
│   ├── model/                   # Domain models
│   ├── handler/                 # HTTP handlers
│   │   ├── user/                #   Auth: register, login, profile
│   │   ├── agent/               #   Public: list, detail
│   │   ├── chat/                #   Chat: send message
│   │   ├── admin/               #   Admin: agent CRUD, toggle status
│   │   ├── knowledge/           #   Admin: KB & document management
│   │   ├── wallet/              #   Balance query
│   │   └── transaction/         #   Transaction records
│   ├── service/                 # Business logic
│   │   ├── user/                #   Auth + profile
│   │   ├── agent/               #   Agent CRUD + admin detail
│   │   ├── chat/                #   LLM loop + tool calling + RAG
│   │   ├── knowledge/           #   KB + document + embedding + search
│   │   ├── wallet/              #   Balance operations
│   │   ├── transaction/         #   Transaction logging
│   │   └── admin/               #   Dashboard stats
│   ├── repository/              # Data access (SQLite)
│   ├── middleware/              # JWT auth, CORS
│   ├── router/                  # Route registration
│   ├── llm/                     # OpenAI-compatible LLM client
│   ├── tool/                    # Tool executors (Coze, n8n)
│   ├── embedding/               # Embedding client + vector ops
│   ├── vectorstore/             # SQLite-based vector store
│   └── chunker/                 # Document chunking
├── pkg/
│   ├── response/                # Unified API response
│   ├── errcode/                 # Error codes
│   └── logger/                  # Structured logging (slog)
└── web/                         # React frontend
    └── src/
        ├── api/                 # API client (Axios)
        ├── components/          # Reusable components
        ├── contexts/            # Auth + Locale contexts
        ├── i18n/                # zh/en translation dictionaries
        ├── layouts/             # User + Admin layouts
        ├── pages/               # Route pages
        │   └── admin/           # Admin pages
        └── types/               # TypeScript interfaces
```

## API Endpoints

### Public
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/login` | User login |
| GET | `/api/v1/agents` | List active agents |
| GET | `/api/v1/agents/:id` | Agent detail |

### Protected (JWT required)
| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/v1/user/profile` | User profile |
| POST | `/api/v1/agents/:id/chat` | Send message to agent |
| GET | `/api/v1/conversations` | List conversations |
| GET | `/api/v1/conversations/:id/messages` | Get messages |
| DELETE | `/api/v1/conversations/:id` | Delete conversation |

### Admin (JWT + admin role)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/admin/agents` | Create agent |
| GET | `/api/v1/admin/agents` | List all agents |
| GET | `/api/v1/admin/agents/:id` | Agent detail (with tool config) |
| PUT | `/api/v1/admin/agents/:id` | Update agent |
| DELETE | `/api/v1/admin/agents/:id` | Delete agent |
| PUT | `/api/v1/admin/agents/:id/toggle` | Toggle agent status |
| POST | `/api/v1/admin/agents/:id/knowledge-bases` | Bind KB |
| GET | `/api/v1/admin/agents/:id/knowledge-bases` | List bound KBs |
| DELETE | `/api/v1/admin/agents/:id/knowledge-bases/:kb_id` | Unbind KB |
| POST | `/api/v1/admin/knowledge-bases` | Create KB |
| GET | `/api/v1/admin/knowledge-bases` | List KBs |
| GET | `/api/v1/admin/knowledge-bases/:id` | KB detail |
| PUT | `/api/v1/admin/knowledge-bases/:id` | Update KB |
| DELETE | `/api/v1/admin/knowledge-bases/:id` | Delete KB |
| POST | `/api/v1/admin/knowledge-bases/:id/documents` | Upload document |
| GET | `/api/v1/admin/knowledge-bases/:id/documents` | List documents |
| DELETE | `/api/v1/admin/knowledge-bases/:id/documents/:doc_id` | Delete document |

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+

### Backend
```bash
# Clone
git clone https://github.com/Ddhjx-code/AgentHub.git
cd AgentHub

# Build & run
make build
make run
# Server starts at http://localhost:8080
```

### Frontend
```bash
cd web
npm install
npm run dev
# Dev server at http://localhost:5173 (proxied to :8080)
```

### Configuration

Edit `configs/config.yaml`:

```yaml
server:
  port: 8080
  mode: debug          # debug / release

database:
  dsn: "data/agenthub.db"

jwt:
  secret: "change-me-in-production"
  expire_hour: 24

coze:
  base_url: "https://api.coze.cn"

n8n:
  default_timeout: 30
```

## Architecture

```
┌──────────┐    ┌──────────┐    ┌──────────┐
│ Frontend │───▶│   Gin    │───▶│  SQLite  │
│  React   │    │ Handlers │    │    DB    │
└──────────┘    └────┬─────┘    └──────────┘
                     │
              ┌──────┴──────┐
              │   Services  │
              ├─────────────┤
              │  Chat Loop  │──▶ LLM API (OpenAI-compatible)
              │  (ReAct)    │
              ├─────────────┤
              │ Tool Exec   │──▶ Coze Workflow / n8n Webhook
              ├─────────────┤
              │ RAG Search  │──▶ Embedding API + Vector Store
              └─────────────┘
```

## License

MIT
