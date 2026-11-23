@echo off
title Orbek - build Vencorda i restart Discorda
cd /d "%~dp0..\.."
echo === Folder: %CD%
echo.
echo === Buduje Vencorda (chwile potrwa)...
call pnpm build
if errorlevel 1 (
    echo.
    echo !!! BLAD BUILDU - zrob screena tego okna i wyslij do Claude'a
    pause
    exit /b 1
)
echo.
echo === Build OK. Zamykam Discorda...
taskkill /F /IM Discord.exe >nul 2>&1
taskkill /F /IM DiscordPTB.exe >nul 2>&1
taskkill /F /IM DiscordCanary.exe >nul 2>&1
timeout /t 2 /nobreak >nul
echo === Uruchamiam Discorda...
if exist "%LOCALAPPDATA%\Discord\Update.exe" (
    start "" "%LOCALAPPDATA%\Discord\Update.exe" --processStart Discord.exe
) else (
    echo Nie znalazlem Discorda w %LOCALAPPDATA%\Discord - odpal go recznie.
)
echo.
echo === Gotowe. Mozesz zamknac to okno.
timeout /t 5
