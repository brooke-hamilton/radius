<#
.SYNOPSIS
Applies the yaml winget configuration to a Windows machine.

.DESCRIPTION
This script invokes winget to apply the yaml configuration file to the Windows machine.

.PARAMETER YamlConfigFilePath
File path to the yaml configuration file to be applied by winget.

.EXAMPLE
Set-WinGetConfiguration.ps1 -ConfigFilePath ".\configuration.dsc.yaml"
#>
param (
    [string]$YamlConfigFilePath = "$PSScriptRoot\configuration.dsc.yaml"
)

Set-StrictMode -Version 3.0
$ErrorActionPreference = 'Stop'

# Check for WinGet
if (-not (Get-Command winget -ErrorAction SilentlyContinue)) {
    Write-Error "WinGet is not installed."
}

Write-Host "Validating WinGet configuration..."
winget configure validate --file $YamlConfigFilePath --disable-interactivity

Write-Host "Starting WinGet configuration from $YamlConfigFilePath..."
winget configure --file $YamlConfigFilePath --accept-configuration-agreements --disable-interactivity
