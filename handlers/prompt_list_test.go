package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestBulkDeletePromptsHandler_Success(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add some prompts
	for i := 0; i < 3; i++ {
		form := url.Values{}
		form.Add("prompt", "Bulk delete test "+string(rune('A'+i)))
		req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		AddPromptHandler(httptest.NewRecorder(), req)
	}

	// Bulk delete indices 0 and 2
	requestBody := map[string][]int{"indices": {0, 2}}
	jsonBody, _ := json.Marshal(requestBody)

	bulkReq := httptest.NewRequest("POST", "/bulk_delete_prompts", bytes.NewReader(jsonBody))
	bulkReq.Header.Set("Content-Type", "application/json")

	bulkRR := httptest.NewRecorder()
	BulkDeletePromptsHandler(bulkRR, bulkReq)

	if bulkRR.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, bulkRR.Code)
	}

	// Verify only 1 prompt remains
	prompts := middleware.ReadPrompts()
	if len(prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(prompts))
	}
}

func TestBulkDeletePromptsHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/bulk_delete_prompts", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestBulkDeletePromptsHandler_EmptyIndices(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add a prompt
	form := url.Values{}
	form.Add("prompt", "Test prompt")
	req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddPromptHandler(httptest.NewRecorder(), req)

	// Bulk delete with empty indices
	requestBody := map[string][]int{"indices": {}}
	jsonBody, _ := json.Marshal(requestBody)

	bulkReq := httptest.NewRequest("POST", "/bulk_delete_prompts", bytes.NewReader(jsonBody))
	bulkReq.Header.Set("Content-Type", "application/json")

	bulkRR := httptest.NewRecorder()
	BulkDeletePromptsHandler(bulkRR, bulkReq)

	if bulkRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, bulkRR.Code)
	}
}

func TestBulkDeletePromptsPageHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/bulk_delete_prompts_page", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsPageHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestBulkDeletePromptsPageHandler_NoIndices(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/bulk_delete_prompts_page", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsPageHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestBulkDeletePromptsPageHandler_InvalidIndicesJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/bulk_delete_prompts_page?indices=not_json", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsPageHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestBulkDeletePromptsPage_RenderError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{RenderError: errors.New("mock render error")})

	req := httptest.NewRequest(http.MethodGet, "/bulk_delete_prompts_page?indices=[0]", nil)
	rr := httptest.NewRecorder()
	handler.BulkDeletePromptsPage(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error rendering template") {
		t.Fatalf("expected render error message, got %q", rr.Body.String())
	}
}

func TestPromptListHandler_InvalidOrderFilter(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/prompts?order_filter=invalid", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestBulkDeletePrompts_InvalidJSONBody(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/bulk_delete_prompts", strings.NewReader("not json"))
	rr := httptest.NewRecorder()
	handler.BulkDeletePrompts(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error decoding request") {
		t.Fatalf("expected decode error message, got %q", rr.Body.String())
	}
}

func TestBulkDeletePrompts_NoPrompts(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/bulk_delete_prompts", strings.NewReader(`{"indices":[0]}`))
	rr := httptest.NewRecorder()
	handler.BulkDeletePrompts(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "No prompts to delete") {
		t.Fatalf("expected no prompts message, got %q", rr.Body.String())
	}
}

type promptListListSuitesErrorDataStore struct {
	MockDataStore
	Err error
}

func (ds *promptListListSuitesErrorDataStore) ListPromptSuites() ([]string, error) {
	return nil, ds.Err
}

type promptListReadSuiteErrorDataStore struct {
	MockDataStore
	Err error
}

func (ds *promptListReadSuiteErrorDataStore) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	return nil, ds.Err
}

type promptListEmptySuiteNameDataStore struct {
	MockDataStore
	GetCalls  int
	ReadCalls int
}

func (ds *promptListEmptySuiteNameDataStore) GetCurrentSuiteName() string {
	ds.GetCalls++
	return ""
}

func (ds *promptListEmptySuiteNameDataStore) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	ds.ReadCalls++
	return nil, nil
}

type promptListSecondReadErrorDataStore struct {
	MockDataStore
	ReadCalls int
}

func (ds *promptListSecondReadErrorDataStore) ReadPromptSuite(suiteName string) ([]middleware.Prompt, error) {
	ds.ReadCalls++
	if ds.ReadCalls == 2 {
		return nil, errors.New("mock read prompt suite error")
	}
	return nil, nil
}

func TestPromptList_InvalidOrderFilter(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts?order_filter=not-an-int", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid order filter") {
		t.Fatalf("expected invalid order filter message, got %q", rr.Body.String())
	}
}

