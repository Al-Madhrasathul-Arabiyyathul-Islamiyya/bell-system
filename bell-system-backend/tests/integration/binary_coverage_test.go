//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeTestConfig writes a config.toml in dir pointing at the testcontainers DB.
func writeTestConfig(t *testing.T, dir string, port int) string {
	t.Helper()
	audioDir := filepath.Join(dir, "audio")
	require.NoError(t, os.MkdirAll(audioDir, 0o755))

	content := fmt.Sprintf(`[server]
port = %d
environment = "test"
readTimeout = 5
writeTimeout = 5

[database]
driver = "sqlserver"
host = "%s"
port = %s
user = "sa"
password = "%s"
name = "%s"
sslmode = "disable"

[jwt]
secret = "%s"
expiresIn = 180

[storage]
audioDir = "%s"

[cors]
allowedOrigins = ["http://localhost:3000"]
maxAge = 300

[websocket]
pingInterval = 30
pongTimeout = 10
maxMessageSize = 512

[scheduler]
checkInterval = 60
enabled = false
timezone = "UTC"
`, port, testDBHost, testDBPort, saPassword, testDBName, testJWTSecret, audioDir)

	path := filepath.Join(dir, "config.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

// TestBinaryCoverage_Server builds cmd/server with -cover, starts it against the
// testcontainers DB, verifies it responds to /status, then shuts it down gracefully
// to collect coverage data.
func TestBinaryCoverage_Server(t *testing.T) {
	if testDBHost == "" || testDBPort == "" {
		t.Skip("testcontainers DB not available")
	}

	projectRoot, err := filepath.Abs("../..")
	require.NoError(t, err)

	tmpDir := t.TempDir()
	coverDir := filepath.Join(tmpDir, "covdata")
	require.NoError(t, os.MkdirAll(coverDir, 0o755))

	// Build server with coverage instrumentation
	binName := "server-cover"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(tmpDir, binName)

	build := exec.Command("go", "build", "-cover", "-o", binPath, "./cmd/server")
	build.Dir = projectRoot
	out, err := build.CombinedOutput()
	require.NoError(t, err, "build failed: %s", string(out))

	// Write config with a free port
	port := findFreePort(t)
	writeTestConfig(t, tmpDir, port)

	// Start the server
	cmd := exec.Command(binPath)
	cmd.Dir = tmpDir
	cmd.Env = append(os.Environ(), "GOCOVERDIR="+coverDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start())

	// Wait for server to be ready
	serverURL := fmt.Sprintf("http://localhost:%d", port)
	ready := waitForServer(t, serverURL+"/status", 10*time.Second)
	if !ready {
		cmd.Process.Kill()
		t.Fatal("server did not become ready in time")
	}

	// Verify /status responds
	resp, err := http.Get(serverURL + "/status")
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var statusResp map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &statusResp))
	assert.Equal(t, "ok", statusResp["status"])

	// Graceful shutdown
	cmd.Process.Signal(os.Interrupt)

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case <-done:
		// Process exited
	case <-time.After(15 * time.Second):
		cmd.Process.Kill()
		t.Fatal("server did not shut down in time")
	}

	// Verify coverage data was written
	entries, err := os.ReadDir(coverDir)
	require.NoError(t, err)
	assert.NotEmpty(t, entries, "expected coverage data files in %s", coverDir)

	// Convert to text format and write to the project's coverage directory
	// so that scripts/coverage.sh can merge it into the final profile.
	textOut := filepath.Join(tmpDir, "server.out")
	convert := exec.Command("go", "tool", "covdata", "textfmt", "-i="+coverDir, "-o="+textOut)
	out, err = convert.CombinedOutput()
	if err != nil {
		t.Logf("covdata convert: %s", string(out))
		return
	}

	info, statErr := os.Stat(textOut)
	if statErr != nil {
		return
	}
	t.Logf("server binary coverage profile: %d bytes", info.Size())

	// Copy profile to coverage/binary.out for merging by coverage.sh
	destDir := filepath.Join(projectRoot, "coverage")
	if mkErr := os.MkdirAll(destDir, 0o755); mkErr != nil {
		t.Logf("could not create coverage dir: %v", mkErr)
		return
	}
	src, readErr := os.ReadFile(textOut)
	if readErr != nil {
		t.Logf("could not read text profile: %v", readErr)
		return
	}
	destPath := filepath.Join(destDir, "binary.out")
	if writeErr := os.WriteFile(destPath, src, 0o644); writeErr != nil {
		t.Logf("could not write binary.out: %v", writeErr)
		return
	}
	t.Logf("binary coverage profile written to %s", destPath)
}

// findFreePort returns an available TCP port.
func findFreePort(t *testing.T) int {
	t.Helper()
	l, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "localhost:0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

// waitForServer polls the URL until it returns 200 or timeout expires.
func waitForServer(t *testing.T, url string, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}
