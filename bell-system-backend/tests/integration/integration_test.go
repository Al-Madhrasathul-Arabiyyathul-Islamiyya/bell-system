//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/config"
	"arabiyya.edu.mv/bell-system-backend/internal/database"
	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/router"
	"arabiyya.edu.mv/bell-system-backend/internal/services"
	ws "arabiyya.edu.mv/bell-system-backend/internal/websocket"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/pkg/logger"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testJWTSecret = "integration-test-jwt-secret"
	testDBName    = "BellScheduleTestDB"
	saPassword    = "hEsyPE5v32R&Mb"
)

var (
	testServer *httptest.Server
	testDB     *sql.DB
	testHub    *ws.Hub
	hubDone    chan struct{}

	// Exposed for binary coverage tests
	testDBHost string
	testDBPort string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "mcr.microsoft.com/mssql/server:2025-CU2-GDR1-ubuntu-22.04",
		ExposedPorts: []string{"1433/tcp"},
		Env: map[string]string{
			"ACCEPT_EULA":       "Y",
			"MSSQL_SA_PASSWORD": saPassword,
		},
		WaitingFor: wait.ForAll(
			wait.ForListeningPort("1433/tcp"),
			wait.ForLog("Recovery is complete."),
		).WithStartupTimeout(5 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start mssql container: %v\n", err)
		os.Exit(1)
	}
	defer container.Terminate(ctx)

	// Get mapped port
	mappedPort, err := container.MappedPort(ctx, "1433")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get mapped port: %v\n", err)
		os.Exit(1)
	}

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get host: %v\n", err)
		os.Exit(1)
	}

	testDBHost = host
	testDBPort = mappedPort.Port()

	connStr := fmt.Sprintf("sqlserver://sa:%s@%s:%s?encrypt=disable&TrustServerCertificate=true",
		saPassword, host, mappedPort.Port())

	masterDB, err := sql.Open("sqlserver", connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open master db: %v\n", err)
		os.Exit(1)
	}

	if _, err := masterDB.ExecContext(ctx, "CREATE DATABASE "+testDBName); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create test database: %v\n", err)
		os.Exit(1)
	}
	masterDB.Close()

	// Reconnect to test database
	testConnStr := fmt.Sprintf("sqlserver://sa:%s@%s:%s?database=%s&encrypt=disable&TrustServerCertificate=true",
		saPassword, host, mappedPort.Port(), testDBName)

	testDB, err = sql.Open("sqlserver", testConnStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open test db: %v\n", err)
		os.Exit(1)
	}

	// Wait for connection
	for i := 0; i < 30; i++ {
		if err := testDB.PingContext(ctx); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	// Run migrations
	if err := runMigrations(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	// Wire the full application stack
	log, err := logger.New("test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	// Repositories
	userRepo := database.NewUserRepository(testDB, log)
	sessionRepo := database.NewSessionRepository(testDB, log)
	scheduleDayRepo := database.NewScheduleDayRepository(testDB, log)
	audioFileRepo := database.NewSystemAudioFileRepository(testDB, log)
	scheduleItemRepo := database.NewScheduleItemRepository(testDB, log, scheduleDayRepo, sessionRepo, audioFileRepo)

	// Services
	tokenSvc := services.NewTokenService(testJWTSecret, 180)
	hasher := services.NewPasswordHasher()

	tmpDir, err := os.MkdirTemp("", "bell-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	fileStore, err := services.NewLocalFileStorage(tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create file storage: %v\n", err)
		os.Exit(1)
	}

	// WebSocket
	testHub = ws.NewHub(log)
	hubDone = make(chan struct{})
	go testHub.Run(hubDone)

	notifier := ws.NewNotifier(testHub)
	wsCfg := config.WebSocketConfig{
		PingInterval:   30,
		PongTimeout:    10,
		MaxMessageSize: 512,
	}
	wsHandler := ws.NewHandler(testHub, tokenSvc, log, wsCfg)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, tokenSvc, hasher)
	userHandler := handlers.NewUserHandler(userRepo, hasher)
	sessionHandler := handlers.NewSessionHandler(sessionRepo)
	scheduleHandler := handlers.NewScheduleHandler(scheduleItemRepo, scheduleDayRepo, sessionRepo, notifier)
	audioHandler := handlers.NewAudioHandler(audioFileRepo, notifier)
	audioHandler.FileStorage = fileStore

	stateRepo := database.NewSystemStateRepository(testDB, log)

	// Router (mirrors cmd/server/main.go)
	systemHandler := handlers.NewSystemHandler(stateRepo, nil, notifier)
	apiRouter := router.New(tokenSvc, authHandler, userHandler, sessionHandler, scheduleHandler, audioHandler, systemHandler)

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Get("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// API documentation (Scalar) — mirrors cmd/server/main.go
	docsHTML, err := scalargo.NewV2(
		scalargo.WithSpecDir("../../api"),
		scalargo.WithBaseFileName("openapi.yaml"),
		scalargo.WithDarkMode(),
		scalargo.WithOperationsSorter(scalargo.SorterMethod),
		scalargo.WithTheme(scalargo.ThemeKepler),
		scalargo.WithPersistAuth(true),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize API docs: %v\n", err)
		os.Exit(1)
	}
	r.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, docsHTML)
	})

	r.Handle("/ws", wsHandler)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(time.Second * 30))
		r.Mount("/", apiRouter)
	})

	testServer = httptest.NewServer(r)

	// Run tests
	code := m.Run()

	// Cleanup
	close(hubDone)
	<-testHub.Done()
	testServer.Close()

	dropCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Reconnect to master to drop test DB
	masterDB2, err := sql.Open("sqlserver", connStr)
	if err == nil {
		masterDB2.ExecContext(dropCtx, "ALTER DATABASE "+testDBName+" SET SINGLE_USER WITH ROLLBACK IMMEDIATE")
		masterDB2.ExecContext(dropCtx, "DROP DATABASE "+testDBName)
		masterDB2.Close()
	}

	testDB.Close()
	container.Terminate(dropCtx)

	os.Exit(code)
}

