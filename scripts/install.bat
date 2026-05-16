@echo off
setlocal enabledelayedexpansion

echo.
echo   WinKeep Installer
echo   nohup for Windows
echo.

set REPO=winkeep/winkeep
set INSTALL_DIR=%LOCALAPPDATA%\winkeep
set BINARY=winkeep.exe

REM Detect architecture
if "%PROCESSOR_ARCHITECTURE%"=="AMD64" (
    set ARCH=amd64
) else (
    set ARCH=386
)
echo [1/4] Detected architecture: %ARCH%

REM Get latest version
echo [2/4] Fetching latest version...
for /f "tokens=2 delims=:, " %%a in ('curl -s "https://api.github.com/repos/%REPO%/releases/latest" ^| findstr "tag_name"') do (
    set VERSION=%%~a
)
if "%VERSION%"=="" (
    echo        Failed to fetch version, using "latest"
    set VERSION=latest
) else (
    echo        Latest version: %VERSION%
)

REM Download
echo [3/4] Downloading WinKeep...
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
set FILENAME=winkeep_windows_%ARCH%.exe
set DOWNLOAD_URL=https://github.com/%REPO%/releases/download/%VERSION%/%FILENAME%

curl -L -o "%INSTALL_DIR%\%BINARY%" "%DOWNLOAD_URL%" 2>nul
if errorlevel 1 (
    echo        GitHub download failed, trying SourceForge...
    curl -L -o "%INSTALL_DIR%\%BINARY%" "https://sourceforge.net/projects/winkeep/files/%VERSION%/%FILENAME%/download" 2>nul
    if errorlevel 1 (
        echo        Download failed!
        echo        Please download manually from: https://sourceforge.net/projects/winkeep/
        exit /b 1
    )
)

REM Add to PATH
echo [4/4] Adding to PATH...
echo %PATH% | findstr /C:"%INSTALL_DIR%" >nul
if errorlevel 1 (
    setx PATH "%PATH%;%INSTALL_DIR%" >nul 2>&1
    echo        Added to PATH
) else (
    echo        Already in PATH
)

echo.
echo   Installation complete!
echo.
"%INSTALL_DIR%\%BINARY%" version
echo.
echo   Usage:
echo     winkeep run -- python server.py
echo     winkeep list
echo     winkeep logs ^<id^>
echo.
echo   Restart your terminal to use 'winkeep' command.
echo.

endlocal
