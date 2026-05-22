param(
    [Parameter(Mandatory = $false)]
    [string]$Makefiles = ""
)

$files = @()
if ($Makefiles) {
    $files = $Makefiles -split '\s+' | Where-Object { $_ -ne "" }
}

if ($files.Count -eq 0) {
    Write-Error "No makefiles provided."
    exit 1
}

Write-Output ""
Write-Output "Usage:"
Write-Output "  make <target>"

foreach ($file in $files) {
    if (-not (Test-Path -LiteralPath $file)) {
        continue
    }

    foreach ($line in Get-Content -LiteralPath $file) {
        if ($line -match '^##@\s*(.+)$') {
            Write-Output ""
            Write-Output $Matches[1]
            continue
        }

        if ($line -match '^([a-zA-Z_0-9-]+):.*##\s*(.*)$') {
            $target = $Matches[1]
            $description = $Matches[2]
            Write-Output ("  {0,-20} {1}" -f $target, $description)
        }
    }
}
