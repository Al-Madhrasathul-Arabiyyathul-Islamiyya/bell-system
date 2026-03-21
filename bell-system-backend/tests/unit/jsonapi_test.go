package unit_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Types ---

func TestResource_JSON_Roundtrip(t *testing.T) {
	r := jsonapi.Resource{
		Type:       "users",
		ID:         "abc-123",
		Attributes: map[string]string{"username": "admin"},
		Links:      &jsonapi.ResourceLinks{Self: "/api/v1/users/abc-123"},
	}

	data, err := json.Marshal(r)
	require.NoError(t, err)

	var decoded jsonapi.Resource
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, "users", decoded.Type)
	assert.Equal(t, "abc-123", decoded.ID)
}

func TestDocument_JSON_WithIncluded(t *testing.T) {
	doc := jsonapi.Document{
		Data: &jsonapi.Resource{Type: "users", ID: "1", Attributes: map[string]string{"name": "test"}},
		Included: []jsonapi.Resource{
			{Type: "sessions", ID: "2", Attributes: map[string]string{"name": "Morning"}},
		},
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	err = json.Unmarshal(data, &raw)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"included"`)
}

func TestCollectionDocument_JSON(t *testing.T) {
	doc := jsonapi.CollectionDocument{
		Data:  []jsonapi.Resource{{Type: "users", ID: "1", Attributes: nil}},
		Meta:  map[string]any{"total": 1},
		Links: map[string]any{"self": "/api/v1/users"},
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"meta"`)
	assert.Contains(t, string(data), `"links"`)
}

func TestErrorDocument_JSON(t *testing.T) {
	doc := jsonapi.ErrorDocument{
		Errors: []jsonapi.Error{
			{Status: "404", Code: "not_found", Title: "Not Found", Detail: "user not found"},
		},
	}

	data, err := json.Marshal(doc)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"errors"`)
	assert.Contains(t, string(data), `"not_found"`)
}

// --- Response writers ---

func TestWriteResource(t *testing.T) {
	w := httptest.NewRecorder()
	doc := jsonapi.Document{
		Data: &jsonapi.Resource{Type: "users", ID: "1", Attributes: map[string]string{"name": "admin"}},
	}

	jsonapi.WriteResource(w, http.StatusOK, doc)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, jsonapi.ContentType, w.Header().Get("Content-Type"))

	var resp jsonapi.Document
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "users", resp.Data.Type)
}

func TestWriteCollection(t *testing.T) {
	w := httptest.NewRecorder()
	doc := jsonapi.CollectionDocument{
		Data:  []jsonapi.Resource{{Type: "users", ID: "1", Attributes: nil}},
		Meta:  map[string]any{"total": 1},
		Links: map[string]any{"self": "/api/v1/users"},
	}

	jsonapi.WriteCollection(w, http.StatusOK, doc)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, jsonapi.ContentType, w.Header().Get("Content-Type"))
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	jsonapi.WriteError(w, http.StatusNotFound, "not_found", "", "user not found")

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, jsonapi.ContentType, w.Header().Get("Content-Type"))

	var resp jsonapi.ErrorDocument
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	require.Len(t, resp.Errors, 1)
	assert.Equal(t, "404", resp.Errors[0].Status)
	assert.Equal(t, "not_found", resp.Errors[0].Code)
	assert.Equal(t, "Not Found", resp.Errors[0].Title)
	assert.Equal(t, "user not found", resp.Errors[0].Detail)
}

func TestWriteError_CustomTitle(t *testing.T) {
	w := httptest.NewRecorder()
	jsonapi.WriteError(w, http.StatusBadRequest, "invalid_request", "Custom Title", "bad data")

	var resp jsonapi.ErrorDocument
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Equal(t, "Custom Title", resp.Errors[0].Title)
}

func TestWriteNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	jsonapi.WriteNoContent(w)
	assert.Equal(t, http.StatusNoContent, w.Code)
}

// --- Request parsing ---

func TestParseRequest_Success(t *testing.T) {
	body := `{"data":{"type":"users","attributes":{"username":"admin","role":"admin"}}}`
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	doc, err := jsonapi.ParseRequest(r, "users")
	require.NoError(t, err)
	assert.Equal(t, "users", doc.Data.Type)

	var attrs struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}
	err = doc.UnmarshalAttributes(&attrs)
	require.NoError(t, err)
	assert.Equal(t, "admin", attrs.Username)
	assert.Equal(t, "admin", attrs.Role)
}

func TestParseRequest_WrongType(t *testing.T) {
	body := `{"data":{"type":"sessions","attributes":{}}}`
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	_, err := jsonapi.ParseRequest(r, "users")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expected resource type")
}

func TestParseRequest_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{bad`))

	_, err := jsonapi.ParseRequest(r, "users")
	assert.Error(t, err)
}

