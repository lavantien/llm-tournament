package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"llm-tournament/middleware"
	"llm-tournament/testutil"

	_ "github.com/mattn/go-sqlite3"
)

// setupModelsTestDB creates a test database for model handler tests
func setupModelsTestDB(t *testing.T) func() {
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

func TestAddModelHandler_Success(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("model", "TestModel")

	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddModelHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify model was added
	results := middleware.ReadResults()
	if _, exists := results["TestModel"]; !exists {
		t.Error("TestModel should exist in results")
	}
}

func TestAddModelHandler_EmptyName(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("model", "")

	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddModelHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestAddModelHandler_DuplicateModel(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Add first model
	form := url.Values{}
	form.Add("model", "DuplicateModel")

	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddModelHandler(rr, req)

	// Try adding the same model again (should still succeed - it's idempotent)
	req2 := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr2 := httptest.NewRecorder()
	AddModelHandler(rr2, req2)

	if rr2.Code != http.StatusSeeOther {
		t.Errorf("expected status %d for duplicate, got %d", http.StatusSeeOther, rr2.Code)
	}
}

func TestDeleteModelHandler_POST_Success(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// First add a model
	form := url.Values{}
	form.Add("model", "ModelToDelete")

	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	AddModelHandler(rr, req)

	// Now delete it
	deleteForm := url.Values{}
	deleteForm.Add("model", "ModelToDelete")

	deleteReq := httptest.NewRequest("POST", "/delete_model", strings.NewReader(deleteForm.Encode()))
	deleteReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	deleteRR := httptest.NewRecorder()
	DeleteModelHandler(deleteRR, deleteReq)

	if deleteRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, deleteRR.Code)
	}

	// Verify model was deleted
	results := middleware.ReadResults()
	if _, exists := results["ModelToDelete"]; exists {
		t.Error("ModelToDelete should not exist after deletion")
	}
}

func TestDeleteModelHandler_POST_EmptyName(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("model", "")

	req := httptest.NewRequest("POST", "/delete_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	DeleteModelHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestDeleteModelHandler_GET_EmptyName(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/delete_model?model=", nil)
	rr := httptest.NewRecorder()
	DeleteModelHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestEditModelHandler_POST_Success(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// First add a model
	addForm := url.Values{}
	addForm.Add("model", "OldModelName")

	addReq := httptest.NewRequest("POST", "/add_model", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addRR := httptest.NewRecorder()
	AddModelHandler(addRR, addReq)

	// Now edit it
	editForm := url.Values{}
	editForm.Add("new_model_name", "NewModelName")

	editReq := httptest.NewRequest("POST", "/edit_model?model=OldModelName", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditModelHandler(editRR, editReq)

	if editRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, editRR.Code)
	}

	// Verify model was renamed
	results := middleware.ReadResults()
	if _, exists := results["OldModelName"]; exists {
		t.Error("OldModelName should not exist after rename")
	}
	if _, exists := results["NewModelName"]; !exists {
		t.Error("NewModelName should exist after rename")
	}
}

func TestEditModelHandler_POST_EmptyNewName(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// First add a model
	addForm := url.Values{}
	addForm.Add("model", "ExistingModel")

	addReq := httptest.NewRequest("POST", "/add_model", strings.NewReader(addForm.Encode()))
	addReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addRR := httptest.NewRecorder()
	AddModelHandler(addRR, addReq)

	// Try to edit with empty name
	editForm := url.Values{}
	editForm.Add("new_model_name", "")

	editReq := httptest.NewRequest("POST", "/edit_model?model=ExistingModel", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditModelHandler(editRR, editReq)

	if editRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, editRR.Code)
	}
}

