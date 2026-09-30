package handlers

import (
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestEvaluateResultHandler_GET_InvalidPromptIndexRedirects(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	cases := []struct {
		name   string
		prompt string
	}{
		{"negative index", "-5"},
		{"non numeric index", "abc"},
		{"index beyond suite", "999"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockDS := &MockDataStore{
				Prompts: []middleware.Prompt{{Text: "P1"}, {Text: "P2"}},
				Results: map[string]middleware.Result{
					"TestModel": {Scores: []int{50, 60}},
				},
			}
			renderer := &captureDataRenderer{}
			handler := &Handler{DataStore: mockDS, Renderer: renderer}

			req := httptest.NewRequest("GET", "/evaluate_result?model=TestModel&prompt="+tc.prompt, nil)
			rr := httptest.NewRecorder()
			handler.EvaluateResultHandler(rr, req)

			if rr.Code != http.StatusSeeOther {
				t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
			}
			if loc := rr.Header().Get("Location"); loc != "/results" {
				t.Errorf("expected redirect to /results, got %q", loc)
			}
			if renderer.data != nil {
				t.Error("expected no render for an invalid prompt index")
			}
		})
	}
}
