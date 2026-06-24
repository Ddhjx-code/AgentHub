# Changelog

## [Unreleased]

### Planned
- Tool calling confirmation mechanism (human-in-the-loop)
- ReAct reasoning loop improvements
- MCP Server integration
- Skill configuration system
- Rich media rendering (Mermaid, code highlight, prototype preview)
- Parallel tool execution
- Anti-hallucination measures

---

## 2024-06-24 — PR #8: Tool Configuration UI

**Branch:** `feat/tool-config-ui`

### Added
- **ToolConfigForm component** — New form component for editing tool configurations, dynamically switches between Coze and n8n config fields
- **AgentModal Tools tab** — 5th tab in agent editor for adding/editing/deleting tools with JSON schema validation
- **ChatMessage tool call display** — Parses `tool_calls` JSON and renders collapsible tool call details (name + arguments)
- **AdminToolView** — Backend returns `admin_tools` with config on admin detail endpoint (config is hidden from public API via `json:"-"`)
- **i18n keys** — 25 new translation keys per language for tool config and chat tool display

### Changed
- Chat.tsx reloads full message chain after send (replaces synthetic temp message) to show tool interactions

### Fixed
- API key field no longer pre-fills with masked value on edit (prevents accidental key corruption)

---

## 2024-06-23 — PR #7: Frontend UI + i18n + Bug Fixes

**Branch:** `feat/frontend-ui`

### Added
- **React frontend** — Complete SPA with React 18 + TypeScript + Vite + TailwindCSS 4
- **Pages** — Landing, Agent Market, Agent Detail, Chat, Dashboard, Admin (Overview, Agent List, KB List, KB Detail)
- **Components** — AgentCard, AgentModal, AuthModal, ChatMessage, ConversationList, ProtectedRoute, Toast
- **Layouts** — UserLayout (public nav), AdminLayout (sidebar)
- **i18n system** — Custom LocaleContext + zh/en translation dictionaries, CN/EN toggle in nav, persisted to localStorage
- **API client** — Axios-based with JWT interceptor, proxy to backend

### Fixed
- Agents API made public (moved GET routes out of protected group)
- API key no longer wiped on agent update (`json:"-"` round-trip fix)
- Field name mismatches between frontend and backend (`short_desc`/`calls`)
- Default tool input schema changed to `{"type":"object","properties":{}}` (LLM requires valid schema)

---

## 2024-06-19 — PR #6: RAG Optimization

**Branch:** `feat/rag-optimization`

### Added
- Similarity threshold filtering (configurable, default 0.3)
- Prompt constraints for RAG context usage
- Hybrid search scoring (keyword + vector)

### Changed
- Improved search result ranking with combined scoring

---

## 2024-06-18 — PR #5: Knowledge Base Module

**Branch:** `feat/knowledge-base`

### Added
- Knowledge base CRUD (admin endpoints)
- Document upload and management
- Text chunking with configurable chunk size and overlap
- Embedding client (OpenAI-compatible API)
- SQLite-based vector store with cosine similarity search
- RAG integration in chat: auto-search knowledge base and inject context into system prompt
- Dynamic `knowledge_search` tool for LLM to query KB during conversation
- Agent-KB binding/unbinding (many-to-many)

---

## 2024-06-17 — PR #4: Chat Module

**Branch:** `feat/chat`

### Added
- Chat service with multi-turn conversation support
- LLM client (OpenAI-compatible API with function calling)
- Tool calling loop (max 5 iterations)
- Coze workflow executor
- n8n webhook executor
- Conversation and message persistence
- Credit deduction on successful chat
- Agent call count tracking

### Changed
- Agent model refactored: added `short_desc`, `full_desc`, `speed`, `precision`, `featured` fields
- Tool definitions moved from separate tables to `agent_tools` (with migration)

---

## 2024-06-16 — PR #3: Agent Management

**Branch:** `feat/agent-management`

### Added
- Agent CRUD with admin-only access
- Agent status toggle (active/inactive)
- Category and tag support
- Admin RBAC middleware (`RequireAdmin`)
- Paginated agent listing with filters

---

## 2024-06-15 — PR #2: User Authentication

**Branch:** `feat/user-auth`

### Added
- User registration with bcrypt password hashing
- JWT-based login and token validation
- Auth middleware for protected routes
- User profile endpoint
- Wallet auto-creation on registration (initial balance: 100 credits)

---

## 2024-06-14 — PR #1: Project Init

**Branch:** `feat/project-init`

### Added
- Project scaffolding with Go module structure
- SQLite database with auto-migration
- Configuration loading (YAML)
- Unified API response format
- Error code system
- Structured logging (slog)
- CORS middleware
- Makefile with build/run/test targets