func TestRequestDocument_RelationshipID(t *testing.T) {
	body := `{"data":{"type":"schedule-items","attributes":{"name":"Bell"},"relationships":{"session":{"data":{"type":"sessions","id":"abc-123"}},"sound":{"data":{"type":"audio-files","id":"def-456"}}}}}`
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	doc, err := jsonapi.ParseRequest(r, "schedule-items")
	require.NoError(t, err)

	assert.Equal(t, "abc-123", doc.RelationshipID("session"))
	assert.Equal(t, "def-456", doc.RelationshipID("sound"))
	assert.Equal(t, "", doc.RelationshipID("nonexistent"))
}

func TestRequestDocument_NullRelationship(t *testing.T) {
	body := `{"data":{"type":"schedule-items","attributes":{},"relationships":{"session":{"data":null}}}}`
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))

	doc, err := jsonapi.ParseRequest(r, "schedule-items")
	require.NoError(t, err)
	assert.Equal(t, "", doc.RelationshipID("session"))
}

func TestRequestDocument_UnmarshalAttributes_Missing(t *testing.T) {
	doc := &jsonapi.RequestDocument{}
	var v struct{}
	err := doc.UnmarshalAttributes(&v)
	assert.Error(t, err)
}

// --- Pagination ---

func TestParsePagination_Defaults(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	p := jsonapi.ParsePagination(r)
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, jsonapi.DefaultPageSize, p.Size)
}

func TestParsePagination_Custom(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page[number]=3&page[size]=10", nil)
	p := jsonapi.ParsePagination(r)
	assert.Equal(t, 3, p.Page)
	assert.Equal(t, 10, p.Size)
}

func TestParsePagination_MaxSize(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page[size]=999", nil)
	p := jsonapi.ParsePagination(r)
	assert.Equal(t, jsonapi.MaxPageSize, p.Size)
}

func TestParsePagination_NegativeValues(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?page[number]=-1&page[size]=-5", nil)
	p := jsonapi.ParsePagination(r)
	assert.Equal(t, 1, p.Page)
	assert.Equal(t, jsonapi.DefaultPageSize, p.Size)
}

func TestPaginationParams_Offset(t *testing.T) {
	p := jsonapi.PaginationParams{Page: 3, Size: 10}
	assert.Equal(t, 20, p.Offset())
}

func TestPaginationMeta(t *testing.T) {
	meta := jsonapi.PaginationMeta(55, 2, 20)
	assert.Equal(t, 55, meta["total"])
	assert.Equal(t, 2, meta["page"])
	assert.Equal(t, 20, meta["pageSize"])
	assert.Equal(t, 3, meta["totalPages"])
}

func TestPaginationLinks(t *testing.T) {
	params := jsonapi.PaginationParams{Page: 2, Size: 10}
	links := jsonapi.PaginationLinks("/api/v1/users", params, 30)

	assert.Contains(t, links["self"], "page[number]=2")
	assert.Contains(t, links["first"], "page[number]=1")
	assert.Contains(t, links["last"], "page[number]=3")
	assert.Contains(t, links["prev"], "page[number]=1")
	assert.Contains(t, links["next"], "page[number]=3")
}

func TestPaginationLinks_FirstPage(t *testing.T) {
	params := jsonapi.PaginationParams{Page: 1, Size: 10}
	links := jsonapi.PaginationLinks("/api/v1/users", params, 30)

	_, hasPrev := links["prev"]
	assert.False(t, hasPrev)
	assert.Contains(t, links["next"], "page[number]=2")
}

func TestPaginationLinks_LastPage(t *testing.T) {
	params := jsonapi.PaginationParams{Page: 3, Size: 10}
	links := jsonapi.PaginationLinks("/api/v1/users", params, 30)

	_, hasNext := links["next"]
	assert.False(t, hasNext)
	assert.Contains(t, links["prev"], "page[number]=2")
}

func TestPaginationLinks_EmptyCollection(t *testing.T) {
	params := jsonapi.PaginationParams{Page: 1, Size: 10}
	links := jsonapi.PaginationLinks("/api/v1/users", params, 0)

	assert.Contains(t, links["last"], "page[number]=1")
}

// --- Query parsing ---

func TestParseSort_Empty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	fields := jsonapi.ParseSort(r)
	assert.Nil(t, fields)
}

func TestParseSort_Single(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?sort=username", nil)
	fields := jsonapi.ParseSort(r)
	require.Len(t, fields, 1)
	assert.Equal(t, "username", fields[0].Field)
	assert.False(t, fields[0].Desc)
}

func TestParseSort_Descending(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?sort=-createdAt", nil)
	fields := jsonapi.ParseSort(r)
	require.Len(t, fields, 1)
	assert.Equal(t, "createdAt", fields[0].Field)
	assert.True(t, fields[0].Desc)
}

