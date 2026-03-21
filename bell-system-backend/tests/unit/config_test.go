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

func TestLoadConfig_InvalidTOML(t *testing.T) {
	dir := t.TempDir()
	content := `[server
	this is not valid toml!!!
`
	err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644)
	require.NoError(t, err)

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg, err := config.LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
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

func TestLoadConfig_UnmarshalError(t *testing.T) {
	dir := t.TempDir()
	// TOML maps allowedOrigins to a nested table instead of a string slice,
	// which causes mapstructure to fail when decoding into []string.
	content := `
[server]
port = 8080

[cors]
[cors.allowedOrigins]
nested = "table"
`
	err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644)
	require.NoError(t, err)

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg, err := config.LoadConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to unmarshal config")
}

func TestLoadConfig_CORSDefaults(t *testing.T) {
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

	assert.Empty(t, cfg.CORS.AllowedOrigins)
	assert.Equal(t, 0, cfg.CORS.MaxAge)
}

func TestLoadConfig_CORSCustom(t *testing.T) {
	dir := t.TempDir()
	content := `
[server]
port = 8080

[cors]
allowedOrigins = ["http://example.com", "http://app.example.com"]
maxAge = 600
`
	err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644)
	require.NoError(t, err)

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, []string{"http://example.com", "http://app.example.com"}, cfg.CORS.AllowedOrigins)
	assert.Equal(t, 600, cfg.CORS.MaxAge)
}

func TestLoadConfig_WebSocketDefaults(t *testing.T) {
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

	assert.Equal(t, 0, cfg.WebSocket.PingInterval)
	assert.Equal(t, 0, cfg.WebSocket.PongTimeout)
	assert.Equal(t, 0, cfg.WebSocket.MaxMessageSize)
}

func TestLoadConfig_WebSocketCustom(t *testing.T) {
	dir := t.TempDir()
	content := `
[server]
port = 8080

[websocket]
pingInterval = 45
pongTimeout = 15
maxMessageSize = 1024
`
	err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(content), 0644)
	require.NoError(t, err)

	origDir, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { os.Chdir(origDir) })

	cfg, err := config.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, 45, cfg.WebSocket.PingInterval)
	assert.Equal(t, 15, cfg.WebSocket.PongTimeout)
	assert.Equal(t, 1024, cfg.WebSocket.MaxMessageSize)
}
