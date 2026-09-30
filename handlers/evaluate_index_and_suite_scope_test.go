package handlers

import (
	"html/template"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// captureDataRenderer records the template data passed to Render
type captureDataRenderer struct {
	data interface{}
}

func (c *captureDataRenderer) Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error {
	c.data = data
	_, _ = w.Write([]byte("mock rendered"))
	return nil
}

func (c *captureDataRenderer) RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	return c.Render(w, tmpl, nil, data, "templates/"+tmpl)
}

func TestEvaluateResultHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Test prompt"}},
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{50}},
		},
	}
	handler := &Handler{DataStore: mockDS, Renderer: &captureDataRenderer{}}

	for _, method := range []string{http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/evaluate_result?model=TestModel&prompt=0", nil)
		rr := httptest.NewRecorder()
		handler.EvaluateResultHandler(rr, req)

		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d for %s, got %d", http.StatusMethodNotAllowed, method, rr.Code)
		}
	}
}

func TestEvaluateResultHandler_GET_ModelLookupIsSuiteScoped(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Create the non-current suite first so its model has the lower suite_id
	// and rowid: without suite scoping the name lookup resolves to it.
	suiteB, err := middleware.GetSuiteID("suite-b")
	if err != nil {
		t.Fatalf("failed to create suite-b: %v", err)
	}
	if err := middleware.SetCurrentSuite("suite-a"); err != nil {
		t.Fatalf("failed to set current suite: %v", err)
	}
	suiteA, err := middleware.GetSuiteID("suite-a")
	if err != nil {
		t.Fatalf("failed to get suite-a id: %v", err)
	}

	db := middleware.GetDB()
	if _, err := db.Exec("INSERT INTO prompts (text, solution, suite_id, display_order) VALUES ('prompt a', 'sol', ?, 1)", suiteA); err != nil {
		t.Fatalf("failed to insert prompt: %v", err)
	}
	resB, err := db.Exec("INSERT INTO models (name, suite_id) VALUES ('Shared', ?)", suiteB)
	if err != nil {
		t.Fatalf("failed to insert suite-b model: %v", err)
	}
	resA, err := db.Exec("INSERT INTO models (name, suite_id) VALUES ('Shared', ?)", suiteA)
	if err != nil {
		t.Fatalf("failed to insert suite-a model: %v", err)
	}
	modelB, _ := resB.LastInsertId()
	modelA, _ := resA.LastInsertId()

	var promptID int
	if err := db.QueryRow("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order LIMIT 1", suiteA).Scan(&promptID); err != nil {
		t.Fatalf("failed to get prompt id: %v", err)
	}
	if _, err := db.Exec("INSERT INTO model_responses (model_id, prompt_id, response_text) VALUES (?, ?, 'response from suite b model')", modelB, promptID); err != nil {
		t.Fatalf("failed to insert suite-b response: %v", err)
	}
	if _, err := db.Exec("INSERT INTO model_responses (model_id, prompt_id, response_text) VALUES (?, ?, 'response from suite a model')", modelA, promptID); err != nil {
		t.Fatalf("failed to insert suite-a response: %v", err)
	}

	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "prompt a"}},
		Results: map[string]middleware.Result{
			"Shared": {Scores: []int{50}},
		},
		CurrentSuite: "suite-a",
	}
	renderer := &captureDataRenderer{}
	handler := &Handler{DataStore: mockDS, Renderer: renderer}

	req := httptest.NewRequest("GET", "/evaluate_result?model=Shared&prompt=0", nil)
	rr := httptest.NewRecorder()

	handler.EvaluateResultHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	data := reflect.ValueOf(renderer.data)
	if got := data.FieldByName("ModelID").Int(); got != modelA {
		t.Errorf("expected model id %d from the current suite, got %d", modelA, got)
	}
	if got := data.FieldByName("ModelResponse").String(); got != "response from suite a model" {
		t.Errorf("expected response from the current suite model, got %q", got)
	}
}
