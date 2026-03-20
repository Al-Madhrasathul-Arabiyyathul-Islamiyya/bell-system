package jsonapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RequestDocument represents an incoming JSON:API request body.
type RequestDocument struct {
	Data RequestResource `json:"data"`
}

// RequestResource is the data portion of a JSON:API request.
type RequestResource struct {
	Type          string                     `json:"type"`
	ID            string                     `json:"id,omitempty"`
	Attributes    json.RawMessage            `json:"attributes"`
	Relationships map[string]RequestRelation `json:"relationships,omitempty"`
}

// RequestRelation is a relationship in a JSON:API request.
type RequestRelation struct {
	Data *ResourceIdentifier `json:"data"`
}

// ParseRequest decodes a JSON:API request body and validates the resource type.
func ParseRequest(r *http.Request, expectedType string) (*RequestDocument, error) {
	var doc RequestDocument
	if err := json.NewDecoder(r.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("invalid JSON body")
	}
	if doc.Data.Type != expectedType {
		return nil, fmt.Errorf("expected resource type %q, got %q", expectedType, doc.Data.Type)
	}
	return &doc, nil
}

// UnmarshalAttributes decodes the raw attributes into a typed struct.
func (d *RequestDocument) UnmarshalAttributes(v any) error {
	if d.Data.Attributes == nil {
		return fmt.Errorf("missing attributes")
	}
	return json.Unmarshal(d.Data.Attributes, v)
}

// RelationshipID extracts the ID from a named relationship, or returns empty string.
func (d *RequestDocument) RelationshipID(name string) string {
	rel, ok := d.Data.Relationships[name]
	if !ok || rel.Data == nil {
		return ""
	}
	return rel.Data.ID
}
