@echo off
setlocal

rem ============================================================
rem  1bot build script
rem  Produces: gateway.exe, plugin-codebuddy.exe, plugin-codex.exe
rem  Usage:    build.bat          -> build both
rem            build.bat clean    -> delete exe then build
rem ============================================================

rem --- locate Go: prefer F:\go\bin\go.exe, fall back to PATH ---
set "GO=F:\go\bin\go.exe"
if not exist "%GO%" set "GO=go"

rem --- pure Go, no CGO (per DESIGN.md ADR-6) ---
set "CGO_ENABLED=0"

rem --- run from this script's directory ---
cd /d "%~dp0"

rem --- optional clean ---
if /i "%~1"=="clean" (
    echo cleaning old binaries ...
    del /q gateway.exe plugin-codebuddy.exe plugin-codex.exe 2>nul
)

echo.
echo [1/3] building gateway.exe ...
"%GO%" build -trimpath -o gateway.exe ./cmd/gateway
if errorlevel 1 (
    echo [FAIL] gateway build failed.
    exit /b 1
)

echo [2/3] building plugin-codebuddy.exe ...
"%GO%" build -trimpath -o plugin-codebuddy.exe ./cmd/plugin-codebuddy
if errorlevel 1 (
    echo [FAIL] plugin-codebuddy build failed.
    exit /b 1
)

echo [3/3] building plugin-codex.exe ...
"%GO%" build -trimpath -o plugin-codex.exe ./cmd/plugin-codex
if errorlevel 1 (
    echo [FAIL] plugin-codex build failed.
    exit /b 1
)

echo.
echo build ok. binaries:
dir /b gateway.exe plugin-codebuddy.exe plugin-codex.exe 2>nul
echo.
echo run:
echo   gateway.exe -plugin .\plugin-codebuddy.exe -model gpt-6.1-sol -y
echo   gateway.exe -plugin .\plugin-codex.exe -model gpt-5.4
echo.
exit /b 0
