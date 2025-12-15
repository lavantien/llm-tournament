package handlers

import (
	"log"
	"net/http"

	"llm-tournament/middleware"
)

// AddModelHandler handles adding a model (backward compatible wrapper)
func AddModelHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.AddModel(w, r)
}

// EditModelHandler handles editing a model (backward compatible wrapper)
func EditModelHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.EditModel(w, r)
}

// DeleteModelHandler handles deleting a model (backward compatible wrapper)
func DeleteModelHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.DeleteModel(w, r)
}

// AddModel handles adding a model
func (h *Handler) AddModel(w http.ResponseWriter, r *http.Request) {
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
	results := h.DataStore.ReadResults()
	if results == nil {
		results = make(map[string]middleware.Result)
	}
	if _, ok := results[modelName]; !ok {
		results[modelName] = middleware.Result{Scores: make([]int, len(h.DataStore.ReadPrompts()))}
	}
	suiteName := h.DataStore.GetCurrentSuiteName()
	err = h.DataStore.WriteResults(suiteName, results)
	if err != nil {
		log.Printf("Error writing results: %v", err)
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}
	log.Println("Model added successfully")
	h.DataStore.BroadcastResults()
	http.Redirect(w, r, "/results", http.StatusSeeOther)
}

// EditModel handles editing a model
func (h *Handler) EditModel(w http.ResponseWriter, r *http.Request) {
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

		results := h.DataStore.ReadResults()
		if _, exists := results[newModelName]; exists {
			http.Error(w, "Model with this name already exists", http.StatusBadRequest)
			return
		}

		results[newModelName] = results[modelName]
		delete(results, modelName)
		suiteName := h.DataStore.GetCurrentSuiteName()
		if err := h.DataStore.WriteResults(suiteName, results); err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}

		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
		return
	}

	// Render the edit model form
	if err := h.Renderer.RenderTemplateSimple(w, "edit_model.html", map[string]string{"Model": modelName}); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
	}
}

// DeleteModel handles deleting a model
func (h *Handler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling delete model")
	if r.Method == "GET" {
		modelName := r.URL.Query().Get("model")
		if modelName == "" {
			http.Error(w, "Model name is required", http.StatusBadRequest)
			return
		}
		if err := h.Renderer.RenderTemplateSimple(w, "delete_model.html", map[string]string{"Model": modelName}); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		modelName := r.FormValue("model")
		if modelName == "" {
			http.Error(w, "Model name is required", http.StatusBadRequest)
			return
		}

		results := h.DataStore.ReadResults()
		delete(results, modelName)
		suiteName := h.DataStore.GetCurrentSuiteName()
		if err := h.DataStore.WriteResults(suiteName, results); err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}

		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}
