# Status — Remaining

This is the “next backlog” derived from [CONTENT_ENGINE_AGENTIC_ARCHITECTURE (1).md](file:///d:/theEngine/claudedocs/CONTENT_ENGINE_AGENTIC_ARCHITECTURE%20(1).md).

## Phase 1 (Days 1–3)

- backend-go: add settings endpoints
  - `/settings/agents/*`
  - `/settings/integrations/*`
- backend-go: seed agent prompts into agent settings (so prompts can change without code deploy)
- ai-orchestration: add shared `http_client.py` and make tools use it (no direct requests/httpx elsewhere)
- celery-worker: add shared `http_client.py` and make tasks use it
- both Python services: remove `psycopg2-binary` and remove any direct DB logic
- backend-go: add intelligence routes (`/intelligence/accounts`, `/intelligence/reels`, trend signals)
- ai-orchestration: implement `OutlierDetectionAgent` reading accounts from backend-go
- celery-worker: add beat schedule for outlier detection (every 6h)

## Weeks 1–2

- ai-orchestration: implement remaining intelligence agents + tools
- backend-go: pipeline routes (brief lifecycle, SSML, attempts, confidence flags)
- ai-orchestration: implement Production agents (Agent1/Agent2, validation/writer loop, selection loop)
- backend-go: production routes (videos, scenes, packages, analytics)

## Weeks 3–4

- celery-worker: analytics tasks
- ai-orchestration: BaselineBenchmarkAgent + ScriptPromotionAgent

## Architecture Corrections (Must Fix)

- Ensure ai-orchestration and celery-worker never connect to Postgres directly; backend-go is the only DB owner.
- Rename/align `global_settings` vs `integration_settings` (pick one convention and standardize routes + DB).
