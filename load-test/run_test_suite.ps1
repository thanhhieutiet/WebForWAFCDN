# ==============================================================================
# CMC Telecom LabShop - Automated Load & Security Test Suite Controller (PowerShell)
# Supports: Performance Benchmark (4 Cases), Multi-ISP Routing Test, WAF Defense Test
# Usage:
#   .\run_test_suite.ps1 [Mode] [Host] [Workers]
# Example:
#   .\run_test_suite.ps1 perf https://app.thanhhieutiet.shop 4
#   .\run_test_suite.ps1 security https://app.thanhhieutiet.shop 4
#   .\run_test_suite.ps1 down
# ==============================================================================

param (
    [string]$Mode = "interactive",
    [string]$TargetHost = "https://app.thanhhieutiet.shop",
    [int]$Workers = 4
)

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

function Print-Header {
    Write-Host "====================================================================" -ForegroundColor Cyan
    Write-Host "  CMC TELECOM LABSHOP - DISTRIBUTED LOAD TESTING CONTROLLER (PS)  " -ForegroundColor Green
    Write-Host "====================================================================" -ForegroundColor Cyan
}

function Start-Stack {
    param (
        [string]$TestMode,
        [string]$UseRange,
        [string]$HostUrl,
        [int]$WorkerCount
    )

    Write-Host "`n>> Configuring Environment:" -ForegroundColor Yellow
    Write-Host "   - TEST_MODE=$TestMode"
    Write-Host "   - USE_RANGE_HEADER=$UseRange"
    Write-Host "   - TARGET_HOST=$HostUrl"
    Write-Host "   - WORKERS=$WorkerCount replicas"

    $env:TEST_MODE = $TestMode
    $env:USE_RANGE_HEADER = $UseRange
    $env:LOCUST_HOST = $HostUrl

    Write-Host "`n>> Starting Locust Master-Worker cluster and Telemetry pipeline..." -ForegroundColor Blue
    docker compose up -d --scale locust-worker=$WorkerCount

    Write-Host "`n====================================================================" -ForegroundColor Green
    Write-Host "  LOAD TEST CLUSTER IS READY & RUNNING!  " -ForegroundColor Green
    Write-Host "====================================================================" -ForegroundColor Green
    Write-Host "  * Locust Web UI:         http://localhost:8089" -ForegroundColor Cyan
    Write-Host "  * Grafana Dashboard:     http://localhost:3000 (admin / admin)" -ForegroundColor Cyan
    Write-Host "  * Prometheus Targets:    http://localhost:9090/targets" -ForegroundColor Cyan
    Write-Host "  * Locust Exporter:       http://localhost:9646/metrics" -ForegroundColor Cyan
    Write-Host "--------------------------------------------------------------------" -ForegroundColor Green
    Write-Host ">> Open http://localhost:8089 in your browser, enter CCU & Ramp-up, and click 'Start swarming'!"
    Write-Host ">> Watch live real-time metrics at http://localhost:3000`n"
}

if ($Mode -eq "interactive") {
    Print-Header
    Write-Host "Select Test Mode to Execute:" -ForegroundColor White
    Write-Host "  1) Performance Benchmark (4 Cases: Direct, WAF, CDN, Full)" -ForegroundColor Green
    Write-Host "     - TEST_MODE=performance | USE_RANGE_HEADER=true (Anti-Bottleneck)"
    Write-Host "     - Recommended Load: 300 - 500 CCU | Ramp-up: 20 | Duration: 5 mins"
    Write-Host ""
    Write-Host "  2) Multi-ISP / Multi-Network Routing Test (Office vs Home vs 4G)" -ForegroundColor Blue
    Write-Host "     - TEST_MODE=performance | USE_RANGE_HEADER=true"
    Write-Host "     - Recommended Load: 30 CCU | Ramp-up: 10 | Duration: 2 mins"
    Write-Host ""
    Write-Host "  3) WAF Defense & Anti-Brute-Force Security Test" -ForegroundColor Yellow
    Write-Host "     - TEST_MODE=security | WafRateLimitProbeUser only"
    Write-Host "     - Recommended Load: 100 CCU | Ramp-up: 20 | Duration: 2 mins"
    Write-Host ""
    Write-Host "  4) Scale Locust Workers (Current default: $Workers workers)" -ForegroundColor Cyan
    Write-Host "  5) Stop All Load Testing Services (docker compose down)" -ForegroundColor Red
    Write-Host "  0) Exit"
    Write-Host "--------------------------------------------------------------------" -ForegroundColor Cyan
    $Choice = Read-Host "Enter choice [1-5]"

    switch ($Choice) {
        "1" {
            $inHost = Read-Host "Enter target host [$TargetHost]"
            if ($inHost) { $TargetHost = $inHost }
            Start-Stack -TestMode "performance" -UseRange "true" -HostUrl $TargetHost -WorkerCount $Workers
        }
        "2" {
            $inHost = Read-Host "Enter target host [$TargetHost]"
            if ($inHost) { $TargetHost = $inHost }
            Start-Stack -TestMode "performance" -UseRange "true" -HostUrl $TargetHost -WorkerCount 2
        }
        "3" {
            $inHost = Read-Host "Enter target host [$TargetHost]"
            if ($inHost) { $TargetHost = $inHost }
            Start-Stack -TestMode "security" -UseRange "false" -HostUrl $TargetHost -WorkerCount $Workers
        }
        "4" {
            $newWorkers = Read-Host "Enter desired number of Locust workers"
            docker compose up -d --scale locust-worker=$newWorkers
            Write-Host "Scaled locust-worker to $newWorkers instances." -ForegroundColor Green
        }
        "5" {
            Write-Host "Stopping all services..." -ForegroundColor Yellow
            docker compose down
            Write-Host "All services stopped successfully." -ForegroundColor Green
        }
        Default {
            Write-Host "Exiting."
        }
    }
}
elseif ($Mode -in @("perf", "performance")) {
    Start-Stack -TestMode "performance" -UseRange "true" -HostUrl $TargetHost -WorkerCount $Workers
}
elseif ($Mode -in @("network", "multi-isp")) {
    Start-Stack -TestMode "performance" -UseRange "true" -HostUrl $TargetHost -WorkerCount 2
}
elseif ($Mode -in @("security", "waf")) {
    Start-Stack -TestMode "security" -UseRange "false" -HostUrl $TargetHost -WorkerCount $Workers
}
elseif ($Mode -in @("down", "stop")) {
    Write-Host "Stopping all services..." -ForegroundColor Yellow
    docker compose down
    Write-Host "All services stopped." -ForegroundColor Green
}
elseif ($Mode -in @("status", "ps")) {
    docker compose ps
}
else {
    Write-Host "Unknown mode: $Mode" -ForegroundColor Red
    Write-Host "Usage: .\run_test_suite.ps1 [perf|network|security|down|status] [TargetHost] [Workers]"
}
