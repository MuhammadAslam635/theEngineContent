# CONTENT ENGINE — AGENTIC ARCHITECTURE
## Supervisor → Sub-Supervisor → Agent → Tool Design
### Repo: github.com/MuhammadAslam635/theEngineContent | April 2026

---

## THE ONE RULE THAT GOVERNS EVERYTHING

```
┌─────────────────────────────────────────────────────────────────┐
│  NEITHER ai-orchestration NOR celery-worker EVER TOUCHES        │
│  THE DATABASE DIRECTLY.                                          │
│                                                                  │
│  Every CREATE, READ, UPDATE, DELETE for every table goes        │
│  through backend-go REST APIs via the shared HTTP client.       │
│                                                                  │
│  backend-go is the ONLY service that owns database access.      │
│  ai-orchestration and celery-worker are HTTP clients of it.     │
└─────────────────────────────────────────────────────────────────┘
```

This means:
- ai-orchestration needs to store a hook it just extracted → calls `POST /library/hooks`
- ai-orchestration needs the persona list to build an agent prompt → calls `GET /library/personas`
- ai-orchestration needs to update a brief after Agent 1 runs → calls `PATCH /pipeline/briefs/{id}/agent1`
- celery-worker needs videos due for analytics collection → calls `GET /production/videos/pending-analytics`
- celery-worker needs to mark a scene as rendered → calls `PATCH /production/scenes/{id}/rendered`

**No exceptions. No `psycopg2` connections. No direct SQL. No SQLAlchemy sessions in ai-orchestration or celery-worker.**

The `psycopg2-binary` entries currently in both `requirements.txt` files are removed. They are not needed.

---

## KAFKA EVENTING — REAL-TIME NOTIFICATIONS (NEW)

Kafka is added to the stack to power live progress toasts and dashboards.

Topics (standardized)
- task.status — every Celery task publishes start/progress/done/fail
- agent.events — every agent step publishes its decision + confidence
- production.events — video/scene render milestones
- analytics.events — performance collection complete, baseline beat/miss

Producer rules
- backend-go publishes after every state change (task, brief, video, analytics)
- ai-orchestration publishes after every agent invocation (via audit event)
- celery-worker publishes at task boundaries

Consumer rules
- Frontend subscribes via WebSocket gateway (or SSE bridge) that reads from Kafka
- No service ever polls the database for status; it listens to Kafka

Kubernetes
- Zookeeper + Kafka deployed as single-broker for Phase 1 (see k8s/kafka-deployment.yml)
- Services connect at kafka-service:9092 (ClusterIP)
- Persistent volumes for Kafka logs and Zookeeper data

---

## TABLE OF CONTENTS