func TestParseSort_Multiple(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?sort=role,-username", nil)
	fields := jsonapi.ParseSort(r)
	require.Len(t, fields, 2)
	assert.Equal(t, "role", fields[0].Field)
	assert.False(t, fields[0].Desc)
	assert.Equal(t, "username", fields[1].Field)
	assert.True(t, fields[1].Desc)
}

func TestParseSort_SkipsEmptySegments(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?sort=name,,startTime", nil)
	fields := jsonapi.ParseSort(r)
	require.Len(t, fields, 2)
	assert.Equal(t, "name", fields[0].Field)
	assert.Equal(t, "startTime", fields[1].Field)
}

func TestParseFilter(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?filter[role]=admin&filter[status]=active&filter[bad]=evil", nil)
	filters := jsonapi.ParseFilter(r, []string{"role", "status"})

	assert.Equal(t, "admin", filters["role"])
	assert.Equal(t, "active", filters["status"])
	_, hasBad := filters["bad"]
	assert.False(t, hasBad)
}

func TestParseFilter_Empty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	filters := jsonapi.ParseFilter(r, []string{"role"})
	assert.Empty(t, filters)
}

func TestParseFilter_EmptyValue(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?filter[role]=", nil)
	filters := jsonapi.ParseFilter(r, []string{"role"})
	_, hasRole := filters["role"]
	assert.False(t, hasRole)
}

func TestSortToSQL_Empty(t *testing.T) {
	result := jsonapi.SortToSQL(nil, map[string]string{"name": "Name"}, "ORDER BY Name")
	assert.Equal(t, "ORDER BY Name", result)
}

func TestSortToSQL_SingleAsc(t *testing.T) {
	sorts := []jsonapi.SortField{{Field: "username", Desc: false}}
	colMap := map[string]string{"username": "Username", "role": "Role"}
	result := jsonapi.SortToSQL(sorts, colMap, "ORDER BY Username")
	assert.Equal(t, "ORDER BY Username ASC", result)
}

func TestSortToSQL_SingleDesc(t *testing.T) {
	sorts := []jsonapi.SortField{{Field: "createdAt", Desc: true}}
	colMap := map[string]string{"createdAt": "CreatedAt"}
	result := jsonapi.SortToSQL(sorts, colMap, "ORDER BY CreatedAt")
	assert.Equal(t, "ORDER BY CreatedAt DESC", result)
}

func TestSortToSQL_Multiple(t *testing.T) {
	sorts := []jsonapi.SortField{
		{Field: "role", Desc: false},
		{Field: "username", Desc: true},
	}
	colMap := map[string]string{"username": "Username", "role": "Role"}
	result := jsonapi.SortToSQL(sorts, colMap, "ORDER BY Username")
	assert.Equal(t, "ORDER BY Role ASC, Username DESC", result)
}

func TestSortToSQL_UnknownFieldsIgnored(t *testing.T) {
	sorts := []jsonapi.SortField{{Field: "unknown", Desc: false}}
	colMap := map[string]string{"name": "Name"}
	result := jsonapi.SortToSQL(sorts, colMap, "ORDER BY Name")
	assert.Equal(t, "ORDER BY Name", result)
}

func TestSortToSQL_MixedKnownAndUnknown(t *testing.T) {
	sorts := []jsonapi.SortField{
		{Field: "unknown", Desc: false},
		{Field: "name", Desc: true},
	}
	colMap := map[string]string{"name": "Name"}
	result := jsonapi.SortToSQL(sorts, colMap, "ORDER BY Name")
	assert.Equal(t, "ORDER BY Name DESC", result)
}

func TestParseInclude_Empty(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	includes := jsonapi.ParseInclude(r)
	assert.Nil(t, includes)
}

func TestParseInclude_Single(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?include=session", nil)
	includes := jsonapi.ParseInclude(r)
	assert.Equal(t, []string{"session"}, includes)
}

func TestParseInclude_Multiple(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?include=session,sound", nil)
	includes := jsonapi.ParseInclude(r)
	assert.Equal(t, []string{"session", "sound"}, includes)
}

// --- Marshalers ---

func TestMarshalUser(t *testing.T) {
	id := uuid.New()
	now := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	user := &models.User{ID: id, Username: "admin", Role: models.RoleAdmin, CreatedAt: now}

	r := jsonapi.MarshalUser(user)

	assert.Equal(t, "users", r.Type)
	assert.Equal(t, id.String(), r.ID)
	assert.Equal(t, "/api/v1/users/"+id.String(), r.Links.Self)

	data, _ := json.Marshal(r.Attributes)
	var attrs map[string]any
	json.Unmarshal(data, &attrs)
	assert.Equal(t, "admin", attrs["username"])
	assert.Equal(t, "admin", attrs["role"])
}

