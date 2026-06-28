@echo off
title RAG Knowledge Base Launcher
echo ==============================
echo   RAG Knowledge Base - Start
echo ==============================
echo.

echo [1/4] Starting Redis...
start "Redis" /MIN "D:\ctfshow\Redis-x64-3.0.504\redis-server.exe"
echo   OK - Redis started

echo [2/4] Starting Qdrant...
start "Qdrant" /MIN "D:\ctfshow\qdrant-x86_64-pc-windows-msvc\qdrant.exe"
echo   OK - Qdrant started

echo [3/4] Waiting for services to be ready (5s)...
timeout /t 5 /nobreak >nul
echo   OK

echo [3/3] All services started.
echo.
echo Redis and Qdrant are running in the background.
echo You can now run: go run .
echo.
