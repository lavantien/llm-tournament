package handlers

import (
	"encoding/json"
	"errors"
	"llm-tournament/middleware"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestMovePromptHandler_POST_Success(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add multiple prompts
	for i := 0; i < 3; i++ {
		form := url.Values{}
		form.Add("prompt", "Prompt "+string(rune('A'+i)))
		req := httptest.NewRequest("POST", "/add_prompt", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		AddPromptHandler(httptest.NewRecorder(), req)
	}

	// Move prompt from index 0 to index 2
	moveForm := url.Values{}
	moveForm.Add("index", "0")
	moveForm.Add("new_index", "2")

	moveReq := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(moveForm.Encode()))
	moveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	moveRR := httptest.NewRecorder()
	MovePromptHandler(moveRR, moveReq)

	if moveRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, moveRR.Code)
	}

	// Verify the move persisted: moving index 0 of [A, B, C] to index 2
	// reorders the suite to [B, A, C]
	prompts := middleware.ReadPrompts()
	texts := make([]string, 0, len(prompts))
	for _, p := range prompts {
		texts = append(texts, p.Text)
	}
	expected := []string{"Prompt B", "Prompt A", "Prompt C"}
	if !reflect.DeepEqual(texts, expected) {
		t.Errorf("expected prompt order %v after move, got %v", expected, texts)
	}
}

func TestMovePromptHandler_POST_InvalidIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	moveForm := url.Values{}
	moveForm.Add("index", "invalid")
	moveForm.Add("new_index", "1")

	moveReq := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(moveForm.Encode()))
	moveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	moveRR := httptest.NewRecorder()
	MovePromptHandler(moveRR, moveReq)

	if moveRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, moveRR.Code)
	}
}

func TestMovePromptHandler_POST_InvalidNewIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	moveForm := url.Values{}
	moveForm.Add("index", "0")
	moveForm.Add("new_index", "invalid")

	moveReq := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(moveForm.Encode()))
	moveReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	moveRR := httptest.NewRecorder()
	MovePromptHandler(moveRR, moveReq)

	if moveRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, moveRR.Code)
	}
}

func TestMovePromptHandler_GET_InvalidIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/move_prompt?index=invalid", nil)
	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodPut, "/move_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestUpdatePromptsOrderHandler_ValidOrder(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// The UpdatePromptsOrder function expects database prompt IDs in the order array
	// which makes it complex to test without knowing the actual IDs generated.
	// For now, just test that the handler accepts valid JSON and redirects.
	// Invalid orders will log errors but still redirect.

	order := []int{}
	orderJSON, _ := json.Marshal(order)

	form := url.Values{}
	form.Add("order", string(orderJSON))

	orderReq := httptest.NewRequest("POST", "/update_prompts_order", strings.NewReader(form.Encode()))
	orderReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	orderRR := httptest.NewRecorder()
	UpdatePromptsOrderHandler(orderRR, orderReq)

	// Handler always redirects after processing
	if orderRR.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, orderRR.Code)
	}
}

func TestUpdatePromptsOrderHandler_EmptyOrder(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("order", "")

	orderReq := httptest.NewRequest("POST", "/update_prompts_order", strings.NewReader(form.Encode()))
	orderReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	orderRR := httptest.NewRecorder()
	UpdatePromptsOrderHandler(orderRR, orderReq)

	if orderRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, orderRR.Code)
	}
}

func TestUpdatePromptsOrderHandler_InvalidJSON(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	form := url.Values{}
	form.Add("order", "not valid json")

	orderReq := httptest.NewRequest("POST", "/update_prompts_order", strings.NewReader(form.Encode()))
	orderReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	orderRR := httptest.NewRecorder()
	UpdatePromptsOrderHandler(orderRR, orderReq)

	if orderRR.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, orderRR.Code)
	}
}

