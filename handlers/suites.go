package handlers

import (
	"fmt"
	"log"
	"net/http"

	"llm-tournament/middleware"
)

// DeletePromptSuiteHandler handles delete prompt suite (backward compatible wrapper)
func DeletePromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.DeletePromptSuite(w, r)
}

// SelectPromptSuiteHandler handles select prompt suite (backward compatible wrapper)
func SelectPromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.SelectPromptSuite(w, r)
}

// NewPromptSuiteHandler handles new prompt suite (backward compatible wrapper)
func NewPromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.NewPromptSuite(w, r)
}

// EditPromptSuiteHandler handles edit prompt suite (backward compatible wrapper)
func EditPromptSuiteHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.EditPromptSuite(w, r)
}

// DeletePromptSuite handles delete prompt suite page
func (h *Handler) DeletePromptSuite(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling delete prompt suite page")
	if r.Method == "GET" {
		suiteName := r.URL.Query().Get("suite_name")
		if err := h.Renderer.Render(w, "delete_prompt_suite.html", nil, map[string]string{"SuiteName": suiteName}, "templates/delete_prompt_suite.html"); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
			return
		}
		log.Println("Delete prompt suite page rendered successfully")
	} else if r.Method == "POST" {
		suiteName := r.FormValue("suite_name")
		if suiteName == "" {
			http.Error(w, "Suite name is required", http.StatusBadRequest)
			return
		}

		currentSuite := h.DataStore.GetCurrentSuiteName()
		if suiteName == currentSuite {
			if err := h.DataStore.SetCurrentSuite("default"); err != nil {
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
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// SelectPromptSuite handles select prompt suite
func (h *Handler) SelectPromptSuite(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling select prompt suite")
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
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

	if err = h.DataStore.SetCurrentSuite(suiteName); err != nil {
		log.Printf("Error setting current suite: %v", err)
		http.Error(w, "Error setting current suite", http.StatusInternalServerError)
		return
	}

	log.Printf("Prompt suite '%s' selected successfully", suiteName)
	h.DataStore.BroadcastResults()
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// NewPromptSuite handles new prompt suite
func (h *Handler) NewPromptSuite(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling new prompt suite")
	if r.Method == "GET" {
		if err := h.Renderer.Render(w, "new_prompt_suite.html", nil, nil, "templates/new_prompt_suite.html"); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
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
		err = h.DataStore.WritePromptSuite(suiteName, []middleware.Prompt{})
		if err != nil {
			log.Printf("Error creating prompt suite: %v", err)
			http.Error(w, "Error creating prompt suite", http.StatusInternalServerError)
			return
		}
		log.Printf("Prompt suite '%s' created successfully", suiteName)
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// EditPromptSuite handles edit prompt suite
func (h *Handler) EditPromptSuite(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling edit prompt suite")
	if r.Method == "GET" {
		suiteName := r.URL.Query().Get("suite_name")
		if err := h.Renderer.Render(w, "edit_prompt_suite.html", nil, map[string]string{"SuiteName": suiteName}, "templates/edit_prompt_suite.html"); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
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
		log.Printf("Prompt suite '%s' edited successfully to '%s'", oldSuiteName, newSuiteName)
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
