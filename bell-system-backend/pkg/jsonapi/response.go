package jsonapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const ContentType = "application/vnd.api+json"

// WriteResource writes a single-resource JSON:API document.
func WriteResource(w http.ResponseWriter, status int, doc Document) {
	writeJSONAPI(w, status, doc)
}

// WriteCollection writes a collection JSON:API document.
func WriteCollection(w http.ResponseWriter, status int, doc CollectionDocument) {
	writeJSONAPI(w, status, doc)
}

// WriteError writes a JSON:API error document.
func WriteError(w http.ResponseWriter, status int, code, title, detail string) {
	doc := ErrorDocument{
		Errors: []Error{
			{
				Status: strconv.Itoa(status),
				Code:   code,
				Title:  codeToTitle(code),
				Detail: detail,
			},
		},
	}
	if title != "" {
		doc.Errors[0].Title = title
	}
	writeJSONAPI(w, status, doc)
}

// WriteNoContent writes a 204 No Content response.
func WriteNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func writeJSONAPI(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", ContentType)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// codeToTitle converts a snake_case error code to Title Case.
func codeToTitle(code string) string {
	parts := strings.Split(code, "_")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
