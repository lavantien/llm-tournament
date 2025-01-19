package handlers

import (
	"html/template"
	"log"
	"net/http"

	"llm-tournament/middleware"
)

// Handle add model
func AddModelHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling add model")
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}
	modelName := r.Form.Get("model")
	if modelName == "" {
		log.Println("Model name cannot be empty")
		http.Error(w, "Model name cannot be empty", http.StatusBadRequest)
		return
	}
	results := middleware.ReadResults()
	if results == nil {
		results = make(map[string]middleware.Result)
	}
	if _, ok := results[modelName]; !ok {
		results[modelName] = middleware.Result{Passes: make([]bool, len(middleware.ReadPrompts()))}
	}
	suiteName := middleware.GetCurrentSuiteName()
	err = middleware.WriteResults(suiteName, results)
	if err != nil {
		log.Printf("Error writing results: %v", err)
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}
	log.Println("Model added successfully")
	middleware.BroadcastResults()
	http.Redirect(w, r, "/results", http.StatusSeeOther)
}

// Handle edit model
func EditModelHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling edit model")
	modelName := r.URL.Query().Get("model")
	if modelName == "" {
		http.Error(w, "Model name is required", http.StatusBadRequest)
		return
	}

	if r.Method == "POST" {
		newModelName := r.FormValue("new_model_name")
		if newModelName == "" {
			http.Error(w, "New model name cannot be empty", http.StatusBadRequest)
			return
		}

		results := middleware.ReadResults()
		if _, exists := results[newModelName]; exists {
			http.Error(w, "Model with this name already exists", http.StatusBadRequest)
			return
		}

		results[newModelName] = results[modelName]
		delete(results, modelName)
		suiteName := middleware.GetCurrentSuiteName()
		middleware.WriteResults(suiteName, results)

		middleware.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
		return
	}

	// Render the edit model form
	tmpl, err := template.ParseFiles("templates/edit_model.html")
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, map[string]string{"Model": modelName})
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// Handle delete model
func DeleteModelHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling delete model")
	if r.Method == "GET" {
		modelName := r.URL.Query().Get("model")
		if modelName == "" {
			http.Error(w, "Model name is required", http.StatusBadRequest)
			return
		}
		tmpl, err := template.ParseFiles("templates/delete_model.html")
		if err != nil {
			http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, map[string]string{"Model": modelName})
		if err != nil {
			http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		modelName := r.FormValue("model")
		if modelName == "" {
			http.Error(w, "Model name is required", http.StatusBadRequest)
			return
		}

		results := middleware.ReadResults()
		delete(results, modelName)
		suiteName := middleware.GetCurrentSuiteName()
		middleware.WriteResults(suiteName, results)

		middleware.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}
