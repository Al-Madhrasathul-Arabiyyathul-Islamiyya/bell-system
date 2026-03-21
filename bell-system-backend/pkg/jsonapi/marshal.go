package jsonapi

import (
	"time"

	"arabiyya.edu.mv/bell-system-backend/internal/models"
)

// --- User ---

type userAttributes struct {
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

func MarshalUser(u *models.User) Resource {
	return Resource{
		Type:       "users",
		ID:         u.ID.String(),
		Attributes: userAttributes{Username: u.Username, Role: string(u.Role), CreatedAt: u.CreatedAt},
		Links:      &ResourceLinks{Self: "/api/v1/users/" + u.ID.String()},
	}
}

// --- Session ---

type sessionAttributes struct {
	Name      string `json:"name"`
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
}

func MarshalSession(s *models.Session) Resource {
	return Resource{
		Type: "sessions",
		ID:   s.ID.String(),
		Attributes: sessionAttributes{
			Name:      s.Name,
			StartTime: s.StartTime.Format("15:04"),
			EndTime:   s.EndTime.Format("15:04"),
		},
		Links: &ResourceLinks{Self: "/api/v1/sessions/" + s.ID.String()},
	}
}

// --- ScheduleItem ---

type scheduleItemAttributes struct {
	Name      string    `json:"name"`
	Time      string    `json:"time"`
	Days      []int     `json:"days"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func MarshalScheduleItem(item *models.ScheduleItem) Resource {
	attrs := scheduleItemAttributes{
		Name:      item.Name,
		Time:      item.Time.Format("15:04"),
		Days:      item.Days,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
	if attrs.Days == nil {
		attrs.Days = []int{}
	}

	rels := map[string]Relation{
		"sound": {Data: &ResourceIdentifier{Type: "audio-files", ID: item.SoundID.String()}},
	}
	if item.SessionID != nil {
		rels["session"] = Relation{Data: &ResourceIdentifier{Type: "sessions", ID: item.SessionID.String()}}
	} else {
		rels["session"] = Relation{Data: nil}
	}

	return Resource{
		Type:          "schedule-items",
		ID:            item.ID.String(),
		Attributes:    attrs,
		Relationships: rels,
		Links:         &ResourceLinks{Self: "/api/v1/schedule/" + item.ID.String()},
	}
}

// --- AudioFile ---

type audioFileAttributes struct {
	Name      string    `json:"name"`
	FileType  string    `json:"fileType"`
	Checksum  string    `json:"checksum"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func MarshalAudioFile(f *models.SystemAudioFile) Resource {
	return Resource{
		Type: "audio-files",
		ID:   f.ID.String(),
		Attributes: audioFileAttributes{
			Name:      f.Name,
			FileType:  string(f.FileType),
			Checksum:  f.Checksum,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		},
		Links: &ResourceLinks{
			Self:    "/api/v1/audio/" + f.ID.String(),
			Content: "/api/v1/audio/" + f.ID.String() + "/content",
		},
	}
}

// --- AudioChecksum (minimal) ---

type audioChecksumAttributes struct {
	FileType string `json:"fileType"`
	Checksum string `json:"checksum"`
}

func MarshalAudioChecksum(f *models.SystemAudioFile) Resource {
	return Resource{
		Type: "audio-files",
		ID:   f.ID.String(),
		Attributes: audioChecksumAttributes{
			FileType: string(f.FileType),
			Checksum: f.Checksum,
		},
	}
}

// --- SystemState ---

type systemStateAttributes struct {
	State       string    `json:"state"`
	LastUpdated time.Time `json:"lastUpdated"`
}

func MarshalSystemState(state string, lastUpdated time.Time) Resource {
	return Resource{
		Type: "system-state",
		ID:   "current",
		Attributes: systemStateAttributes{
			State:       state,
			LastUpdated: lastUpdated,
		},
	}
}
