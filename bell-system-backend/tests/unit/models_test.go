package unit_test

import (
	"encoding/json"
	"testing"
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Role constants per spec: admin, morning_user, afternoon_user ---

func TestRole_Constants(t *testing.T) {
	assert.Equal(t, models.Role("admin"), models.RoleAdmin)
	assert.Equal(t, models.Role("morning_user"), models.RoleMorningUser)
	assert.Equal(t, models.Role("afternoon_user"), models.RoleAfternoonUser)
}

// --- FileType constants per spec: bell, anthem, school_song, other ---

func TestFileType_Constants(t *testing.T) {
	assert.Equal(t, models.FileType("bell"), models.FileTypeBell)
	assert.Equal(t, models.FileType("anthem"), models.FileTypeAnthem)
	assert.Equal(t, models.FileType("school_song"), models.FileTypeSchoolSong)
	assert.Equal(t, models.FileType("other"), models.FileTypeOther)
}

// --- User JSON serialization ---

func TestUser_JSON_PasswordHashOmitted(t *testing.T) {
	user := models.User{
		ID:           uuid.New(),
		Username:     "admin",
		PasswordHash: "secret-hash",
		Role:         models.RoleAdmin,
		CreatedAt:    time.Now(),
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Contains(t, result, "id")
	assert.Contains(t, result, "username")
	assert.Contains(t, result, "role")
	assert.Contains(t, result, "createdAt")
	assert.NotContains(t, result, "passwordHash", "PasswordHash must not appear in JSON output")
}

func TestUser_JSON_RoundTrip(t *testing.T) {
	id := uuid.New()
	now := time.Now().Truncate(time.Millisecond)
	user := models.User{
		ID:        id,
		Username:  "morning_user",
		Role:      models.RoleMorningUser,
		CreatedAt: now,
	}

	data, err := json.Marshal(user)
	require.NoError(t, err)

	var decoded models.User
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, id, decoded.ID)
	assert.Equal(t, "morning_user", decoded.Username)
	assert.Equal(t, models.RoleMorningUser, decoded.Role)
}

// --- Session JSON ---

func TestSession_JSON_Fields(t *testing.T) {
	s := models.Session{
		ID:        uuid.New(),
		Name:      "Morning Session",
		StartTime: time.Date(0, 1, 1, 7, 0, 0, 0, time.UTC),
		EndTime:   time.Date(0, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Contains(t, result, "id")
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "startTime")
	assert.Contains(t, result, "endTime")
}

// --- ScheduleItem JSON ---

func TestScheduleItem_JSON_SessionIDOmittedWhenNil(t *testing.T) {
	item := models.ScheduleItem{
		ID:      uuid.New(),
		Name:    "First Bell",
		SoundID: uuid.New(),
		Days:    []int{2, 3, 4, 5, 6},
	}

	data, err := json.Marshal(item)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.NotContains(t, result, "sessionId", "sessionId should be omitted when nil")
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "soundId")
	assert.Contains(t, result, "days")
}

func TestScheduleItem_JSON_SessionIDPresentWhenSet(t *testing.T) {
	sessionID := uuid.New()
	item := models.ScheduleItem{
		ID:        uuid.New(),
		SessionID: &sessionID,
		Name:      "Morning Bell",
		SoundID:   uuid.New(),
	}

	data, err := json.Marshal(item)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Contains(t, result, "sessionId")
}

// --- ScheduleDay ---

func TestScheduleDay_DayOfWeek_Range(t *testing.T) {
	for day := 1; day <= 7; day++ {
		sd := models.ScheduleDay{ScheduleItemID: uuid.New(), DayOfWeek: day}
		assert.GreaterOrEqual(t, sd.DayOfWeek, 1)
		assert.LessOrEqual(t, sd.DayOfWeek, 7)
	}
}

// --- ScheduleDayInfo ---

func TestScheduleDayInfo_JSON(t *testing.T) {
	info := models.ScheduleDayInfo{DayNumber: 2, DayName: "Monday"}

	data, err := json.Marshal(info)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, float64(2), result["dayNumber"])
	assert.Equal(t, "Monday", result["dayName"])
}

// --- SystemAudioFile JSON ---

func TestSystemAudioFile_JSON_Fields(t *testing.T) {
	audio := models.SystemAudioFile{
		ID:       uuid.New(),
		Name:     "bell.mp3",
		FilePath: "/audio/bell.mp3",
		FileType: models.FileTypeBell,
		Checksum: "abc123def456",
	}

	data, err := json.Marshal(audio)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, "bell.mp3", result["name"])
	assert.Equal(t, "/audio/bell.mp3", result["filePath"])
	assert.Equal(t, "bell", result["fileType"])
	assert.Equal(t, "abc123def456", result["checksum"])
}

// --- WebSocket message models ---

func TestWebSocketMessage_JSON(t *testing.T) {
	msg := models.WebSocketMessage{
		Type:      "bell_triggered",
		Timestamp: time.Date(2026, 3, 10, 8, 0, 0, 0, time.UTC),
		Payload: models.BellTriggeredPayload{
			ScheduleItemID: "item-1",
			SoundID:        "sound-1",
			Name:           "First Period",
		},
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, "bell_triggered", result["type"])
	assert.Contains(t, result, "timestamp")
	assert.Contains(t, result, "payload")
}

