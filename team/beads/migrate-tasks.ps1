# task-pool to beads migration script
# Usage: Run from projects/orig-cms/ directory

$ProjectDir = "D:/workspace/project/golang/origadmin/framework/projects/orig-cms"
Set-Location $ProjectDir

# Active tasks to import (Review + Doing status)
$tasks = @(
    @{ ID="A001"; Title="RESTful relationship API improvement plan"; Type="task"; Priority=1; Desc="SPEC+AC+R1/R2/R3 design complete, awaiting user confirmation" },
    @{ ID="F014"; Title="Unified pagination+field naming+Proto-first response"; Type="feature"; Priority=1; Desc="Phase 2 all complete(R1+R2+R5+R4), awaiting confirmation" },
    @{ ID="F013"; Title="Dynamic theme system: runtime loading+15 preset themes"; Type="feature"; Priority=1; Desc="Phase 2 implementation complete, awaiting QA" },
    @{ ID="C013"; Title="Auth refactor: AuthProvider+Router Context+auth layout route"; Type="task"; Priority=0; Desc="R1+R2+R3 all implemented, build passing" },
    @{ ID="B069"; Title="Featured button white-on-white regression"; Type="bug"; Priority=0; Desc="R2 fix: index.css [data-theme] + admin_handler double wrap fix" },
    @{ ID="B075"; Title="User field data loss: create_author/update_author empty"; Type="bug"; Priority=0; Desc="Fixed: user_repo→convpb.ConvertUserToUserPBFull" },
    @{ ID="B077"; Title="Watch page shows deleted user + unknown time"; Type="bug"; Priority=0; Desc="Fixed: normalizeMedia GetMediaResponse unwrap" },
    @{ ID="B078"; Title="Admin users: role column icon-only + status toggle broken"; Type="bug"; Priority=0; Desc="R5 fix: Badge children + UpdateUserStatus 5-value enum" },
    @{ ID="B079"; Title="Media CRUD all stub handlers + listable field error"; Type="bug"; Priority=0; Desc="R2: 9 admin CRUD handlers + AdminMode=true + ProtoOK" },
    @{ ID="B080"; Title="Channels/me returns 404 instead of 200+null"; Type="bug"; Priority=0; Desc="Fixed: GetMyChannel 404→200+null channel" },
    @{ ID="B082"; Title="Upload refactoring issues"; Type="bug"; Priority=0; Desc="Fixed: POST /uploads/simple + uploadPart unwrap" },
    @{ ID="B083"; Title="Category status toggle broken"; Type="bug"; Priority=1; Desc="Fixed: categoryStatusFromInt case 0 + newStatus 0→2" },
    @{ ID="B084"; Title="Transcoding status SSE returns 404"; Type="bug"; Priority=1; Desc="Fixed: getSSEUrl URL path correction" },
    @{ ID="B085"; Title="Parent category disabled but child still selectable"; Type="bug"; Priority=1; Desc="Fixed: parent selector filter + cascade status hint" },
    @{ ID="B086"; Title="Watch page comments not working"; Type="bug"; Priority=0; Desc="Fixed: double wrap + OptionalJWT + like_count sort" },
    @{ ID="B087"; Title="Admin comments page cannot display"; Type="bug"; Priority=1; Desc="Fixed: ListByMedia + field mapping + DELETE route" },
    @{ ID="B089"; Title="Media privacy status cannot select"; Type="bug"; Priority=1; Desc="Protojson string enum vs frontend number mismatch" },
    @{ ID="B090"; Title="Portal /@username page style issues"; Type="bug"; Priority=1; Desc="Default icon + empty state layout issues" },
    @{ ID="B091"; Title="Watch page view_count not incrementing"; Type="bug"; Priority=1; Desc="Frontend not calling IncrementViewCount" },
    @{ ID="B092"; Title="Frontend requests non-existent channels/me"; Type="bug"; Priority=1; Desc="GET /api/v1/channels/me returns 404" },
    @{ ID="B093"; Title="Explore page shows Failed to load trending"; Type="bug"; Priority=1; Desc="Fixed: Trending.tsx use res.items directly" },
    @{ ID="B094"; Title="Tag creation fails: field name mismatch"; Type="bug"; Priority=0; Desc="Frontend-backend field name mismatch" },
    @{ ID="B095"; Title="Admin media edit: encoding task display chaos"; Type="bug"; Priority=2; Desc="Profile field mismatch + unclear button" },
    @{ ID="C015"; Title="Rule enhancement: 6 issues"; Type="task"; Priority=0; Desc="All 6 fixes complete" },
    @{ ID="C016"; Title="Unified response API: ProtoOK/ProtoOKPage"; Type="task"; Priority=1; Desc="Design complete" },
    @{ ID="C017"; Title="Image asset reorganization + logo replacement"; Type="task"; Priority=1; Desc="Implementation complete" },
    @{ ID="C018"; Title="User table: remove redundant uuid field"; Type="task"; Priority=1; Desc="Full chain cleanup complete" },
    @{ ID="C019"; Title="convpb time field mapping fix: 11 entities"; Type="task"; Priority=0; Desc="Fixed: 11 entities created_at→create_time" },
    @{ ID="C020"; Title="Ent Schema time field rename"; Type="task"; Priority=0; Desc="Root cause fix in Ent schema" },
    @{ ID="C021"; Title="create_author/update_author Int64→String"; Type="task"; Priority=0; Desc="Match UUID type, full chain fix" },
    @{ ID="C022"; Title="Frontend admin time display: truncate to seconds"; Type="task"; Priority=2; Desc="formatDateTime + 8 page migrations" },
    @{ ID="C023"; Title="Theme color scale enhancement: 50-950"; Type="task"; Priority=1; Desc="Awaiting implementation" },
    @{ ID="C024"; Title="i18n expansion: Intl API + language switcher"; Type="task"; Priority=1; Desc="Awaiting implementation" },
    @{ ID="F015"; Title="Video sprite preview (YouTube-style)"; Type="feature"; Priority=2; Desc="Phase 2 + Bug fix B1 complete" },
    @{ ID="F016"; Title="@username profile page + sidebar My entry"; Type="feature"; Priority=1; Desc="Phase 2 frontend complete" },
    @{ ID="F017"; Title="Category hierarchical display refactor"; Type="feature"; Priority=1; Desc="Awaiting implementation" },
    @{ ID="F018"; Title="Admin comment management enhancement"; Type="feature"; Priority=1; Desc="Phase 2 implementation complete" },
    @{ ID="F019"; Title="Channel creation (v2)"; Type="feature"; Priority=1; Desc="Phase 1 design complete" },
    @{ ID="F020"; Title="User-side article creation/editing"; Type="feature"; Priority=1; Desc="Phase 1 design complete"; Status="doing" },
    @{ ID="F021"; Title="Tag color: auto-assign + admin picker"; Type="feature"; Priority=2; Desc="Awaiting tech-lead design" },
    @{ ID="F022"; Title="Article create/edit page (video site style)"; Type="feature"; Priority=1; Desc="Phase 1 design complete"; Status="doing" },
    @{ ID="F023"; Title="Hashtag auto-tag + slug auto-generate"; Type="feature"; Priority=1; Desc="Phase 2 code complete, awaiting QA" },
    @{ ID="A008"; Title="Quality check enhancement analysis"; Type="task"; Priority=1; Desc="Analysis complete" },
    @{ ID="A009"; Title="Channel creation feature analysis"; Type="task"; Priority=1; Desc="Analysis complete" },
    @{ ID="A010"; Title="Watch history architecture analysis"; Type="task"; Priority=1; Desc="Analysis complete" }
)

