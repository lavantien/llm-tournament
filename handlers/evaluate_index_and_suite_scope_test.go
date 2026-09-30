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

func TestEvaluateResultHandler_GET_NegativePromptIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{{Text: "Test prompt"}},
		Results: map[string]middleware.Result{
			"TestModel": {Scores: []int{50}},
		},
		CurrentSuite: "default",
	}
	renderer := &captureDataRenderer{}
	handler := &Handler{DataStore: mockDS, Renderer: renderer}

	req := httptest.NewRequest("GET", "/evaluate_result?model=TestModel&prompt=-1", nil)
	rr := httptest.NewRecorder()

	handler.EvaluateResultHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d for negative prompt index, got %d", http.StatusOK, rr.Code)
	}
	currentScore := reflect.ValueOf(renderer.data).FieldByName("CurrentScore").Int()
	if currentScore != 0 {
		t.Errorf("expected current score 0 for negative prompt index, got %d", currentScore)
	}
}
