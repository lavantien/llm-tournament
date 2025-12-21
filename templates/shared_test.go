package templates

import (
	"html/template"
	"testing"
)

func TestFuncMap_Inc(t *testing.T) {
	inc := FuncMap["inc"].(func(int) int)
	tests := []struct {
		input    int
		expected int
	}{
		{0, 1},
		{99, 100},
		{-1, 0},
	}
	for _, tt := range tests {
		if got := inc(tt.input); got != tt.expected {
			t.Errorf("inc(%d) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestFuncMap_Add(t *testing.T) {
	add := FuncMap["add"].(func(int, int) int)
	tests := []struct {
		a, b     int
		expected int
	}{
		{2, 3, 5},
		{0, 0, 0},
		{-1, 1, 0},
		{100, -50, 50},
	}
	for _, tt := range tests {
		if got := add(tt.a, tt.b); got != tt.expected {
			t.Errorf("add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestFuncMap_Sub(t *testing.T) {
	sub := FuncMap["sub"].(func(int, int) int)
	tests := []struct {
		a, b     int
		expected int
	}{
		{5, 3, 2},
		{0, 0, 0},
		{1, -1, 2},
		{50, 100, -50},
	}
	for _, tt := range tests {
		if got := sub(tt.a, tt.b); got != tt.expected {
			t.Errorf("sub(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestFuncMap_Atoi(t *testing.T) {
	atoi := FuncMap["atoi"].(func(string) int)
	tests := []struct {
		input    string
		expected int
	}{
		{"42", 42},
		{"0", 0},
		{"-10", -10},
		{"invalid", 0},
		{"", 0},
		{"12abc", 0},
	}
	for _, tt := range tests {
		if got := atoi(tt.input); got != tt.expected {
			t.Errorf("atoi(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}

func TestFuncMap_Markdown(t *testing.T) {
	markdown := FuncMap["markdown"].(func(string) template.HTML)
	tests := []struct {
		input string
		check func(template.HTML) bool
	}{
		{"**bold**", func(h template.HTML) bool { return len(h) > 0 }},
		{"# Heading", func(h template.HTML) bool { return len(h) > 0 }},
		{"plain text", func(h template.HTML) bool { return len(h) > 0 }},
		{"", func(h template.HTML) bool { return len(h) == 0 }},
		{"<script>alert('xss')</script>", func(h template.HTML) bool {
			// Should sanitize out script tags.
			return string(h) != "<script>alert('xss')</script>"
		}},
	}
	for _, tt := range tests {
		result := markdown(tt.input)
		if !tt.check(result) {
			t.Errorf("markdown(%q) = %q, failed check", tt.input, result)
		}
	}
}

func TestFuncMap_ToLower(t *testing.T) {
	tolower := FuncMap["tolower"].(func(string) string)
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"Hello World", "hello world"},
		{"already lower", "already lower"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := tolower(tt.input); got != tt.expected {
			t.Errorf("tolower(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestFuncMap_Contains(t *testing.T) {
	contains := FuncMap["contains"].(func(string, string) bool)
	tests := []struct {
		s, substr string
		expected  bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"", "", true},
		{"hello", "", true},
		{"", "x", false},
	}
	for _, tt := range tests {
		if got := contains(tt.s, tt.substr); got != tt.expected {
			t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.expected)
		}
	}
}

func TestFuncMap_JSON(t *testing.T) {
	jsonFn := FuncMap["json"].(func(interface{}) (string, error))

	t.Run("simple map", func(t *testing.T) {
		result, err := jsonFn(map[string]int{"a": 1})
		if err != nil {
			t.Errorf("json failed: %v", err)
		}
		if result != `{"a":1}` {
			t.Errorf("expected {\"a\":1}, got %q", result)
		}
	})

	t.Run("array", func(t *testing.T) {
		result, err := jsonFn([]int{1, 2, 3})
		if err != nil {
			t.Errorf("json failed: %v", err)
		}
		if result != `[1,2,3]` {
			t.Errorf("expected [1,2,3], got %q", result)
		}
	})

	t.Run("nil", func(t *testing.T) {
		result, err := jsonFn(nil)
		if err != nil {
			t.Errorf("json failed: %v", err)
		}
		if result != "null" {
			t.Errorf("expected null, got %q", result)
		}
	})

	t.Run("unmarshalable value", func(t *testing.T) {
		// Channels cannot be marshaled to JSON.
		ch := make(chan int)
		_, err := jsonFn(ch)
		if err == nil {
			t.Error("expected error for channel, got nil")
		}
	})
}

func TestScoreOptions(t *testing.T) {
	expectedScores := map[string]int{
		"0/5 (0)":   0,
		"1/5 (20)":  20,
		"2/5 (40)":  40,
		"3/5 (60)":  60,
		"4/5 (80)":  80,
		"5/5 (100)": 100,
	}

	if len(ScoreOptions) != len(expectedScores) {
		t.Errorf("expected %d score options, got %d", len(expectedScores), len(ScoreOptions))
	}

	for key, expected := range expectedScores {
		if got, ok := ScoreOptions[key]; !ok {
			t.Errorf("missing score option %q", key)
		} else if got != expected {
			t.Errorf("ScoreOptions[%q] = %d, want %d", key, got, expected)
		}
	}
}

func TestPageNameConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"PageNameResults", PageNameResults, "Results"},
		{"PageNamePrompts", PageNamePrompts, "Prompts"},
		{"PageNameProfiles", PageNameProfiles, "Profiles"},
		{"PageNameEvaluate", PageNameEvaluate, "Evaluate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestFuncMapCompleteness(t *testing.T) {
	expectedFuncs := []string{
		"inc", "add", "sub",
		"eqs", "atoi", "markdown", "tolower", "contains", "json",
	}

	for _, name := range expectedFuncs {
		if _, ok := FuncMap[name]; !ok {
			t.Errorf("FuncMap missing expected function %q", name)
		}
	}
}
