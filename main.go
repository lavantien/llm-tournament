package main

import (
	"log"
	"net/http"

	"llm-tournament/handlers"
	"llm-tournament/middleware"
)

func main() {
	log.Println("Starting the server...")
	http.HandleFunc("/", router)
	http.HandleFunc("/ws", middleware.HandleWebSocket)
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	log.Println("Server is listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func router(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)
	switch r.URL.Path {
	case "/import_error":
		middleware.ImportErrorHandler(w, r)
	case "/prompts":
		handlers.PromptListHandler(w, r)
	case "/add_model":
		handlers.AddModelHandler(w, r)
	case "/edit_model":
		handlers.EditModelHandler(w, r)
	case "/delete_model":
		handlers.DeleteModelHandler(w, r)
	case "/add_prompt":
		handlers.AddPromptHandler(w, r)
	case "/edit_prompt":
		handlers.EditPromptHandler(w, r)
	case "/delete_prompt":
		handlers.DeletePromptHandler(w, r)
	case "/move_prompt":
		handlers.MovePromptHandler(w, r)
	case "/import_results":
		handlers.ImportResultsHandler(w, r)
	case "/export_prompts":
		handlers.ExportPromptsHandler(w, r)
	case "/import_prompts":
		handlers.ImportPromptsHandler(w, r)
	case "/update_prompts_order":
		handlers.UpdatePromptsOrderHandler(w, r)
	case "/reset_prompts":
		handlers.ResetPromptsHandler(w, r)
	case "/results":
		handlers.ResultsHandler(w, r)
	case "/update_result":
		handlers.UpdateResultHandler(w, r)
	case "/reset_results":
		handlers.ResetResultsHandler(w, r)
	case "/confirm_refresh_results":
		handlers.ConfirmRefreshResultsHandler(w, r)
	case "/refresh_results":
		handlers.RefreshResultsHandler(w, r)
	case "/export_results":
		handlers.ExportResultsHandler(w, r)
	default:
		log.Printf("Redirecting to /prompts from %s", r.URL.Path)
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}