func runMigrations(ctx context.Context) error {
	schemaSQL, err := os.ReadFile("../../migrations/001_create_tables_up.sql")
	if err != nil {
		return fmt.Errorf("read schema migration: %w", err)
	}
	if _, err := testDB.ExecContext(ctx, string(schemaSQL)); err != nil {
		return fmt.Errorf("exec schema migration: %w", err)
	}

	seedSQL, err := os.ReadFile("../../migrations/002_seed_data_up.sql")
	if err != nil {
		return fmt.Errorf("read seed migration: %w", err)
	}
	if _, err := testDB.ExecContext(ctx, string(seedSQL)); err != nil {
		return fmt.Errorf("exec seed migration: %w", err)
	}

	stateSQL, err := os.ReadFile("../../migrations/003_create_system_state_up.sql")
	if err != nil {
		return fmt.Errorf("read system state migration: %w", err)
	}
	if _, err := testDB.ExecContext(ctx, string(stateSQL)); err != nil {
		return fmt.Errorf("exec system state migration: %w", err)
	}

	return nil
}

// cleanAndSeed truncates all tables in FK-safe order and re-inserts seed data.
func cleanAndSeed(t *testing.T) {
	t.Helper()
	ctx := context.Background()

	tables := []string{"ScheduleDays", "ScheduleItems", "SystemAudioFiles", "Sessions", "Users"}
	for _, table := range tables {
		_, err := testDB.ExecContext(ctx, "DELETE FROM "+table)
		require.NoError(t, err, "failed to clean table %s", table)
	}

	// Reset system state to active
	_, err := testDB.ExecContext(ctx, "UPDATE SystemState SET Value = 'active', UpdatedAt = GETDATE() WHERE [Key] = 'system_state'")
	require.NoError(t, err, "failed to reset system state")

	seedSQL, err := os.ReadFile("../../migrations/002_seed_data_up.sql")
	require.NoError(t, err, "failed to read seed SQL")
	_, err = testDB.ExecContext(ctx, string(seedSQL))
	require.NoError(t, err, "failed to re-seed data")
}

// loginAs logs in with the given credentials and returns a JWT token.
func loginAs(t *testing.T, username, password string) string {
	t.Helper()

	body := fmt.Sprintf(`{"username":%q,"password":%q}`, username, password)
	resp := doRequest(t, http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body), "")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "login failed for %s", username)

	var result map[string]any
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	token, ok := result["token"].(string)
	require.True(t, ok, "token not found in login response")
	return token
}

// doRequest builds and executes an HTTP request against the test server.
func doRequest(t *testing.T, method, path string, body io.Reader, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, testServer.URL+path, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// readJSON decodes the response body into the given target.
func readJSON(t *testing.T, resp *http.Response, target any) {
	t.Helper()
	defer resp.Body.Close()
	err := json.NewDecoder(resp.Body).Decode(target)
	require.NoError(t, err)
}

// getDayOfWeek returns the current day of week as 1=Sunday..7=Saturday.
func getDayOfWeek() int {
	return int(time.Now().Weekday()) + 1
}

// adminToken creates a fresh admin user directly in the DB (bypassing the API auth gate),
// then logs in via the API to obtain a valid JWT. Use this to authenticate protected requests.
func adminToken(t *testing.T) string {
	t.Helper()

	const username = "integration_admin"
	const password = "integrationpass123"

	hasher := services.NewPasswordHasher()
	hash, err := hasher.Hash(password)
	require.NoError(t, err)

	_, err = testDB.ExecContext(context.Background(),
		`MERGE Users AS target
		 USING (SELECT @p1 AS Username) AS source ON target.Username = source.Username
		 WHEN MATCHED THEN UPDATE SET PasswordHash = @p2, Role = 'admin'
		 WHEN NOT MATCHED THEN INSERT (Username, PasswordHash, Role) VALUES (@p1, @p2, 'admin');`,
		username, hash)
	require.NoError(t, err, "failed to upsert integration admin user")

	return loginAs(t, username, password)
}

// doJSONAPIRequest builds and executes an HTTP request with JSON:API content type.
func doJSONAPIRequest(t *testing.T, method, path string, body io.Reader, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, testServer.URL+path, body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", jsonapi.ContentType)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// readDocument decodes a JSON:API single-resource response.
func readDocument(t *testing.T, resp *http.Response) jsonapi.Document {
	t.Helper()
	defer resp.Body.Close()
	var doc jsonapi.Document
	err := json.NewDecoder(resp.Body).Decode(&doc)
	require.NoError(t, err)
	return doc
}

// readCollection decodes a JSON:API collection response.
func readCollection(t *testing.T, resp *http.Response) jsonapi.CollectionDocument {
	t.Helper()
	defer resp.Body.Close()
	var doc jsonapi.CollectionDocument
	err := json.NewDecoder(resp.Body).Decode(&doc)
	require.NoError(t, err)
	return doc
}

// createTestUser creates a user via the API and returns the user ID.
func createTestUser(t *testing.T, username, password, role string) string {
	t.Helper()

	token := adminToken(t)
	body := fmt.Sprintf(`{"data":{"type":"users","attributes":{"username":%q,"password":%q,"role":%q}}}`, username, password, role)
	resp := doJSONAPIRequest(t, http.MethodPost, "/api/v1/users", bytes.NewBufferString(body), token)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "failed to create test user %s", username)

	doc := readDocument(t, resp)
	return doc.Data.ID
}
