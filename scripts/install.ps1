$ErrorActionPreference = "Stop"

$REPO = "winkeep/winkeep"
$INSTALL_DIR = "$env:LOCALAPPDATA\winkeep"
$BINARY = "winkeep.exe"

Write-Host ""
Write-Host "  WinKeep Installer" -ForegroundColor Cyan
Write-Host "  nohup for Windows" -ForegroundColor DarkGray
Write-Host ""

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
Write-Host "[1/4] Detected architecture: $arch" -ForegroundColor Gray

# Get latest version from GitHub API
Write-Host "[2/4] Fetching latest version..." -ForegroundColor Gray
$version = $null
try {
    $release = Invoke-RestMethod `
        -Uri "https://api.github.com/repos/$REPO/releases/latest" `
        -Headers @{
            "Accept"     = "application/vnd.github.v3+json"
            "User-Agent" = "winkeep-installer"
        } `
        -TimeoutSec 15
    $version = $release.tag_name
    Write-Host "       Latest version: $version" -ForegroundColor Green
} catch {
    Write-Host "       Failed to fetch latest version from GitHub API." -ForegroundColor Red
    Write-Host "       Check your internet connection or visit: https://github.com/$REPO/releases" -ForegroundColor Yellow
    exit 1
}

# Validate version tag
if (-not $version -or $version -eq "") {
    Write-Host "       Could not determine latest version. Aborting." -ForegroundColor Red
    exit 1
}

# Download binary
Write-Host "[3/4] Downloading WinKeep $version..." -ForegroundColor Gray
$filename = "winkeep_windows_${arch}.exe"
$downloadUrl = "https://github.com/$REPO/releases/download/$version/$filename"
$sfUrl = "https://sourceforge.net/projects/winkeep/files/$version/$filename/download"

New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
$destPath = Join-Path $INSTALL_DIR $BINARY

$downloaded = $false
try {
    Write-Host "       Trying GitHub..." -ForegroundColor Gray
    Invoke-WebRequest `
        -Uri $downloadUrl `
        -OutFile $destPath `
        -UseBasicParsing `
        -Headers @{ "User-Agent" = "winkeep-installer" } `
        -TimeoutSec 60
    $downloaded = $true
} catch {
    Write-Host "       GitHub download failed, trying SourceForge..." -ForegroundColor Yellow
}

if (-not $downloaded) {
    try {
        Invoke-WebRequest `
            -Uri $sfUrl `
            -OutFile $destPath `
            -UseBasicParsing `
            -Headers @{ "User-Agent" = "winkeep-installer" } `
            -TimeoutSec 60
        $downloaded = $true
    } catch {
        Write-Host "       All download sources failed." -ForegroundColor Red
        Write-Host "       Please download manually from: https://github.com/$REPO/releases" -ForegroundColor Yellow
        exit 1
    }
}

# Verify downloaded file exists and is non-empty
if (-not (Test-Path $destPath) -or (Get-Item $destPath).Length -eq 0) {
    Write-Host "       Downloaded file is missing or empty. Aborting." -ForegroundColor Red
    exit 1
}

# Add to PATH
Write-Host "[4/4] Adding to PATH..." -ForegroundColor Gray
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$INSTALL_DIR*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$INSTALL_DIR", "User")
    $env:Path = "$env:Path;$INSTALL_DIR"
    Write-Host "       Added to PATH" -ForegroundColor Green
} else {
    Write-Host "       Already in PATH" -ForegroundColor Gray
}

# Verify binary runs
Write-Host ""
Write-Host "  Installation complete!" -ForegroundColor Green
Write-Host ""
try {
    & $destPath version
} catch {
    Write-Host "  Warning: Could not run 'winkeep version'. Try restarting your terminal." -ForegroundColor Yellow
}
Write-Host ""
Write-Host "  Usage:" -ForegroundColor Cyan
Write-Host "    winkeep run -- python server.py" -ForegroundColor White
Write-Host "    winkeep list" -ForegroundColor White
Write-Host "    winkeep logs <id>" -ForegroundColor White
Write-Host ""
Write-Host "  Restart your terminal to use 'winkeep' command." -ForegroundColor Yellow
Write-Host ""
