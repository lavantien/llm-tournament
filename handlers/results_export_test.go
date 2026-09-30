package handlers

import (
	"encoding/json"
	"errors"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestExportResultsHandler(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add a model with results
	form := url.Values{}
	form.Add("model", "ExportTestModel")
	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req)

	// Export results
	exportReq := httptest.NewRequest("GET", "/export_results", nil)
	exportRR := httptest.NewRecorder()
	ExportResultsHandler(exportRR, exportReq)

	if exportRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, exportRR.Code)
	}

	// Verify JSON content type
	contentType := exportRR.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected content type 'application/json', got %q", contentType)
	}
}

func TestExportResultsHandler_GET(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test prompts and results
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt"},
	})
	if err != nil {
		t.Fatalf("failed to write test prompts: %v", err)
	}

	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, map[string]middleware.Result{
		"TestModel": {Scores: []int{80}},
	})
	if err != nil {
		t.Fatalf("failed to write test results: %v", err)
	}

	req := httptest.NewRequest("GET", "/export_results", nil)
	rr := httptest.NewRecorder()
	ExportResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Verify CSV content
	body := rr.Body.String()
	if !strings.Contains(body, "Model") {
		t.Error("expected 'Model' header in CSV output")
	}
}

func TestExportResultsHandler_GET_WithData(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	// Add test data
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Test prompt"}})
	_ = middleware.WriteResults("default", map[string]middleware.Result{
		"Model1": {Scores: []int{80}},
	})

	req := httptest.NewRequest("GET", "/export_results", nil)
	rr := httptest.NewRecorder()
	ExportResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", contentType)
	}

	// Verify JSON is valid
	var data map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &data)
	if err != nil {
		t.Errorf("expected valid JSON, got error: %v", err)
	}
}

func TestExportResultsHandler_EmptyResults(t *testing.T) {
	cleanup := setupResultsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/export_results", nil)
	rr := httptest.NewRecorder()
	ExportResultsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

// recordingFailWriter fails every write like FailingResponseWriter but counts
// the attempts, so tests can prove the handler actually tried to write.
type recordingFailWriter struct {
	http.ResponseWriter
	writeCalls int
}

func (w *recordingFailWriter) Write(p []byte) (int, error) {
	w.writeCalls++
	return 0, errors.New("mock write error")
}

func TestExportResultsHandler_WriteError(t *testing.T) {
	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{80}},
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	rr := httptest.NewRecorder()
	failingWriter := &recordingFailWriter{ResponseWriter: rr}

	req := httptest.NewRequest("GET", "/export_results", nil)
	handler.ExportResults(failingWriter, req)

	if failingWriter.writeCalls == 0 {
		t.Fatal("expected the handler to attempt writing the export body")
	}
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d after the write fails, got %d", http.StatusInternalServerError, rr.Code)
	}
}
