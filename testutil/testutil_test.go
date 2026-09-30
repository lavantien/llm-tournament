package testutil

import (
	"errors"
	"net/http/httptest"
	"testing"
)

func TestMockRenderer_Render_RecordsCallAndWritesContent(t *testing.T) {
	rec := httptest.NewRecorder()

	renderer := &MockRenderer{}
	if err := renderer.Render(rec, "example.tmpl", nil, map[string]string{"k": "v"}, "a.tmpl", "b.tmpl"); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}

	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}
	call := renderer.RenderCalls[0]
	if call.Name != "example.tmpl" {
		t.Fatalf("expected call.Name=%q, got %q", "example.tmpl", call.Name)
	}
	if got := rec.Body.String(); got != "mock rendered" {
		t.Fatalf("expected response body %q, got %q", "mock rendered", got)
	}
}

func TestMockRenderer_Render_ReturnsErrorWithoutWriting(t *testing.T) {
	rec := httptest.NewRecorder()

	renderer := &MockRenderer{RenderError: errors.New("boom")}
	if err := renderer.Render(rec, "example.tmpl", nil, nil, "a.tmpl"); err == nil {
		t.Fatalf("expected Render to return an error")
	}

	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}
	if got := rec.Body.String(); got != "" {
		t.Fatalf("expected response body to be empty on error, got %q", got)
	}
}

func TestMockRenderer_RenderTemplateSimple_DelegatesToRender(t *testing.T) {
	rec := httptest.NewRecorder()

	renderer := &MockRenderer{}
	if err := renderer.RenderTemplateSimple(rec, "example.tmpl", 123); err != nil {
		t.Fatalf("RenderTemplateSimple returned error: %v", err)
	}

	if len(renderer.RenderCalls) != 1 {
		t.Fatalf("expected 1 render call, got %d", len(renderer.RenderCalls))
	}
	call := renderer.RenderCalls[0]
	if call.Name != "example.tmpl" {
		t.Fatalf("expected call.Name=%q, got %q", "example.tmpl", call.Name)
	}
	if len(call.Files) != 1 || call.Files[0] != "templates/example.tmpl" {
		t.Fatalf("expected call.Files to include templates path, got %#v", call.Files)
	}
}