func TestPromptList_ListPromptSuitesError(t *testing.T) {
	handler := NewHandlerWithDeps(&promptListListSuitesErrorDataStore{
		Err: errors.New("mock list suites error"),
	}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error listing prompt suites") {
		t.Fatalf("expected list suites error message, got %q", rr.Body.String())
	}
}

func TestPromptList_ReadPromptSuiteError(t *testing.T) {
	handler := NewHandlerWithDeps(&promptListReadSuiteErrorDataStore{
		Err: errors.New("mock read suite error"),
	}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error reading prompt suite") {
		t.Fatalf("expected read suite error message, got %q", rr.Body.String())
	}
}

func TestPromptList_RenderError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{RenderError: errors.New("mock render error")})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error rendering template") {
		t.Fatalf("expected render error message, got %q", rr.Body.String())
	}
}

func TestPromptList_DefaultSuiteFallback(t *testing.T) {
	ds := &promptListEmptySuiteNameDataStore{}
	handler := NewHandlerWithDeps(ds, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if ds.GetCalls != 1 {
		t.Fatalf("expected GetCurrentSuiteName to be called once, got %d", ds.GetCalls)
	}
	if ds.ReadCalls != 2 {
		t.Fatalf("expected ReadPromptSuite to be called twice, got %d", ds.ReadCalls)
	}
}

func TestPromptList_WithPrompts(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Prompt 1", Solution: "Solution 1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestPromptList_ReadDefaultPromptSuiteError(t *testing.T) {
	handler := NewHandlerWithDeps(&promptListSecondReadErrorDataStore{}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error reading default prompt suite") {
		t.Fatalf("expected default suite read error message, got %q", rr.Body.String())
	}
}

func TestPromptList_OrderFilterParseSuccess(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest("GET", "/prompts?order_filter=2", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestPromptListHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add test profile and prompt
	err := middleware.WriteProfiles([]middleware.Profile{
		{Name: "TestProfile", Description: "Test"},
	})
	if err != nil {
		t.Fatalf("failed to write profile: %v", err)
	}

	err = middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt", Profile: "TestProfile"},
	})
	if err != nil {
		t.Fatalf("failed to write prompt: %v", err)
	}

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Test Prompt") {
		t.Error("expected prompt text in response body")
	}
}

func TestBulkDeletePromptsPageHandler_GET(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add a prompt first
	err := middleware.WritePromptSuite("default", []middleware.Prompt{
		{Text: "Test Prompt"},
	})
	if err != nil {
		t.Fatalf("failed to write prompt: %v", err)
	}

	// Request with indices parameter
	req := httptest.NewRequest("GET", "/bulk_delete_prompts?indices=[0]", nil)
	rr := httptest.NewRecorder()
	BulkDeletePromptsPageHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestPromptListHandler_WithSearch(t *testing.T) {
	restoreDir := changeToProjectRootPrompts(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts with different text
	prompts := []middleware.Prompt{
		{Text: "Find this prompt"},
		{Text: "Another prompt"},
	}
	_ = middleware.WritePrompts(prompts)

	req := httptest.NewRequest("GET", "/prompts?search_query=Find", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestPromptListHandler_WithOrderFilter(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	if err := middleware.WritePrompts([]middleware.Prompt{
		{Text: "First Prompt"},
		{Text: "Second Prompt"},
	}); err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}

	req := httptest.NewRequest("GET", "/prompts?order_filter=1", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
}

func TestPromptListHandler_WithProfileFilter(t *testing.T) {
	restoreDir := changeToProjectRoot(t)
	defer restoreDir()

	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts with profiles
	_ = middleware.WriteProfiles([]middleware.Profile{{Name: "TestProfile"}})
	if err := middleware.WritePrompts([]middleware.Prompt{
		{Text: "Filtered Prompt", Profile: "TestProfile"},
		{Text: "Other Prompt", Profile: "Other"},
	}); err != nil {
		t.Fatalf("failed to write prompts: %v", err)
	}

	req := httptest.NewRequest("GET", "/prompts?profile_filter=TestProfile", nil)
	rr := httptest.NewRecorder()
	PromptListHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Filtered Prompt") {
		t.Error("expected 'Filtered Prompt' in response body")
	}
}

func TestPromptListHandler_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts:      []middleware.Prompt{{Text: "Test"}},
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/prompts", nil)
	rr := httptest.NewRecorder()
	handler.PromptList(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestBulkDeletePromptsHandler_WritePromptsError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
			{Text: "Prompt 3"},
		},
		CurrentSuite: "test-suite",
		WritePromptsFunc: func(prompts []middleware.Prompt) error {
			return errors.New("mock write error")
		},
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{},
	}

	// BulkDeletePrompts expects JSON input
	jsonBody := `{"indices": [0, 1]}`

	req := httptest.NewRequest("POST", "/bulk_delete_prompts", strings.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.BulkDeletePrompts(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}
