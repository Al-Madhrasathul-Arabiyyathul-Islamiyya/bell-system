package jsonapi

// Resource represents a JSON:API resource object.
type Resource struct {
	Type          string              `json:"type"`
	ID            string              `json:"id"`
	Attributes    any                 `json:"attributes"`
	Relationships map[string]Relation `json:"relationships,omitempty"`
	Links         *ResourceLinks      `json:"links,omitempty"`
}

// ResourceIdentifier is a JSON:API resource identifier (type + id only).
type ResourceIdentifier struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// Relation represents a JSON:API relationship object.
type Relation struct {
	Data any `json:"data"` // *ResourceIdentifier, []ResourceIdentifier, or nil
}

// ResourceLinks contains self/related links for a resource.
type ResourceLinks struct {
	Self    string `json:"self,omitempty"`
	Content string `json:"content,omitempty"`
}

// Document is a JSON:API single-resource response document.
type Document struct {
	Data     *Resource  `json:"data"`
	Included []Resource `json:"included,omitempty"`
}

// CollectionDocument is a JSON:API collection response document.
type CollectionDocument struct {
	Data     []Resource     `json:"data"`
	Meta     map[string]any `json:"meta"`
	Links    map[string]any `json:"links"`
	Included []Resource     `json:"included,omitempty"`
}

// ErrorDocument is a JSON:API error response document.
type ErrorDocument struct {
	Errors []Error `json:"errors"`
}

// Error represents a single JSON:API error object.
type Error struct {
	Status string       `json:"status"`
	Code   string       `json:"code"`
	Title  string       `json:"title"`
	Detail string       `json:"detail,omitempty"`
	Source *ErrorSource `json:"source,omitempty"`
}

// ErrorSource identifies the source of a JSON:API error.
type ErrorSource struct {
	Pointer   string `json:"pointer,omitempty"`
	Parameter string `json:"parameter,omitempty"`
}
