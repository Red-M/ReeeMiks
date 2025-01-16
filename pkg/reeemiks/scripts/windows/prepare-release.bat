@ECHO OFF

IF "%1"=="" GOTO NOTAG

ECHO Preparing release (%1)...
ECHO.

git tag --delete %1 >NUL 2>&1
git tag %1

REM set windows scripts dir root in relation to script path to avoid cwd dependency
SET "WIN_SCRIPTS_ROOT=%~dp0"

CALL "%WIN_SCRIPTS_ROOT%build-dev.bat"

ECHO.

CALL "%WIN_SCRIPTS_ROOT%build-release.bat"

REM make this next part nicer by setting the repo root
SET "REEEMIKS_ROOT=%WIN_SCRIPTS_ROOT%..\..\..\.."
PUSHD "%REEEMIKS_ROOT%"
SET "REEEMIKS_ROOT=%CD%"
POPD

MKDIR "%REEEMIKS_ROOT%\releases\%1" 2> NUL
MOVE /Y "%REEEMIKS_ROOT%\reeemiks-release.exe" "%REEEMIKS_ROOT%\releases\%1\reeemiks.exe" >NUL 2>&1
MOVE /Y "%REEEMIKS_ROOT%\reeemiks-dev.exe" "%REEEMIKS_ROOT%\releases\%1\reeemiks-debug.exe" >NUL 2>&1
COPY /Y "%REEEMIKS_ROOT%\pkg\reeemiks\scripts\misc\default-config.yaml" "%REEEMIKS_ROOT%\releases\%1\config.yaml" >NUL 2>&1
COPY /Y "%REEEMIKS_ROOT%\pkg\reeemiks\scripts\misc\release-notes.txt" "%REEEMIKS_ROOT%\releases\%1\notes.txt" >NUL 2>&1

ECHO.
ECHO Release binaries created in %REEEMIKS_ROOT%\releases\%1
ECHO Opening release directory and notes for editing.
ECHO When you're done, run "git push origin %1" and draft the release on GitHub.

START explorer.exe "%REEEMIKS_ROOT%\releases\%1"
START notepad.exe "%REEEMIKS_ROOT%\releases\%1\notes.txt"

GOTO DONE

:NOTAG
ECHO usage: %0 ^<tag name^>    (use semver i.e. v0.9.3)
GOTO DONE

:DONE
