package handlers

import (
	"html/template"
	"log"
	"net/http"
)

// Handle new prompt suite page
func NewPromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling new prompt suite page")
	t, err := template.ParseFiles("templates/new_prompt_suite.html", "templates/nav.html")
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
	log.Println("New prompt suite page rendered successfully")
}

// Handle edit prompt suite page
func EditPromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling edit prompt suite page")
	t, err := template.ParseFiles("templates/edit_prompt_suite.html", "templates/nav.html")
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
	log.Println("Edit prompt suite page rendered successfully")
}

// Handle delete prompt suite page
func DeletePromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling delete prompt suite page")
	t, err := template.ParseFiles("templates/delete_prompt_suite.html", "templates/nav.html")
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
	log.Println("Delete prompt suite page rendered successfully")
}

// Handle select prompt suite
func SelectPromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling select prompt suite")
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	suiteName := r.Form.Get("suite_name")
	if suiteName == "" {
		log.Println("Suite name cannot be empty")
		http.Error(w, "Suite name cannot be empty", http.StatusBadRequest)
		return
	}

	prompts, err := middleware.ReadPromptSuite(suiteName)
	if err != nil {
		log.Printf("Error reading prompt suite: %v", err)
		http.Error(w, "Error reading prompt suite", http.StatusInternalServerError)
		return
	}

	err = middleware.WritePrompts(prompts)
	if err != nil {
		log.Printf("Error writing prompts: %v", err)
		http.Error(w, "Error writing prompts", http.StatusInternalServerError)
		return
	}

	log.Printf("Prompt suite '%s' selected successfully", suiteName)
	middleware.BroadcastResults()
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}
