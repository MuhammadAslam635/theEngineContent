# Project Structure & Conventions

## Top-Level Layout

```
├── backend-go/          # Go REST API (Gin/GORM) — public-facing via Kong
├── ai-orchestration/    # Python FastAPI + LangGraph multi-agent system — internal only
├── celery-worker/       # Python Celery distributed task workers
├── k8s/                 # Kubernetes manifests (namespace, deployments, services, etc.)
├── migrations/          # Raw SQL migration files for PostgreSQL
├── orchestration.json   # Full LangGraph agent/supervisor blueprint (state, prompts, tools, routing)
├── build_images.ps1     # Docker image build script (PowerShell)
├── apply.ps1            # K8s manifest apply script (PowerShell)
├── migrate_db.ps1       # Database migration script
└── .env                 # Local environment variables (not committed with secrets)
```

## Go Backend (`backend-go/`)

Follows a layered architecture with dependency injection:

```
backend-go/
├── main.go                      # Entrypoint: wires config → DB → repos → services → handlers → routes
├── config/config.go             # Env-based configuration struct
├── db/db.go                     # GORM DB init + AutoMigrate (add new models here)
├── http-client/http-client.go   # HTTP client for internal service calls (e.g., AI orchestration)
├── internal/
│   ├── dto/                     # Request/response data transfer objects
│   ├── models/                  # GORM model structs (one file per domain entity)
│   ├── repositories/            # Data access layer (interface + implementation per domain)
│   │   └── base_repository.go   # Shared helpers: paginate, applyActiveFilter, countRows
│   ├── services/                # Business logic layer (interface + implementation)
│   ├── handlers/                # HTTP handlers (Gin context binding)
│   ├── routes/routes.go         # All route registration in one file
│   └── helpers/                 # Utilities: OTP generation, JWT tokens, email sending
└── templates/                   # HTML email templates
```

### Go Conventions
- Repositories define an interface and a private struct implementing it (e.g., `AuthRepository` interface + `authRepository` struct).
- Services follow the same interface pattern and receive repositories via constructor injection.
- Handlers receive services via constructor injection.
- Models use GORM tags for schema definition. Status/type fields use string constants defined in the same file.
- `db.go` AutoMigrate list must be kept in dependency order (referenced tables before referencing tables).
- All models use `int32` primary keys with `autoIncrement`, `CreatedAt`/`UpdatedAt` timestamps, and JSON tags.
- Sensitive fields use `json:"-"` to exclude from API responses.

## AI Orchestration (`ai-orchestration/`)

Multi-agent system using LangGraph's supervisor/agent pattern:

```
ai-orchestration/
├── main.py                          # FastAPI app with /health, /ready, /generate endpoints
├── langgraph/
│   ├── content_engine_graph.py      # StateGraph construction: nodes, edges, routing
│   ├── state/state.py               # TypedDict state contract (ContentengineState)
│   ├── supervisors/{domain}/        # Supervisor nodes (route to agents or FINISH)
│   ├── agents/{domain}/
│   │   ├── agent.py                 # Agent chain: prompt | llm.bind_tools(tools)
│   │   └── tools.py                 # @tool decorated functions
│   └── utils/utils.py               # Shared: OpenRouter LLM init, supervisor chain factory, agent runner
```

### AI Conventions
- Supervisors use structured output (`route` function) to set `next` field in state.
- Agents bind tools via `llm.bind_tools()` and return to their supervisor after execution.
- Tools are decorated with `@tool` from `langchain_core.tools`.
- State merges use `operator.add` for messages and custom merge functions for `usage_stats` and `valid_companies`.
- LLM calls go through OpenRouter (`openai_api_base="https://openrouter.ai/api/v1"`).
- The full agent topology, prompts, and tool mappings are documented in `orchestration.json` at the project root.

## Database Migrations (`migrations/`)

- Naming: `YYYYMMDD_NN_create_{table_name}_table.sql`
- Pure SQL with `CREATE TABLE IF NOT EXISTS` and explicit foreign key constraints.
- Tables use `SERIAL PRIMARY KEY`, `TIMESTAMPTZ DEFAULT NOW()` for timestamps.
- Indexes created alongside tables in the same migration file.
- GORM AutoMigrate in `db.go` handles schema sync at runtime; raw migrations are the source of truth for initial schema.

## Kubernetes (`k8s/`)

- All resources live in the `cont-gen` namespace.
- Flat file structure (one YAML per concern: `services.yml`, `configmaps.yml`, etc.).
- Deployments are grouped: `infra-deployments.yml` (postgres, redis), `admin-deployments.yml` (kong, konga, pgadmin), `celery-deployments.yml` (worker, beat, flower).
- Network policies enforce that AI orchestration is internal-only (only backend-go can reach it).

## Inter-Service Communication

- External → Kong (port 8000) → backend-go (port 9001): JWT-authenticated
- backend-go → ai-orchestration (port 8086): Internal HTTP with `x-user-id` header
- backend-go → Redis: Celery task enqueuing
- All services → PostgreSQL: Direct connection
