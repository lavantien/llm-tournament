package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter_UnknownRoute_RedirectsToPrompts(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rr := httptest.NewRecorder()

	router(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
	if location := rr.Header().Get("Location"); location != "/prompts" {
		t.Fatalf("expected redirect to %q, got %q", "/prompts", location)
	}
}

func TestRouter_KnownRoute_DoesNotRedirect(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/import_error", nil)
	rr := httptest.NewRecorder()

	router(rr, req)

	if rr.Code == http.StatusSeeOther {
		t.Fatalf("expected known route to not redirect, got %d", rr.Code)
	}
}

func TestRegisterRoutes_InstallsRouterOnMux(t *testing.T) {
	mux := http.NewServeMux()
	registerRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Fatalf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
	if location := rr.Header().Get("Location"); location != "/prompts" {
		t.Fatalf("expected redirect to %q, got %q", "/prompts", location)
	}
}