1. [Three-Service Architecture](#1-three-service-architecture)
2. [Hierarchy Map](#2-hierarchy-map)
3. [The Shared HTTP Client — Both Services](#3-the-shared-http-client--both-services)
4. [Service Responsibility Split](#4-service-responsibility-split)
5. [What Each Service Calls](#5-what-each-service-calls)
6. [Main Supervisor](#6-main-supervisor)
7. [Sub-Supervisors](#7-sub-supervisors)
8. [Agents — System Prompts, Models, Dynamic Data](#8-agents--system-prompts-models-dynamic-data)
9. [Tools — How They Use the HTTP Client](#9-tools--how-they-use-the-http-client)
10. [Celery vs Direct Execution Matrix](#10-celery-vs-direct-execution-matrix)
11. [backend-go: What Gets Added](#11-backend-go-what-gets-added)
12. [ai-orchestration: What Gets Added](#12-ai-orchestration-what-gets-added)
13. [celery-worker: What Gets Added](#13-celery-worker-what-gets-added)
14. [Global Reusable Functions Per Service](#14-global-reusable-functions-per-service)
15. [Phase 1 Task Breakdown](#15-phase-1-task-breakdown)
16. [Complete File Structure](#16-complete-file-structure)

---

## 1. THREE-SERVICE ARCHITECTURE

```
+------------------------------------------------------------------------------+
|  backend-go  (port 9001)  -- Gin + GORM                                     |
|                                                                              |
|  THE ONLY SERVICE WITH DATABASE ACCESS.                                      |
|  Owns: all CRUD, all tables, all config, all API keys, all agent prompts.   |
|  Exposes: REST API consumed by ai-orchestration and celery-worker.           |
|  Does NOT: run LLMs, call HeyGen/ElevenLabs, queue background tasks.        |
+-----------------------------------+------------------------------------------+
                                    |  HTTP (REST)
                    +---------------+---------------+
                    |                               |
       +------------v-----------+    +--------------v-----------+
       |  ai-orchestration      |    |  celery-worker           |
       |  (port 8086)           |    |  (Redis broker)          |
       |  FastAPI + LangGraph   |    |  Celery + Redis          |
       |                        |    |                          |
       |  Calls backend-go for: |    |  Calls backend-go for:   |
       |  - all data reads      |    |  - all data reads        |
       |  - all data writes     |    |  - all data writes       |
       |  after every agent     |    |  before/after tasks      |
       |  step                  |    |                          |
       |  Does NOT: own DB.     |    |  Does NOT: own DB.       |
       +------------------------+    +--------------------------+
```

**Data flow:** Celery fires → ai-orchestration executes agents → every result stored via backend-go API call. No shortcut.

---

## 2. HIERARCHY MAP

```
ContentEngineSupervisor                       [ai-orchestration / LangGraph root]
|
+-- IntelligenceSupervisor                    [SUB-SUPERVISOR 1 — all Celery]
|   +-- OutlierDetectionAgent
|   |   +-- tool_get_accounts()              GET  /intelligence/accounts
|   |   +-- tool_get_platform_credentials()  GET  /settings/integrations/{service}
|   |   +-- tool_sociavault_fetch()          external: SociaVault API
|   |   +-- tool_youtube_fetch()             external: YouTube Data API v3
|   |   +-- tool_tiktok_fetch()              external: TikTok Research API
|   |   +-- tool_ingest_reel()              POST  /intelligence/reels
|   |
|   +-- TranscriptionAgent
|   |   +-- tool_get_reel()                  GET  /intelligence/reels/{id}
|   |   +-- tool_audio_transcribe()          external: Whisper API
|   |   +-- tool_update_reel_transcript()   PATCH /intelligence/reels/{id}/transcript
|   |
|   +-- RelevanceFilterAgent
|   |   +-- tool_get_reel()                  GET  /intelligence/reels/{id}
|   |   +-- tool_get_credentials()           GET  /library/credentials
|   |   +-- tool_update_reel_relevance()    PATCH /intelligence/reels/{id}/relevance
|   |
|   +-- HookAngleExtractionAgent
|   |   +-- tool_get_reel()                  GET  /intelligence/reels/{id}
|   |   +-- tool_get_angles()                GET  /library/angles
|   |   +-- tool_create_hook()              POST  /library/hooks
|   |   +-- tool_create_angle()             POST  /library/angles
|   |
|   +-- TrendSignalAgent
|       +-- tool_get_platforms()             GET  /platforms
|       +-- tool_google_trends_fetch()       external: Google Trends API
|       +-- tool_hashtag_volume_fetch()      external: SociaVault API
|       +-- tool_upsert_trend_signal()      POST  /intelligence/trend-signals
|
+-- ProductionSupervisor                     [SUB-SUPERVISOR 2 — mixed]
|   +-- StrategicDecisionsAgent  [DIRECT]
|   |   +-- tool_get_brief()                 GET  /pipeline/briefs/{id}
|   |   +-- tool_get_platforms()             GET  /platforms
|   |   +-- tool_get_personas()              GET  /library/personas
|   |   +-- tool_get_scripts_health()        GET  /library/scripts/health
|   |   +-- tool_record_agent1()            PATCH /pipeline/briefs/{id}/agent1
|   |
|   +-- CreativeDecisionsAgent  [DIRECT]
|   |   +-- tool_get_brief()                 GET  /pipeline/briefs/{id}
|   |   +-- tool_get_angles()                GET  /library/angles
|   |   +-- tool_get_hooks()                 GET  /library/hooks
|   |   +-- tool_get_persona()               GET  /library/personas/{id}
|   |   +-- tool_record_agent2()            PATCH /pipeline/briefs/{id}/agent2
|   |
|   +-- ScriptValidationAgent  [DIRECT, in loop]
|   |   +-- tool_get_brief()                 GET  /pipeline/briefs/{id}
|   |   +-- tool_get_voicedna_rules()        GET  /settings/voicedna-rules
|   |   +-- tool_get_credentials()           GET  /library/credentials
|   |   +-- tool_get_persona()               GET  /library/personas/{id}
|   |   +-- tool_get_angle()                 GET  /library/angles/{id}
|   |   +-- tool_get_hook()                  GET  /library/hooks/{id}
|   |   +-- tool_record_attempt()           POST  /pipeline/briefs/{id}/attempts
|   |
|   +-- ScriptWriterAgent  [DIRECT, in loop]
|   |   +-- tool_get_brief()                 GET  /pipeline/briefs/{id}
|   |   +-- tool_get_voicedna_rules()        GET  /settings/voicedna-rules
|   |   +-- tool_get_credentials()           GET  /library/credentials
|   |   +-- tool_get_last_attempt()          GET  /pipeline/briefs/{id}/attempts/latest
|   |
|   +-- ScriptSelectionLoopAgent  [DIRECT]
|   |   +-- tool_get_script_candidates()     GET  /library/scripts/candidates
|   |   +-- tool_select_script()            PATCH /pipeline/briefs/{id}/select-script
|   |
|   +-- SSMLFormattingAgent  [CELERY QUEUED]
|   |   +-- tool_get_brief()                 GET  /pipeline/briefs/{id}
|   |   +-- tool_get_selected_script()       GET  /library/scripts/{id}
|   |   +-- tool_get_elevenlabs_config()     GET  /settings/integrations/elevenlabs
|   |   +-- tool_create_ssml()              POST  /pipeline/ssml
|   |   +-- tool_elevenlabs_submit()         external: ElevenLabs API
|   |   +-- tool_mark_audio_ready()         PATCH /pipeline/ssml/{id}/audio-ready
|   |
|   +-- AvatarVideoProductionAgent  [CELERY QUEUED]
|   |   +-- tool_get_brief()                 GET  /pipeline/briefs/{id}
|   |   +-- tool_get_ssml()                  GET  /pipeline/ssml/{brief_id}
|   |   +-- tool_get_heygen_config()         GET  /settings/integrations/heygen
|   |   +-- tool_create_video()             POST  /production/videos
|   |   +-- tool_create_scenes()            POST  /production/videos/{id}/scenes
|   |   +-- tool_heygen_render_scene()       external: HeyGen API
|   |   +-- tool_update_scene_rendered()    PATCH /production/scenes/{id}/rendered
|   |   +-- tool_update_scene_coherence()   PATCH /production/scenes/{id}/coherence
|   |   +-- tool_heygen_stitch()             external: HeyGen API
|   |   +-- tool_update_video_status()      PATCH /production/videos/{id}/render-status
|   |
|   +-- FacelessVideoProductionAgent  [CELERY QUEUED]
|   |   +-- tool_get_brief()                 GET  /pipeline/briefs/{id}
|   |   +-- tool_get_ssml()                  GET  /pipeline/ssml/{brief_id}
|   |   +-- tool_get_kling_config()          GET  /settings/integrations/kling
|   |   +-- tool_create_video()             POST  /production/videos
|   |   +-- tool_kling_generate()            external: Kling AI API
|   |   +-- tool_update_video_status()      PATCH /production/videos/{id}/render-status
|   |
|   +-- PostingPackageAgent  [DIRECT]
|       +-- tool_get_brief()                 GET  /pipeline/briefs/{id}
|       +-- tool_get_video()                 GET  /production/videos/{id}
|       +-- tool_get_posting_times()         GET  /settings/posting-times/{platform}
|       +-- tool_create_package()           POST  /production/videos/{id}/packages
|
+-- AnalyticsSupervisor                      [SUB-SUPERVISOR 3 — all Celery]
    +-- PerformanceCollectionAgent
    |   +-- tool_get_video()                 GET  /production/videos/{id}
    |   +-- tool_get_platform_credentials()  GET  /settings/integrations/{service}
    |   +-- tool_instagram_fetch()           external: Instagram Graph API
    |   +-- tool_youtube_analytics_fetch()   external: YouTube Data API v3
    |   +-- tool_tiktok_analytics_fetch()    external: TikTok API
    |   +-- tool_record_analytics()         POST  /production/videos/{id}/analytics
    |
    +-- BaselineBenchmarkAgent
    |   +-- tool_get_phase1_status()         GET  /production/analytics/phase1-status
    |   +-- tool_log_benchmark()            POST  /production/analytics/benchmark-log
    |
    +-- ScriptPromotionAgent
        +-- tool_get_video_analytics()       GET  /production/videos/{id}/analytics
        +-- tool_get_script()                GET  /library/scripts/{id}
        +-- tool_promote_script()           POST  /library/scripts/{id}/promote
```

---

## 3. THE SHARED HTTP CLIENT — BOTH SERVICES

Both ai-orchestration and celery-worker use a shared HTTP client module to talk to backend-go. **This is the only way either service communicates with the database.** It is the single most important file in the system.

### ai-orchestration HTTP client

```python
# ai-orchestration/http_client.py
#
# THE ONLY WAY ai-orchestration reads or writes any data.
# Every tool, every agent, every supervisor uses this.
# Never import requests or httpx directly anywhere else.
# Never connect to postgres directly anywhere.

import os
import requests
from typing import Optional

BACKEND_URL = os.getenv("BACKEND_GO_URL", "http://backend-go-service:9001")
INTERNAL_TOKEN = os.getenv("INTERNAL_SERVICE_TOKEN", "")


def _headers(authorization: Optional[str] = None) -> dict:
    """
    Global: builds the Authorization header for every backend-go call.
    Uses the per-request user token if provided (from LangGraph state),
    otherwise falls back to the internal service token for system tasks.
    """
    token = authorization if authorization else f"Bearer {INTERNAL_TOKEN}"
    return {
        "Authorization": token,
        "Content-Type": "application/json",
        "X-Service": "ai-orchestration",
    }


def backend_get(path: str, authorization: Optional[str] = None, params: dict = None) -> dict:
    """
    Global: GET any resource from backend-go.
    Used by every tool that reads data — closed lists, configs, records.
    """
    resp = requests.get(
        f"{BACKEND_URL}{path}",
        headers=_headers(authorization),
        params=params or {},
        timeout=15,
    )
    resp.raise_for_status()
    return resp.json()


def backend_post(path: str, payload: dict, authorization: Optional[str] = None) -> dict:
    """
    Global: POST (create) any resource in backend-go.
    Used by every tool that creates a new record — hooks, angles, reels,
    ssml scripts, videos, scenes, packages, analytics records, etc.
    """
    resp = requests.post(
        f"{BACKEND_URL}{path}",
        json=payload,
        headers=_headers(authorization),
        timeout=30,
    )
    resp.raise_for_status()
    return resp.json()


def backend_patch(path: str, payload: dict, authorization: Optional[str] = None) -> dict:
    """
    Global: PATCH (update) any resource in backend-go.
    Used by every tool that updates a record — relevance status, agent decisions,
    audio status, render status, coherence checks, analytics, etc.
    """
    resp = requests.patch(
        f"{BACKEND_URL}{path}",
        json=payload,
        headers=_headers(authorization),
        timeout=15,
    )
    resp.raise_for_status()
    return resp.json()


def backend_put(path: str, payload: dict, authorization: Optional[str] = None) -> dict:
    """
    Global: PUT (full replace) any resource in backend-go.
    Used for settings updates — agent prompts, integration configs, voicedna rules.
    """
    resp = requests.put(
        f"{BACKEND_URL}{path}",
        json=payload,
        headers=_headers(authorization),
        timeout=15,
    )
    resp.raise_for_status()
    return resp.json()


def backend_delete(path: str, authorization: Optional[str] = None) -> dict:
    """
    Global: DELETE (soft-delete) any resource in backend-go.
    Used for deactivating competitor accounts, scripts, etc.
    """
    resp = requests.delete(
        f"{BACKEND_URL}{path}",
        headers=_headers(authorization),
        timeout=15,
    )
    resp.raise_for_status()
    return resp.json()
```

### celery-worker HTTP client

```python
# celery-worker/http_client.py
#
# THE ONLY WAY celery-worker reads or writes any data.
# Every task uses this — never requests or httpx directly, never psycopg2.
# Identical contract to ai-orchestration's http_client.py.

import os
import requests
from typing import Optional

BACKEND_URL = os.getenv("BACKEND_GO_URL", "http://backend-go-service:9001")
AI_URL = os.getenv("AI_ORCHESTRATION_URL", "http://ai-orchestration-service:8086")
INTERNAL_TOKEN = os.getenv("INTERNAL_SERVICE_TOKEN", "")


def _backend_headers() -> dict:
    return {
        "Authorization": f"Bearer {INTERNAL_TOKEN}",
        "Content-Type": "application/json",
        "X-Service": "celery-worker",
    }

def _ai_headers() -> dict:
    return {
        "Authorization": f"Bearer {INTERNAL_TOKEN}",
        "Content-Type": "application/json",
        "X-Service": "celery-worker",
    }


# ── backend-go calls ────────────────────────────────────────────────────────

def backend_get(path: str, params: dict = None) -> dict:
    """Global: GET any resource from backend-go. Used by all celery tasks."""
    resp = requests.get(
        f"{BACKEND_URL}{path}",
        headers=_backend_headers(),
        params=params or {},
        timeout=15,
    )
    resp.raise_for_status()
    return resp.json()


def backend_post(path: str, payload: dict) -> dict:
    """Global: POST (create) any resource in backend-go. Used by all celery tasks."""
    resp = requests.post(
        f"{BACKEND_URL}{path}",
        json=payload,
        headers=_backend_headers(),
        timeout=30,
    )
    resp.raise_for_status()
    return resp.json()


def backend_patch(path: str, payload: dict) -> dict:
    """Global: PATCH (update) any resource in backend-go. Used by all celery tasks."""
    resp = requests.patch(
        f"{BACKEND_URL}{path}",
        json=payload,
        headers=_backend_headers(),
        timeout=15,
    )
    resp.raise_for_status()
    return resp.json()


# ── ai-orchestration calls ──────────────────────────────────────────────────

def trigger_ai(task_type: str, payload: dict) -> dict:
    """
    Global: trigger ai-orchestration to run a graph.
    ALL celery tasks that need AI execution call this — never call ai-orchestration
    endpoints directly in task code.
    """
    resp = requests.post(
        f"{AI_URL}/run",
        json={"task_type": task_type, "user_id": "system", "payload": payload},
        headers=_ai_headers(),
        timeout=600,   # AI tasks can run for minutes
    )
    resp.raise_for_status()
    return resp.json()


# ── escalation ──────────────────────────────────────────────────────────────

def escalate(task_name: str, error: str, payload: dict) -> None:
    """
    Global: after max retries, store an escalation record in backend-go
    so the project manager dashboard surfaces it.
    Never raises — escalation must not cause a second failure.
    """
    try:
        backend_post("/settings/escalations", {
            "source": task_name,
            "error": error,
            "payload": payload,
        })
    except Exception:
        pass  # escalation failure must be silent
```

---

## 4. SERVICE RESPONSIBILITY SPLIT

| Operation | Who does it | How |
|-----------|------------|-----|
| Create a competitor account | backend-go handler | direct GORM write |
| ai-orchestration needs to ingest a reel it detected | calls `POST /intelligence/reels` | via `backend_post()` |
| celery-worker needs the list of active accounts | calls `GET /intelligence/accounts` | via `backend_get()` |
| ai-orchestration extracts a hook, needs to store it | calls `POST /library/hooks` | via `backend_post()` |
| ai-orchestration scores a script, needs to record the attempt | calls `POST /pipeline/briefs/{id}/attempts` | via `backend_post()` |
| ai-orchestration approves/rejects a reel | calls `PATCH /intelligence/reels/{id}/relevance` | via `backend_patch()` |
| ai-orchestration marks ElevenLabs audio ready | calls `PATCH /pipeline/ssml/{id}/audio-ready` | via `backend_patch()` |
| ai-orchestration creates a video record | calls `POST /production/videos` | via `backend_post()` |
| ai-orchestration updates scene coherence result | calls `PATCH /production/scenes/{id}/coherence` | via `backend_patch()` |
| ai-orchestration creates a posting package | calls `POST /production/videos/{id}/packages` | via `backend_post()` |
| celery-worker fetches videos due for analytics | calls `GET /production/videos/pending-analytics` | via `backend_get()` |
| celery-worker records 24h analytics | calls `POST /production/videos/{id}/analytics` | via `backend_post()` |
| ai-orchestration promotes a generated script | calls `POST /library/scripts/{id}/promote` | via `backend_post()` |
| Updating an agent system prompt | admin calls `PUT /settings/agents/{key}` | backend-go handler |
| ai-orchestration reads agent system prompt at task start | calls `GET /settings/agents/{key}` | via `backend_get()` |

**The pattern is the same every time:**
- Need data → `backend_get(path)`
- Need to create a record → `backend_post(path, payload)`
- Need to update a record → `backend_patch(path, payload)`
- Need to replace a config → `backend_put(path, payload)`
- Need to deactivate something → `backend_delete(path)`

---

## 5. WHAT EACH SERVICE CALLS

Complete map of every backend-go endpoint, which service calls it, and in which direction.

### Intelligence endpoints

| Endpoint | Called by | Direction |
|----------|-----------|-----------|
| `GET /intelligence/accounts` | celery-worker (OutlierDetectionTask) | read account list before scan |
| `GET /intelligence/accounts/count` | ai-orchestration (IntelligenceSupervisor prompt) | read for dynamic prompt value |
| `POST /intelligence/accounts` | frontend / admin API | direct backend-go call |
| `PUT /intelligence/accounts/{id}` | frontend / admin API | direct backend-go call |
| `DELETE /intelligence/accounts/{id}` | frontend / admin API | direct backend-go call |
| `PATCH /intelligence/accounts/{id}/baseline` | ai-orchestration (OutlierDetectionAgent) | write updated baseline after scan |
| `POST /intelligence/reels` | ai-orchestration (OutlierDetectionAgent) | write each detected outlier reel |
| `GET /intelligence/reels/{id}` | ai-orchestration (TranscriptionAgent, RelevanceFilterAgent, HookAngleExtractionAgent) | read reel data before processing |
| `PATCH /intelligence/reels/{id}/transcript` | ai-orchestration (TranscriptionAgent) | write verbatim transcript after Whisper |
| `PATCH /intelligence/reels/{id}/relevance` | ai-orchestration (RelevanceFilterAgent) | write approve/reject decision |
| `GET /intelligence/reels/pending-filter` | celery-worker (RelevanceFilterTask dispatch) | read queue of reels awaiting filter |
| `POST /intelligence/trend-signals` | ai-orchestration (TrendSignalAgent) | write upserted trend signal |
| `GET /intelligence/trend-signals/actionable` | celery-worker (ProductionDispatchTask) | read emerging/peaking signals |

### Content library endpoints

| Endpoint | Called by | Direction |
|----------|-----------|-----------|
| `GET /library/hooks` | ai-orchestration (CreativeDecisionsAgent) | read hook closed list for agent prompt |
| `POST /library/hooks` | ai-orchestration (HookAngleExtractionAgent) | write extracted hook from outlier reel |
| `GET /library/hooks/{id}` | ai-orchestration (ScriptValidationAgent) | read hook detail for validation context |
| `GET /library/angles` | ai-orchestration (CreativeDecisionsAgent, HookAngleExtractionAgent) | read angle closed list |
| `POST /library/angles` | ai-orchestration (HookAngleExtractionAgent) | write new angle candidate from outlier |
| `GET /library/angles/{id}` | ai-orchestration (ScriptValidationAgent) | read angle detail for validation context |
| `GET /library/personas` | ai-orchestration (StrategicDecisionsAgent) | read persona closed list for agent prompt |
| `GET /library/personas/{id}` | ai-orchestration (CreativeDecisionsAgent, ScriptValidationAgent, ScriptWriterAgent) | read persona detail |
| `GET /library/credentials` | ai-orchestration (RelevanceFilterAgent, ScriptValidationAgent, ScriptWriterAgent) | read Nick's verified credentials |
| `POST /library/credentials` | frontend / admin API | direct backend-go call |
| `GET /library/scripts` | frontend / admin API | direct backend-go call |
| `GET /library/scripts/health` | ai-orchestration (StrategicDecisionsAgent) | read short/long form count for prompt |
| `GET /library/scripts/candidates` | ai-orchestration (ScriptSelectionLoopAgent) | read banked scripts matching brief |
| `GET /library/scripts/{id}` | ai-orchestration (SSMLFormattingAgent, ScriptPromotionAgent) | read script body |
| `POST /library/scripts` | ai-orchestration (ScriptWriterAgent — promoted via service) | write ai-generated script to storehouse |
| `POST /library/scripts/{id}/promote` | ai-orchestration (ScriptPromotionAgent) | write promotion status after 30d analytics |

### Pipeline endpoints

| Endpoint | Called by | Direction |
|----------|-----------|-----------|
| `POST /pipeline/briefs` | backend-go (triggered by Idea Engine webhook) | direct |
| `GET /pipeline/briefs/{id}` | ai-orchestration (every production agent) | read brief + all join data |
| `PATCH /pipeline/briefs/{id}/agent1` | ai-orchestration (StrategicDecisionsAgent) | write Agent 1 decisions |
| `PATCH /pipeline/briefs/{id}/agent2` | ai-orchestration (CreativeDecisionsAgent) | write Agent 2 decisions |
| `POST /pipeline/briefs/{id}/attempts` | ai-orchestration (ScriptValidationAgent) | write each script attempt record |
| `GET /pipeline/briefs/{id}/attempts/latest` | ai-orchestration (ScriptWriterAgent) | read last attempt for feedback loop |
| `PATCH /pipeline/briefs/{id}/select-script` | ai-orchestration (ScriptSelectionLoopAgent) | write selected script ID |
| `POST /pipeline/ssml` | ai-orchestration (SSMLFormattingAgent) | write SSML script record |
| `GET /pipeline/ssml/{brief_id}` | ai-orchestration (AvatarVideoProductionAgent, FacelessVideoProductionAgent) | read SSML + audio URL |
| `PATCH /pipeline/ssml/{id}/audio-ready` | ai-orchestration (SSMLFormattingAgent) | write ElevenLabs audio URL + duration |
| `GET /pipeline/confidence-flags` | frontend / admin API | direct backend-go call |
| `POST /pipeline/confidence-flags` | ai-orchestration (any agent with confidence < 70) | write flag for human review |
| `PATCH /pipeline/confidence-flags/{id}/resolve` | frontend / admin API | direct backend-go call |
| `PATCH /pipeline/confidence-flags/{id}/escalate` | ai-orchestration (adversarial audit disagreement) | write escalation |

### Production endpoints

| Endpoint | Called by | Direction |
|----------|-----------|-----------|
| `POST /production/videos` | ai-orchestration (AvatarVideoProductionAgent, FacelessVideoProductionAgent) | write new video record |
| `GET /production/videos/{id}` | ai-orchestration (PostingPackageAgent) | read video record |
| `PATCH /production/videos/{id}/render-status` | ai-orchestration (both video production agents) | write render progress |
| `POST /production/videos/{id}/scenes` | ai-orchestration (AvatarVideoProductionAgent) | write all scene records in one call |
| `PATCH /production/scenes/{id}/rendered` | ai-orchestration (AvatarVideoProductionAgent) | write scene render URL |
| `PATCH /production/scenes/{id}/coherence` | ai-orchestration (AvatarVideoProductionAgent) | write coherence check pass/fail |
| `POST /production/videos/{id}/packages` | ai-orchestration (PostingPackageAgent) | write completed posting package |
| `PATCH /production/packages/{id}/review` | frontend / admin dashboard | direct backend-go call |
| `PATCH /production/packages/{id}/posted` | frontend / admin dashboard (after manual posting) | direct backend-go call |
| `GET /production/videos/pending-analytics` | celery-worker (AnalyticsDispatchTask) | read videos due for collection |
| `POST /production/videos/{id}/analytics` | ai-orchestration (PerformanceCollectionAgent) | write analytics per window |
| `GET /production/videos/{id}/analytics` | ai-orchestration (ScriptPromotionAgent) | read 30d analytics for promotion check |
| `GET /production/analytics/phase1-status` | ai-orchestration (BaselineBenchmarkAgent) | read Phase 1 running success rate |
| `POST /production/analytics/benchmark-log` | ai-orchestration (BaselineBenchmarkAgent) | write benchmark assessment record |

### Settings endpoints

| Endpoint | Called by | Direction |
|----------|-----------|-----------|
| `GET /settings/agents` | ai-orchestration (at every graph init) | read all agent settings |
| `GET /settings/agents/{key}` | ai-orchestration (per-agent settings fetch) | read one agent's prompt + model |
| `PUT /settings/agents/{key}` | frontend / admin API | direct backend-go call |
| `GET /settings/integrations/sociavault` | ai-orchestration (OutlierDetectionAgent, TrendSignalAgent) | read API key |
| `GET /settings/integrations/elevenlabs` | ai-orchestration (SSMLFormattingAgent) | read voice ID + API key |
| `GET /settings/integrations/heygen` | ai-orchestration (AvatarVideoProductionAgent) | read avatar ID + API key |
| `GET /settings/integrations/kling` | ai-orchestration (FacelessVideoProductionAgent) | read API key |
| `GET /settings/integrations/{platform}` | ai-orchestration (PerformanceCollectionAgent) | read platform analytics API credentials |
| `PUT /settings/integrations/{service}` | frontend / admin API | direct backend-go call |
| `GET /settings/voicedna-rules` | ai-orchestration (ScriptValidationAgent, ScriptWriterAgent) | read VoiceDNA rules for prompt injection |
| `PUT /settings/voicedna-rules` | frontend / admin API | direct backend-go call |
| `GET /settings/posting-times/{platform}` | ai-orchestration (PostingPackageAgent) | read peak engagement windows |
| `POST /settings/escalations` | celery-worker (after max retries) | write escalation record |

---

## 6. MAIN SUPERVISOR

**File:** `ai-orchestration/langgraph/content_engine_graph.py`

**System prompt (stored in backend-go `agent_settings`, key: `content_engine_supervisor`):**

```
You are the Content Engine Main Supervisor for 7 Figure Cartel.
You orchestrate AI agents that produce video content at scale for Nick Perry.

Non-negotiable rules:
- All selections come from closed lists fetched from the backend API. Never invent.
- Every agent output carries a confidence score 0-100.
- Confidence below 70 -> flag for human review via POST /pipeline/confidence-flags.
- Critical decisions require adversarial audit before proceeding.
- After 2 failed retries on any step -> escalate to project manager.
- Script confidence threshold is 90%. Hard gate. No video proceeds below it
  unless 5 Writer Agent attempts have been exhausted.

Current task: {task_type}
Content brief ID: {brief_id}
User: {user_id}
```

**Fetches this prompt at init:**
```python
# ai-orchestration/langgraph/content_engine_graph.py
from http_client import backend_get

def create_content_engine_graph(task_type: str, authorization: str):
    settings = backend_get("/settings/agents", authorization)
    settings_map = {s["agent_key"]: s for s in settings["data"]}
    # Build LLM per agent using settings_map[agent_key]["model_name"] etc.
    # Build prompts using settings_map[agent_key]["system_prompt"]
    ...
```

---

## 7. SUB-SUPERVISORS

### `IntelligenceSupervisor`

**File:** `ai-orchestration/langgraph/supervisors/intelligence/supervisor.py`

**System prompt (key: `intelligence_supervisor`) — dynamic values fetched before building:**

```python
# Fetch dynamic values from backend-go before constructing the prompt
accounts_resp = backend_get("/intelligence/accounts/count", authorization)
platforms_resp = backend_get("/platforms", authorization)
settings = backend_get("/settings/agents/intelligence_supervisor", authorization)

prompt = settings["system_prompt"].format(
    account_count=accounts_resp["count"],
    platform_list=", ".join(p["name"] for p in platforms_resp["data"]),
)
```

**Stored prompt template:**
```
You are the Intelligence Supervisor for the Content Engine.
Manage the continuous background loop feeding the Approved Scripts Database.

Monitored accounts: {account_count}
Platforms: {platform_list}

Critical rules:
- 5X threshold is per-account-relative. Never apply a fixed global number.
- Scripts only approved if Nick can speak from his verified credentials.
- Hooks and angles added to library ONLY from real proven outlier content.
- Storehouse must never be empty: short_form always banked, long_form from day one.
```

---

### `ProductionSupervisor`

**File:** `ai-orchestration/langgraph/supervisors/production/supervisor.py`

**Dynamic values fetched before building:**

```python
brief = backend_get(f"/pipeline/briefs/{brief_id}", authorization)
settings = backend_get("/settings/agents/production_supervisor", authorization)

prompt = settings["system_prompt"].format(
    brief_id=brief_id,
    idea_source=brief["data"]["idea_source"] or "unknown",
    trend_stage=brief["data"]["trend_signal"]["stage"] if brief["data"]["trend_signal"] else "unknown",
    trend_urgency=brief["data"]["trend_signal"]["urgency_rating"] if brief["data"]["trend_signal"] else "unknown",
    outlier_reel_id=brief["data"]["outlier_reel_ref_id"] or "none",
)
```

**Stored prompt template:**
```
You are the Production Supervisor for the Content Engine.
Take approved content brief {brief_id} and produce a fully finished posting package.

Idea source: {idea_source}
Trend stage: {trend_stage} | Urgency: {trend_urgency}
Outlier reel reference: {outlier_reel_id}

Strict sequence: Agent1 -> Agent2 -> ScriptLoop -> SSML -> VideoProduction -> Package.
Hard gates: 70% confidence before next step. 90% for scripts. Coherence check mandatory.
```

---

### `AnalyticsSupervisor`

**File:** `ai-orchestration/langgraph/supervisors/analytics/supervisor.py`

**Dynamic values fetched before building:**

```python
video = backend_get(f"/production/videos/{video_id}", authorization)
settings = backend_get("/settings/agents/analytics_supervisor", authorization)

prompt = settings["system_prompt"].format(
    video_id=video_id,
    platform=video["data"]["content_brief"]["primary_platform"]["name"],
    posted_at=video["data"]["content_brief"]["posting_packages"][0]["posted_at"],
    window=window,
)
```

**Stored prompt template:**
```
You are the Analytics Supervisor for the Content Engine.
Collect performance data for video {video_id} on {platform}.
Posted at: {posted_at} | Collection window: {window}
Phase 1 success metric: consistently beat Nick's 38,000 average view baseline.
Promotion threshold: beat baseline AND completion_rate >= 70% at 30d window.
```

---

## 8. AGENTS — SYSTEM PROMPTS, MODELS, DYNAMIC DATA

Every agent follows this exact pattern:

```python
# Pattern used by EVERY agent in ai-orchestration
from http_client import backend_get, backend_post, backend_patch

def build_agent(agent_key: str, authorization: str, context_ids: dict) -> str:
    """
    1. Fetch agent settings (prompt template + model config) from backend-go.
    2. Fetch all dynamic data needed to fill {placeholders} from backend-go.
    3. Fill the template.
    4. Build the LangChain agent with the filled prompt.
    Never hardcode a prompt or a model name.
    """
    settings = backend_get(f"/settings/agents/{agent_key}", authorization)
    # fetch additional dynamic data based on context_ids
    # fill settings["system_prompt"].format(**dynamic_values)
    # return filled prompt
```

### Agent settings table record per agent

| agent_key | model_name | temperature | notes |
|-----------|-----------|-------------|-------|
| `content_engine_supervisor` | `anthropic/claude-sonnet-4-5` | 0.2 | routing only |
| `intelligence_supervisor` | `anthropic/claude-sonnet-4-5` | 0.2 | routing only |
| `production_supervisor` | `anthropic/claude-sonnet-4-5` | 0.2 | routing only |
| `analytics_supervisor` | `anthropic/claude-sonnet-4-5` | 0.1 | routing only |
| `outlier_detection_agent` | `anthropic/claude-sonnet-4-5` | 0.1 | factual scan |
| `transcription_agent` | `anthropic/claude-haiku-4-5` | 0.0 | tool-only |
| `relevance_filter_agent` | `anthropic/claude-sonnet-4-5` | 0.1 | binary decision |
| `hook_angle_extraction_agent` | `anthropic/claude-sonnet-4-5` | 0.2 | structured extraction |
| `trend_signal_agent` | `anthropic/claude-sonnet-4-5` | 0.1 | classification |
| `strategic_decisions_agent` | `anthropic/claude-sonnet-4-5` | 0.3 | structured decision |
| `creative_decisions_agent` | `anthropic/claude-sonnet-4-5` | 0.3 | structured decision |
| `script_validation_agent` | `anthropic/claude-sonnet-4-5` | 0.1 | precise scoring |
| `script_writer_agent` | `anthropic/claude-opus-4` | 0.7 | creative writing |
| `ssml_formatting_agent` | `anthropic/claude-sonnet-4-5` | 0.2 | structured format |
| `avatar_video_production_agent` | `anthropic/claude-sonnet-4-5` | 0.1 | procedural |
| `faceless_video_production_agent` | `anthropic/claude-sonnet-4-5` | 0.3 | creative |
| `posting_package_agent` | `anthropic/claude-sonnet-4-5` | 0.4 | copywriting |
| `performance_collection_agent` | `anthropic/claude-haiku-4-5` | 0.0 | tool-only |
| `baseline_benchmark_agent` | `anthropic/claude-haiku-4-5` | 0.1 | factual |
| `script_promotion_agent` | `anthropic/claude-haiku-4-5` | 0.0 | tool-only |

### How dynamic values are injected — concrete example for `ScriptValidationAgent`

```python
# ai-orchestration/langgraph/agents/production/script_validation_agent.py

from http_client import backend_get, backend_post

def build_script_validation_agent(brief_id: int, script_text: str, authorization: str):
    # Step 1: fetch agent settings (prompt template + model)
    settings = backend_get("/settings/agents/script_validation_agent", authorization)

    # Step 2: fetch ALL dynamic data from backend-go — never hardcoded
    brief        = backend_get(f"/pipeline/briefs/{brief_id}", authorization)["data"]
    credentials  = backend_get("/library/credentials", authorization)["data"]
    voicedna     = backend_get("/settings/voicedna-rules", authorization)["data"]
    persona      = backend_get(f"/library/personas/{brief['persona_id']}", authorization)["data"]
    angle        = backend_get(f"/library/angles/{brief['primary_angle_id']}", authorization)["data"]
    hook         = backend_get(f"/library/hooks/{brief['primary_hook_id']}", authorization)["data"]

    # Step 3: format credentials and voicedna as readable lists
    credentials_text = "\n".join(
        f"{i+1}. {c['credential_key']}: {c['display_value']}"
        for i, c in enumerate(credentials)
    )
    voicedna_text = "\n".join(
        f"{i+1}. {rule['name']}: {rule['description']}"
        for i, rule in enumerate(voicedna["rules"])
    )

    # Step 4: fill the prompt template stored in backend-go
    filled_prompt = settings["system_prompt"].format(
        idea_source      = brief.get("idea_source", ""),
        persona_name     = persona["display_name"],
        persona_core_pain= persona["core_pain"],
        angle_name       = angle["display_name"],
        angle_mechanism  = angle["psychological_mechanism"],
        hook_text        = hook["hook_text"],
        content_type     = brief["content_type"],
        nick_credentials = credentials_text,
        voicedna_rules   = voicedna_text,
        script_text      = script_text,
    )

    # Step 5: build the LangChain agent with filled prompt + correct model
    llm = get_openrouter_llm(settings["model_name"], settings["temperature"])
    return build_langchain_agent(llm, filled_prompt, get_validation_tools(authorization))
```

**The stored prompt template (in `agent_settings.system_prompt`):**

```
You are the Script Validation Agent for the Content Engine.
Score a script across five dimensions and produce an adjusted version
with all issues corrected. A script CANNOT reach 90% without VoiceDNA compliance.

Content brief context:
- Idea: {idea_source}
- Persona: {persona_name} | Core pain: {persona_core_pain}
- Angle: {angle_name} | Mechanism: {angle_mechanism}
- Hook pattern: {hook_text}
- Media type: {content_type}

Nick's verified credentials (only these -- never invented stats):
{nick_credentials}

VoiceDNA Rules -- check every rule, flag every violation with line number:
{voicedna_rules}

Script to validate:
"{script_text}"

Score each dimension 0-100. Return:
{{
  "score_idea_alignment": float, "score_angle_match": float,
  "score_hook_match": float, "score_persona_fit": float,
  "score_voicedna": float, "overall_score": float,
  "passed_threshold": bool,
  "issues_found": ["Line 3: uses 'might' -- VoiceDNA rule 1 violation"],
  "adjusted_script": str
}}
```

This same pattern — fetch settings, fetch dynamic data, fill template, build agent — applies to every agent. The template is always stored in `agent_settings.system_prompt`. The dynamic data always comes from backend-go API calls.

---

## 9. TOOLS — HOW THEY USE THE HTTP CLIENT

Every tool in ai-orchestration imports from `http_client.py` only. No tool opens its own HTTP connection. No tool touches a database.

```python
# ai-orchestration/langgraph/tools/intelligence/reel_tools.py

from langchain_core.tools import tool
from http_client import backend_get, backend_post, backend_patch

# The authorization token flows through every tool via the tool's closure
# or via tool factory pattern — never hardcoded.

def get_reel_tools(authorization: str):
    """Factory: returns tool functions bound to the request's authorization token."""

    @tool
    def tool_get_reel(reel_id: int) -> dict:
        """Fetch a single outlier reel record from backend-go."""
        return backend_get(f"/intelligence/reels/{reel_id}", authorization)

    @tool
    def tool_update_reel_transcript(reel_id: int, transcript: str) -> dict:
        """Store the verbatim transcript for an outlier reel in backend-go."""
        return backend_patch(
            f"/intelligence/reels/{reel_id}/transcript",
            {"raw_transcript": transcript},
            authorization,
        )

    @tool
    def tool_update_reel_relevance(reel_id: int, decision: str, reason: str) -> dict:
        """Store the relevance filter decision (approve/reject) in backend-go."""
        return backend_patch(
            f"/intelligence/reels/{reel_id}/relevance",
            {"relevance_status": decision, "relevance_reason": reason},
            authorization,
        )

    return [tool_get_reel, tool_update_reel_transcript, tool_update_reel_relevance]
```

```python
# ai-orchestration/langgraph/tools/intelligence/hook_tools.py

from langchain_core.tools import tool
from http_client import backend_get, backend_post

def get_hook_tools(authorization: str):

    @tool
    def tool_get_hooks(order_by: str = "avg_performance") -> dict:
        """Fetch the hook library closed list from backend-go for agent prompt building."""
        return backend_get("/library/hooks", authorization, params={"order": order_by})

    @tool
    def tool_create_hook(
        outlier_reel_id: int,
        hook_text: str,
        visual_hook: str,
        timing_seconds: float,
        pacing: str,
        emotional_trigger: str,
        topic_tags: list,
    ) -> dict:
        """Store a newly extracted hook from an outlier reel into backend-go."""
        return backend_post("/library/hooks", {
            "outlier_reel_id": outlier_reel_id,
            "hook_text": hook_text,
            "visual_hook": visual_hook,
            "timing_seconds": timing_seconds,
            "pacing": pacing,
            "emotional_trigger": emotional_trigger,
            "topic_tags": topic_tags,
        }, authorization)

    return [tool_get_hooks, tool_create_hook]
```

```python
# ai-orchestration/langgraph/tools/production/brief_tools.py

from langchain_core.tools import tool
from http_client import backend_get, backend_patch

def get_brief_tools(authorization: str):

    @tool
    def tool_get_brief(brief_id: int) -> dict:
        """Fetch a content brief with all joins (platform, persona, angle, hook, trend signal)."""
        return backend_get(f"/pipeline/briefs/{brief_id}", authorization)

    @tool
    def tool_record_agent1(brief_id: int, decision: dict) -> dict:
        """Write Agent 1's strategic decisions to the content brief record in backend-go."""
        return backend_patch(f"/pipeline/briefs/{brief_id}/agent1", decision, authorization)

    @tool
    def tool_record_agent2(brief_id: int, decision: dict) -> dict:
        """Write Agent 2's creative decisions to the content brief record in backend-go."""
        return backend_patch(f"/pipeline/briefs/{brief_id}/agent2", decision, authorization)

    return [tool_get_brief, tool_record_agent1, tool_record_agent2]
```

---

## 10. CELERY VS DIRECT EXECUTION MATRIX

| Agent | Mode | Trigger | Why |
|-------|------|---------|-----|
| OutlierDetectionAgent | Celery Scheduled (6h) | beat | Always-on background scan |
| TranscriptionAgent | Celery Queued | reel ingested | Heavy audio fetch -- async |
| RelevanceFilterAgent | Celery Queued | transcript stored | LLM call -- async |
| HookAngleExtractionAgent | Celery Queued | reel approved | LLM call -- async |
| TrendSignalAgent | Celery Scheduled (4h) | beat | Independent of idea flow |
| StrategicDecisionsAgent | Direct (Sync) | in production graph | Gate -- Agent 2 blocked until done |
| CreativeDecisionsAgent | Direct (Sync) | after Agent 1 | Gate -- ScriptLoop blocked until done |
| ScriptValidationAgent | Direct (Sync) | inside ScriptLoop | Loop control requires sync score |
| ScriptWriterAgent | Direct (Sync) | inside ScriptLoop | Iterative feedback -- must be sync |
| ScriptSelectionLoopAgent | Direct (Sync) | after Agent 2 | Orchestrates the sync loop |
| SSMLFormattingAgent | Celery Queued | after loop exits | ElevenLabs API is slow |
| AvatarVideoProductionAgent | Celery Queued | after audio ready | HeyGen renders take minutes |
| FacelessVideoProductionAgent | Celery Queued | after audio ready | Kling AI is slow |
| PostingPackageAgent | Direct (Sync) | after video ready | Fast LLM -- no slow external API |
| PerformanceCollectionAgent | Celery Scheduled (per video, 24h/7d/30d) | time-interval | Time-based per posted video |
| BaselineBenchmarkAgent | Celery Queued | after 7d analytics | Triggered by data availability |
| ScriptPromotionAgent | Celery Queued | after 30d analytics | Needs full 30d picture |

---

## 11. BACKEND-GO: WHAT GETS ADDED

Follows the existing `handler -> service -> repository -> model` pattern exactly.
`baseRepository` already has `paginate()`, `applyActiveFilter()`, `applyStatusFilter()`.
All new repositories embed it. No duplication.

### New models

```
backend-go/internal/models/
+-- agent_settings.go          agent system prompts + model_name + temperature per agent
+-- integration_settings.go    SociaVault/ElevenLabs/HeyGen/Kling API keys and account IDs
+-- voicedna_rules.go          VoiceDNA rules stored as structured JSON in DB
+-- posting_time_config.go     peak engagement windows per platform + persona
```

### `agent_settings` model

```go
type AgentSettings struct {
    ID            int32     `gorm:"primaryKey;autoIncrement" json:"id"`
    AgentKey      string    `gorm:"uniqueIndex;size:100;not null" json:"agent_key"`
    DisplayName   string    `gorm:"size:150;not null" json:"display_name"`
    SupervisorKey string    `gorm:"size:100;not null" json:"supervisor_key"`
    SystemPrompt  string    `gorm:"type:text;not null" json:"system_prompt"`
    ModelProvider string    `gorm:"size:50;default:openrouter" json:"model_provider"`
    ModelName     string    `gorm:"size:150;not null" json:"model_name"`
    Temperature   float64   `gorm:"type:numeric(3,2);default:0.20" json:"temperature"`
    MaxTokens     int       `gorm:"default:1000" json:"max_tokens"`
    IsActive      bool      `gorm:"default:true" json:"is_active"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
}
```

### `integration_settings` model

```go
type IntegrationSettings struct {
    ID          int32     `gorm:"primaryKey;autoIncrement" json:"id"`
    ServiceKey  string    `gorm:"uniqueIndex;size:100;not null" json:"service_key"`
    // sociavault | elevenlabs | heygen | kling | instagram | youtube | tiktok
    ApiKey      string    `gorm:"not null" json:"api_key"`
    ApiSecret   *string   `json:"api_secret,omitempty"`
    AccountID   *string   `gorm:"size:255" json:"account_id,omitempty"`
    // For ElevenLabs: Nick's voice ID. For HeyGen: avatar ID.
    ExtraConfig string    `gorm:"type:jsonb;default:'{}'" json:"extra_config"`
    IsActive    bool      `gorm:"default:true" json:"is_active"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### New handlers (each follows the exact `AuthHandler` pattern)

```
backend-go/internal/handlers/
+-- intelligence_handler.go      competitor accounts, reels, trend signals CRUD
+-- content_library_handler.go   hooks, angles, personas, credentials, scripts CRUD
+-- pipeline_handler.go          content briefs, SSML, attempts, confidence flags
+-- production_handler.go        videos, scenes, packages, analytics
+-- settings_handler.go          agent settings, integration keys, voicedna, posting times
```

### New global helpers (never duplicated across handlers/services)

```go
// backend-go/internal/helpers/confidence_gate.go
const ScriptPassThreshold      = 90.0
const ConfidenceReviewThreshold = 70.0
func IsConfidenceFlagged(score float64) bool   { return score < ConfidenceReviewThreshold }
func PassesScriptThreshold(score float64) bool  { return score >= ScriptPassThreshold }

// backend-go/internal/helpers/outlier_threshold.go
func ComputeOutlierThreshold(avgViewCount int64) int64 { return avgViewCount * 5 }
func IsOutlier(viewCount, threshold int64) bool        { return viewCount >= threshold }
func ComputeMultiplier(viewCount, avg int64) float64 {
    if avg == 0 { return 0 }
    return float64(viewCount) / float64(avg)
}

// backend-go/internal/helpers/beat_baseline.go
const NickBaseline int64 = 38000
func BeatBaseline(viewCount int64) bool      { return viewCount > NickBaseline }
func BaselineMargin(viewCount int64) int64   { return viewCount - NickBaseline }
```

---

## 12. AI-ORCHESTRATION: WHAT GETS ADDED

### `requirements.txt` change

```
# REMOVE: psycopg2-binary  (no direct DB access -- ever)
# KEEP:   requests          (already present -- used for all backend-go calls)

fastapi
uvicorn
requests          # for http_client.py -- all backend-go calls
langgraph>=0.1.0
langchain-openai
langchain-core
```

### New files

```
ai-orchestration/
+-- http_client.py             THE HTTP CLIENT -- all backend-go calls go through here
|
+-- langgraph/
    +-- content_engine_graph.py    REPLACED -- full graph
    +-- state/state.py             REPLACED -- ContentEngineState extended
    |
    +-- supervisors/
    |   +-- intelligence/supervisor.py   NEW
    |   +-- production/supervisor.py     NEW
    |   +-- analytics/supervisor.py      NEW
    |
    +-- agents/
    |   +-- intelligence/
    |   |   +-- outlier_detection_agent.py      NEW
    |   |   +-- transcription_agent.py          NEW
    |   |   +-- relevance_filter_agent.py       NEW
    |   |   +-- hook_angle_extraction_agent.py  NEW
    |   |   +-- trend_signal_agent.py           NEW
    |   +-- production/
    |   |   +-- strategic_decisions_agent.py    NEW  (Agent 1)
    |   |   +-- creative_decisions_agent.py     NEW  (Agent 2)
    |   |   +-- script_selection_loop_agent.py  NEW
    |   |   +-- script_validation_agent.py      NEW
    |   |   +-- script_writer_agent.py          NEW
    |   |   +-- ssml_formatting_agent.py        NEW
    |   |   +-- avatar_video_production_agent.py NEW
    |   |   +-- faceless_video_production_agent.py NEW
    |   |   +-- posting_package_agent.py        NEW
    |   +-- analytics/
    |       +-- performance_collection_agent.py NEW
    |       +-- baseline_benchmark_agent.py     NEW
    |       +-- script_promotion_agent.py       NEW
    |
    +-- tools/
    |   +-- intelligence/
    |   |   +-- account_tools.py          GET /intelligence/accounts
    |   |   +-- reel_tools.py             GET/PATCH /intelligence/reels/*
    |   |   +-- sociavault_tools.py       external: SociaVault API
    |   |   +-- youtube_tools.py          external: YouTube Data API v3
    |   |   +-- tiktok_tools.py           external: TikTok Research API
    |   |   +-- transcription_tools.py    external: Whisper API
    |   |   +-- hook_tools.py             GET/POST /library/hooks
    |   |   +-- angle_tools.py            GET/POST /library/angles
    |   |   +-- credential_tools.py       GET /library/credentials
    |   |   +-- trend_tools.py            GET/POST /intelligence/trend-signals + external
    |   +-- production/
    |   |   +-- brief_tools.py            GET/PATCH /pipeline/briefs/*
    |   |   +-- platform_tools.py         GET /platforms
    |   |   +-- persona_tools.py          GET /library/personas/*
    |   |   +-- script_tools.py           GET /library/scripts/*
    |   |   +-- attempt_tools.py          GET/POST /pipeline/briefs/{id}/attempts
    |   |   +-- ssml_tools.py             GET/POST/PATCH /pipeline/ssml/*
    |   |   +-- elevenlabs_tools.py       external: ElevenLabs API
    |   |   +-- video_tools.py            GET/POST/PATCH /production/videos/*
    |   |   +-- scene_tools.py            POST/PATCH /production/scenes/*
    |   |   +-- heygen_tools.py           external: HeyGen API
    |   |   +-- kling_tools.py            external: Kling AI API
    |   |   +-- package_tools.py          POST /production/videos/{id}/packages
    |   |   +-- settings_tools.py         GET /settings/integrations/* + voicedna-rules
    |   +-- analytics/
    |       +-- analytics_tools.py        GET/POST /production/videos/{id}/analytics
    |       +-- instagram_tools.py        external: Instagram Graph API
    |       +-- youtube_tools.py          external: YouTube Data API v3
    |       +-- tiktok_tools.py           external: TikTok API
    |       +-- promotion_tools.py        POST /library/scripts/{id}/promote
    |
    +-- utils/
        +-- utils.py              existing: get_openrouter_llm, create_supervisor_chain
        +-- settings_loader.py    NEW: load_agent_settings(authorization) via backend_get
        +-- confidence.py         NEW: compute_confidence_flag, build_agent_output
        +-- retry.py              NEW: agent_retry_wrapper (2 retries then escalate)
        +-- prompt_builder.py     NEW: build_prompt(template, **dynamic_values)
        +-- adversarial.py        NEW: run_adversarial_audit(llm, decision, context)
```

---

## 13. CELERY-WORKER: WHAT GETS ADDED

### `requirements.txt` change

```
# REMOVE: psycopg2-binary  (no direct DB access -- ever)
# KEEP:   requests          (already present -- used for all calls)

celery[redis]
requests          # for http_client.py
```

### New files

```
celery-worker/
+-- http_client.py             THE HTTP CLIENT -- all backend-go and ai-orchestration calls
|
+-- tasks/
    +-- __init__.py            existing (Celery app)
    +-- intelligence_tasks.py  NEW
    +-- production_tasks.py    NEW
    +-- analytics_tasks.py     NEW
```

### Beat schedule (`celeryconfig.py` addition)

```python
from celery.schedules import crontab

beat_schedule = {
    "outlier-detection-6h": {
        "task": "tasks.intelligence_tasks.run_outlier_detection",
        "schedule": crontab(minute=0, hour="*/6"),
    },
    "trend-signal-4h": {
        "task": "tasks.intelligence_tasks.run_trend_signal_scan",
        "schedule": crontab(minute=30, hour="*/4"),
    },
    "analytics-dispatch-daily": {
        "task": "tasks.analytics_tasks.dispatch_due_analytics_collections",
        "schedule": crontab(minute=0, hour=6),
    },
}
```

### Task files — use `http_client` exclusively

```python
# celery-worker/tasks/intelligence_tasks.py

from . import app
from http_client import backend_get, trigger_ai, escalate


@app.task(bind=True, max_retries=2, default_retry_delay=300)
def run_outlier_detection(self):
    """
    Scheduled every 6 hours.
    Reads account list from backend-go, then triggers ai-orchestration
    to run OutlierDetectionAgent.
    All reel writes happen inside ai-orchestration via POST /intelligence/reels.
    """
    try:
        return trigger_ai("intelligence_scan", {"scan_type": "outlier_detection"})
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_outlier_detection", str(exc), {})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=60)
def run_transcription(self, outlier_reel_id: int):
    """
    Queued after a new outlier reel is stored by OutlierDetectionAgent.
    Triggers TranscriptionAgent in ai-orchestration.
    Transcript written back via PATCH /intelligence/reels/{id}/transcript.
    """
    try:
        return trigger_ai("intelligence_scan", {
            "scan_type": "transcription",
            "outlier_reel_id": outlier_reel_id,
        })
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_transcription", str(exc), {"outlier_reel_id": outlier_reel_id})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=60)
def run_relevance_filter(self, outlier_reel_id: int):
    """
    Queued after TranscriptionAgent stores a transcript.
    Relevance decision written back via PATCH /intelligence/reels/{id}/relevance.
    """
    try:
        return trigger_ai("intelligence_scan", {
            "scan_type": "relevance_filter",
            "outlier_reel_id": outlier_reel_id,
        })
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_relevance_filter", str(exc), {"outlier_reel_id": outlier_reel_id})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=60)
def run_hook_angle_extraction(self, outlier_reel_id: int):
    """
    Queued after RelevanceFilterAgent approves a reel.
    Hook written via POST /library/hooks.
    Angle written via POST /library/angles or matched to existing.
    """
    try:
        return trigger_ai("intelligence_scan", {
            "scan_type": "hook_angle_extraction",
            "outlier_reel_id": outlier_reel_id,
        })
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_hook_angle_extraction", str(exc), {"outlier_reel_id": outlier_reel_id})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=300)
def run_trend_signal_scan(self):
    """Scheduled every 4 hours. Trend signals written via POST /intelligence/trend-signals."""
    try:
        return trigger_ai("intelligence_scan", {"scan_type": "trend_signal"})
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_trend_signal_scan", str(exc), {})
        raise self.retry(exc=exc)
```

```python
# celery-worker/tasks/production_tasks.py

from . import app
from http_client import trigger_ai, escalate


@app.task(bind=True, max_retries=2, default_retry_delay=60)
def run_production_pipeline(self, brief_id: int):
    """
    Triggered when a content brief is created (Idea Engine webhook -> backend-go -> this task).
    Runs Agent 1 -> Agent 2 -> ScriptLoop inside ai-orchestration.
    Every decision written back to backend-go by agents via their tools.
    """
    try:
        return trigger_ai("production", {"brief_id": brief_id, "stage": "strategic_creative_script"})
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_production_pipeline", str(exc), {"brief_id": brief_id})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=120)
def run_ssml_and_audio(self, brief_id: int):
    """
    Queued after ScriptSelectionLoop exits with a selected script.
    SSMLFormattingAgent formats and submits to ElevenLabs.
    Audio URL written back via PATCH /pipeline/ssml/{id}/audio-ready.
    """
    try:
        return trigger_ai("production", {"brief_id": brief_id, "stage": "ssml_audio"})
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_ssml_and_audio", str(exc), {"brief_id": brief_id})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=300)
def run_avatar_video_production(self, brief_id: int):
    """
    Queued after ElevenLabs audio is ready and video_type == avatar.
    Scenes created via POST /production/videos/{id}/scenes.
    Coherence + render status updated via PATCH calls.
    Posting package created via POST /production/videos/{id}/packages.
    """
    try:
        return trigger_ai("production", {"brief_id": brief_id, "stage": "avatar_video"})
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_avatar_video_production", str(exc), {"brief_id": brief_id})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=300)
def run_faceless_video_production(self, brief_id: int):
    """Queued after audio ready and video_type == faceless."""
    try:
        return trigger_ai("production", {"brief_id": brief_id, "stage": "faceless_video"})
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_faceless_video_production", str(exc), {"brief_id": brief_id})
        raise self.retry(exc=exc)
```

```python
# celery-worker/tasks/analytics_tasks.py

from . import app
from http_client import backend_get, trigger_ai, escalate


@app.task(bind=True, max_retries=2)
def dispatch_due_analytics_collections(self):
    """
    Daily 6am beat task.
    Reads all videos due for analytics collection from backend-go.
    Dispatches one collect_video_analytics task per video per window.
    Never queries the DB directly -- backend-go returns the due list.
    """
    try:
        resp = backend_get("/production/videos/pending-analytics")
        for entry in resp["data"]:
            collect_video_analytics.delay(entry["video_id"], entry["window"])
    except Exception as exc:
        escalate("dispatch_due_analytics_collections", str(exc), {})


@app.task(bind=True, max_retries=2, default_retry_delay=300)
def collect_video_analytics(self, video_id: int, window: str):
    """
    Triggered per video per window (24h, 7d, 30d).
    PerformanceCollectionAgent fetches from platform APIs.
    Results written via POST /production/videos/{id}/analytics.
    After 7d window: queues run_baseline_benchmark.
    After 30d window: queues run_script_promotion if applicable.
    """
    try:
        result = trigger_ai("analytics_collect", {"video_id": video_id, "window": window})
        if window == "7d":
            run_baseline_benchmark.delay(video_id)
        if window == "30d":
            run_script_promotion.delay(video_id)
        return result
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("collect_video_analytics", str(exc), {"video_id": video_id, "window": window})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=60)
def run_baseline_benchmark(self, video_id: int):
    """
    Queued after 7d analytics stored.
    BaselineBenchmarkAgent reads phase1 status and logs a benchmark record.
    Written via POST /production/analytics/benchmark-log.
    """
    try:
        return trigger_ai("analytics_collect", {"video_id": video_id, "stage": "benchmark"})
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_baseline_benchmark", str(exc), {"video_id": video_id})
        raise self.retry(exc=exc)


@app.task(bind=True, max_retries=2, default_retry_delay=60)
def run_script_promotion(self, video_id: int):
    """
    Queued after 30d analytics stored.
    ScriptPromotionAgent checks criteria and promotes if both are met.
    Written via POST /library/scripts/{id}/promote.
    """
    try:
        return trigger_ai("analytics_collect", {"video_id": video_id, "stage": "script_promotion"})
    except Exception as exc:
        if self.request.retries >= self.max_retries:
            escalate("run_script_promotion", str(exc), {"video_id": video_id})
        raise self.retry(exc=exc)
```

---

## 14. GLOBAL REUSABLE FUNCTIONS PER SERVICE

### backend-go — reusable, never duplicated

```
internal/repositories/base_repository.go   EXISTING: paginate(), applyActiveFilter(), applyStatusFilter(), countRows()
internal/helpers/confidence_gate.go        NEW: IsConfidenceFlagged(), PassesScriptThreshold()
internal/helpers/outlier_threshold.go      NEW: ComputeOutlierThreshold(), IsOutlier(), ComputeMultiplier()
internal/helpers/beat_baseline.go          NEW: BeatBaseline(), BaselineMargin(), const NickBaseline
```

### ai-orchestration — reusable, never duplicated

```
http_client.py                             THE CLIENT: backend_get/post/patch/put/delete
langgraph/utils/utils.py                   EXISTING: get_openrouter_llm, create_supervisor_chain, run_agent_node
langgraph/utils/settings_loader.py         NEW: load_agent_settings() -- fetches all via backend_get
langgraph/utils/prompt_builder.py          NEW: build_prompt(template, **values) -- fills {placeholders}
langgraph/utils/confidence.py              NEW: compute_confidence_flag(), build_agent_output()
langgraph/utils/retry.py                   NEW: agent_retry_wrapper() -- 2 retries then escalate
langgraph/utils/adversarial.py             NEW: run_adversarial_audit() -- second LLM review
```

### celery-worker — reusable, never duplicated

```
http_client.py                             THE CLIENT: backend_get/post/patch + trigger_ai + escalate
tasks/__init__.py                          EXISTING: Celery app instance
```

---

## 15. PHASE 1 TASK BREAKDOWN

### Days 1-3

| Task | Service | Deliverable |
|------|---------|-------------|
| Add `agent_settings` + `integration_settings` migrations | backend-go | Settings tables live |
| `settings_handler.go` + `settings_service.go` + routes | backend-go | `/settings/agents/*` + `/settings/integrations/*` |
| Seed all 20 agent system prompts into `agent_settings` | backend-go | Admin updates prompts without code deploy |
| Add `http_client.py` to ai-orchestration | ai-orchestration | All data access goes through it |
| Add `http_client.py` to celery-worker | celery-worker | All data + trigger access goes through it |
| Remove `psycopg2-binary` from both `requirements.txt` | both | No direct DB connections possible |
| `intelligence_handler.go` + routes | backend-go | `/intelligence/accounts` + `/intelligence/reels` + trend signals |
| Load 20-50 competitor accounts via `POST /intelligence/accounts` | admin | Database seeded |
| Configure SociaVault via `PUT /settings/integrations/sociavault` | admin | API key stored |
| `OutlierDetectionAgent` using `http_client.backend_get` for account list | ai-orchestration | Agent reading accounts from backend-go |
| Celery beat `run_outlier_detection` every 6h | celery-worker | First scan running |
| Configure ElevenLabs via `PUT /settings/integrations/elevenlabs` | admin | Voice ID confirmed |
| HeyGen Avatar Shots (April 8th update) tested | engineering | Team experienced with new API |

### Weeks 1-2

| Task | Service | Deliverable |
|------|---------|-------------|
| All intelligence agents + tools using `http_client` | ai-orchestration | Full intelligence pipeline running |
| `pipeline_handler.go` + routes | backend-go | Brief, SSML, attempts, confidence flags endpoints |
| Agent 1 + Agent 2 with dynamic prompts from `agent_settings` | ai-orchestration | Strategic + creative decisions working |
| ScriptValidationAgent + ScriptWriterAgent + loop | ai-orchestration | Script loop operational, 90% threshold enforced |
| SSMLFormattingAgent + ElevenLabs via `http_client` | ai-orchestration | Audio generating end-to-end |
| AvatarVideoProductionAgent + HeyGen + coherence check | ai-orchestration | First avatar video end-to-end |
| `production_handler.go` + routes | backend-go | Videos, scenes, packages, analytics endpoints |
| Manual review dashboard wired to `/production/packages` | backend-go | Nick reviews and approves |
| Long-form scripts seeded (competitor research complete) | Nick + engineering | Long-form active Day 1 |

### Weeks 3-4

| Task | Service | Deliverable |
|------|---------|-------------|
| Analytics celery tasks live | celery-worker | 24h/7d/30d collection per posted video |
| `BaselineBenchmarkAgent` + phase1-status endpoint | all | Beat/missed tracked per video |
| `ScriptPromotionAgent` live | ai-orchestration | Storehouse auto-growing via promotion |
| Week 4 Phase 1 assessment | Nick + engineering | AI consistently beating 38K? |

---

## 16. COMPLETE FILE STRUCTURE

```
theEngineContent/
|
+-- backend-go/
|   +-- config/config.go                           existing
|   +-- http-client/http-client.go                 existing (calls ai-orchestration health)
|   +-- internal/
|   |   +-- dto/content_engine_dto.go              existing (complete DTOs)
|   |   +-- models/
|   |   |   +-- ai_task.go                         existing
|   |   |   +-- audit_log.go                       existing
|   |   |   +-- competitor_account.go              existing
|   |   |   +-- content_intelligence.go            existing (hooks, angles, personas, credentials)
|   |   |   +-- content_pipeline.go                existing (trend, scripts, briefs, attempts)
|   |   |   +-- content_production.go              existing (SSML, video, scenes, packages, analytics)
|   |   |   +-- outlier_reel.go                    existing
|   |   |   +-- platform.go                        existing
|   |   |   +-- user.go                            existing
|   |   |   +-- agent_settings.go                  NEW
|   |   |   +-- integration_settings.go            NEW
|   |   |   +-- voicedna_rules.go                  NEW
|   |   |   +-- posting_time_config.go             NEW
|   |   +-- repositories/
|   |   |   +-- base_repository.go                 existing (paginate, applyActiveFilter)
|   |   |   +-- auth_repository.go                 existing
|   |   |   +-- intelligence_repository.go         existing
|   |   |   +-- content_library_repository.go      existing
|   |   |   +-- pipeline_repository.go             existing
|   |   |   +-- production_repository.go           existing
|   |   |   +-- settings_repository.go             NEW
|   |   +-- services/
|   |   |   +-- auth_service.go                    existing
|   |   |   +-- intelligence_service.go            existing
|   |   |   +-- content_library_service.go         existing
|   |   |   +-- pipeline_service.go                existing
|   |   |   +-- production_service.go              existing
|   |   |   +-- settings_service.go                NEW
|   |   +-- handlers/
|   |   |   +-- auth_handler.go                    existing
|   |   |   +-- intelligence_handler.go            NEW
|   |   |   +-- content_library_handler.go         NEW
|   |   |   +-- pipeline_handler.go                NEW
|   |   |   +-- production_handler.go              NEW
|   |   |   +-- settings_handler.go                NEW
|   |   +-- helpers/
|   |   |   +-- generate_otp.go                    existing
|   |   |   +-- secret_token.go                    existing
|   |   |   +-- send_mails.go                      existing
|   |   |   +-- confidence_gate.go                 NEW
|   |   |   +-- outlier_threshold.go               NEW
|   |   |   +-- beat_baseline.go                   NEW
|   |   +-- routes/routes.go                       existing + all new routes
|   +-- main.go                                    existing + all new handlers wired
|
+-- ai-orchestration/
|   +-- main.py                                    existing + /run endpoint added
|   +-- http_client.py                             NEW -- all backend-go calls
|   +-- requirements.txt                           UPDATED: remove psycopg2-binary
|   +-- langgraph/
|       +-- content_engine_graph.py                REPLACED
|       +-- state/state.py                         REPLACED (ContentEngineState)
|       +-- supervisors/
|       |   +-- intelligence/supervisor.py         NEW
|       |   +-- production/supervisor.py           NEW
|       |   +-- analytics/supervisor.py            NEW
|       +-- agents/
|       |   +-- intelligence/
|       |   |   +-- outlier_detection_agent.py     NEW
|       |   |   +-- transcription_agent.py         NEW
|       |   |   +-- relevance_filter_agent.py      NEW
|       |   |   +-- hook_angle_extraction_agent.py NEW
|       |   |   +-- trend_signal_agent.py          NEW
|       |   +-- production/
|       |   |   +-- strategic_decisions_agent.py   NEW
|       |   |   +-- creative_decisions_agent.py    NEW
|       |   |   +-- script_selection_loop_agent.py NEW
|       |   |   +-- script_validation_agent.py     NEW
|       |   |   +-- script_writer_agent.py         NEW
|       |   |   +-- ssml_formatting_agent.py       NEW
|       |   |   +-- avatar_video_production_agent.py NEW
|       |   |   +-- faceless_video_production_agent.py NEW
|       |   |   +-- posting_package_agent.py       NEW
|       |   +-- analytics/
|       |       +-- performance_collection_agent.py NEW
|       |       +-- baseline_benchmark_agent.py    NEW
|       |       +-- script_promotion_agent.py      NEW
|       +-- tools/
|       |   +-- intelligence/
|       |   |   +-- account_tools.py    backend_get /intelligence/accounts
|       |   |   +-- reel_tools.py       backend_get/patch /intelligence/reels/*
|       |   |   +-- sociavault_tools.py external
|       |   |   +-- youtube_tools.py    external
|       |   |   +-- tiktok_tools.py     external
|       |   |   +-- transcription_tools.py external (Whisper)
|       |   |   +-- hook_tools.py       backend_get/post /library/hooks
|       |   |   +-- angle_tools.py      backend_get/post /library/angles
|       |   |   +-- credential_tools.py backend_get /library/credentials
|       |   |   +-- trend_tools.py      backend_post /intelligence/trend-signals + external
|       |   +-- production/
|       |   |   +-- brief_tools.py      backend_get/patch /pipeline/briefs/*
|       |   |   +-- platform_tools.py   backend_get /platforms
|       |   |   +-- persona_tools.py    backend_get /library/personas/*
|       |   |   +-- script_tools.py     backend_get/post /library/scripts/*
|       |   |   +-- attempt_tools.py    backend_get/post /pipeline/briefs/{id}/attempts
|       |   |   +-- ssml_tools.py       backend_get/post/patch /pipeline/ssml/*
|       |   |   +-- elevenlabs_tools.py external
|       |   |   +-- video_tools.py      backend_get/post/patch /production/videos/*
|       |   |   +-- scene_tools.py      backend_post/patch /production/scenes/*
|       |   |   +-- heygen_tools.py     external
|       |   |   +-- kling_tools.py      external
|       |   |   +-- package_tools.py    backend_post /production/videos/{id}/packages
|       |   |   +-- settings_tools.py   backend_get /settings/integrations/* + voicedna
|       |   +-- analytics/
|       |       +-- analytics_tools.py  backend_get/post /production/videos/{id}/analytics
|       |       +-- instagram_tools.py  external
|       |       +-- youtube_tools.py    external
|       |       +-- tiktok_tools.py     external
|       |       +-- promotion_tools.py  backend_post /library/scripts/{id}/promote
|       +-- utils/
|           +-- utils.py                existing
|           +-- settings_loader.py      NEW: backend_get("/settings/agents")
|           +-- prompt_builder.py       NEW: build_prompt(template, **values)
|           +-- confidence.py           NEW: compute_confidence_flag, build_agent_output
|           +-- retry.py               NEW: agent_retry_wrapper
|           +-- adversarial.py         NEW: run_adversarial_audit
|
+-- celery-worker/
    +-- http_client.py                             NEW -- all backend-go + ai calls
    +-- celeryconfig.py                            existing + beat_schedule
    +-- requirements.txt                           UPDATED: remove psycopg2-binary
    +-- tasks/
        +-- __init__.py                            existing (Celery app)
        +-- intelligence_tasks.py                  NEW
        +-- production_tasks.py                    NEW
        +-- analytics_tasks.py                     NEW
```

---

## THE RULE ONE MORE TIME

```
backend-go    = the only service with a database connection
               = the source of truth for ALL data
               = exposes REST API for everything

ai-orchestration = HTTP client of backend-go
                 = reads data before building agent prompts
                 = writes data after every agent decision
                 = calls external AI APIs (ElevenLabs, HeyGen, Kling, Whisper)
                 = never connects to postgres

celery-worker    = HTTP client of backend-go + trigger of ai-orchestration
                 = reads data to dispatch tasks
                 = never connects to postgres
                 = never runs LLM logic
```

*Prepared by Cyberify | April 2026 | Repo: github.com/MuhammadAslam635/theEngineContent*
*Based on Content_Engine_v2_Nick.docx -- Nick Perry / 7 Figure Cartel*
