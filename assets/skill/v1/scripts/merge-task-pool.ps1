# Rebuild .team/task-pool.md clean from {TEAM_PATH}/task-pool.md
# Conflict IDs: A001,B061,C010,C011,F013,F015,F018,F020 -> T_ prefix

[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
$OutputEncoding = [System.Text.Encoding]::UTF8

$conflictIds = @("A001","B061","C010","C011","F013","F015","F018","F020")
$projectsDir = "{PROJECT_PATH}"
$docsDir = "{DOCS_INTERNAL}"
$requirementsDir = "$docsDir\requirements"

# Read {TEAM_PATH}/task-pool.md
$teamContent = Get-Content "{TEAM_PATH}\task-pool.md" -Encoding UTF8

# Parse {TEAM_PATH} active table (from header to ## 任务详情)
$inTable = $false
$teamTable = @()

foreach ($line in $teamContent) {
    if ($line -match 'ID.*任务描述') { $inTable = $true; continue }
    if ($line -match '任务详情') { break }
    if ($inTable -and $line -match '^\|\s+([A-Z]\d+)\s*\|') {
        $id = $Matches[1]
        $cols = $line -split '\|' | ForEach-Object { $_.Trim() }
        # cols[0]="", cols[1]=ID, cols[2]=Desc, cols[3]=Type, cols[4]=Deps, cols[5]=Milestone, cols[6]=Owner, cols[7]=Priority, cols[8]=Status, cols[9]=Role, cols[10]=Docs
        if ($cols.Count -ge 11) {
            $newId = if ($id -in $conflictIds) { "T_$id" } else { $id }
            $teamTable += @{
                id = $id
                newId = $newId
                desc = $cols[2]
                type = $cols[3]
                deps = $cols[4]
                milestone = $cols[5]
                owner = $cols[6]
                priority = $cols[7]
                status = $cols[8]
                nextRole = $cols[9]
                docs = $cols[10]
            }
        }
    }
}
Write-Host "Parsed {TEAM_PATH} tasks: $($teamTable.Count)"
$conflictTasks = $teamTable | Where-Object { $_.newId -match '^T_' }
$newTasks = $teamTable | Where-Object { $_.newId -notmatch '^T_' }
Write-Host "  Conflict (T_): $($conflictTasks.Count)"
Write-Host "  New: $($newTasks.Count)"

# Build merged table
$merged = @()
$merged += "# 任务池 (Task Pool)"
$merged += ""
$merged += "**版本**: 8.0 | **Owner**: Triage | **最后更新**: 2026-05-01"
$merged += ""
$merged += "## 活跃任务"
$merged += ""
$merged += "| ID | 任务描述 | 类型 | 关联任务 | 里程碑 | 负责人 | 优先级 | 状态 | 阶段 | 建议角色 | 关联文档 |"
$merged += "|----|----------|------|----------|--------|--------|--------|------|------|----------|----------|"

$addedIds = @()

# Conflict tasks first
foreach ($t in $conflictTasks) {
    $merged += "| $($t.newId) | $($t.desc) | $($t.type) | $($t.deps) | $($t.milestone) | $($t.owner) | $($t.priority) | $($t.status) | - | $($t.nextRole) | $($t.docs) |"
    $addedIds += $t.newId
}

# New tasks
foreach ($t in $newTasks) {
    $merged += "| $($t.newId) | $($t.desc) | $($t.type) | $($t.deps) | $($t.milestone) | $($t.owner) | $($t.priority) | $($t.status) | - | $($t.nextRole) | $($t.docs) |"
    $addedIds += $t.newId
}

Write-Host "Total tasks: $($addedIds.Count)"
$merged | Out-File -FilePath "$projectsDir\.team\task-pool.md" -Encoding UTF8

# Verify
$verify = Get-Content "$projectsDir\.team\task-pool.md" -Encoding UTF8
$dupes = ($verify | Where-Object { $_ -match '^\|\s+([A-Z]\d+)\s*\|\s*([A-Z]\d+)\s*\|' }).Count
$rows = ($verify | Where-Object { $_ -match '^\|\s+[A-Z]\d+\s*\|' }).Count
Write-Host ""
Write-Host "=== Result ===" -ForegroundColor Green
Write-Host "Rows: $rows, Duplicate-ID: $dupes"
if ($dupes -eq 0 -and $rows -eq $teamTable.Count) {
    Write-Host "Status: CLEAN" -ForegroundColor Green
} else {
    Write-Host "Status: NEEDS REVIEW" -ForegroundColor Yellow
}
Write-Host "Written: $projectsDir\.team\task-pool.md" -ForegroundColor Cyan
Write-Host ""
Write-Host "Conflict IDs (T_ prefix):"
$conflictTasks | ForEach-Object { Write-Host "  $($_.id) -> $($_.newId)" }