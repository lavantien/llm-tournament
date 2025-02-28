package handlers

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"llm-tournament/middleware"
)

// Handle delete prompt suite page
func DeletePromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling delete prompt suite page")
	if r.Method == "GET" {
		suiteName := r.URL.Query().Get("suite_name")
		t, err := template.ParseFiles("templates/delete_prompt_suite.html", "templates/nav.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, map[string]string{"SuiteName": suiteName})
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
		log.Println("Delete prompt suite page rendered successfully")
	} else if r.Method == "POST" {
		suiteName := r.FormValue("suite_name")
		if suiteName == "" {
			http.Error(w, "Suite name is required", http.StatusBadRequest)
			return
		}

		currentSuite := middleware.GetCurrentSuiteName()
		if suiteName == currentSuite {
			err := os.WriteFile("data/current_suite.txt", []byte("default"), 0644)
			if err != nil {
				log.Printf("Error updating current suite: %v", err)
				http.Error(w, "Error updating current suite", http.StatusInternalServerError)
				return
			}
		}

		err := middleware.DeletePromptSuite(suiteName)
		if err != nil {
			log.Printf("Error deleting prompt suite: %v", err)
			http.Error(w, "Error deleting prompt suite", http.StatusInternalServerError)
			return
		}

		log.Printf("Prompt suite '%s' deleted successfully", suiteName)
		middleware.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
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

	err = os.WriteFile("data/current_suite.txt", []byte(suiteName), 0644)
	if err != nil {
		log.Printf("Error writing current suite: %v", err)
		http.Error(w, "Error writing current suite", http.StatusInternalServerError)
		return
	}

	log.Printf("Prompt suite '%s' selected successfully", suiteName)
	middleware.BroadcastResults()
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// Handle new prompt suite
func NewPromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling new prompt suite")
	if r.Method == "GET" {
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
	} else if r.Method == "POST" {
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
		err = middleware.WritePromptSuite(suiteName, []middleware.Prompt{})
		if err != nil {
			log.Printf("Error creating prompt suite: %v", err)
			http.Error(w, "Error creating prompt suite", http.StatusInternalServerError)
			return
		}
		log.Printf("Prompt suite '%s' created successfully", suiteName)
		middleware.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle edit prompt suite
func EditPromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling edit prompt suite")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/edit_prompt_suite.html", "templates/nav.html")
		if err != nil {
			log.Printf("Error parsing template: %v", err)
			http.Error(w, "Error parsing template", http.StatusInternalServerError)
			return
		}
		suiteName := r.URL.Query().Get("suite_name")
		err = t.Execute(w, map[string]string{"SuiteName": suiteName})
		if err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Error executing template", http.StatusInternalServerError)
			return
		}
		log.Println("Edit prompt suite page rendered successfully")
	} else if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		oldSuiteName := r.Form.Get("suite_name")
		newSuiteName := r.Form.Get("new_suite_name")

		if oldSuiteName == "" || newSuiteName == "" {
			log.Println("Both old and new names required")
			http.Error(w, "Both original and new suite names are required", http.StatusBadRequest)
			return
		}

		if err := middleware.RenameSuiteFiles(oldSuiteName, newSuiteName); err != nil {
			log.Printf("Error renaming suite: %v", err)
			http.Error(w, fmt.Sprintf("Error renaming suite: %v", err), http.StatusBadRequest)
			return
		}
		log.Printf("Prompt suite '%s' edited successfully to '%s'", suiteName, newSuiteName)
		middleware.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}
