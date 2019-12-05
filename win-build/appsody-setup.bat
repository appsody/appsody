@echo Checking prerequisites...
@echo off

docker ps >nul 2>nul
IF %ERRORLEVEL% EQU 0 goto CheckDocker
@ECHO [Warning] Docker not running or not installed
@ECHO [Warning] Appsody could not check if Docker has the appropriate write permissions on your file system
goto AfterDocker
:CheckDocker
@ECHO Checking whether Docker has the appropriate write permissions on your file system 
docker run --rm -it -v "%USERPROFILE%\.appsody":/data alpine /bin/sh -c "mkdir /data/test-write-permission && echo Success - Docker has write permissions; rmdir /data/test-write-permission"
IF %ERRORLEVEL% EQU 0 goto AfterDocker
@ECHO [Warning] Docker may not have the appropriate write permissions on your file system 
@ECHO [Warning] Please refer to https://appsody.dev/docs/docker-windows-aad/ for more info on this
:AfterDocker
@echo Adding %~dp0 to your Path environment variable if not already present....
@echo off

setx APPSODY_PATH "%~dp0
set APPSODY_PATH=%~dp0
set lastPathChar=%PATH:~-1%
if NOT "%lastPathChar%" == ";" set "PATH=%PATH%;"

for /F "skip=2 tokens=1,2*" %%N in ('%SystemRoot%\System32\reg.exe query "HKCU\Environment" /v "Path" 2^>nul') do if /I "%%N" == "Path" set "UserPath=%%P"
IF DEFINED UserPath goto UserPathRead
REM If no user path env var is set, we just set it to %APPSODY_PATH%
setx PATH "%%APPSODY_PATH%%

goto :SkipSetx

:UserPathRead
REM If the user path env var is already populated, add %APPSODY_PATH%, unless it is there already
SET UserPathTest=%UserPath:APPSODY_PATH=NONE%
REM echo UserPathTest = %UserPathTest%
REM echo UserPath = %UserPath%
if NOT "%UserPathTest%" == "%UserPath%" goto SkipSetx

set lastUserPathChar=%UserPath:~-1%
if NOT "%lastUserPathChar%" == ";" set "UserPath=%UserPath%;"
setx PATH "%UserPath%%%APPSODY_PATH%%
:SkipSetx
REM Append the value of %APPSODY_PATH% to the PATH env var, unless it's already there
CALL SET TestPath=%%PATH:%APPSODY_PATH%=NONE%%
REM echo TestPath = %TestPath%
REM echo PATH = %PATH%
if NOT "%TestPath%" == "%PATH%" goto :done
set PATH=%PATH%%APPSODY_PATH%
:done
REM Checking if the docker user has the correct write permissions

@echo Done - enjoy appsody!