@echo off
setlocal
cd /d "%~dp0"
if not defined VS_VCVARS set "VS_VCVARS=C:\Program Files (x86)\Microsoft Visual Studio\2019\Community\VC\Auxiliary\Build\vcvarsall.bat"
if not defined QT_ROOT set "QT_ROOT=F:\qt\msvc2019-x86-5.15.18"
if not defined GO_EXE set "GO_EXE=F:\go\bin\go.exe"
call "%VS_VCVARS%" x86
if errorlevel 1 exit /b 1
set "PATH=%QT_ROOT%\bin;%PATH%"
set "BUILD=%~dp0build\msvc2019-x86"
set "DIST=%~dp0dist"
cmake -S "%~dp0." -B "%BUILD%" -G "Visual Studio 16 2019" -A Win32 -DCMAKE_PREFIX_PATH="%QT_ROOT%"
if errorlevel 1 exit /b 1
cmake --build "%BUILD%" --config Release -- /m
if errorlevel 1 exit /b 1
cmake --install "%BUILD%" --config Release --prefix "%DIST%"
if errorlevel 1 exit /b 1
"%QT_ROOT%\bin\windeployqt.exe" --release --compiler-runtime --no-translations "%DIST%\1bot-control.exe"
if errorlevel 1 exit /b 1
set "CGO_ENABLED=0"
set "GOARCH=386"
set "GOOS=windows"
pushd "%~dp0.."
"%GO_EXE%" build -o "%DIST%\gateway.exe" ./cmd/gateway
if errorlevel 1 goto :failed
for %%P in (codebuddy codex zcode) do (
  "%GO_EXE%" build -o "%DIST%\plugin-%%P.exe" ./plugin/cmd/plugin-%%P
  if errorlevel 1 goto :failed
)
popd
echo [OK] %DIST%\1bot-control.exe
echo HTTPS updates use libcurl and OpenSSL with the deployed CA bundle.
exit /b 0
:failed
popd
exit /b 1
