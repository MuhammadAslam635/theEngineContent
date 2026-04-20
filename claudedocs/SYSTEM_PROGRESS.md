# System Progress + Rules (Source of Truth)

This document is derived from the system plan in [CONTENT_ENGINE_AGENTIC_ARCHITECTURE (1).md](file:///d:/theEngine/claudedocs/CONTENT_ENGINE_AGENTIC_ARCHITECTURE%20(1).md).

## Core Rules (Must Follow)

### Rule 1 — Database Ownership (Non‑Negotiable)

- backend-go is the only service that connects to the database.
- ai-orchestration and celery-worker never use `psycopg2`, never run SQL, never use SQLAlchemy.
- Every read/write goes through backend-go REST APIs via a shared HTTP client module.

### Rule 2 — Shared HTTP Client Only

- ai-orchestration and celery-worker must use a single shared `http_client.py` module for backend-go calls.
- Tools and tasks must call `backend_get/post/patch/put/delete` from that client, not raw `requests` scattered across the codebase.

### Rule 3 — Clear Service Responsibilities

- backend-go: CRUD + config + prompts + API keys + validation + confidence gating helpers.
- ai-orchestration: LangGraph supervisors/agents + external AI vendors (LLM, Whisper, HeyGen, ElevenLabs, Kling) + calls backend-go for state.
- celery-worker: scheduling + long-running tasks + triggers ai-orchestration; never runs LLM logic.

### Rule 6 — Kafka Eventing for Real-time Notifications

- All long-running tasks should publish status events (started, progress, done, failed) to Kafka topics.
- Frontend subscribes via gateway to Kafka (or bridge service) for toast/notifications.
- Topics to standardize: `task.status`, `agent.events`, `production.events`, `analytics.events`.

### Rule 4 — Code Structure Pattern (backend-go)

- `internal/models`: GORM structs only.
- `internal/repositories`: DB operations only (query/update/create).
- `internal/services`: orchestration/business logic; depends on repositories.
- `internal/handlers`: HTTP layer; depends on services; maps DTOs ↔ models.
- `internal/routes`: registers handlers and middleware.

### Rule 5 — Settings Source of Truth

- Provider keys and integration configs live in backend-go settings tables and are fetched by ai-orchestration at runtime.
- Agent prompts, model/provider selection live in backend-go agent settings and are fetched by ai-orchestration at runtime.

## Current Progress (High-Level)

### Implemented

- backend-go: settings models created
  - [GlobalSettings](file:///d:/theEngine/backend-go/internal/models/global_settings.go)
  - [AgentSettings](file:///d:/theEngine/backend-go/internal/models/agent_settings.go)
- migrations: created for both tables
  - [20260420_01_create_global_settings_table.sql](file:///d:/theEngine/migrations/20260420_01_create_global_settings_table.sql)
  - [20260420_02_create_agent_settings_table.sql](file:///d:/theEngine/migrations/20260420_02_create_agent_settings_table.sql)
- backend-go: dependency fix for `github.com/lib/pq` added to [go.mod](file:///d:/theEngine/backend-go/go.mod)

### Not Implemented Yet (Most of Phase 1)

- backend-go: `settings_handler.go`, `settings_service.go`, `settings_repository.go`, routes for:
  - `/settings/agents/*`
  - `/settings/integrations/*`
- ai-orchestration + celery-worker: shared `http_client.py`
- ai-orchestration: replace current simplified graph with Supervisor → Sub-supervisor → Agent topology
- celery-worker: beat schedule + tasks per the matrix

### Violations to Fix (To Match Architecture)

- Any direct DB access inside ai-orchestration/celery-worker must be removed and replaced with backend-go API calls.
- Standardize naming: the plan uses `integration_settings`; current code uses `global_settings`.
