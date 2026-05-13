@echo off
cd ..\web && npm run build && cd ..\src && go build -ldflags="-s -w" -o v2rayN-Go.exe . && echo successful build && pause