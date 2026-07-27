@echo off
setlocal
cd /d "%~dp0"
set "MC=mapcreator"
set "OBF=mapcreator\Whitehall-farringdon.obf"
set "OUT=osmand-junction-dump.json"
set "JAVA_OPTS=-Xms512M -Xmx2G"

echo Compiling DumpOsmandJunction.java ...
javac -cp "mapcreator\OsmAndMapCreator.jar;mapcreator\lib\*" -d . DumpOsmandJunction.java
if errorlevel 1 exit /b 1

echo Running dump ...
java %JAVA_OPTS% -cp ".;mapcreator\OsmAndMapCreator.jar;mapcreator\lib\*" DumpOsmandJunction "%OBF%" "%OUT%"
exit /b %ERRORLEVEL%