func TestMarshalSession(t *testing.T) {
	id := uuid.New()
	s := &models.Session{
		ID:        id,
		Name:      "Morning",
		StartTime: time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC),
		EndTime:   time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	r := jsonapi.MarshalSession(s)

	assert.Equal(t, "sessions", r.Type)
	assert.Equal(t, id.String(), r.ID)

	data, _ := json.Marshal(r.Attributes)
	var attrs map[string]string
	json.Unmarshal(data, &attrs)
	assert.Equal(t, "Morning", attrs["name"])
	assert.Equal(t, "07:00", attrs["startTime"])
	assert.Equal(t, "12:00", attrs["endTime"])
}

func TestMarshalScheduleItem(t *testing.T) {
	id := uuid.New()
	sessionID := uuid.New()
	soundID := uuid.New()
	now := time.Now()

	item := &models.ScheduleItem{
		ID:        id,
		SessionID: &sessionID,
		Name:      "Morning Bell",
		Time:      time.Date(0, 1, 1, 7, 30, 0, 0, time.UTC),
		SoundID:   soundID,
		Days:      []int{1, 2, 3},
		CreatedAt: now,
		UpdatedAt: now,
	}

	r := jsonapi.MarshalScheduleItem(item)

	assert.Equal(t, "schedule-items", r.Type)
	assert.Equal(t, id.String(), r.ID)
	assert.Contains(t, r.Relationships, "session")
	assert.Contains(t, r.Relationships, "sound")

	// Session relationship
	sessionRel := r.Relationships["session"]
	sessionRI := sessionRel.Data.(*jsonapi.ResourceIdentifier)
	assert.Equal(t, "sessions", sessionRI.Type)
	assert.Equal(t, sessionID.String(), sessionRI.ID)

	// Sound relationship
	soundRel := r.Relationships["sound"]
	soundRI := soundRel.Data.(*jsonapi.ResourceIdentifier)
	assert.Equal(t, "audio-files", soundRI.Type)
	assert.Equal(t, soundID.String(), soundRI.ID)
}

func TestMarshalScheduleItem_NilSession(t *testing.T) {
	item := &models.ScheduleItem{
		ID:      uuid.New(),
		Name:    "General Bell",
		Time:    time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
		SoundID: uuid.New(),
	}

	r := jsonapi.MarshalScheduleItem(item)

	sessionRel := r.Relationships["session"]
	assert.Nil(t, sessionRel.Data)
}

func TestMarshalScheduleItem_NilDays(t *testing.T) {
	item := &models.ScheduleItem{
		ID:      uuid.New(),
		Name:    "Test",
		Time:    time.Date(0, 1, 1, 8, 0, 0, 0, time.UTC),
		SoundID: uuid.New(),
		Days:    nil,
	}

	r := jsonapi.MarshalScheduleItem(item)

	data, _ := json.Marshal(r.Attributes)
	assert.Contains(t, string(data), `"days":[]`)
}

func TestMarshalAudioFile(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	f := &models.SystemAudioFile{
		ID:        id,
		Name:      "bell.wav",
		FilePath:  "/audio/bell.wav",
		FileType:  models.FileTypeBell,
		Checksum:  "abc123",
		CreatedAt: now,
		UpdatedAt: now,
	}

	r := jsonapi.MarshalAudioFile(f)

	assert.Equal(t, "audio-files", r.Type)
	assert.Equal(t, id.String(), r.ID)
	assert.Equal(t, "/api/v1/audio/"+id.String(), r.Links.Self)
	assert.Equal(t, "/api/v1/audio/"+id.String()+"/content", r.Links.Content)
}

func TestMarshalAudioChecksum(t *testing.T) {
	id := uuid.New()
	f := &models.SystemAudioFile{
		ID:       id,
		Name:     "bell.wav",
		FileType: models.FileTypeBell,
		Checksum: "abc123",
	}

	r := jsonapi.MarshalAudioChecksum(f)

	assert.Equal(t, "audio-files", r.Type)
	assert.Equal(t, id.String(), r.ID)
	assert.Nil(t, r.Links)

	data, _ := json.Marshal(r.Attributes)
	assert.NotContains(t, string(data), "name")
	assert.Contains(t, string(data), "abc123")
}

func TestMarshalSystemState(t *testing.T) {
	now := time.Now()
	r := jsonapi.MarshalSystemState("active", now)

	assert.Equal(t, "system-state", r.Type)
	assert.Equal(t, "current", r.ID)
	assert.Equal(t, "/api/v1/system/state", r.Links.Self)

	data, _ := json.Marshal(r.Attributes)
	assert.Contains(t, string(data), `"active"`)
}
