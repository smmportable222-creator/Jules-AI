@echo off
echo Building Pavuk Crawler...
go build -o pavuk5.exe pavuk5_refactored.go web_gui.go
if %errorlevel% neq 0 (
    echo Build failed. Check errors above.
    pause
    goto :eof
)
echo Build successful!
echo Starting Web GUI on http://localhost:8080...
start http://localhost:8080
pavuk5.exe -web
pause
