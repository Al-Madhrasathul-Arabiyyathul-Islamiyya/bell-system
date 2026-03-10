$RepoRoot = git rev-parse --show-toplevel
$HooksDir = "$RepoRoot/bell-system-backend/scripts/hooks"

git config core.hooksPath $HooksDir

Write-Host "Git hooks configured. Hooks path set to: $HooksDir"
