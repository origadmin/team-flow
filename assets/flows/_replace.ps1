$dir = "d:\workspace\project\golang\origadmin\framework\projects\team-flow\v3\flows"
Get-ChildItem -Path $dir -Filter "*.json" | ForEach-Object {
    $file = $_.FullName
    $content = Get-Content -Path $file -Raw -Encoding UTF8
    $original = $content

    $content = $content -replace '"docs_path": ""', '"DOCS_INTERNAL": ""'
    $content = $content -replace '\{docs_path\}/', '{DOCS_INTERNAL}/'
    $content = $content -replace '\{docs_path\}"', '{DOCS_INTERNAL}"'

    if ($content -ne $original) {
        Set-Content -Path $file -Value $content -NoNewline -Encoding UTF8

        $c1 = ([regex]::Matches($original, '"docs_path": ""')).Count
        $c2 = ([regex]::Matches($original, '\{docs_path\}/')).Count
        $c3 = ([regex]::Matches($original, '\{docs_path\}"')).Count
        $total = $c1 + $c2 + $c3
        Write-Output "$($_.Name): $total changes (key=$c1, path_slash=$c2, path_quote=$c3)"
    } else {
        Write-Output "$($_.Name): 0 changes"
    }
}