function Get-BeadsType($t) {
    switch ($t) {
        "bugfix" { "bug" }
        "feature" { "feature" }
        "change" { "task" }
        "analysis" { "task" }
        "analyze" { "task" }
        "task" { "task" }
        default { "task" }
    }
}

$created = @()
$failed = @()

foreach ($task in $tasks) {
    $beadsType = Get-BeadsType $task.Type
    $priority = $task.Priority
    $title = "$($task.ID): $($task.Title)"
    $desc = $task.Desc
    
    try {
        $result = bd create $title -t $beadsType -p $priority -d $desc --json 2>&1
        $json = $result | ConvertFrom-Json
        $created += @{ OldId=$task.ID; NewId=$json.id; Title=$title }
        Write-Host "OK: $($task.ID) -> $($json.id)" -ForegroundColor Green
    } catch {
        $failed += $task.ID
        Write-Host "FAIL: $($task.ID) - $_" -ForegroundColor Red
    }
}

Write-Host "`n=== Migration Summary ===" -ForegroundColor Cyan
Write-Host "Created: $($created.Count)" -ForegroundColor Green
Write-Host "Failed: $($failed.Count)" -ForegroundColor Red

$mappingPath = Join-Path $ProjectDir ".beads\task-pool-mapping.json"
$created | ConvertTo-Json | Out-File -FilePath $mappingPath -Encoding UTF8
Write-Host "Mapping saved to: $mappingPath" -ForegroundColor Yellow
