package unit_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	scalargo "github.com/bdpiprava/scalar-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScalarDocsHandler(t *testing.T) {
	docsHTML, err := scalargo.NewV2(
		scalargo.WithSpecDir("../../api"),
		scalargo.WithBaseFileName("openapi.yaml"),
		scalargo.WithDarkMode(),
		scalargo.WithTheme(scalargo.ThemeKepler),
	)
	require.NoError(t, err)
	require.NotEmpty(t, docsHTML)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, docsHTML)
	})

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "text/html; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.True(t, strings.Contains(rec.Body.String(), "<html"), "response should contain HTML")
	assert.True(t, strings.Contains(rec.Body.String(), "scalar"), "response should reference scalar")
}
