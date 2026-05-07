#!/usr/bin/env pwsh
# migrate-tasks.ps1 — Migrate task-pool.md to beads
# Usage: ./migrate-tasks.ps1 -ProjectPath <path> [-DryRun]

param(
    [Parameter(Mandatory=$true)]
    [string]$ProjectPath,

    [switch]$DryRun,
    [switch]$Force,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

if ($Help) {
    Write-Host @"
Migrate tasks from .team/task-pool.md to beads.

Usage:
  ./migrate-tasks.ps1 -ProjectPath {PROJECT_PATH}
  ./migrate-tasks.ps1 -ProjectPath {PROJECT_PATH} -DryRun
  ./migrate-tasks.ps1 -ProjectPath {PROJECT_PATH} -Force

Parameters:
  -ProjectPath  Path to project directory (required)
  -DryRun       Preview without creating issues
  -Force        Recreate mapping even if exists
  -Help         Show this help
"@
    exit 0
}

# Validate project path
$FullPath = Resolve-Path $ProjectPath -ErrorAction SilentlyContinue
if (-not $FullPath) {
    Write-Error "Project path not found: $ProjectPath"
    exit 1
}

$TaskPoolPath = Join-Path $FullPath ".team/task-pool.md"
$BeadsPath = Join-Path $FullPath ".beads"
$MappingPath = Join-Path $BeadsPath "task-pool-mapping.json"

# Check prerequisites
if (-not (Test-Path $TaskPoolPath)) {
    Write-Error "task-pool.md not found at: $TaskPoolPath"
    exit 1
}

if (-not (Test-Path $BeadsPath)) {
    Write-Error "beads not initialized. Run 'bd init' in $ProjectPath first"
    exit 1
}

if ((Test-Path $MappingPath) -and -not $Force) {
    Write-Host "Mapping already exists: $MappingPath" -ForegroundColor Yellow
    Write-Host "Use -Force to recreate" -ForegroundColor Yellow
    exit 0
}

Write-Host "Migrating tasks from: $TaskPoolPath" -ForegroundColor Cyan
Write-Host "To beads database: $BeadsPath" -ForegroundColor Cyan
if ($DryRun) {
    Write-Host "** DRY RUN - No changes will be made **" -ForegroundColor Yellow
}

# Parse task-pool.md
$Content = Get-Content $TaskPoolPath -Raw
$Lines = $Content -split "`n"

$Tasks = @()
$InTable = $false

foreach ($Line in $Lines) {
    # Detect table start
    if ($Line -match '^\|.*Task ID.*Task\|') {
        $InTable = $true
        continue
    }

    # Skip separator
    if ($Line -match '^\|[-\s|]+\|$') {
        continue
    }

    # Parse table rows
    if ($InTable -and $Line -match '^\|') {
        $Fields = $Line -split '\|' | Where-Object { $_.Trim() } | ForEach-Object { $_.Trim() }

        # Skip if not enough fields or header row
        if ($Fields.Count -lt 3) { continue }
        if ($Fields[0] -eq "Task ID") { continue }

        # Parse task
        $TaskID = $Fields[0] -replace '^\s+|\s+$', ''

        # Skip empty or malformed IDs
        if ($TaskID -notmatch '[A-Z]+\d+') { continue }

        # Skip T_ prefixed IDs (these are already migrated duplicates)
        if ($TaskID -match '^T_') { continue }

        $Task = @{
            id = $TaskID
            title = if ($Fields.Count -gt 1) { $Fields[1] } else { "Untitled" }
            type = if ($Fields.Count -gt 2) { $Fields[2] } else { "task" }
            deps = if ($Fields.Count -gt 3) { $Fields[3] } else { "" }
            milestone = if ($Fields.Count -gt 4) { $Fields[4] } else { "" }
            owner = if ($Fields.Count -gt 5) { $Fields[5] } else { "" }
            priority = if ($Fields.Count -gt 6) { $Fields[6] } else { "P2" }
            status = if ($Fields.Count -gt 7) { $Fields[7] } else { "todo" }
            phase = if ($Fields.Count -gt 8) { $Fields[8] } else { "" }
            subcol = if ($Fields.Count -gt 9) { $Fields[9] } else { "" }
            doc = if ($Fields.Count -gt 10) { $Fields[10] } else { "" }
        }

        $Tasks += $Task
    }

    # End of table
    if ($InTable -and $Line -notmatch '^\|' -and $Line -match '\S') {
        $InTable = $false
    }
}

Write-Host "Parsed $($Tasks.Count) tasks from task-pool.md" -ForegroundColor Green

if ($Tasks.Count -eq 0) {
    Write-Warning "No tasks found. Check task-pool.md format."
    exit 0
}

# Map type to beads type
function Get-BeadsType {
    param($Type)
    switch -Regex ($Type) {
        'Bug|bug|B' { 'bug' }
        'Feature|feature|F' { 'feature' }
        'Analysis|analysis|A' { 'task' }
        'Change|change|C' { 'task' }
        'Task|task|T' { 'task' }
        default { 'task' }
    }
}

# Map priority
function Get-BeadsPriority {
    param($Priority)
    switch -Regex ($Priority) {
        'P0|0' { 0 }
        'P1|1' { 1 }
        'P2|2' { 2 }
        'P3|3' { 3 }
        'P4|4' { 4 }
        default { 2 }
    }
}

# Map status to beads status + phase label
function Get-BeadsStatusAndPhase {
    param($Status, $Phase)

    $Result = @{
        status = 'open'
        phaseLabel = 'phase:ready'
    }

    switch -Regex ($Status) {
        'Doing|doing|In Progress|in_progress' {
            $Result.status = 'in_progress'
            $Result.phaseLabel = 'phase:implement'
        }
        'Done|done|Closed|closed' {
            $Result.status = 'closed'
            $Result.phaseLabel = $null
        }
        'Blocked|blocked' {
            $Result.status = 'blocked'
            $Result.phaseLabel = 'phase:implement'
        }
        'Review|review' {
            $Result.status = 'open'
            $Result.phaseLabel = 'phase:review'
        }
        default {
            # Use phase column if available
            switch -Regex ($Phase) {
                'todo|Todo|TODO' { $Result.phaseLabel = 'phase:ready' }
                'doing|Doing|DOING' { $Result.phaseLabel = 'phase:implement'; $Result.status = 'in_progress' }
                'done|Done|DONE' { $Result.status = 'closed'; $Result.phaseLabel = $null }
                'analyze|Analyze' { $Result.phaseLabel = 'phase:analyze' }
                'verify|Verify' { $Result.phaseLabel = 'phase:verify' }
                default { $Result.phaseLabel = 'phase:ready' }
            }
        }
    }

    return $Result
}

# Create mapping
$Mapping = @{}

# Process each task
$Created = 0
$Skipped = 0

foreach ($Task in $Tasks) {
    $BeadsType = Get-BeadsType $Task.type
    $BeadsPriority = Get-BeadsPriority $Task.priority
    $StatusPhase = Get-BeadsStatusAndPhase $Task.status $Task.phase

    # Build command
    $Cmd = "bd create `"$($Task.id): $($Task.title)`" -t $BeadsType -p $BeadsPriority --external-ref `"$($Task.id)`""

    if ($StatusPhase.phaseLabel) {
        $Cmd += " --add-label $($StatusPhase.phaseLabel)"
    }

    # Add subsystem label if subcol exists
    if ($Task.subcol -and $Task.subcol -ne '-') {
        $Subsystem = $Task.subcol -replace '^\s+|\s+$', ''
        $Cmd += " --add-label subsystem:$Subsystem"
    }

    # Add notes with additional context
    $Notes = @()
    if ($Task.deps -and $Task.deps -ne '-') {
        $Notes += "Dependencies: $($Task.deps)"
    }
    if ($Task.milestone -and $Task.milestone -ne '-') {
        $Notes += "Milestone: $($Task.milestone)"
    }
    if ($Task.doc -and $Task.doc -ne '-') {
        $Notes += "Doc: $($Task.doc)"
    }
    if ($Task.owner -and $Task.owner -ne '-') {
        $Notes += "Owner: $($Task.owner)"
    }

    if ($Notes.Count -gt 0) {
        $NotesStr = $Notes -join '; '
        $Cmd += " --notes `"$NotesStr`""
    }

    if ($StatusPhase.status -eq 'closed') {
        $Cmd += " --status closed"
    }

    $Cmd += " --json"

    Write-Host "  [$($Task.id)] $($Task.title)" -ForegroundColor Gray

    if ($DryRun) {
        Write-Host "    Would run: $Cmd" -ForegroundColor DarkGray
        $Mapping[$Task.id] = "cms-dryrun-$Created"
        $Created++
    } else {
        # Execute
        $Result = Invoke-Expression $Cmd 2>&1

        if ($LASTEXITCODE -eq 0) {
            $Issue = $Result | ConvertFrom-Json
            $Mapping[$Task.id] = $Issue.id
            $Created++
            Write-Host "    ✓ Created $($Issue.id)" -ForegroundColor Green
        } else {
            $Skipped++
            Write-Warning "    ✗ Failed: $Result"
        }
    }
}

# Save mapping
if (-not $DryRun) {
    $Mapping | ConvertTo-Json | Out-File $MappingPath -Encoding UTF8
    Write-Host "`n✓ Mapping saved to: $MappingPath" -ForegroundColor Green
}

# Summary
Write-Host "`n=== Migration Summary ===" -ForegroundColor Cyan
Write-Host "Total tasks: $($Tasks.Count)" -ForegroundColor White
Write-Host "Created:     $Created" -ForegroundColor Green
Write-Host "Skipped:     $Skipped" -ForegroundColor $(if ($Skipped -gt 0) { 'Red' } else { 'Green' })

if ($DryRun) {
    Write-Host "`n** This was a dry run. Remove -DryRun to perform actual migration. **" -ForegroundColor Yellow
}
