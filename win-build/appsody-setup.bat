@echo Checking prerequisites...
@echo off

docker ps >nul 2>nul
IF %ERRORLEVEL% NEQ 0 ECHO [Warning] Docker not running or not installed 
@echo Adding %~dp0 to your Path environment variable....
@echo off
setx APPSODY_PATH "%~dp0
set APPSODY_PATH=%~dp0


set lastPathChar=%PATH:~-1%

if NOT "%lastPathChar%" == ";" set "PATH=%PATH%;"
for /F "skip=2 tokens=1,2*" %%N in ('%SystemRoot%\System32\reg.exe query "HKCU\Environment" /v "Path" 2^>nul') do if /I "%%N" == "Path" set "UserPath=%%P"
IF DEFINED UserPath goto UserPathRead
setx PATH "%%APPSODY_PATH%%
set PATH=%PATH%%APPSODY_PATH%
goto :done

:UserPathRead
SET UserPathTest=%UserPath:APPSODY_PATH=NONE%
REM echo UserPathTest = %UserPathTest%
REM echo UserPath = %UserPath%
if NOT "%UserPathTest%" == "%UserPath%" goto SkipSetx
echo doing setx
set lastUserPathChar=%UserPath:~-1%
if NOT "%lastUserPathChar%" == ";" set "UserPath=%UserPath%;"
setx PATH "%UserPath%%%APPSODY_PATH%%
:SkipSetx
set PATH=%PATH%%APPSODY_PATH%
:done

@echo Done - enjoy appsody!