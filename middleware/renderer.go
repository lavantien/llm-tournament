package middleware

import (
	"html/template"
	"net/http"
)

// TemplateRenderer defines the interface for rendering templates
type TemplateRenderer interface {
	Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error
	RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error
}

// FileRenderer renders templates from files
type FileRenderer struct{}

// Render parses and executes templates from the given files
func (r *FileRenderer) Render(w http.ResponseWriter, name string, funcMap template.FuncMap, data interface{}, files ...string) error {
	t, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		return err
	}
	return t.Execute(w, data)
}

// RenderTemplateSimple renders a single template file without funcMap
func (r *FileRenderer) RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	return r.Render(w, tmpl, nil, data, "templates/"+tmpl)
}

// DefaultRenderer is the default TemplateRenderer instance
var DefaultRenderer TemplateRenderer = &FileRenderer{}
