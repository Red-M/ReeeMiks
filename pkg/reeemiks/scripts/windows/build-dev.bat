@ECHO OFF

ECHO Building reeemiks (development)...

REM set repo root in relation to script path to avoid cwd dependency
SET "REEEMIKS_ROOT=%~dp0..\..\..\.."

REM shove git commit, version tag into env
for /f "delims=" %%a in ('git rev-list -1 --abbrev-commit HEAD') do @set GIT_COMMIT=%%a
for /f "delims=" %%a in ('git describe --tags --always') do @set VERSION_TAG=%%a
set BUILD_TYPE=dev
ECHO Embedding build-time parameters:
ECHO - gitCommit %GIT_COMMIT%
ECHO - versionTag %VERSION_TAG%
ECHO - buildType %BUILD_TYPE%

go build -o "%REEEMIKS_ROOT%\reeemiks-dev.exe" -ldflags "-X main.gitCommit=%GIT_COMMIT% -X main.versionTag=%VERSION_TAG% -X main.buildType=%BUILD_TYPE%" "%REEEMIKS_ROOT%\pkg\reeemiks\cmd"
if %ERRORLEVEL% NEQ 0 GOTO BUILDERROR
ECHO Done.
GOTO DONE

:BUILDERROR
ECHO Failed to build reeemiks in development mode! See above output for details.
EXIT /B 1

:DONE
