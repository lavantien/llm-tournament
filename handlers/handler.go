package handlers

import (
	"llm-tournament/middleware"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	DataStore middleware.DataStore
	Renderer  middleware.TemplateRenderer
}

// NewHandler creates a new Handler with default dependencies
func NewHandler() *Handler {
	return &Handler{
		DataStore: middleware.DefaultDataStore,
		Renderer:  middleware.DefaultRenderer,
	}
}

// NewHandlerWithDeps creates a new Handler with custom dependencies (for testing)
func NewHandlerWithDeps(ds middleware.DataStore, r middleware.TemplateRenderer) *Handler {
	return &Handler{
		DataStore: ds,
		Renderer:  r,
	}
}

// DefaultHandler is the default handler instance used by HTTP routes
var DefaultHandler = NewHandler()
