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

var routes = map[string]http.HandlerFunc{
	"/import_error":            middleware.ImportErrorHandler,
	"/prompts":                 handlers.PromptListHandler,
	"/add_model":               handlers.AddModelHandler,
	"/edit_model":              handlers.EditModelHandler,
	"/delete_model":            handlers.DeleteModelHandler,
	"/add_prompt":              handlers.AddPromptHandler,
	"/edit_prompt":             handlers.EditPromptHandler,
	"/delete_prompt":           handlers.DeletePromptHandler,
	"/move_prompt":             handlers.MovePromptHandler,
	"/import_results":          handlers.ImportResultsHandler,
	"/export_prompts":          handlers.ExportPromptsHandler,
	"/import_prompts":          handlers.ImportPromptsHandler,
	"/update_prompts_order":    handlers.UpdatePromptsOrderHandler,
	"/reset_prompts":           handlers.ResetPromptsHandler,
	"/bulk_delete_prompts":     handlers.BulkDeletePromptsHandler,
	"/prompts/suites/new":      handlers.NewPromptSuiteHandler,
	"/prompts/suites/edit":     handlers.EditPromptSuiteHandler,
	"/prompts/suites/delete":   handlers.DeletePromptSuiteHandler,
	"/prompts/suites/select":   handlers.SelectPromptSuiteHandler,
	"/results":                 handlers.ResultsHandler,
	"/update_result":           handlers.UpdateResultHandler,
	"/reset_results":           handlers.ResetResultsHandler,
	"/confirm_refresh_results": handlers.ConfirmRefreshResultsHandler,
	"/refresh_results":         handlers.RefreshResultsHandler,
	"/export_results":          handlers.ExportResultsHandler,
	"/profiles":                handlers.ProfilesHandler,
	"/add_profile":             handlers.AddProfileHandler,
	"/edit_profile":            handlers.EditProfileHandler,
	"/delete_profile":          handlers.DeleteProfileHandler,
	"/reset_profiles":          handlers.ResetProfilesHandler,
}

func router(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)
	
	if handler, exists := routes[r.URL.Path]; exists {
		handler(w, r)
	} else {
		log.Printf("Redirecting to /prompts from %s", r.URL.Path)
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}
