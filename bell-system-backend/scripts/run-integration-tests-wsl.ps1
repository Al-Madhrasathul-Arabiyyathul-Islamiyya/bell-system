$projectPath = "/mnt/d/Stuff/projects/bell-system/bell-system-backend"

wsl -e bash -lc "cd $projectPath && GODEBUG=x509negativeserial=1 go test -tags integration -v -count=1 -timeout 300s ./tests/integration/ 2>&1"
exit $LASTEXITCODE