package unit_test

import (
	"os"
	"path/filepath"
	"testing"

	"arabiyya.edu.mv/bell-system-backend/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ValidFile(t *testing.T) {
	dir := t.TempDir()
	content := `
[server]
port = 9090
environment = "testing"
readTimeout = 10
writeTimeout = 10

[database]
driver = "sqlserver"
host = "localhost"
port = 1433
user = "sa"
password = "testpass"
name = "TestDB"
sslMode = "disable"

[jwt]
secret = "test-secret-key"
expiresIn = 60

[storage]
audioDir = "/tmp/audio"
`
	err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644)
	require.NoError(t, err)

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 9090, cfg.Server.Port)
	assert.Equal(t, "testing", cfg.Server.Environment)
	assert.Equal(t, 10, cfg.Server.ReadTimeout)
	assert.Equal(t, 10, cfg.Server.WriteTimeout)

	assert.Equal(t, "sqlserver", cfg.Database.Driver)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, 1433, cfg.Database.Port)
	assert.Equal(t, "sa", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "TestDB", cfg.Database.Name)
	assert.Equal(t, "disable", cfg.Database.SSLMode)

	assert.Equal(t, "test-secret-key", cfg.JWT.Secret)
	assert.Equal(t, 60, cfg.JWT.ExpiresIn)

	assert.Equal(t, "/tmp/audio", cfg.Storage.AudioDir)
}

func TestLoadConfig_MissingFile_ReturnsError(t *testing.T) {
	dir := t.TempDir()

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg, err := config.LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoadConfig_PartialConfig_DefaultsToZeroValues(t *testing.T) {
	dir := t.TempDir()
	content := `
[server]
port = 8080
`
	err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644)
	require.NoError(t, err)

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 8080, cfg.Server.Port)
	assert.Equal(t, "", cfg.Server.Environment)
	assert.Equal(t, 0, cfg.Database.Port)
	assert.Equal(t, "", cfg.JWT.Secret)
	assert.Equal(t, 0, cfg.JWT.ExpiresIn)
	assert.Equal(t, "", cfg.Storage.AudioDir)
}