func TestWebSocketMessage_JSON_PayloadOmittedWhenNil(t *testing.T) {
	msg := models.WebSocketMessage{
		Type:      "schedules_updated",
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(msg)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.NotContains(t, result, "payload")
}

func TestRegistrationPayload_JSON(t *testing.T) {
	payload := models.RegistrationPayload{
		ClientType: "admin",
		ClientName: "Admin Panel",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded models.RegistrationPayload
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "admin", decoded.ClientType)
	assert.Equal(t, "Admin Panel", decoded.ClientName)
}

func TestRegistrationPayload_ClientNameOmittedWhenEmpty(t *testing.T) {
	payload := models.RegistrationPayload{
		ClientType: "client",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.NotContains(t, result, "client_name")
}

func TestSystemStatePayload_JSON(t *testing.T) {
	tests := []struct {
		name  string
		state string
	}{
		{"active state", "active"},
		{"paused state", "paused"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := models.SystemStatePayload{State: tt.state}
			data, err := json.Marshal(payload)
			require.NoError(t, err)

			var decoded models.SystemStatePayload
			require.NoError(t, json.Unmarshal(data, &decoded))
			assert.Equal(t, tt.state, decoded.State)
		})
	}
}

func TestSystemLogPayload_JSON(t *testing.T) {
	payload := models.SystemLogPayload{
		Level:   "error",
		Message: "Audio playback failed",
		Source:  "scheduler",
	}

	data, err := json.Marshal(payload)
	require.NoError(t, err)

	var decoded models.SystemLogPayload
	require.NoError(t, json.Unmarshal(data, &decoded))

	assert.Equal(t, "error", decoded.Level)
	assert.Equal(t, "Audio playback failed", decoded.Message)
	assert.Equal(t, "scheduler", decoded.Source)
}

// --- Request models ---

func TestLoginRequest_JSON(t *testing.T) {
	body := `{"username":"admin","password":"admin123"}`
	var req models.LoginRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))
	assert.Equal(t, "admin", req.Username)
	assert.Equal(t, "admin123", req.Password)
}

func TestUserCreateRequest_JSON(t *testing.T) {
	body := `{"username":"newuser","password":"password123","role":"morning_user"}`
	var req models.UserCreateRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))
	assert.Equal(t, "newuser", req.Username)
	assert.Equal(t, "password123", req.Password)
	assert.Equal(t, models.RoleMorningUser, req.Role)
}

func TestScheduleItemCreateRequest_JSON_WithOptionalSessionID(t *testing.T) {
	sessionID := uuid.New()
	body, _ := json.Marshal(models.ScheduleItemCreateRequest{
		SessionID: &sessionID,
		Name:      "First Bell",
		Time:      "07:45",
		SoundID:   uuid.New(),
		Days:      []int{2, 3, 4, 5, 6},
	})

	var req models.ScheduleItemCreateRequest
	require.NoError(t, json.Unmarshal(body, &req))
	require.NotNil(t, req.SessionID)
	assert.Equal(t, sessionID, *req.SessionID)
	assert.Equal(t, []int{2, 3, 4, 5, 6}, req.Days)
}

func TestScheduleItemCreateRequest_JSON_WithoutSessionID(t *testing.T) {
	body := `{"name":"School Opening","time":"06:45","soundId":"` + uuid.New().String() + `","days":[2,3,4,5,6]}`
	var req models.ScheduleItemCreateRequest
	require.NoError(t, json.Unmarshal([]byte(body), &req))
	assert.Nil(t, req.SessionID, "sessionId should be nil when not provided")
}

func TestSystemStateRequest_JSON(t *testing.T) {
	tests := []struct {
		body  string
		state string
	}{
		{`{"state":"active"}`, "active"},
		{`{"state":"paused"}`, "paused"},
	}
	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			var req models.SystemStateRequest
			require.NoError(t, json.Unmarshal([]byte(tt.body), &req))
			assert.Equal(t, tt.state, req.State)
		})
	}
}

// --- Response models ---

func TestErrorResponse_JSON(t *testing.T) {
	resp := jsonapi.ErrorDocument{
		Errors: []jsonapi.Error{
			{Status: "404", Code: "not_found", Title: "Not Found", Detail: "User not found"},
		},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	errArr, ok := result["errors"].([]any)
	require.True(t, ok)
	require.Len(t, errArr, 1)

	errObj, ok := errArr[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "not_found", errObj["code"])
	assert.Equal(t, "User not found", errObj["detail"])
}

func TestSuccessResponse_JSON_MessageOmittedWhenEmpty(t *testing.T) {
	resp := models.SuccessResponse{Success: true}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, true, result["success"])
	assert.NotContains(t, result, "message")
}

func TestListResponse_JSON(t *testing.T) {
	resp := models.ListResponse{
		Total: 2,
		Items: []string{"a", "b"},
	}

	data, err := json.Marshal(resp)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(data, &result))

	assert.Equal(t, float64(2), result["total"])
	items, ok := result["items"].([]any)
	require.True(t, ok)
	assert.Len(t, items, 2)
}

func TestCurrentScheduleItem_StatusValues(t *testing.T) {
	validStatuses := []string{"pending", "current", "completed"}
	for _, status := range validStatuses {
		item := models.CurrentScheduleItem{
			ID:     uuid.New(),
			Name:   "Test Bell",
			Time:   "08:00",
			Status: status,
		}
		data, err := json.Marshal(item)
		require.NoError(t, err)

		var decoded models.CurrentScheduleItem
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, status, decoded.Status)
	}
}
