# Apply script for cont-gen microservices Kubernetes manifests
# Run from the project root: d:\theEngine\

$ScriptDir = $PSScriptRoot
$Namespace = "cont-gen"

# Check for Kubernetes connectivity
Write-Host "Checking Kubernetes cluster connectivity..." -ForegroundColor Cyan
try {
    kubectl cluster-info | Out-Null
} catch {
    Write-Error "Cannot connect to Kubernetes cluster. Please ensure your context is set correctly."
    exit 1
}

if ($LASTEXITCODE -ne 0) {
    Write-Error "Cannot connect to Kubernetes cluster. Please ensure your context is set correctly."
    exit 1
}

Write-Host "Cluster connected. Proceeding..." -ForegroundColor Green

# 1. Namespace
Write-Host "Applying Namespace..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir/k8s/namespace.yml"

# 2. Config & Secrets
Write-Host "Applying Secrets and ConfigMaps..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir/k8s/secrets.yml"
kubectl apply -f "$ScriptDir/k8s/configmaps.yml"
kubectl apply -f "$ScriptDir/k8s/pvcs.yml"

# 3. Infrastructure (Postgres, Redis)
Write-Host "Applying Infrastructure (Postgres, Redis)..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir/k8s/infra-deployments.yml"

Write-Host "Waiting for Postgres to be ready..." -ForegroundColor Cyan
kubectl wait --for=condition=ready pod -l app=postgres -n $Namespace --timeout=120s

Write-Host "Waiting for Redis to be ready..." -ForegroundColor Cyan
kubectl wait --for=condition=ready pod -l app=redis -n $Namespace --timeout=60s

# 4. Initialize Kong Databases (as per SETUP.md)
Write-Host "Initializing databases and fixing auth for old clients..." -ForegroundColor Yellow
try {
    # Fix for old clients (Konga) using Postgres 14+
    kubectl -n $Namespace exec deploy/postgres -- psql -U postgres -c "ALTER USER postgres WITH PASSWORD 'postgres_secret_password';"
    
    kubectl -n $Namespace exec deploy/postgres -- psql -U postgres -c "CREATE DATABASE kong;"
    kubectl -n $Namespace exec deploy/postgres -- psql -U postgres -c "CREATE DATABASE konga;"
} catch {
    Write-Warning "Database initialization might have failed or databases already exist."
}

# 5. Application Services
Write-Host "Applying Application Services..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir/k8s/backend-go-deployment.yml"
kubectl apply -f "$ScriptDir/k8s/ai-orchestration-deployment.yml"
kubectl apply -f "$ScriptDir/k8s/celery-deployments.yml"
kubectl apply -f "$ScriptDir/k8s/admin-deployments.yml"

# Restart to pick up DB changes
Write-Host "Restarting gateway services..." -ForegroundColor Cyan
kubectl -n $Namespace rollout restart deploy/kong
kubectl -n $Namespace rollout restart deploy/konga

# 6. Networking & HPA
Write-Host "Applying Services, Network Policies, and HPA..." -ForegroundColor Yellow
kubectl apply -f "$ScriptDir/k8s/services.yml"
kubectl apply -f "$ScriptDir/k8s/network-policies.yml"
kubectl apply -f "$ScriptDir/k8s/hpa.yml"
kubectl apply -f "$ScriptDir/k8s/ingress.yml"

Write-Host "All manifests applied! Check status with: kubectl get pods -n $Namespace" -ForegroundColor Green