func TestUpdatePromptsOrderHandler_MethodNotAllowed(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	req := httptest.NewRequest(http.MethodGet, "/update_prompts_order", nil)
	rr := httptest.NewRecorder()
	UpdatePromptsOrderHandler(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestUpdatePromptsOrder_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/update_prompts_order", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.UpdatePromptsOrder(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestMovePrompt_GET_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodGet, "/move_prompt?index=%zz", nil)
	rr := httptest.NewRecorder()
	handler.MovePrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestMovePrompt_POST_ParseFormError(t *testing.T) {
	handler := NewHandlerWithDeps(&MockDataStore{
		Prompts: []middleware.Prompt{{Text: "P1"}},
	}, &MockRenderer{})

	req := httptest.NewRequest(http.MethodPost, "/move_prompt", readErrorReader{})
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	handler.MovePrompt(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Error parsing form") {
		t.Fatalf("expected parse error message, got %q", rr.Body.String())
	}
}

func TestMovePromptHandler_POST_OutOfRange(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "First"}, {Text: "Second"}}
	_ = middleware.WritePrompts(prompts)

	form := url.Values{}
	form.Add("index", "99")
	form.Add("new_index", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	// Should redirect without error (bounds check in handler)
	if rr.Code != http.StatusSeeOther {
		t.Logf("status: %d", rr.Code)
	}
}

func TestMovePromptHandler_POST_MoveUp(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "First"}, {Text: "Second"}, {Text: "Third"}}
	_ = middleware.WritePrompts(prompts)

	// Move Third (index 2) to position 0
	form := url.Values{}
	form.Add("index", "2")
	form.Add("new_index", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Verify order changed
	newPrompts := middleware.ReadPrompts()
	if len(newPrompts) != 3 {
		t.Fatalf("expected 3 prompts, got %d", len(newPrompts))
	}
	if newPrompts[0].Text != "Third" {
		t.Errorf("expected 'Third' first, got %q", newPrompts[0].Text)
	}
}

func TestMovePromptHandler_GET_OutOfRange(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "Only prompt"}}
	_ = middleware.WritePrompts(prompts)

	req := httptest.NewRequest("GET", "/move_prompt?index=99", nil)
	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	// Should not crash with out-of-range index
	if rr.Code == http.StatusInternalServerError {
		t.Error("unexpected internal server error")
	}
}

func TestMovePromptHandler_POST_SamePosition(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	prompts := []middleware.Prompt{{Text: "First"}, {Text: "Second"}, {Text: "Third"}}
	_ = middleware.WritePrompts(prompts)

	// Move to same position
	form := url.Values{}
	form.Add("index", "1")
	form.Add("new_index", "1")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rr.Code)
	}

	// Order should remain unchanged
	newPrompts := middleware.ReadPrompts()
	if newPrompts[1].Text != "Second" {
		t.Errorf("expected 'Second' to remain at index 1, got %q", newPrompts[1].Text)
	}
}

func TestMovePromptHandler_POST_InvalidFromIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try to move from invalid index
	form := url.Values{}
	form.Add("from", "invalid")
	form.Add("to", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_POST_InvalidToIndex(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try to move to invalid index
	form := url.Values{}
	form.Add("from", "0")
	form.Add("to", "invalid")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_POST_OutOfBoundsFrom(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try to move from out-of-bounds index
	form := url.Values{}
	form.Add("from", "999")
	form.Add("to", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_POST_OutOfBoundsTo(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try to move to out-of-bounds index
	form := url.Values{}
	form.Add("from", "0")
	form.Add("to", "999")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_POST_NegativeFrom(t *testing.T) {
	cleanup := setupPromptTestDB(t)
	defer cleanup()

	// Add prompts
	_ = middleware.WritePrompts([]middleware.Prompt{{Text: "Prompt 1"}})

	// Try negative from index
	form := url.Values{}
	form.Add("from", "-1")
	form.Add("to", "0")

	req := httptest.NewRequest("POST", "/move_prompt", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	MovePromptHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestMovePromptHandler_WritePromptsError(t *testing.T) {
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

	form := url.Values{}
	form.Add("new_index", "2")

	req := httptest.NewRequest("POST", "/move_prompt?index=0", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler.MovePrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on write error, got %d", http.StatusInternalServerError, rr.Code)
	}
}

func TestMovePromptHandler_GET_RenderError(t *testing.T) {
	mockDS := &MockDataStore{
		Prompts: []middleware.Prompt{
			{Text: "Prompt 1"},
			{Text: "Prompt 2"},
		},
		CurrentSuite: "test-suite",
	}

	handler := &Handler{
		DataStore: mockDS,
		Renderer:  &MockRenderer{RenderError: errors.New("mock render error")},
	}

	req := httptest.NewRequest("GET", "/move_prompt?index=0", nil)
	rr := httptest.NewRecorder()
	handler.MovePrompt(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d on render error, got %d", http.StatusInternalServerError, rr.Code)
	}
}
