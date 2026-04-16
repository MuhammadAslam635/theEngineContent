#!/usr/bin/env pwsh 
 <# 
 .SYNOPSIS 
     Database Migration Script for cont-gen Platform 
 .DESCRIPTION 
     Runs pending SQL migrations in the migrations/ folder against the running postgres pod. 
 .PARAMETER Namespace 
     Kubernetes namespace (default: cont-gen) 
 .PARAMETER Database 
     Database name (default: contgen) 
 .PARAMETER Fresh 
     Drop all tables and run all migrations from scratch (DANGEROUS!) 
 .EXAMPLE 
     .\migrate_db.ps1 
     .\migrate_db.ps1 -Fresh 
 #> 
 
 param( 
     [string]$Namespace = "cont-gen", 
     [string]$Database  = "contgen", 
     [switch]$Fresh 
 ) 
 
 $ErrorActionPreference = "Stop" 
 # Set location to the script directory
 Set-Location $PSScriptRoot 
 
 Write-Host "============================================" -ForegroundColor Cyan 
 Write-Host "  cont-gen Database Migration Script" -ForegroundColor Cyan 
 Write-Host "============================================" -ForegroundColor Cyan 
 Write-Host "Namespace: $Namespace" -ForegroundColor Yellow 
 Write-Host "Database:  $Database"  -ForegroundColor Yellow 
 Write-Host "" 
 
 # --------------------------------------------------------------------------- 
 # Find PostgreSQL pod 
 # --------------------------------------------------------------------------- 
 Write-Host "Finding PostgreSQL pod..." -ForegroundColor Yellow 
 $postgresPod = kubectl get pods -n $Namespace -l app=postgres -o jsonpath='{.items[0].metadata.name}' 2>$null 
 
 if (-not $postgresPod) { 
     Write-Host "ERROR: PostgreSQL pod not found in namespace '$Namespace'" -ForegroundColor Red 
     exit 1 
 } 
 
 $podStatus = kubectl get pod $postgresPod -n $Namespace -o jsonpath='{.status.phase}' 2>$null 
 if ($podStatus -ne "Running") { 
     Write-Host "ERROR: PostgreSQL pod is not running (status: $podStatus)" -ForegroundColor Red 
     exit 1 
 } 
 
 Write-Host "Pod: $postgresPod  [Running]" -ForegroundColor Green 
 Write-Host "" 
 
 # --------------------------------------------------------------------------- 
 # Fresh: drop all tables 
 # --------------------------------------------------------------------------- 
 if ($Fresh) { 
     Write-Host "WARNING: FRESH MIGRATION - THIS WILL DROP ALL TABLES AND DATA!" -ForegroundColor Red 
     $confirmation = Read-Host "Type YES to confirm" 
     if ($confirmation -ne "YES") { 
         Write-Host "Cancelled." -ForegroundColor Yellow 
         exit 0 
     } 
 
     Write-Host "Dropping all tables..." -ForegroundColor Yellow 
 
     # Write drop SQL to a temp file to avoid PowerShell string escaping issues with $$ 
     $dropFile = Join-Path $PSScriptRoot "drop_all_temp.sql" 
     $dropLines = @( 
         "DO", 
         "$$", 
         "DECLARE r RECORD;", 
         "BEGIN", 
         "  FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP", 
         "    EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';", 
         "  END LOOP;", 
         "END", 
         "$$;" 
     ) 
     $dropLines | Set-Content -Path $dropFile -Encoding UTF8 
 
     kubectl cp "$dropFile" "${postgresPod}:/tmp/drop_all.sql" -n $Namespace -c postgres
    kubectl exec -n $Namespace $postgresPod -c postgres -- sh -c "PGPASSWORD=postgres psql -U postgres -d $Database -f /tmp/drop_all.sql 2>&1" 
     kubectl exec -n $Namespace $postgresPod -- rm -f /tmp/drop_all.sql 
     Remove-Item $dropFile -ErrorAction SilentlyContinue 
 
     Write-Host "All tables dropped." -ForegroundColor Green 
     Write-Host "" 
 } 
 
 # --------------------------------------------------------------------------- 
 # Ensure schema_migrations table exists 
 # --------------------------------------------------------------------------- 
 Write-Host "Ensuring schema_migrations table exists..." -ForegroundColor Yellow 
 
 $createMigrationsTable = "CREATE TABLE IF NOT EXISTS schema_migrations (id SERIAL PRIMARY KEY, version VARCHAR(255) UNIQUE NOT NULL, description TEXT, applied_at TIMESTAMPTZ DEFAULT NOW());" 
 kubectl exec -n $Namespace $postgresPod -c postgres -- sh -c "PGPASSWORD=postgres psql -U postgres -d $Database -c `"$createMigrationsTable`" 2>&1" | Out-Null 
 
 Write-Host "schema_migrations table ready." -ForegroundColor Green 
 Write-Host "" 
 
 # --------------------------------------------------------------------------- 
 # Scan migration files 
 # --------------------------------------------------------------------------- 
 Write-Host "============================================" -ForegroundColor Cyan 
 Write-Host "  Scanning migration files..." -ForegroundColor Cyan 
 Write-Host "============================================" -ForegroundColor Cyan 
 
 $migrationsPath = Join-Path $PSScriptRoot "migrations" 
 
 if (-not (Test-Path $migrationsPath)) { 
     Write-Host "ERROR: Migrations folder not found: $migrationsPath" -ForegroundColor Red 
     exit 1 
 } 
 
 $migrationFiles = Get-ChildItem -Path $migrationsPath -Filter "*.sql" | Sort-Object Name 
 
 if ($migrationFiles.Count -eq 0) { 
     Write-Host "ERROR: No .sql files found in $migrationsPath" -ForegroundColor Red 
     exit 1 
 } 
 
 Write-Host "Found $($migrationFiles.Count) migration file(s):" -ForegroundColor Green 
 foreach ($f in $migrationFiles) { 
     Write-Host "  - $($f.Name)" -ForegroundColor Gray 
 } 
 Write-Host "" 
 
 # --------------------------------------------------------------------------- 
 # Check which migrations are already applied 
 # --------------------------------------------------------------------------- 
 Write-Host "Checking applied migrations..." -ForegroundColor Yellow 
 
 $appliedRaw = kubectl exec -n $Namespace $postgresPod -c postgres -- sh -c "PGPASSWORD=postgres psql -U postgres -d $Database -t -c 'SELECT version FROM schema_migrations ORDER BY id;' 2>&1" 
 $appliedMigrations = @() 
 if ($appliedRaw) { 
     $appliedMigrations = $appliedRaw -split "`n" | 
         ForEach-Object { $_.Trim() } | 
         Where-Object { $_ -ne "" -and $_ -notmatch "NOTICE:" } 
 } 
 
 if ($appliedMigrations.Count -gt 0) { 
     Write-Host "Already applied ($($appliedMigrations.Count)):" -ForegroundColor Green 
     foreach ($m in $appliedMigrations) { 
         Write-Host "  - $m" -ForegroundColor Gray 
     } 
 } else { 
     Write-Host "No migrations applied yet." -ForegroundColor Green 
 } 
 Write-Host "" 
 
 # --------------------------------------------------------------------------- 
 # Run pending migrations 
 # --------------------------------------------------------------------------- 
 Write-Host "============================================" -ForegroundColor Cyan 
 Write-Host "  Running pending migrations..." -ForegroundColor Cyan 
 Write-Host "============================================" -ForegroundColor Cyan 
 Write-Host "" 
 
 $pendingCount = 0 
 $successCount = 0 
 $failedCount  = 0 
 
 foreach ($file in $migrationFiles) { 
     $version = $file.BaseName 
 
     if ($appliedMigrations -contains $version) { 
         Write-Host "SKIP  $($file.Name) (already applied)" -ForegroundColor Gray 
         continue 
     } 
 
     $pendingCount++ 
     Write-Host "RUN   $($file.Name)" -ForegroundColor Yellow 
 
     $tempFile    = "/tmp/migration_${version}.sql" 
     $relativePath = "migrations/$($file.Name)" 
 
     kubectl cp "$relativePath" "${postgresPod}:${tempFile}" -n $Namespace -c postgres 2>$null 
 
     if ($LASTEXITCODE -ne 0) { 
         Write-Host "  ERROR: Failed to copy file to pod" -ForegroundColor Red 
         $failedCount++ 
         continue 
     } 
 
     $output   = kubectl exec -n $Namespace $postgresPod -c postgres -- sh -c "PGPASSWORD=postgres psql -U postgres -d $Database -f $tempFile 2>&1" 
     $psqlExit = $LASTEXITCODE 
     $errors   = $output | Where-Object { $_ -match "^ERROR:" } 
 
     if ($psqlExit -eq 0 -and $errors.Count -eq 0) { 
         Write-Host "  OK" -ForegroundColor Green 
         
         # Record migration in schema_migrations table 
         $description = $file.BaseName -replace '^\d+_', '' -replace '_', ' ' 
         
         # Use a temp file to avoid PowerShell string escaping issues 
         $recordFile = Join-Path $PSScriptRoot "record_migration_temp.sql" 
         $recordLines = @( 
             "INSERT INTO schema_migrations (version, description)", 
             "VALUES ('$version', '$description');" 
         ) 
         $recordLines | Set-Content -Path $recordFile -Encoding UTF8 
         
         kubectl cp "$recordFile" "${postgresPod}:/tmp/record_migration.sql" -n $Namespace -c postgres 2>$null 
         $recordOutput = kubectl exec -n $Namespace $postgresPod -c postgres -- sh -c "PGPASSWORD=postgres psql -U postgres -d $Database -f /tmp/record_migration.sql 2>&1" 
         Remove-Item $recordFile -ErrorAction SilentlyContinue 
         kubectl exec -n $Namespace $postgresPod -c postgres -- rm -f /tmp/record_migration.sql 2>$null 
         
         if ($LASTEXITCODE -eq 0) { 
             Write-Host "  RECORDED in schema_migrations" -ForegroundColor Green 
         } else { 
             Write-Host "  WARNING: Failed to record migration (but SQL executed successfully)" -ForegroundColor Yellow 
         } 
         
         $successCount++ 
     } else { 
         Write-Host "  FAILED:" -ForegroundColor Red 
         $output | ForEach-Object { Write-Host "    $_" -ForegroundColor Red } 
         $failedCount++ 
     } 
 
     kubectl exec -n $Namespace $postgresPod -c postgres -- rm -f $tempFile 2>$null 
     Write-Host "" 
 } 
 
 # --------------------------------------------------------------------------- 
 # Summary 
 # --------------------------------------------------------------------------- 
 Write-Host "============================================" -ForegroundColor Cyan 
 Write-Host "  Summary" -ForegroundColor Cyan 
 Write-Host "============================================" -ForegroundColor Cyan 
 Write-Host "Total files:      $($migrationFiles.Count)" 
 Write-Host "Already applied:  $($appliedMigrations.Count)" 
 Write-Host "Pending:          $pendingCount" 
 Write-Host "Applied now:      $successCount" -ForegroundColor Green 
 
 if ($failedCount -gt 0) { 
     Write-Host "Failed:           $failedCount" -ForegroundColor Red 
 } else { 
     Write-Host "Failed:           $failedCount" -ForegroundColor Green 
 } 
 Write-Host "" 
 
 if ($failedCount -gt 0) { 
     Write-Host "Some migrations failed. Check errors above." -ForegroundColor Red 
     exit 1 
 } 
 
 if ($successCount -eq 0 -and $pendingCount -eq 0) { 
     Write-Host "Database is up to date. Nothing to do." -ForegroundColor Green 
 } else { 
     Write-Host "All pending migrations applied successfully." -ForegroundColor Green 
 } 
 
 Write-Host "" 
 Write-Host "Verify with:" -ForegroundColor Yellow 
 Write-Host "  kubectl exec -n $Namespace $postgresPod -c postgres -- sh -c ""PGPASSWORD=postgres psql -U postgres -d $Database -c 'SELECT * FROM schema_migrations;'"" " -ForegroundColor Cyan 
 Write-Host "" 
