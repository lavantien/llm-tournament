package middleware

import (
	"encoding/json"
	"log"
	"net/http"

	"llm-tournament/templates"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	err := DefaultRenderer.Render(w, tmpl, templates.FuncMap, data, "templates/"+tmpl, "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// RenderTemplateSimple renders a single template without nav.html
func RenderTemplateSimple(w http.ResponseWriter, tmpl string, data interface{}) error {
	err := DefaultRenderer.Render(w, tmpl, nil, data, "templates/"+tmpl)
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
	return err
}

func HandleFormError(w http.ResponseWriter, err error) {
	log.Printf("Error parsing form: %v", err)
	http.Error(w, "Error parsing form", http.StatusBadRequest)
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
	}
}
