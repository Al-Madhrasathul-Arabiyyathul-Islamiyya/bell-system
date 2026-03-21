package jsonapi

import (
	"net/http"
	"strings"
)

// SortField represents a parsed sort parameter.
type SortField struct {
	Field string
	Desc  bool
}

// ParseSort parses the `sort` query parameter. Prefix `-` means descending.
func ParseSort(r *http.Request) []SortField {
	raw := r.URL.Query().Get("sort")
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	fields := make([]SortField, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "-") {
			fields = append(fields, SortField{Field: p[1:], Desc: true})
		} else {
			fields = append(fields, SortField{Field: p, Desc: false})
		}
	}
	return fields
}

// ParseFilter extracts filter[key] parameters, only allowing specified keys.
func ParseFilter(r *http.Request, allowedKeys []string) map[string]string {
	allowed := make(map[string]bool, len(allowedKeys))
	for _, k := range allowedKeys {
		allowed[k] = true
	}

	filters := make(map[string]string)
	for key, values := range r.URL.Query() {
		if strings.HasPrefix(key, "filter[") && strings.HasSuffix(key, "]") {
			name := key[7 : len(key)-1]
			if allowed[name] && len(values) > 0 && values[0] != "" {
				filters[name] = values[0]
			}
		}
	}
	return filters
}

// SortToSQL converts parsed sort fields into a SQL ORDER BY clause using a column whitelist.
// Unknown field names are silently ignored. Returns defaultOrder if no valid sorts found.
func SortToSQL(sorts []SortField, columnMap map[string]string, defaultOrder string) string {
	var parts []string
	for _, s := range sorts {
		col, ok := columnMap[s.Field]
		if !ok {
			continue
		}
		if s.Desc {
			parts = append(parts, col+" DESC")
		} else {
			parts = append(parts, col+" ASC")
		}
	}
	if len(parts) == 0 {
		return defaultOrder
	}
	return "ORDER BY " + strings.Join(parts, ", ")
}

// ParseInclude parses the `include` query parameter into a slice of relationship names.
func ParseInclude(r *http.Request) []string {
	raw := r.URL.Query().Get("include")
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	includes := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			includes = append(includes, p)
		}
	}
	return includes
}
