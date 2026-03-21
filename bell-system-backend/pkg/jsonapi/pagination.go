package jsonapi

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// PaginationParams holds parsed pagination parameters.
type PaginationParams struct {
	Page int
	Size int
}

// Offset returns the SQL offset for the current page.
func (p PaginationParams) Offset() int {
	return (p.Page - 1) * p.Size
}

// ParsePagination extracts page[number] and page[size] from the request.
func ParsePagination(r *http.Request) PaginationParams {
	page, _ := strconv.Atoi(r.URL.Query().Get("page[number]"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page[size]"))

	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = DefaultPageSize
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}

	return PaginationParams{Page: page, Size: size}
}

// PaginationMeta returns the meta object for a paginated response.
func PaginationMeta(total, page, size int) map[string]any {
	totalPages := int(math.Ceil(float64(total) / float64(size)))
	return map[string]any{
		"total":      total,
		"page":       page,
		"pageSize":   size,
		"totalPages": totalPages,
	}
}

// PaginationLinks returns the links object for a paginated response.
func PaginationLinks(basePath string, params PaginationParams, total int) map[string]any {
	totalPages := max(int(math.Ceil(float64(total)/float64(params.Size))), 1)

	link := func(page int) string {
		return fmt.Sprintf("%s?page[number]=%d&page[size]=%d", basePath, page, params.Size)
	}

	links := map[string]any{
		"self":  link(params.Page),
		"first": link(1),
		"last":  link(totalPages),
	}

	if params.Page > 1 {
		links["prev"] = link(params.Page - 1)
	}
	if params.Page < totalPages {
		links["next"] = link(params.Page + 1)
	}

	return links
}
