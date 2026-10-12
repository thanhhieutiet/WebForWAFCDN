# PowerShell development script for LabShop
# Handles Windows path with '&' safely.

param (
    [Parameter(Position=0)]
    [ValidateSet("help", "backend", "frontend", "test", "lint", "build", "docker-build", "docker-up", "docker-down")]
    [string]$Action = "help"
)

$ErrorActionPreference = "Stop"
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$RootDir = Split-Path -Parent $ScriptDir

function Show-Help {
    Write-Host "LabShop Dev CLI" -ForegroundColor Cyan
    Write-Host "Usage: .\scripts\dev.ps1 <command>"
    Write-Host "Commands:"
    Write-Host "  backend       Run Go backend (:8080)"
    Write-Host "  frontend      Run Vite dev server (:5173)"
    Write-Host "  test          Run Go tests & Vitest"
    Write-Host "  lint          Run go vet and ESLint"
    Write-Host "  build         Build Go binaries and Vite production assets"
    Write-Host "  docker-build  Build Docker images via Docker Compose"
    Write-Host "  docker-up     Start Docker containers in background"
    Write-Host "  docker-down   Stop and remove Docker containers"
}

switch ($Action) {
    "backend" {
        Write-Host "Starting Go backend..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir"
        go run ./cmd/server
    }
    "frontend" {
        Write-Host "Starting Vite frontend dev server..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir\frontend"
        npm run dev
    }
    "test" {
        Write-Host "Running Go tests..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir"
        go test -race ./...
        
        Write-Host "Running Frontend tests..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir\frontend"
        npm test
    }
    "lint" {
        Write-Host "Linting Go code..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir"
        go vet ./...

        Write-Host "Linting Frontend code..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir\frontend"
        npm run lint
    }
    "build" {
        Write-Host "Building Go binaries..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir"
        go build -o bin/server.exe ./cmd/server

        Write-Host "Building Frontend..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir\frontend"
        npm run build
    }
    "docker-build" {
        Write-Host "Building Docker containers..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir"
        docker compose build
    }
    "docker-up" {
        Write-Host "Starting Docker containers..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir"
        docker compose up -d
    }
    "docker-down" {
        Write-Host "Stopping Docker containers..." -ForegroundColor Green
        Set-Location -LiteralPath "$RootDir"
        docker compose down
    }
    Default {
        Show-Help
    }
}
