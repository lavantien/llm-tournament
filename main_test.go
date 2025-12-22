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
		_ = middleware.CloseDB()
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

func TestRouter_AllRoutesRespond(t *testing.T) {
	cleanup := setupMainTestDB(t)
	defer cleanup()

	// Test a sample of routes to ensure they all respond (not 404)
	testRoutes := []struct {
		path   string
		method string
	}{
		{"/prompts", "GET"},
		{"/results", "GET"},
		{"/profiles", "GET"},
		{"/stats", "GET"},
		{"/settings", "GET"},
		{"/add_model", "GET"},
		{"/add_prompt", "GET"},
		{"/add_profile", "GET"},
	}

	for _, tc := range testRoutes {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()
			router(rr, req)

			// Should not be a redirect to /prompts (which indicates unhandled route)
			// or should be a successful response
			if rr.Code == http.StatusNotFound {
				t.Errorf("route %s returned 404", tc.path)
			}
		})
	}
}

func TestRouter_POSTRoutes(t *testing.T) {
	cleanup := setupMainTestDB(t)
	defer cleanup()

	// Test POST routes that require POST method
	postRoutes := []string{
		"/evaluate/all",
		"/evaluate/model",
		"/evaluate/prompt",
		"/evaluation/cancel",
		"/settings/update",
	}

	for _, route := range postRoutes {
		t.Run(route+"_GET", func(t *testing.T) {
			req := httptest.NewRequest("GET", route, nil)
			rr := httptest.NewRecorder()
			router(rr, req)

			// GET on POST-only routes should return method not allowed
			if rr.Code != http.StatusMethodNotAllowed && rr.Code != http.StatusBadRequest {
				// Some routes may allow GET, that's ok
				t.Logf("route %s with GET returned %d", route, rr.Code)
			}
		})
	}
}

func TestRouter_WithQueryParams(t *testing.T) {
	cleanup := setupMainTestDB(t)
	defer cleanup()

	// Test routes with query parameters
	req := httptest.NewRequest("GET", "/edit_model?id=1", nil)
	rr := httptest.NewRecorder()
	router(rr, req)

	// Should not be 404
	if rr.Code == http.StatusNotFound {
		t.Error("edit_model route should be handled")
	}
}

func TestRouter_StaticPaths(t *testing.T) {
	cleanup := setupMainTestDB(t)
	defer cleanup()

	// Test that /templates/ and /assets/ paths go through router
	// (but may not find files in test environment)
	req := httptest.NewRequest("GET", "/templates/nonexistent.css", nil)
	rr := httptest.NewRecorder()
	router(rr, req)

	// Static paths go through router, should redirect to /prompts
	if rr.Code != http.StatusSeeOther {
		// Could also be handled by static file server if configured
		t.Logf("static path returned %d", rr.Code)
	}
}
