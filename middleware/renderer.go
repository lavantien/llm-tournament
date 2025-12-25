package middleware

import (
	"html/template"
	"net/http"
	"reflect"
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
	// Wrap data with common template variables if it's a map
	wrappedData := wrapTemplateData(data)
	t, err := template.New(name).Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		return err
	}
	return t.Execute(w, wrappedData)
}

// RenderTemplateSimple renders a single template file without funcMap
func (r *FileRenderer) RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	return r.Render(w, tmpl, nil, data, "templates/"+tmpl)
}

// DefaultRenderer is the default TemplateRenderer instance
var DefaultRenderer TemplateRenderer = &FileRenderer{}

// CommonTemplateData contains data available to all templates
type CommonTemplateData struct {
	Suites       []string
	CurrentSuite string
	CurrentPath  string
}

// wrapTemplateData wraps the provided data with common template variables
func wrapTemplateData(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Get suite information
	suites, _ := ListPromptSuites()
	currentSuite := GetCurrentSuiteName()

	// Add common data
	result["Suites"] = suites
	result["CurrentSuite"] = currentSuite

	// Add the original data
	v := reflect.ValueOf(data)
	if data != nil && v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			strKey := key.String()
			result[strKey] = v.MapIndex(key).Interface()
		}
		// Extract CurrentPath if present in the data
		if path, ok := result["CurrentPath"].(string); ok {
			result["CurrentPath"] = path
		}
	} else if data != nil && v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if field.IsExported() {
				result[field.Name] = v.Field(i).Interface()
			}
		}
		// Extract CurrentPath if present in the data
		if _, found := t.FieldByName("CurrentPath"); found {
			result["CurrentPath"] = v.FieldByName("CurrentPath").Interface()
		}
	}

	return result
}
