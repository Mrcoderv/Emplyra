@echo off
setlocal
title Emplyra Launcher
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0launch.ps1"
if errorlevel 1 (
  echo.
  echo Emplyra failed to start. Review the message above.
)
pause
