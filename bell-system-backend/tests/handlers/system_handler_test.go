package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/handlers"
	"arabiyya.edu.mv/bell-system-backend/internal/router"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
	"arabiyya.edu.mv/bell-system-backend/tests/mocks"
	"arabiyya.edu.mv/bell-system-backend/tests/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSystemRouter(stateRepo handlers.SystemStateRepository, schedSvc handlers.SchedulerService, notifier handlers.EventNotifier) http.Handler {
	h := handlers.NewSystemHandler(stateRepo, schedSvc, notifier)
	return router.SystemRoutes(h, testutil.PermissiveTokenService)
}

func authSystemReq(req *http.Request) *http.Request {
	testutil.SetAuthHeader(req)
	return req
}

func systemStateBody(state string) *bytes.Buffer {
	body, _ := json.Marshal(map[string]any{
		"data": map[string]any{
			"type":       "system-state",
			"attributes": map[string]any{"state": state},
		},
	})
	return bytes.NewBuffer(body)
}

// --- GET /state ---

func TestSystemHandler_GetState_Success(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	repo := &mocks.MockSystemStateRepo{
		GetStateFunc: func(_ context.Context) (string, time.Time, error) {
			return "active", now, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/state", nil)
	rr := httptest.NewRecorder()
	newSystemRouter(repo, &mocks.MockSchedulerService{}, nil).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "system-state", resp.Data.Type)
	assert.Equal(t, "current", resp.Data.ID)
	attrs := resp.Data.Attributes.(map[string]any)
	assert.Equal(t, "active", attrs["state"])
}

func TestSystemHandler_GetState_Error(t *testing.T) {
	repo := &mocks.MockSystemStateRepo{
		GetStateFunc: func(_ context.Context) (string, time.Time, error) {
			return "", time.Time{}, errors.New("db error")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/state", nil)
	rr := httptest.NewRecorder()
	newSystemRouter(repo, &mocks.MockSchedulerService{}, nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// --- POST /state ---

func TestSystemHandler_SetState_Active(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	notifier := &mocks.MockEventNotifier{}
	var setState string
	repo := &mocks.MockSystemStateRepo{
		SetStateFunc: func(_ context.Context, state string) error {
			setState = state
			return nil
		},
		GetStateFunc: func(_ context.Context) (string, time.Time, error) {
			return setState, now, nil
		},
	}

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/state", systemStateBody("active")))
	rr := httptest.NewRecorder()
	newSystemRouter(repo, &mocks.MockSchedulerService{}, notifier).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "system-state", resp.Data.Type)
	attrs := resp.Data.Attributes.(map[string]any)
	assert.Equal(t, "active", attrs["state"])
	assert.True(t, notifier.SystemStateChangedCalled)
}

func TestSystemHandler_SetState_Paused(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	notifier := &mocks.MockEventNotifier{}
	repo := &mocks.MockSystemStateRepo{
		SetStateFunc: func(_ context.Context, state string) error {
			return nil
		},
		GetStateFunc: func(_ context.Context) (string, time.Time, error) {
			return "paused", now, nil
		},
	}

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/state", systemStateBody("paused")))
	rr := httptest.NewRecorder()
	newSystemRouter(repo, &mocks.MockSchedulerService{}, notifier).ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	attrs := resp.Data.Attributes.(map[string]any)
	assert.Equal(t, "paused", attrs["state"])
	assert.True(t, notifier.SystemStateChangedCalled)
}

func TestSystemHandler_SetState_InvalidBody(t *testing.T) {
	repo := &mocks.MockSystemStateRepo{}

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/state", bytes.NewReader([]byte("not json"))))
	rr := httptest.NewRecorder()
	newSystemRouter(repo, &mocks.MockSchedulerService{}, nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSystemHandler_SetState_InvalidState(t *testing.T) {
	repo := &mocks.MockSystemStateRepo{}

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/state", systemStateBody("unknown")))
	rr := httptest.NewRecorder()
	newSystemRouter(repo, &mocks.MockSchedulerService{}, nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestSystemHandler_SetState_DBError(t *testing.T) {
	repo := &mocks.MockSystemStateRepo{
		SetStateFunc: func(_ context.Context, state string) error {
			return errors.New("db error")
		},
	}

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/state", systemStateBody("active")))
	rr := httptest.NewRecorder()
	newSystemRouter(repo, &mocks.MockSchedulerService{}, nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSystemHandler_SetState_NilNotifier(t *testing.T) {
	now := time.Date(2026, 3, 17, 10, 0, 0, 0, time.UTC)
	repo := &mocks.MockSystemStateRepo{
		SetStateFunc: func(_ context.Context, state string) error {
			return nil
		},
		GetStateFunc: func(_ context.Context) (string, time.Time, error) {
			return "active", now, nil
		},
	}

	h := handlers.NewSystemHandler(repo, &mocks.MockSchedulerService{}, nil)
	r := router.SystemRoutes(h, testutil.PermissiveTokenService)

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/state", systemStateBody("active")))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var resp jsonapi.Document
	err := json.NewDecoder(rr.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "system-state", resp.Data.Type)
}

func TestSystemHandler_SetState_GetStateErrorAfterSet(t *testing.T) {
	repo := &mocks.MockSystemStateRepo{
		SetStateFunc: func(_ context.Context, state string) error {
			return nil
		},
		GetStateFunc: func(_ context.Context) (string, time.Time, error) {
			return "", time.Time{}, errors.New("db error on re-read")
		},
	}

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/state", systemStateBody("active")))
	rr := httptest.NewRecorder()
	newSystemRouter(repo, &mocks.MockSchedulerService{}, nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestSystemHandler_CancelNextBell_NilScheduler(t *testing.T) {
	h := handlers.NewSystemHandler(&mocks.MockSystemStateRepo{}, nil, nil)
	r := router.SystemRoutes(h, testutil.PermissiveTokenService)

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/cancel-next-bell", nil))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// --- POST /cancel-next-bell ---

func TestSystemHandler_CancelNextBell_Success(t *testing.T) {
	svc := &mocks.MockSchedulerService{
		CancelNextBellFunc: func() (string, error) {
			return "abc-123", nil
		},
	}

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/cancel-next-bell", nil))
	rr := httptest.NewRecorder()
	newSystemRouter(&mocks.MockSystemStateRepo{}, svc, nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestSystemHandler_CancelNextBell_NoneToCancel(t *testing.T) {
	svc := &mocks.MockSchedulerService{
		CancelNextBellFunc: func() (string, error) {
			return "", errors.New("no pending bells to cancel")
		},
	}

	req := authSystemReq(httptest.NewRequest(http.MethodPost, "/cancel-next-bell", nil))
	rr := httptest.NewRecorder()
	newSystemRouter(&mocks.MockSystemStateRepo{}, svc, nil).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}
