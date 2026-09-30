package templates

import (
	"bytes"
	"html/template"
	"testing"
)

func TestFuncMap_DoesNotShadowEqForStrings(t *testing.T) {
	tmpl, err := template.New("t").Funcs(template.FuncMap(FuncMap)).Parse(`{{if eq .A ""}}empty{{else}}nonempty{{end}}`)
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]any{"A": ""}); err != nil {
		t.Fatalf("execute template: %v", err)
	}
	if got := buf.String(); got != "empty" {
		t.Fatalf("output = %q, want %q", got, "empty")
	}
}

func TestFuncMap_ScoreColor(t *testing.T) {
	cases := map[int]string{
		0:   "#808080",
		20:  "#ffa500",
		40:  "#ffd700",
		60:  "#00bfff",
		80:  "#a77bff",
		100: "#7cff6b",
	}
	for score, want := range cases {
		fn, ok := FuncMap["scoreColor"].(func(int) string)
		if !ok {
			t.Fatalf("scoreColor has unexpected type %T", FuncMap["scoreColor"])
		}
		if got := fn(score); got != want {
			t.Errorf("scoreColor(%d) = %q, want %q", score, got, want)
		}
	}

	fn := FuncMap["scoreColor"].(func(int) string)
	if got := fn(37); got != "#808080" {
		t.Errorf("scoreColor(37) = %q, want fallback %q", got, "#808080")
	}
}
