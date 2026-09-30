package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

func ExportResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ExportResults(w, r)
}

// UpdateMockResultsHandler handles updating mock results (backward compatible wrapper)

// ExportResults handles export results
func (h *Handler) ExportResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling export results")
	results := h.DataStore.ReadResults()

	// Convert results to JSON
	jsonData, _ := json.MarshalIndent(results, "", "  ")

	// Set headers for JSON download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment;filename=results.json")

	// Write JSON to response
	_, err := w.Write(jsonData)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("Results exported successfully as JSON")
}
