package templates

import "testing"

func TestFuncMap_EqString(t *testing.T) {
	eqs := FuncMap["eqs"].(func(string, string) bool)

	tests := []struct {
		a, b string
		want bool
	}{
		{"Results", "Results", true},
		{"Results", "Prompts", false},
		{"", "", true},
	}

	for _, tt := range tests {
		if got := eqs(tt.a, tt.b); got != tt.want {
			t.Errorf("eqs(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
