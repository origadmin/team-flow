$source = '{PROJECT_PATH}\.team\task-pool.md'
$target = '{PROJECT_PATH}\.team\task-pool-new.md'
$lines = Get-Content $source -Encoding UTF8
$kept = $lines[0..108]
$kept | Out-File -FilePath $target -Encoding UTF8
Write-Host "Kept $($kept.Count) lines out of $($lines.Count) total lines"
