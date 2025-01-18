package middleware

import (
	"html/template"
	"log"
	"net/http"
)

func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.New(tmpl).Funcs(templates.FuncMap).ParseFiles("templates/"+tmpl, "templates/nav.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	
	err = t.Execute(w, data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}

func HandleFormError(w http.ResponseWriter, err error) {
	log.Printf("Error parsing form: %v", err)
	http.Error(w, "Error parsing form", http.StatusBadRequest)
}
