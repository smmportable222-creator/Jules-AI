#!/bin/bash
echo "Building Pavuk Crawler..."
go build -o pavuk5 pavuk5_refactored.go web_gui.go
if [ $? -ne 0 ]; then
    echo "Build failed. Check errors above."
    read -p "Press enter to continue"
    # End
else
    echo "Build successful!"
    echo "Starting Web GUI on http://localhost:8080..."

    if which xdg-open > /dev/null
    then
      xdg-open http://localhost:8080 &
    elif which open > /dev/null
    then
      open http://localhost:8080 &
    fi

    ./pavuk5 -web
fi
