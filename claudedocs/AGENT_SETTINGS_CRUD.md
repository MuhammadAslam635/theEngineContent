# Agent Settings CRUD — Implementation Log

## Date: 2026-04-20

## Summary

Added full CRUD for dynamic agent configuration. Agents, prompts, variables, and output types are now DB-driven instead of hardcoded.

## Migration

**File:** `migrations/20260420_02_create_agent_settings_table.sql`

**Table:** `agent_settings`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Auto-increment ID |
| agent_name | VARCHAR(255) | Agent identifier |
| supervisor | VARCHAR(255) | Parent supervisor (nullable) |
| provider | VARCHAR(100) | LLM provider (e.g. openai) |
| model_name | VARCHAR(255) | Model identifier (e.g. gpt-4o) |
| prompt | TEXT | System prompt with `{{variable}}` placeholders |
| variables | JSONB | Array of `{key, description, required}` objects |
| output_type | VARCHAR(10) | `text` or `json` |
| output_schema | JSONB | JSON schema for structured output (when output_type=json) |
| temperature | REAL | Default 0.7 |
| max_tokens | INT | Default 2048 |
| is_active | BOOLEAN | Default true |
| created_at | TIMESTAMPTZ | Auto-set |
| updated_at | TIMESTAMPTZ | Auto-set |

## Go Backend Files

| File | Purpose |
|------|---------|
| `backend-go/internal/models/agent_setting.go` | GORM model |
| `backend-go/internal/dto/agent_setting_dto.go` | Create/Update request DTOs |
| `backend-go/internal/repositories/agent_setting_repository.go` | Data access layer |
| `backend-go/internal/services/agent_setting_service.go` | Business logic |
| `backend-go/internal/handlers/agent_setting_handler.go` | HTTP handlers |

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/agents/all` | List all agent settings |
| GET | `/agents/:id` | Get single agent by ID |
| POST | `/agents/create` | Create new agent setting |
| PUT | `/agents/update/:id` | Partial update agent setting |
| DELETE | `/agents/delete/:id` | Delete agent setting |

## Wiring

- Repository, service, and handler initialized in `backend-go/main.go`
- Routes registered in `backend-go/internal/routes/routes.go`
- Model added to `backend-go/db/db.go` AutoMigrate list

## Key Design Decisions

- `variables` field uses `json.RawMessage` (JSONB) for flexibility — allows defining prompt placeholders that get injected at runtime
- `output_type` supports `text` (free-form) or `json` (structured) — when `json`, the `output_schema` field defines expected structure
- Update endpoint uses partial updates (only provided fields are changed)
- Follows existing project patterns: interface-based repository/service, constructor injection