func TestEditModelHandler_POST_DuplicateName(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Add two models
	form1 := url.Values{}
	form1.Add("model", "Model1")

	req1 := httptest.NewRequest("POST", "/add_model", strings.NewReader(form1.Encode()))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req1)

	form2 := url.Values{}
	form2.Add("model", "Model2")

	req2 := httptest.NewRequest("POST", "/add_model", strings.NewReader(form2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	AddModelHandler(httptest.NewRecorder(), req2)

	// Try to rename Model1 to Model2
	editForm := url.Values{}
	editForm.Add("new_model_name", "Model2")

	editReq := httptest.NewRequest("POST", "/edit_model?model=Model1", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditModelHandler(editRR, editReq)

	if editRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for duplicate name, got %d", http.StatusBadRequest, editRR.Code)
	}
}

func TestEditModelHandler_GET_MissingModelParam(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/edit_model", nil)
	rr := httptest.NewRecorder()
	EditModelHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestDeleteModelHandler_POST_NonExistent(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Try to delete a model that doesn't exist
	form := url.Values{}
	form.Add("model", "NonExistentModel")

	req := httptest.NewRequest("POST", "/delete_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	DeleteModelHandler(rr, req)

	// Should redirect even for non-existent model
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestEditModelHandler_POST_NonExistent(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Try to edit a model that doesn't exist
	form := url.Values{}
	form.Add("new_model_name", "NewName")

	req := httptest.NewRequest("POST", "/edit_model?model=NonExistent", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	EditModelHandler(rr, req)

	// Should redirect even for non-existent model
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}
}

func TestAddModelHandler_WithPrompts(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Add some prompts first to test that new model gets correct number of score slots
	middleware.WritePrompts([]middleware.Prompt{
		{Text: "Prompt 1"},
		{Text: "Prompt 2"},
		{Text: "Prompt 3"},
	})

	form := url.Values{}
	form.Add("model", "ModelWithScores")

	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	AddModelHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify model was added with correct number of score slots
	results := middleware.ReadResults()
	if result, exists := results["ModelWithScores"]; !exists {
		t.Error("ModelWithScores should exist in results")
	} else if len(result.Scores) != 3 {
		t.Errorf("expected 3 score slots, got %d", len(result.Scores))
	}
}

func changeToProjectRootModels(t *testing.T) func() {
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

func TestDeleteModelHandler_GET_Success(t *testing.T) {
	restoreDir := changeToProjectRootModels(t)
	defer restoreDir()

	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Add a model first
	suiteName := middleware.GetCurrentSuiteName()
	middleware.WriteResults(suiteName, map[string]middleware.Result{
		"ModelToDelete": {Scores: []int{80}},
	})

	req := httptest.NewRequest("GET", "/delete_model?model=ModelToDelete", nil)
	rr := httptest.NewRecorder()
	DeleteModelHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "ModelToDelete") {
		t.Error("expected model name in response body")
	}
}

func TestAddModel_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{},
	}
	mockDS.WriteResultsFunc = func(suiteName string, results map[string]middleware.Result) error {
		return errors.New("database write error")
	}
	mockRenderer := &MockRenderer{}

	handler := NewHandlerWithDeps(mockDS, mockRenderer)

	form := url.Values{}
	form.Add("model", "TestModel")

	req := httptest.NewRequest("POST", "/add_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.AddModel(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEditModel_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{
			"OldModel": {Scores: []int{80}},
		},
	}
	mockDS.WriteResultsFunc = func(suiteName string, results map[string]middleware.Result) error {
		return errors.New("database write error")
	}
	mockRenderer := &MockRenderer{}

	handler := NewHandlerWithDeps(mockDS, mockRenderer)

	form := url.Values{}
	form.Add("new_model_name", "NewModel")

	req := httptest.NewRequest("POST", "/edit_model?model=OldModel", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.EditModel(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeleteModel_WriteResultsError(t *testing.T) {
	mockDS := &MockDataStore{
		Results: map[string]middleware.Result{
			"ModelToDelete": {Scores: []int{80}},
		},
	}
	mockDS.WriteResultsFunc = func(suiteName string, results map[string]middleware.Result) error {
		return errors.New("database write error")
	}
	mockRenderer := &MockRenderer{}

	handler := NewHandlerWithDeps(mockDS, mockRenderer)

	form := url.Values{}
	form.Add("model", "ModelToDelete")

	req := httptest.NewRequest("POST", "/delete_model", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.DeleteModel(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d for write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestEditModelHandler_GET_Success(t *testing.T) {
	restoreDir := changeToProjectRootModels(t)
	defer restoreDir()

	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Add a model first
	suiteName := middleware.GetCurrentSuiteName()
	middleware.WriteResults(suiteName, map[string]middleware.Result{
		"ModelToEdit": {Scores: []int{80}},
	})

	req := httptest.NewRequest("GET", "/edit_model?model=ModelToEdit", nil)
	rr := httptest.NewRecorder()
	EditModelHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "ModelToEdit") {
		t.Error("expected model name in response body")
	}
}

func TestEditModelHandler_GET_RenderError(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	// Add a model first
	suiteName := middleware.GetCurrentSuiteName()
	middleware.WriteResults(suiteName, map[string]middleware.Result{
		"ModelToEdit": {Scores: []int{80}},
	})

	req := httptest.NewRequest("GET", "/edit_model?model=ModelToEdit", nil)
	rr := httptest.NewRecorder()
	EditModelHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestDeleteModelHandler_GET_RenderError(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	// Save original renderer and restore after test
	original := middleware.DefaultRenderer
	defer func() { middleware.DefaultRenderer = original }()

	// Swap in mock that returns error
	middleware.DefaultRenderer = &testutil.MockRenderer{RenderError: errors.New("mock render error")}

	// Add a model first
	suiteName := middleware.GetCurrentSuiteName()
	middleware.WriteResults(suiteName, map[string]middleware.Result{
		"ModelToDelete": {Scores: []int{80}},
	})

	req := httptest.NewRequest("GET", "/delete_model?model=ModelToDelete", nil)
	rr := httptest.NewRecorder()
	DeleteModelHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}
