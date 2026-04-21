# Tech Stack & Build

## Services

| Service | Language | Framework | Port |
|---|---|---|---|
| backend-go | Go 1.25 | Gin + GORM | 9001 |
| ai-orchestration | Python 3.11 | FastAPI + LangGraph | 8086 |
| celery-worker | Python 3.11 | Celery (Redis broker) | — |

## Infrastructure

- **Database**: PostgreSQL 16 (database name: `contgen`)
- **Cache/Broker**: Redis 7
- **API Gateway**: Kong (port 8000) with JWT plugin, managed via Konga UI
- **Orchestration**: Kubernetes (namespace: `cont-gen`)
- **Container Runtime**: Docker
- **LLM Provider**: OpenRouter (default model: `gpt-4o`, search: `gpt-4o-search-preview`)

## Key Libraries

### Go Backend
- `gin-gonic/gin` — HTTP router and middleware
- `gorm.io/gorm` + `gorm.io/driver/postgres` — ORM with AutoMigrate
- `golang-jwt/jwt/v5` — JWT authentication
- `golang.org/x/crypto` — bcrypt password hashing
- `google/uuid` — UUID generation

### Python AI Orchestration
- `fastapi` + `uvicorn` — async HTTP API
- `langgraph` — multi-agent graph orchestration (supervisor/agent pattern)
- `langchain-openai` — LLM integration via OpenRouter
- `psycopg2-binary` — direct PostgreSQL access

### Celery Worker
- `celery[redis]` — distributed task queue
- `psycopg2-binary` — PostgreSQL access
- `requests` — HTTP calls

## Common Commands

### Build Docker Images
```powershell
# Build all app services (backend-go, ai-orchestration, celery-worker)
./build_images.ps1

# Build specific service
./build_images.ps1 -BackendGoOnly
./build_images.ps1 -AiOrchestrationOnly
./build_images.ps1 -CeleryWorkerOnly

# Build everything including infra images
./build_images.ps1 -All

# Build without cache
./build_images.ps1 -NoCache
```

### Deploy to Kubernetes
```powershell
# Apply all K8s manifests in correct order
./apply.ps1
```

### Run Database Migrations
```powershell
./migrate_db.ps1
```

### Go Backend (local dev)
```bash
cd backend-go
go build -o main .
go run .
```

### AI Orchestration (local dev)
```bash
cd ai-orchestration
pip install -r requirements.txt
python main.py
```

### Useful kubectl Commands
```bash
kubectl -n cont-gen get pods
kubectl -n cont-gen logs -f deploy/backend-go
kubectl -n cont-gen logs -f deploy/ai-orchestration
kubectl -n cont-gen exec -it deploy/postgres -- psql -U postgres -d contgen
```
