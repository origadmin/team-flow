#!/usr/bin/env pwsh
# export-task-pool.ps1 — Export beads issues to human-readable Markdown
# Usage: ./export-task-pool.ps1 [-Output <path>] [-Priority <0-4>] [-Status <status>]

param(
    [string]$Output = ".team/task-pool-export.md",
    [int[]]$Priority = @(0, 1, 2, 3, 4),
    [string[]]$Status = @("open", "in_progress", "blocked"),
    [switch]$All,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

if ($Help) {
    Write-Host @"
Export beads issues to human-readable Markdown.

Usage:
  ./export-task-pool.ps1                     # Export all open issues
  ./export-task-pool.ps1 -Priority 0,1       # Export P0 and P1 only
  ./export-task-pool.ps1 -Status closed      # Export closed issues
  ./export-task-pool.ps1 -All                # Export all issues regardless of status

Parameters:
  -Output   Output file path (default: .team/task-pool-export.md)
  -Priority Filter by priority (0-4, default: all)
  -Status   Filter by status (default: open, in_progress, blocked)
  -All      Export all issues (ignores -Status filter)
  -Help     Show this help
"@
    exit 0
}

# Ensure we're in a project directory with beads
if (-not (Test-Path ".beads")) {
    Write-Error "No .beads directory found. Run this script from a beads-enabled project."
    exit 1
}

# Build command
$Cmd = "bd list --json"

if (-not $All) {
    $StatusFilter = $Status -join ","
    $Cmd += " --status $StatusFilter"
}

$PriorityFilter = $Priority -join ","
$Cmd += " --priority $PriorityFilter"

Write-Host "Fetching issues..." -ForegroundColor Cyan
Write-Host "  Command: $Cmd" -ForegroundColor Gray

# Execute and parse
$IssuesJson = Invoke-Expression $Cmd 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Error "Failed to fetch issues: $IssuesJson"
    exit 1
}

$Issues = $IssuesJson | ConvertFrom-Json

if ($Issues.Count -eq 0) {
    Write-Host "No issues found matching criteria." -ForegroundColor Yellow
    exit 0
}

Write-Host "Found $($Issues.Count) issues" -ForegroundColor Green

# Build output
$OutputContent = @()
$OutputContent += "# Task Pool Export"
$OutputContent += ""
$OutputContent += "> **Generated**: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
$OutputContent += "> **Filters**: Priority=$($Priority -join ',') | Status=$($Status -join ',')"
$OutputContent += ""
$OutputContent += "## Summary"
$OutputContent += ""
$OutputContent += "| Priority | Count |"
$OutputContent += "|----------|-------|"

# Group by priority
$PriorityGroups = $Issues | Group-Object priority | Sort-Object Name
foreach ($Group in $PriorityGroups) {
    $OutputContent += "| P$($Group.Name) | $($Group.Count) |"
}
$OutputContent += ""

# Detailed sections
foreach ($Prio in ($Priority | Sort-Object)) {
    $PrioIssues = $Issues | Where-Object { $_.priority -eq $Prio }
    if ($PrioIssues.Count -eq 0) { continue }
    
    $OutputContent += "## P$Prio Issues"
    $OutputContent += ""
    $OutputContent += "| Team ID | Beads ID | Title | Type | Status | Phase | Subsystem |"
    $OutputContent += "|---------|----------|-------|------|--------|-------|-----------|"
    
    foreach ($Issue in ($PrioIssues | Sort-Object title)) {
        # Extract labels
        $Phase = ($Issue.labels | Where-Object { $_ -match '^phase:' }) -replace 'phase:', ''
        if (-not $Phase) { $Phase = "-" }
        
        $Subsystem = ($Issue.labels | Where-Object { $_ -match '^subsystem:' }) -replace 'subsystem:', ''
        if (-not $Subsystem) { $Subsystem = "-" }
        
        # Get team ID from externalRef or extract from title
        $TeamID = $Issue.externalRef
        if (-not $TeamID -and $Issue.title -match '^([FBCAT]-?\d+|[TBFC]\d+):') {
            $TeamID = $Matches[1]
        }
        if (-not $TeamID) { $TeamID = "-" }
        
        # Type emoji
        $TypeEmoji = switch ($Issue.issue_type) {
            "bug" { "🐛" }
            "feature" { "✨" }
            "task" { "📋" }
            "chore" { "🔧" }
            "epic" { "📚" }
            "decision" { "⚖️" }
            default { "○" }
        }
        
        # Status emoji
        $StatusEmoji = switch ($Issue.status) {
            "open" { "○" }
            "in_progress" { "◐" }
            "blocked" { "●" }
            "closed" { "✓" }
            "deferred" { "❄" }
            default { "○" }
        }
        
        # Truncate title if too long
        $Title = $Issue.title
        if ($Title.Length -gt 50) {
            $Title = $Title.Substring(0, 47) + "..."
        }
        
        $OutputContent += "| $TeamID | $($Issue.id) | $Title | $TypeEmoji$($Issue.issue_type) | $StatusEmoji$($Issue.status) | $Phase | $Subsystem |"
    }
    $OutputContent += ""
}

# Add detailed notes section
$OutputContent += "---"
$OutputContent += ""
$OutputContent += "<!-- Run: $Cmd -->"
$OutputContent += "<!-- Script: _team/v2/scripts/export-task-pool.ps1 -->"

# Write output
$OutputDir = Split-Path $Output -Parent
if ($OutputDir -and -not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

$OutputContent | Out-File -FilePath $Output -Encoding UTF8

Write-Host "✓ Exported $($Issues.Count) issues to $Output" -ForegroundColor Green
