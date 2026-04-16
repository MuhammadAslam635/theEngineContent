# cont-gen Microservices — Kubernetes Setup Guide

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│                    PUBLIC (LoadBalancer)                      │
│                                                              │
│  :8000 Kong Proxy   :1337 Konga   :5555 Flower  :5050 pgAdmin│
└────────┬─────────────────────────────────────────────────────┘
         │ All API traffic
         ▼
┌────────────────────┐
│   Kong API Gateway  │  JWT validation plugin, rate-limiting
│     (port 8000)    │
└────────┬───────────┘
         │ routes /api/*
         ▼
┌────────────────────┐        ┌─────────────────────────────┐
│   Backend Go       │─HTTP──▶│   AI Orchestration          │
│   port: 9001       │internal│   port: 8086                │
│   Auth: JWT        │  only  │   Auth: x-user-id header    │
└────────┬───────────┘        │   (NOT reachable externally)│
         │                    └─────────────────────────────┘
         │ enqueue jobs
         ▼
┌────────────────────┐        ┌─────────────────────────────┐
│   Redis (broker)   │◀──────▶│   Celery Workers            │
│   port: 6379       │        │   Celery Beat (scheduler)   │
└────────────────────┘        └─────────────────────────────┘
         │
         ▼
┌────────────────────┐
│   PostgreSQL        │  ◀── all services read/write
│   port: 5432        │
└────────────────────┘
```

## Port Reference

| Service           | Internal Port | External (NodePort) | Access             |
|-------------------|--------------|---------------------|--------------------|
| Backend Go        | 9001         | —                   | via Kong :8000     |
| AI Orchestration  | 8086         | —                   | internal only      |
| Kong Proxy        | 8000         | 30800               | `localhost:8000`   |
| Kong Admin        | 8001         | —                   | internal only      |
| Konga             | 1337         | 31337               | `localhost:1337`   |
| Flower            | 5555         | 31555               | `localhost:5555`   |
| pgAdmin           | 80           | 31050               | `localhost:5050`   |
| PostgreSQL        | 5432         | —                   | internal only      |
| Redis             | 6379         | —                   | internal only      |

---

## Prerequisites

```bash
# 1. Install tools
brew install kubectl helm docker

# 2. For local dev — choose ONE:
#    Option A: Docker Desktop with Kubernetes enabled (Settings → Kubernetes)
#    Option B: minikube
brew install minikube
minikube start --driver=docker --memory=4096 --cpus=4

# 3. Verify cluster is running
kubectl cluster-info
kubectl get nodes
```

---

## Directory Structure

```
k8s/
├── namespace/
│   └── namespace.yml
├── configmaps/
│   └── configmaps.yml
├── secrets/
│   └── secrets.yml
├── pvcs/
│   └── pvcs.yml
├── deployments/
│   ├── infra-deployments.yml        # postgres, redis
│   ├── backend-go-deployment.yml
│   ├── ai-orchestration-deployment.yml
│   ├── celery-deployments.yml       # worker, beat, flower
│   └── admin-deployments.yml        # kong, konga, pgadmin
├── services/
│   └── services.yml
├── network-policies/
│   └── network-policies.yml
├── hpa/
│   └── hpa.yml
└── ingress/
    └── ingress.yml
```

---

## Step-by-Step Deployment

### Step 1 — Update Secrets

Before deploying, encode your real values:

```bash
# Encode a value
echo -n "your_actual_password" | base64

# Decode to verify
echo "dGVzdA==" | base64 --decode
```

Edit `k8s/secrets/secrets.yml` and replace the base64 values for:
- `POSTGRES_PASSWORD`
- `JWT_SECRET`
- `PGADMIN_DEFAULT_PASSWORD`

### Step 2 — Update Image Names

Replace placeholder image names in deployments with your actual images:

```bash
# In backend-go-deployment.yml
image: contgen/backend-go:latest          # → your-registry/backend-go:tag

# In ai-orchestration-deployment.yml
image: contgen/ai-orchestration:latest    # → your-registry/ai-orchestration:tag

# In celery-deployments.yml
image: contgen/celery-worker:latest       # → your-registry/celery-worker:tag
```

### Step 3 — Apply All Manifests

```bash
# Create namespace first
kubectl apply -f k8s/namespace/namespace.yml

# Infrastructure (order matters)
kubectl apply -f k8s/secrets/secrets.yml
kubectl apply -f k8s/configmaps/configmaps.yml
kubectl apply -f k8s/pvcs/pvcs.yml

# Deploy infrastructure services first, wait for them to be ready
kubectl apply -f k8s/deployments/infra-deployments.yml
kubectl -n cont-gen wait --for=condition=ready pod -l app=postgres --timeout=120s
kubectl -n cont-gen wait --for=condition=ready pod -l app=redis --timeout=60s

# Create a kong database in postgres
kubectl -n cont-gen exec -it deploy/postgres -- \
  psql -U postgres -c "CREATE DATABASE kong;"
kubectl -n cont-gen exec -it deploy/postgres -- \
  psql -U postgres -c "CREATE DATABASE konga;"

# Deploy application services
kubectl apply -f k8s/deployments/backend-go-deployment.yml
kubectl apply -f k8s/deployments/ai-orchestration-deployment.yml
kubectl apply -f k8s/deployments/celery-deployments.yml
kubectl apply -f k8s/deployments/admin-deployments.yml

# Services and networking
kubectl apply -f k8s/services/services.yml
kubectl apply -f k8s/network-policies/network-policies.yml

# Autoscaling
kubectl apply -f k8s/hpa/hpa.yml

# Ingress (optional — skip if using pure Kong routing)
kubectl apply -f k8s/ingress/ingress.yml
```

### Step 4 — Verify Everything is Running

```bash
# Check all pods
kubectl -n cont-gen get pods

# Check services
kubectl -n cont-gen get services

# Check network policies
kubectl -n cont-gen get networkpolicies

# Watch pod startup
kubectl -n cont-gen get pods --watch
```

Expected output — all pods should be `Running`:
```
NAME                               READY   STATUS    RESTARTS
postgres-xxx                       1/1     Running   0
redis-xxx                          1/1     Running   0
backend-go-xxx                     1/1     Running   0
backend-go-yyy                     1/1     Running   0
ai-orchestration-xxx               1/1     Running   0
ai-orchestration-yyy               1/1     Running   0
celery-worker-xxx                  1/1     Running   0
celery-worker-yyy                  1/1     Running   0
celery-beat-xxx                    1/1     Running   0
flower-xxx                         1/1     Running   0
kong-xxx                           1/1     Running   0
konga-xxx                          1/1     Running   0
pgadmin-xxx                        1/1     Running   0
```

---

## Accessing Services (Local Dev)

### Option A — NodePort (minikube / Docker Desktop)

```bash
# Get minikube IP
minikube ip   # e.g. 192.168.49.2

# Or with Docker Desktop use 127.0.0.1
```

| UI / API         | URL                          |
|------------------|------------------------------|
| Kong API Gateway | `http://localhost:8000`      |
| Konga Admin UI   | `http://localhost:1337`      |
| Flower           | `http://localhost:5555`      |
| pgAdmin          | `http://localhost:5050`      |

### Option B — kubectl port-forward (always works)

```bash
# Kong proxy
kubectl -n cont-gen port-forward svc/kong-proxy-service 8000:8000

# Konga
kubectl -n cont-gen port-forward svc/konga-service 1337:1337

# Flower
kubectl -n cont-gen port-forward svc/flower-service 5555:5555

# pgAdmin
kubectl -n cont-gen port-forward svc/pgadmin-service 5050:5050
```

---

## Kong Setup (via Konga UI)

1. Open `http://localhost:1337`
2. Create admin user on first run
3. Connect to Kong Admin API: `http://kong-admin-service:8001`
4. Add a **Service** pointing to Backend Go:
   - Name: `backend-go`
   - URL: `http://backend-go-service:9001`
5. Add a **Route**:
   - Paths: `/api`
   - Strip Path: enabled
6. Add JWT Plugin to the route:
   - Plugin: `JWT`
   - Config: use your `JWT_SECRET`

> **Important:** Do NOT add any route for AI Orchestration in Kong.
> It is internal-only and protected by NetworkPolicy.

---

## Authentication Flow

### Frontend → Backend Go (JWT)

```
Frontend
  │
  │  POST /api/login   →  Backend Go returns JWT token
  │
  │  GET /api/...
  │  Authorization: Bearer <jwt_token>
  │
  ▼
Kong validates JWT → forwards to Backend Go
```

### Backend Go → AI Orchestration (internal x-user-id)

```go
// Example Go HTTP client call to AI Orchestration
// AI_ORCHESTRATION_URL is injected via env var:
// http://ai-orchestration-service:8086

func callAIOrchestration(userID string, payload interface{}) {
    client := &http.Client{Timeout: 30 * time.Second}
    
    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST",
        os.Getenv("AI_ORCHESTRATION_URL")+"/generate",
        bytes.NewBuffer(body),
    )
    
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("x-user-id", userID)  // extracted from JWT claims
    
    resp, err := client.Do(req)
    // handle response...
}
```

### AI Orchestration — Validate x-user-id

```python
# In your AI Orchestration middleware (FastAPI example)
from fastapi import Request, HTTPException

async def verify_internal_caller(request: Request):
    user_id = request.headers.get("x-user-id")
    if not user_id:
        raise HTTPException(status_code=401, detail="Missing x-user-id")
    return user_id
```

---

## pgAdmin Setup

1. Open `http://localhost:5050`
2. Login: `admin@contgen.com` / `admin123` (from secret)
3. Add Server:
   - Host: `postgres-service`
   - Port: `5432`
   - Database: `contgen`
   - Username: `postgres`
   - Password: from your secret

---

## Celery Task Structure

Your Celery Python project should look like:

```
celery-worker/
├── tasks/
│   ├── __init__.py
│   └── your_tasks.py
├── celeryconfig.py
└── Dockerfile
```

```python
# celeryconfig.py
broker_url = os.environ["CELERY_BROKER_URL"]
result_backend = os.environ["CELERY_RESULT_BACKEND"]
task_serializer = "json"
result_serializer = "json"
accept_content = ["json"]
timezone = "UTC"
```

```python
# tasks/your_tasks.py
from celery import Celery
import os

app = Celery("tasks")
app.config_from_object("celeryconfig")

@app.task(bind=True, max_retries=3)
def process_job(self, job_id: str, payload: dict):
    try:
        # do work
        pass
    except Exception as exc:
        raise self.retry(exc=exc, countdown=60)
```

Backend Go enqueues tasks via Celery's Redis broker directly or via an HTTP endpoint in the celery worker service.

---

## Network Policy Summary

| Source              | Destination         | Port | Allowed |
|---------------------|---------------------|------|---------|
| External / Kong     | backend-go          | 9001 | ✅ Yes  |
| backend-go          | ai-orchestration    | 8086 | ✅ Yes  |
| Kong / Konga / ANY  | ai-orchestration    | 8086 | ❌ No   |
| All tiers           | postgres            | 5432 | ✅ Yes  |
| celery / flower     | redis               | 6379 | ✅ Yes  |
| backend-go          | redis               | 6379 | ✅ Yes  |
| External            | flower              | 5555 | ✅ Yes  |
| External            | konga               | 1337 | ✅ Yes  |
| External            | pgadmin             | 80   | ✅ Yes  |
| konga               | kong-admin          | 8001 | ✅ Yes  |

---

## Useful kubectl Commands

```bash
# View logs
kubectl -n cont-gen logs -f deploy/backend-go
kubectl -n cont-gen logs -f deploy/ai-orchestration
kubectl -n cont-gen logs -f deploy/celery-worker
kubectl -n cont-gen logs -f deploy/flower

# Shell into a pod
kubectl -n cont-gen exec -it deploy/backend-go -- sh
kubectl -n cont-gen exec -it deploy/postgres -- psql -U postgres -d contgen

# Restart a deployment (rolling restart)
kubectl -n cont-gen rollout restart deploy/backend-go

# Scale manually
kubectl -n cont-gen scale deploy/celery-worker --replicas=4

# Check HPA status
kubectl -n cont-gen get hpa

# Check network policies
kubectl -n cont-gen describe networkpolicy allow-backend-go-to-ai-orchestration

# Test network isolation — this should FAIL (connection refused):
kubectl -n cont-gen run test-pod --image=alpine --rm -it -- \
  wget -T5 http://ai-orchestration-service:8086/health

# Delete everything in namespace
kubectl delete namespace cont-gen
```

---

## Apply All at Once (One-liner)

```bash
kubectl apply -f k8s/namespace/namespace.yml && \
kubectl apply -f k8s/secrets/ && \
kubectl apply -f k8s/configmaps/ && \
kubectl apply -f k8s/pvcs/ && \
kubectl apply -f k8s/deployments/infra-deployments.yml && \
sleep 30 && \
kubectl -n cont-gen exec -it deploy/postgres -- psql -U postgres -c "CREATE DATABASE IF NOT EXISTS kong;" && \
kubectl -n cont-gen exec -it deploy/postgres -- psql -U postgres -c "CREATE DATABASE IF NOT EXISTS konga;" && \
kubectl apply -f k8s/deployments/ && \
kubectl apply -f k8s/services/ && \
kubectl apply -f k8s/network-policies/ && \
kubectl apply -f k8s/hpa/
```

---

## Environment Variables Reference

### Backend Go

| Variable              | Source     | Example                                |
|-----------------------|------------|----------------------------------------|
| `APP_PORT`            | ConfigMap  | `9001`                                 |
| `JWT_SECRET`          | Secret     | `super_secret_key`                     |
| `AI_ORCHESTRATION_URL`| ConfigMap  | `http://ai-orchestration-service:8086` |
| `DB_HOST`             | ConfigMap  | `postgres-service`                     |
| `DB_PORT`             | ConfigMap  | `5432`                                 |
| `DB_NAME`             | ConfigMap  | `contgen`                              |
| `DB_USER`             | Secret     | `postgres`                             |
| `DB_PASSWORD`         | Secret     | `your_password`                        |

### AI Orchestration

| Variable     | Source    | Example            |
|--------------|-----------|--------------------|
| `APP_PORT`   | ConfigMap | `8086`             |
| `DB_HOST`    | ConfigMap | `postgres-service` |
| `DB_USER`    | Secret    | `postgres`         |
| `DB_PASSWORD`| Secret    | `your_password`    |

### Celery Worker

| Variable               | Source    | Example                       |
|------------------------|-----------|-------------------------------|
| `CELERY_BROKER_URL`    | ConfigMap | `redis://redis-service:6379/0`|
| `CELERY_RESULT_BACKEND`| ConfigMap | `redis://redis-service:6379/0`|

---

## Troubleshooting

**Pod stuck in `Pending`:**
```bash
kubectl -n cont-gen describe pod <pod-name>
# Usually: PVC not bound, image pull error, or resource limits
```

**Pod in `CrashLoopBackOff`:**
```bash
kubectl -n cont-gen logs <pod-name> --previous
```

**Backend Go can't reach AI Orchestration:**
```bash
# Check network policy is applied
kubectl -n cont-gen get networkpolicies
# Check service exists
kubectl -n cont-gen get svc ai-orchestration-service
# Check env var in backend-go pod
kubectl -n cont-gen exec deploy/backend-go -- env | grep AI_ORCHESTRATION
```

**Kong migration fails:**
```bash
# Ensure kong database was created in postgres
kubectl -n cont-gen exec -it deploy/postgres -- psql -U postgres -l
# Re-run migration manually
kubectl -n cont-gen exec deploy/kong -- kong migrations bootstrap
```

**Konga not connecting to Kong:**
```bash
# Verify Kong admin is reachable from Konga pod
kubectl -n cont-gen exec -it deploy/konga -- wget -O- http://kong-admin-service:8001/status
```
