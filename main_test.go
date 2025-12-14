package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"llm-tournament/middleware"

	_ "github.com/mattn/go-sqlite3"
)

// setupMainTestDB creates a test database for main tests
func setupMainTestDB(t *testing.T) func() {
	t.Helper()
	dbPath := t.TempDir() + "/test.db"
	err := middleware.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}
	return func() {
		middleware.CloseDB()
	}
}

func TestRoutes_AllRoutesDefined(t *testing.T) {
	expectedRoutes := []string{
		"/import_error",
		"/prompts",
		"/add_model",
		"/edit_model",
		"/delete_model",
		"/add_prompt",
		"/edit_prompt",
		"/delete_prompt",
		"/move_prompt",
		"/import_results",
		"/export_prompts",
		"/import_prompts",
		"/update_prompts_order",
		"/reset_prompts",
		"/bulk_delete_prompts",
		"/prompts/suites/new",
		"/prompts/suites/edit",
		"/prompts/suites/delete",
		"/prompts/suites/select",
		"/results",
		"/update_result",
		"/reset_results",
		"/confirm_refresh_results",
		"/refresh_results",
		"/export_results",
		"/update_mock_results",
		"/evaluate",
		"/profiles",
		"/add_profile",
		"/edit_profile",
		"/delete_profile",
		"/reset_profiles",
		"/stats",
		"/settings",
		"/settings/update",
		"/settings/test_key",
		"/evaluate/all",
		"/evaluate/model",
		"/evaluate/prompt",
		"/evaluation/progress",
		"/evaluation/cancel",
	}

	for _, route := range expectedRoutes {
		if _, exists := routes[route]; !exists {
			t.Errorf("expected route %q to be defined", route)
		}
	}
}

func TestRouter_KnownRoute(t *testing.T) {
	cleanup := setupMainTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	router(rr, req)

	// Prompts handler should return OK (or redirect)
	// The actual status depends on template availability
	if rr.Code == http.StatusNotFound {
		t.Errorf("expected router to handle /prompts route")
	}
}

func TestRouter_UnknownRoute_Redirects(t *testing.T) {
	cleanup := setupMainTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/unknown_route", nil)
	rr := httptest.NewRecorder()
	router(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for unknown route, got %d", http.StatusSeeOther, rr.Code)
	}

	// Check redirect location
	location := rr.Header().Get("Location")
	if location != "/prompts" {
		t.Errorf("expected redirect to /prompts, got %q", location)
	}
}

func TestRouter_RootRoute_Redirects(t *testing.T) {
	cleanup := setupMainTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	router(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for root route, got %d", http.StatusSeeOther, rr.Code)
	}

	location := rr.Header().Get("Location")
	if location != "/prompts" {
		t.Errorf("expected redirect to /prompts, got %q", location)
	}
}

func TestRoutesCount(t *testing.T) {
	// Ensure we have the expected number of routes
	expectedCount := 41
	if len(routes) != expectedCount {
		t.Errorf("expected %d routes, got %d", expectedCount, len(routes))
	}
}
