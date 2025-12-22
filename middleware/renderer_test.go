package middleware

import (
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestDefaultRenderer_IsFileRenderer(t *testing.T) {
	if DefaultRenderer == nil {
		t.Fatal("DefaultRenderer should not be nil")
	}
	_, ok := DefaultRenderer.(*FileRenderer)
	if !ok {
		t.Error("DefaultRenderer should be a *FileRenderer")
	}
}

func TestDefaultRenderer_CanBeSwapped(t *testing.T) {
	// Save original
	original := DefaultRenderer
	defer func() { DefaultRenderer = original }()

	// Create mock that returns error
	mock := &mockRenderer{err: errors.New("mock error")}
	DefaultRenderer = mock

	// Verify it was swapped
	if DefaultRenderer != mock {
		t.Error("DefaultRenderer should be the mock")
	}
}

// mockRenderer is a simple mock for testing DefaultRenderer swapping
type mockRenderer struct {
	err error
}

func (m *mockRenderer) Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error {
	if m.err != nil {
		return m.err
	}
	_, _ = w.Write([]byte("mock content"))
	return nil
}

func (m *mockRenderer) RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	return m.Render(w, tmpl, nil, data, "templates/"+tmpl)
}

func TestFileRenderer_Render_ParseError(t *testing.T) {
	renderer := &FileRenderer{}
	rr := httptest.NewRecorder()

	err := renderer.Render(rr, "nonexistent.html", nil, nil, "templates/nonexistent.html")

	if err == nil {
		t.Error("expected error for nonexistent template, got nil")
	}
}

// changeToProjectRootRenderer changes to the project root for template access
func changeToProjectRootRenderer(t *testing.T) func() {
	t.Helper()
	originalDir, _ := os.Getwd()
	_ = os.Chdir("..")
	return func() { _ = os.Chdir(originalDir) }
}

func TestFileRenderer_Render_Success(t *testing.T) {
	restoreDir := changeToProjectRootRenderer(t)
	defer restoreDir()

	renderer := &FileRenderer{}
	rr := httptest.NewRecorder()

	err := renderer.Render(rr, "import_error.html", nil, nil, "templates/import_error.html")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if rr.Body.Len() == 0 {
		t.Error("expected response body to have content")
	}
	// Verify actual template content - should contain "Import Error"
	body := rr.Body.String()
	if !strings.Contains(body, "Import Error") {
		t.Errorf("expected body to contain 'Import Error' from template, got %q", body)
	}
}
