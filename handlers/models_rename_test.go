package handlers

import (
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestEditModelHandler_POST_MissingModelReturns404(t *testing.T) {
	cleanup := setupModelsTestDB(t)
	defer cleanup()

	editForm := url.Values{}
	editForm.Add("new_model_name", "NewName")

	editReq := httptest.NewRequest("POST", "/edit_model?model=MissingModel", strings.NewReader(editForm.Encode()))
	editReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	editRR := httptest.NewRecorder()
	EditModelHandler(editRR, editReq)

	if editRR.Code != http.StatusNotFound {
		t.Errorf("expected status %d for missing model, got %d", http.StatusNotFound, editRR.Code)
	}

	results := middleware.ReadResults()
	if _, exists := results["NewName"]; exists {
		t.Error("NewName should not be created when the source model is missing")
	}
	if _, exists := results["MissingModel"]; exists {
		t.Error("MissingModel should not appear after a rejected rename")
	}
}
