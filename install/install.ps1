# PowerShell script to install SSL Certificate Checker

# Enable strict mode
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Output colors
$Blue = [System.ConsoleColor]::Blue
$Green = [System.ConsoleColor]::Green
$Red = [System.ConsoleColor]::Red

Write-Host "Installing SSL Certificate Checker..." -ForegroundColor $Blue

# Determine system architecture
$arch = if ([Environment]::Is64BitOperatingSystem) {
    if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
        "arm64"
    } else {
        "amd64"
    }
} else {
    Write-Host "Unsupported architecture: 32-bit systems are not supported" -ForegroundColor $Red
    exit 1
}

# Set installation directory
$installDir = Join-Path $env:USERPROFILE ".certchecker"
Write-Host "Creating directory structure in $installDir..." -ForegroundColor $Blue

# Create directory structure
@("bin", "config", "logs", "data") | ForEach-Object {
    $path = Join-Path $installDir $_
    New-Item -ItemType Directory -Path $path -Force | Out-Null
}

# Get latest release URL
Write-Host "Fetching latest release..." -ForegroundColor $Blue
$latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/mchl18/ssl-expiration-check-bot/releases/latest"
$assetPattern = "windows-$arch.zip$"
$downloadUrl = $latestRelease.assets | Where-Object { $_.browser_download_url -match $assetPattern } | Select-Object -ExpandProperty browser_download_url

if (-not $downloadUrl) {
    Write-Host "Failed to find release for windows-$arch" -ForegroundColor $Red
    exit 1
}

# Download and extract release
Write-Host "Downloading windows-$arch release..." -ForegroundColor $Blue
$zipPath = Join-Path $installDir "release.zip"
Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath

Write-Host "Extracting release..." -ForegroundColor $Blue
Expand-Archive -Path $zipPath -DestinationPath (Join-Path $installDir "temp") -Force
Move-Item -Path (Join-Path $installDir "temp\bin\*") -Destination (Join-Path $installDir "bin") -Force
Remove-Item -Path $zipPath
Remove-Item -Path (Join-Path $installDir "temp") -Recurse -Force

Write-Host "Installation complete!" -ForegroundColor $Green
Write-Host

Write-Host "To complete the installation:" -ForegroundColor $Blue
Write-Host
Write-Host "1. Add to your PATH by running this command in PowerShell as Administrator:" -ForegroundColor $Blue
Write-Host "   [Environment]::SetEnvironmentVariable('Path', `$env:Path + ';$installDir\bin', 'User')" -ForegroundColor $Green
Write-Host
Write-Host "2. Configure the service by running:" -ForegroundColor $Blue
Write-Host "   certchecker.exe -configure" -ForegroundColor $Green 