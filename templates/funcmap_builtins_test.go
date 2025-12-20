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

