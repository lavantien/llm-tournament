package testutil

import (
	"html/template"
	"net/http"
)

// MockRenderer implements TemplateRenderer for testing with error injection
type MockRenderer struct {
	RenderError error
	RenderCalls []MockRenderCall
}

// MockRenderCall records a call to Render
type MockRenderCall struct {
	Name  string
	Data  interface{}
	Files []string
}

// Render records the call and returns any configured error
func (m *MockRenderer) Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error {
	m.RenderCalls = append(m.RenderCalls, MockRenderCall{
		Name:  name,
		Data:  data,
		Files: files,
	})
	if m.RenderError != nil {
		return m.RenderError
	}
	// Write minimal content to satisfy tests expecting output
	_, _ = w.Write([]byte("mock rendered"))
	return nil
}

// RenderTemplateSimple records the call and returns any configured error
func (m *MockRenderer) RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	return m.Render(w, tmpl, nil, data, "templates/"+tmpl)
}
