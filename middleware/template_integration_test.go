package middleware

import (
	"html/template"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderTemplate_MissingRequiredField tests template rendering when required data fields are missing
func TestRenderTemplate_MissingRequiredField(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	// Pass data with missing/nil fields that templates might expect
	data := map[string]interface{}{
		"PageName": "Test",
		// Missing other fields that templates might reference
	}

	// Should not panic even with missing fields
	RenderTemplate(rr, "import_error.html", data)

	// Should complete successfully (templates handle missing fields gracefully)
	if rr.Code != 200 {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// TestRenderTemplate_EmptyData tests template rendering with completely empty data
func TestRenderTemplate_EmptyData(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	// Empty data map
	data := map[string]interface{}{}

	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != 200 {
		t.Errorf("expected status 200 with empty data, got %d", rr.Code)
	}
}

// TestRenderTemplate_NilData tests template rendering with nil data
func TestRenderTemplate_NilData(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	// Nil data
	RenderTemplate(rr, "import_error.html", nil)

	if rr.Code != 200 {
		t.Errorf("expected status 200 with nil data, got %d", rr.Code)
	}
}

// TestRenderTemplate_LargeDataStructure tests rendering with a large dataset
func TestRenderTemplate_LargeDataStructure(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer func() { _ = CloseDB() }()

	// Create 100 prompts
	prompts := make([]Prompt, 100)
	for i := 0; i < 100; i++ {
		prompts[i] = Prompt{
			Text:     "Test prompt " + string(rune('A'+i%26)),
			Profile:  "Profile " + string(rune('A'+i%10)),
			Solution: "Solution " + string(rune('A'+i%26)),
		}
	}

	// Create 50 models
	models := make([]string, 50)
	for i := 0; i < 50; i++ {
		models[i] = "Model-" + string(rune('A'+i%26)) + string(rune('0'+i%10))
	}

	// Create results structure
	results := make(map[string]Result)
	for _, model := range models {
		scores := make([]int, len(prompts))
		for j := range scores {
			scores[j] = (j * 17) % 100 // Varied scores
		}
		results[model] = Result{Scores: scores}
	}

	data := map[string]interface{}{
		"PageName": "Large Data Test",
		"Prompts":  prompts,
		"Models":   models,
		"Results":  results,
	}

	rr := httptest.NewRecorder()

	// This tests performance and memory handling
	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != 200 {
		t.Errorf("expected status 200 with large data, got %d", rr.Code)
	}

	// Verify some content was rendered
	if rr.Body.Len() == 0 {
		t.Error("expected non-empty response body")
	}
}

// TestFileRenderer_MalformedTemplate tests error handling for malformed template syntax
func TestFileRenderer_MalformedTemplate(t *testing.T) {
	renderer := &FileRenderer{}
	rr := httptest.NewRecorder()

	// Create a temporary malformed template
	malformedTemplate := `{{define "test"}}
		{{ .Name
		Missing closing braces
	{{end}}`

	// Try to parse malformed template directly
	_, err := template.New("test").Parse(malformedTemplate)

	if err == nil {
		t.Error("expected error for malformed template syntax")
	}

	// The renderer would fail when trying to parse such a template
	err = renderer.Render(rr, "test", nil, nil, "nonexistent_malformed.html")

	if err == nil {
		t.Error("expected error when rendering malformed template")
	}
}

// TestRenderTemplate_WithFuncMap tests template rendering with FuncMap
func TestRenderTemplate_WithFuncMap(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	data := map[string]interface{}{
		"PageName": "FuncMap Test",
		"Score":    85,
		"Text":     "**bold** _italic_",
	}

	// RenderTemplate uses templates.FuncMap which includes markdown, inc, etc.
	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != 200 {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

// TestRenderTemplate_NestedData tests rendering with deeply nested data structures
func TestRenderTemplate_NestedData(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	// Create nested data structure
	data := map[string]interface{}{
		"PageName": "Nested Data Test",
		"Outer": map[string]interface{}{
			"Middle": map[string]interface{}{
				"Inner": map[string]interface{}{
					"Value": "deeply nested",
				},
			},
		},
		"Array": []map[string]interface{}{
			{"Name": "Item 1", "Value": 10},
			{"Name": "Item 2", "Value": 20},
			{"Name": "Item 3", "Value": 30},
		},
	}

	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != 200 {
		t.Errorf("expected status 200 with nested data, got %d", rr.Code)
	}
}

// TestRenderTemplate_SpecialCharactersInData tests handling of special characters
func TestRenderTemplate_SpecialCharactersInData(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	data := map[string]interface{}{
		"PageName": "Special <script>alert('xss')</script>",
		"Text":     `"quotes" & <tags> and 'apostrophes'`,
		"HTML":     `<div onclick="alert(1)">Click me</div>`,
	}

	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != 200 {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Verify rendering completes successfully with special characters
	// Go templates auto-escape in HTML context - detailed XSS testing in Phase 3.2
	body := rr.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty response body")
	}
}

// TestRenderTemplate_UnicodeData tests handling of unicode characters
func TestRenderTemplate_UnicodeData(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	data := map[string]interface{}{
		"PageName": "Unicode Test",
		"Text":     "Hello 世界 🌍 Привет мир 🚀",
		"Emoji":    "😀 😃 😄 😁 🎉",
	}

	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != 200 {
		t.Errorf("expected status 200 with unicode data, got %d", rr.Code)
	}

	// Verify unicode is preserved (check that at least the page was rendered)
	body := rr.Body.String()
	if body == "" {
		t.Error("expected non-empty response body with unicode data")
	}
}

// TestFileRenderer_Render_ExecuteError tests error during template execution
func TestFileRenderer_Render_ExecuteError(t *testing.T) {
	dbPath, cleanup := setupTestDB(t)
	defer cleanup()

	if err := InitDB(dbPath); err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	// A template that parses cleanly but dereferences a field on a plain
	// int, so the real renderer must fail at execute time. (A missing map
	// key alone is lenient: templates render it as nil without error.)
	tmplPath := filepath.Join(t.TempDir(), "execfail.html")
	if err := os.WriteFile(tmplPath, []byte(`{{ .Count.Nested }}`), 0o644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	renderer := &FileRenderer{}
	rr := httptest.NewRecorder()
	err := renderer.Render(rr, "execfail.html", nil, map[string]interface{}{"Count": 0}, tmplPath)
	if err == nil {
		t.Fatal("expected FileRenderer.Render to return the execute error")
	}
	if !strings.Contains(err.Error(), "can't evaluate field Nested") {
		t.Fatalf("expected error to reference the missing field, got %v", err)
	}
}

// TestRenderTemplate_ZeroValues tests rendering with zero values
func TestRenderTemplate_ZeroValues(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	data := map[string]interface{}{
		"PageName": "",
		"Count":    0,
		"Score":    0.0,
		"Enabled":  false,
		"Items":    []string{},
		"Map":      map[string]string{},
	}

	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != 200 {
		t.Errorf("expected status 200 with zero values, got %d", rr.Code)
	}
}

// TestRenderTemplate_MultipleTemplates tests rendering with multiple template files
func TestRenderTemplate_MultipleTemplates(t *testing.T) {
	restoreDir := changeToProjectRootMiddleware(t)
	defer restoreDir()

	rr := httptest.NewRecorder()

	data := map[string]interface{}{
		"PageName": "Multi-Template Test",
	}

	// RenderTemplate loads both main template and nav.html
	RenderTemplate(rr, "import_error.html", data)

	if rr.Code != 200 {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	// Both templates should be loaded and rendered
	if rr.Body.Len() == 0 {
		t.Error("expected non-empty response from multi-template render")
	}
}
