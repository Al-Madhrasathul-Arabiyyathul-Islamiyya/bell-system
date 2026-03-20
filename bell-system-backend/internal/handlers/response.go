package handlers

import (
	"encoding/json"
	"net/http"

	"arabiyya.edu.mv/bell-system-backend/pkg/jsonapi"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeJSONAPI(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", jsonapi.ContentType)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	jsonapi.WriteError(w, status, code, "", message)
}
