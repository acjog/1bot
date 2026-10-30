@echo off
setlocal
cd /d "%~dp0"
if not defined QT_ROOT set "QT_ROOT=F:\qt\msvc2019-x86-5.15.18"
"%QT_ROOT%\bin\lupdate.exe" -no-obsolete -source-language en_US main.cpp modelspage.cpp pluginspage.cpp accesscard.cpp updater.cpp wiring.cpp localization.cpp -ts zh_CN.ts en_US.ts ja_JP.ts ko_KR.ts zh_TW.ts
if errorlevel 1 exit /b 1
"%QT_ROOT%\bin\lrelease.exe" zh_CN.ts en_US.ts ja_JP.ts ko_KR.ts zh_TW.ts
exit /b %errorlevel%
