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

// changeToProjectRootMiddleware changes to the project root directory for tests that need templates
func changeToProjectRootMiddleware(t *testing.T) func() {
	t.Helper()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(".."); err != nil {
		t.Fatalf("failed to change to project root: %v", err)
	}
	return func() {
		os.Chdir(originalDir)
	}
}

func TestHandleFormError(t *testing.T) {
	rr := httptest.NewRecorder()
	testErr := errors.New("test form parsing error")

	HandleFormError(rr, testErr)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Errorf("expected body to contain 'Error parsing form', got %q", rr.Body.String())
	}
}

func TestRespondJSON_Success(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]interface{}{
		"success": true,
		"message": "test message",
	}

	RespondJSON(rr, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "success") {
		t.Errorf("expected body to contain 'success', got %q", body)
	}
	if !strings.Contains(body, "test message") {
		t.Errorf("expected body to contain 'test message', got %q", body)
	}
}

func TestRespondJSON_Array(t *testing.T) {
	rr := httptest.NewRecorder()
	data := []string{"item1", "item2", "item3"}

	RespondJSON(rr, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "item1") {
		t.Errorf("expected body to contain 'item1', got %q", body)
	}
}

func TestRespondJSON_EmptyObject(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]interface{}{}

	RespondJSON(rr, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := strings.TrimSpace(rr.Body.String())
	if body != "{}" {
		t.Errorf("expected body '{}', got %q", body)
	}
}

func TestRespondJSON_Null(t *testing.T) {
	rr := httptest.NewRecorder()

	RespondJSON(rr, nil)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := strings.TrimSpace(rr.Body.String())
	if body != "null" {
		t.Errorf("expected body 'null', got %q", body)
	}
}

func TestRespondJSON_Struct(t *testing.T) {
	rr := httptest.NewRecorder()
	data := struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}{
		Name:  "test",
		Value: 42,
	}

	RespondJSON(rr, data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, `"name":"test"`) {
		t.Errorf("expected body to contain '\"name\":\"test\"', got %q", body)
	}
	if !strings.Contains(body, `"value":42`) {
		t.Errorf("expected body to contain '\"value\":42', got %q", body)
	}
}

func TestRenderTemplate_InvalidTemplate(t *testing.T) {
	rr := httptest.NewRecorder()

	// Try to render a non-existent template
	RenderTemplate(rr, "nonexistent.html", nil)

	// Should return error since template file doesn't exist
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for invalid template, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestImportErrorHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	req := httptest.NewRequest("GET", "/import_error", nil)
	rr := httptest.NewRecorder()
	ImportErrorHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestRenderTemplate_Success(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()
	data := map[string]interface{}{
		"PageName": "Test Page",
	}

	// Render a simple template that exists
	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestRenderTemplateSimple_Success(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()
	data := map[string]string{"Model": "test-model"}

	err := RenderTemplateSimple(rr, "edit_model.html", data)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestRenderTemplateSimple_InvalidTemplate(t *testing.T) {
	rr := httptest.NewRecorder()

	err := RenderTemplateSimple(rr, "nonexistent.html", nil)

	if err == nil {
		t.Error("expected error for nonexistent template")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestRenderTemplate_WithMockRenderer_Error(t *testing.T) {
	// Save original and restore after test
	original := DefaultRenderer
	defer func() { DefaultRenderer = original }()

	// Create mock that returns error
	mock := &testMockRenderer{err: errors.New("mock render error")}
	DefaultRenderer = mock

	rr := httptest.NewRecorder()
	RenderTemplate(rr, "test.html", nil)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error rendering template") {
		t.Errorf("expected error message in body, got %q", rr.Body.String())
	}
}

func TestRenderTemplateSimple_WithMockRenderer_Error(t *testing.T) {
	// Save original and restore after test
	original := DefaultRenderer
	defer func() { DefaultRenderer = original }()

	// Create mock that returns error
	mock := &testMockRenderer{err: errors.New("mock render error")}
	DefaultRenderer = mock

	rr := httptest.NewRecorder()
	err := RenderTemplateSimple(rr, "test.html", nil)

	if err == nil {
		t.Error("expected error from mock renderer")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
}

// testMockRenderer is a simple mock for testing error paths
type testMockRenderer struct {
	err error
}

func (m *testMockRenderer) Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error {
	if m.err != nil {
		return m.err
	}
	w.Write([]byte("mock content"))
	return nil
}

func (m *testMockRenderer) RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	return m.Render(w, tmpl, nil, data, "templates/"+tmpl)
}

func TestRespondJSON_Unmarshalable(t *testing.T) {
	rr := httptest.NewRecorder()

	// Create a channel which cannot be marshaled to JSON
	ch := make(chan int)
	RespondJSON(rr, ch)

	// Should return error status
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for unmarshalable data, got %d", http.StatusInternalServerError, rr.Code)
	}
}
