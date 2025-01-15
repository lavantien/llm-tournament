package middleware

import (
	"html/template"
	"log"
	"net/http"
)

// ImportErrorHandler handles the import error page
func ImportErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling import error page")
	t, err := template.ParseFiles("templates/import_error.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
	log.Println("Import error page rendered successfully")
}
