param(
    [Parameter(Mandatory = $true)]
    [string]$ReleaseChannel,

    [Parameter(Mandatory = $true)]
    [string]$OutputBinaryPath,

    [Parameter(Mandatory = $true)]
    [string]$Arch
)

if ($ReleaseChannel -eq "edge") {
    $ReleaseChannel = "latest"
}

$outputDir = Split-Path -Parent $OutputBinaryPath
if (-not $outputDir) {
    Write-Error "Output directory could not be derived from '$OutputBinaryPath'."
    exit 1
}

if (-not (Get-Command Invoke-WebRequest -ErrorAction SilentlyContinue)) {
    Write-Error "Invoke-WebRequest is required but was not found."
    exit 1
}

New-Item -ItemType Directory -Path $outputDir -Force | Out-Null

$bicepConfigPath = Join-Path $outputDir "bicepconfig.json"
@"
{
  "extensions": {
    "radius": "br:biceptypes.azurecr.io/radius:$ReleaseChannel",
    "aws": "br:biceptypes.azurecr.io/aws:$ReleaseChannel"
  }
}
"@ | Set-Content -LiteralPath $bicepConfigPath -Encoding ascii

$bicepArch = "x64"
if ($Arch -eq "arm64") {
    $bicepArch = "arm64"
}

# Bicep CLI version. Pinned because Bicep v0.43+ tightened
# ContainerRegistryClientFactory.ThrowIfRegistryNotTrusted to reject br:localhost:5000/... targets,
# breaking publish-extension to local registries used by our CI and local dev workflows.
$bicepVersion = "v0.42.1"
$bicepDownloadUrl = "https://github.com/Azure/bicep/releases/download/$bicepVersion/bicep-win-$bicepArch.exe"

if (Test-Path -LiteralPath $OutputBinaryPath) {
    Write-Host "Bicep CLI already exists at $OutputBinaryPath, skipping download."
    exit 0
}

Write-Host "Downloading Bicep CLI $bicepVersion..."
try {
    Invoke-WebRequest -Uri $bicepDownloadUrl -OutFile $OutputBinaryPath
}
catch {
    Write-Error "Failed to download Bicep CLI from $bicepDownloadUrl. $_"
    exit 1
}

Write-Host "Bicep CLI installed successfully at $OutputBinaryPath"
