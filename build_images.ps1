# Build script for cont-gen microservices
# Run this script from the project root: d:\theEngine\

param (
    [switch]$BackendGoOnly,
    [switch]$AiOrchestrationOnly,
    [switch]$CeleryWorkerOnly,
    [switch]$PostgresOnly,
    [switch]$RedisOnly,
    [switch]$PgAdminOnly,
    [switch]$KongOnly,
    [switch]$KongaOnly,
    [switch]$All,
    [switch]$NoCache,
    [string]$Namespace = "cont-gen",
    [string]$Registry = "contgen"
)

$ErrorActionPreference = "Stop"

# Ensure we are in the project root
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
Set-Location $ScriptDir
Write-Host "Working directory: $PWD" -ForegroundColor Gray

# Generate a dynamic tag for this build
$Timestamp = Get-Date -Format "yyyyMMddHHmmss"
Write-Host "Build Timestamp: $Timestamp" -ForegroundColor Yellow

function Build-Image {
    param (
        [string]$ContextPath,
        [string]$DockerfilePath,
        [string]$ImageName
    )
    
    $BuildCmd = "docker build -t $ImageName"
    if ($NoCache) {
        $BuildCmd += " --no-cache"
    }
    if ($DockerfilePath) {
        $BuildCmd += " -f $DockerfilePath"
    }
    $BuildCmd += " $ContextPath"
    
    Write-Host "Executing: $BuildCmd" -ForegroundColor Cyan
    Invoke-Expression $BuildCmd
    
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to build $ImageName"
        exit 1
    }

    # Also tag as :latest so manifests with :latest always work
    $LatestTag = ($ImageName -replace ":.*$", ":latest")
    if ($LatestTag -ne $ImageName) {
        docker tag $ImageName $LatestTag
        Write-Host "Tagged $ImageName as $LatestTag" -ForegroundColor Gray
    }

    Write-Host "Successfully built $ImageName" -ForegroundColor Green
    Write-Host "----------------------------------------"
}

function Load-Image-To-Cluster {
    param (
        [string]$ImageName
    )
    
    # Discover Kind/Docker Desktop node containers
    $NodeNames = docker ps --format "{{.Names}}\t{{.Image}}" | `
        Where-Object { $_ -match "kindest/node" } | `
        ForEach-Object { ($_ -split "\t")[0] }

    if (-not $NodeNames -or $NodeNames.Count -eq 0) {
        Write-Host "No Kind/Docker Desktop nodes detected. Skipping manual image load." -ForegroundColor Gray
        return
    }

    $SafeName = ($ImageName -replace '[^a-zA-Z0-9._-]', '_')
    $TarFile = "image-load-$($SafeName)-$Timestamp.tar"
    Write-Host "Saving $ImageName to $TarFile..." -ForegroundColor Cyan
    docker save -o $TarFile $ImageName

    foreach ($Node in $NodeNames) {
        try {
            Write-Host "Loading image into node: $Node" -ForegroundColor Cyan
            docker cp $TarFile "$($Node):/$TarFile"
            docker exec $Node ctr -n k8s.io images import "/$TarFile"
            docker exec $Node rm "/$TarFile" | Out-Null
            Write-Host "Loaded into $Node successfully." -ForegroundColor Green
        } catch {
            Write-Warning "Failed to load image into $($Node): $($_.Exception.Message)"
        }
    }

    Remove-Item $TarFile -ErrorAction SilentlyContinue
}

# --- Build Logic ---

$Tags = @{}
$StandardImages = @{
    "postgres" = "postgres:16-alpine"
    "redis"    = "redis:7-alpine"
    "pgadmin"  = "dpage/pgadmin4:8"
    "kong"     = "kong:3.5"
    "konga"    = "pantsel/konga:latest"
}

if ($All) {
    $BackendGoOnly = $AiOrchestrationOnly = $CeleryWorkerOnly = $PostgresOnly = $RedisOnly = $PgAdminOnly = $KongOnly = $KongaOnly = $true
}

# If no specific switches provided, default to core app services
if (-not ($BackendGoOnly -or $AiOrchestrationOnly -or $CeleryWorkerOnly -or $PostgresOnly -or $RedisOnly -or $PgAdminOnly -or $KongOnly -or $KongaOnly)) {
    $BackendGoOnly = $AiOrchestrationOnly = $CeleryWorkerOnly = $true
}

# Custom App Images
if ($BackendGoOnly) { $Tags["backend-go"] = "$Registry/backend-go:v$Timestamp" }
if ($AiOrchestrationOnly) { $Tags["ai-orchestration"] = "$Registry/ai-orchestration:v$Timestamp" }
if ($CeleryWorkerOnly) { $Tags["celery-worker"] = "$Registry/celery-worker:v$Timestamp" }

# Standard Infra Images (Tagged with Registry)
if ($PostgresOnly) { $Tags["postgres"] = "$Registry/postgres:v$Timestamp" }
if ($RedisOnly) { $Tags["redis"] = "$Registry/redis:v$Timestamp" }
if ($PgAdminOnly) { $Tags["pgadmin"] = "$Registry/pgadmin:v$Timestamp" }
if ($KongOnly) { $Tags["kong"] = "$Registry/kong:v$Timestamp" }
if ($KongaOnly) { $Tags["konga"] = "$Registry/konga:v$Timestamp" }

foreach ($Service in $Tags.Keys) {
    $ImageTag = $Tags[$Service]
    Write-Host "Processing $Service..." -ForegroundColor Yellow

    if ($StandardImages.ContainsKey($Service)) {
        # Pull standard image and re-tag
        $SourceImage = $StandardImages[$Service]
        Write-Host "Pulling standard image $SourceImage..." -ForegroundColor Cyan
        docker pull $SourceImage
        docker tag $SourceImage $ImageTag
        
        # Also tag as :latest
        $LatestTag = ($ImageTag -replace ":.*$", ":latest")
        docker tag $SourceImage $LatestTag
        Write-Host "Re-tagged $SourceImage as $ImageTag and $LatestTag" -ForegroundColor Gray
    } else {
        # Build custom app image
        Build-Image -ContextPath "./$Service" -ImageName $ImageTag
    }

    Load-Image-To-Cluster -ImageName $ImageTag

    # Update K8s deployment
    Write-Host "Updating K8s deployment for $Service..." -ForegroundColor Cyan
    try {
        kubectl set image "deployment/$Service" "$Service=$ImageTag" -n $Namespace
        Write-Host "Deployment $Service updated successfully." -ForegroundColor Green
    } catch {
        Write-Warning "Could not update deployment $Service. It might not exist in namespace '$Namespace'."
    }
}

Write-Host "All requested tasks completed successfully!" -ForegroundColor Green
