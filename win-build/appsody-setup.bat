@echo Checking prerequisites...
@echo off
docker ps >nul 2>nul
IF %ERRORLEVEL% NEQ 0 ECHO [Warning] Docker not running or not installed 
@echo Adding %~dp0 to your Path environment variable....
@echo off
set lastPathChar=%PATH:~-1%

if NOT "%lastPathChar%" == ";" set "PATH=%PATH%;"
for /F "skip=2 tokens=1,2*" %%N in ('%SystemRoot%\System32\reg.exe query "HKCU\Environment" /v "Path" 2^>nul') do if /I "%%N" == "Path" call set "UserPath=%%P" & goto UserPathRead
setx PATH %~dp0 > nul 2>&1
set PATH=%PATH%%~dp0
goto :done

:UserPathRead
set lastUserPathChar=%UserPath:~-1%
if NOT "%lastUserPathChar%" == ";" set "UserPath=%UserPath%;"
setx PATH %UserPath%%~dp0 > nul 2>&1
set PATH=%PATH%%~dp0
:done
REM Removed CLI initialization - no longer needed after flattening
REM @echo Initializing the Command Line Interface...
REM appsody init > nul 2>&1
@echo Done - enjoy appsody!