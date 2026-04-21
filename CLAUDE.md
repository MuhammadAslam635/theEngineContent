# CLAUDE.md — Project Guidelines for AI Assistants

## Project

cont-gen (Content Engine) — multi-tenant AI content generation platform. Three services: Go backend (Gin/GORM), Python AI orchestration (FastAPI/LangGraph), Python Celery workers. Deployed on Kubernetes (`cont-gen` namespace) behind Kong API gateway.

## Commands

```bash
# Build images
./build_images.ps1                    # all app services
./build_images.ps1 -BackendGoOnly     # just Go backend
./build_images.ps1 -NoCache           # clean build

# Deploy
./apply.ps1                           # apply all K8s manifests

# Migrations
./migrate_db.ps1

# Go backend local
cd backend-go && go run .

# AI orchestration local
cd ai-orchestration && python main.py

# K8s
kubectl -n cont-gen get pods
kubectl -n cont-gen logs -f deploy/backend-go
```

## Architecture Rules

- External traffic → Kong (:8000) → backend-go (:9001) via JWT
- backend-go → ai-orchestration (:8086) via internal HTTP with `x-user-id` header
- AI orchestration is NEVER exposed externally (NetworkPolicy enforced)
- All data scoped by `tenant_id` and `user_id`
- Celery tasks enqueued via Redis broker

## Go Backend Conventions (`backend-go/`)

### Layered Architecture
- `config/` → `db/` → `repositories/` → `services/` → `handlers/` → `routes/`
- All wiring happens in `main.go` via constructor injection

### Code Style
- Repositories: define an exported interface + unexported struct implementing it
- Services: same pattern, receive repo interfaces via constructor
- Handlers: receive service interfaces, bind JSON from Gin context
- DTOs: separate request/response structs in `internal/dto/`
- Routes: all registered in `internal/routes/routes.go`

### Models
- One file per domain entity in `internal/models/`
- Use `int32` primary keys with `gorm:"primaryKey;autoIncrement"`
- Always include `CreatedAt time.Time` and `UpdatedAt time.Time`
- Use `json:"-"` for sensitive fields (passwords, OTPs)
- Define status/type constants as string consts in the same file
- Use `*Type` (pointer) for nullable fields, with `json:"...,omitempty"`
- Use `pq.StringArray` with `gorm:"type:text[]"` for PostgreSQL arrays
- Add GORM associations at the bottom of the struct

### Database
- `db.go` AutoMigrate list must be in dependency order (referenced tables first)
- Never remove models from AutoMigrate — only append
- Raw SQL migrations in `migrations/` are source of truth for initial schema
- Migration naming: `YYYYMMDD_NN_create_{table_name}_table.sql`
- Use `CREATE TABLE IF NOT EXISTS`, `SERIAL PRIMARY KEY`, `TIMESTAMPTZ DEFAULT NOW()`

### Error Handling
- Return generic error messages to clients ("invalid email or password")
- Never expose internal errors or stack traces in API responses
- Use `errors.New()` for service-layer errors

## Python AI Orchestration Conventions (`ai-orchestration/`)

### LangGraph Pattern
- Supervisor → Agent → Tools (hierarchical multi-agent)
- Supervisors use structured output (`route` function) to set `next` in state
- Agents: `prompt | llm.bind_tools(tools)` composition
- Tools: `@tool` decorator from `langchain_core.tools`
- Agent nodes always return to their parent supervisor
- Loop guard: if last AI message has no tool_calls, supervisor emits FINISH

### State Contract
- `ContentengineState` TypedDict in `langgraph/state/state.py`
- Required fields: `messages`, `tenant_id`, `user_id`, `user_name`, `authorization`
- Merge functions: `operator.add` for messages, `combine_usage` for stats, `limit_list` for rolling lists

### File Organization
- `langgraph/agents/{domain}/agent.py` — agent chain construction
- `langgraph/agents/{domain}/tools.py` — tool definitions
- `langgraph/supervisors/{domain}/supervisor.py` — supervisor chain
- `langgraph/utils/utils.py` — shared utilities (LLM init, supervisor factory, agent runner)
- `global_services/` — cross-cutting concerns (audit logging)

### LLM Configuration
- All LLM calls go through OpenRouter (`openai_api_base="https://openrouter.ai/api/v1"`)
- Default model: `gpt-4o`, web search: `gpt-4o-search-preview`
- `orchestration.json` at project root is the canonical blueprint for all agents, prompts, tools, and routing

## Celery Worker Conventions (`celery-worker/`)

- Tasks defined in `tasks/` package, imported in `__init__.py`
- Use `@app.task(bind=True, max_retries=3)` pattern
- Retry with `self.retry(exc=exc, countdown=60)`
- Config in `celeryconfig.py` (JSON serialization, UTC timezone)

## Documentation

- All implementation changes must be documented in `claudedocs/` directory
- Create a new markdown file per feature/change (e.g. `claudedocs/FEATURE_NAME.md`)
- Include: summary, files changed, API endpoints, migration details, design decisions

## General Rules

- Never commit secrets or real credentials — use K8s secrets or `.env`
- All services must expose `/health` and `/ready` endpoints
- Use snake_case for Python, camelCase for Go variables, PascalCase for Go types/exports
- PostgreSQL column names use snake_case
- JSON response keys use snake_case
- Keep `orchestration.json` in sync when adding/modifying agents or tools
- When adding a new model: create migration SQL, add GORM model, register in `db.go` AutoMigrate
