package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun_ConfigError(t *testing.T) {
	// Run from a temp dir with no config.toml
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	err := run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

func TestRun_BadTimezone(t *testing.T) {
	dir := t.TempDir()
	content := `
[server]
port = 0
environment = "test"
readTimeout = 5
writeTimeout = 5

[database]
driver = "sqlserver"
host = "localhost"
port = 1433
user = "sa"
password = "test"
name = "testdb"

[jwt]
secret = "test-secret"
expiresIn = 60

[storage]
audioDir = "` + filepath.ToSlash(filepath.Join(dir, "audio")) + `"

[scheduler]
timezone = "Invalid/Not_A_Timezone"
checkInterval = 30
enabled = false
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	err := run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load timezone")
}

func TestRun_DatabaseError(t *testing.T) {
	dir := t.TempDir()
	content := `
[server]
port = 0
environment = "test"
readTimeout = 5
writeTimeout = 5

[database]
driver = "sqlserver"
host = "192.0.2.1"
port = 1
user = "sa"
password = "test"
name = "testdb"

[jwt]
secret = "test-secret"
expiresIn = 60

[storage]
audioDir = "` + filepath.ToSlash(filepath.Join(dir, "audio")) + `"

[scheduler]
timezone = "UTC"
checkInterval = 30
enabled = false
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644))

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	err := run(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}
