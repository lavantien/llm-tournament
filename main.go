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
		pageName = "Import Error"
		middleware.ImportErrorHandler(w, r)
	case "/prompts":
		pageName = "Prompts"
		handlers.PromptListHandler(w, r)
	case "/add_model":
		pageName = "Add Model"
		handlers.AddModelHandler(w, r)
	case "/edit_model":
		pageName = "Edit Model"
		handlers.EditModelHandler(w, r)
	case "/delete_model":
		pageName = "Delete Model"
		handlers.DeleteModelHandler(w, r)
	case "/add_prompt":
		pageName = "Add Prompt"
		handlers.AddPromptHandler(w, r)
	case "/edit_prompt":
		pageName = "Edit Prompt"
		handlers.EditPromptHandler(w, r)
	case "/delete_prompt":
		pageName = "Delete Prompt"
		handlers.DeletePromptHandler(w, r)
	case "/move_prompt":
		pageName = "Move Prompt"
		handlers.MovePromptHandler(w, r)
	case "/import_results":
		pageName = "Import Results"
		handlers.ImportResultsHandler(w, r)
	case "/export_prompts":
		pageName = "Export Prompts"
		handlers.ExportPromptsHandler(w, r)
	case "/import_prompts":
		pageName = "Import Prompts"
		handlers.ImportPromptsHandler(w, r)
	case "/update_prompts_order":
		pageName = "Update Prompts Order"
		handlers.UpdatePromptsOrderHandler(w, r)
	case "/reset_prompts":
		pageName = "Reset Prompts"
		handlers.ResetPromptsHandler(w, r)
	case "/bulk_delete_prompts":
		pageName = "Bulk Delete Prompts"
		if r.Method == "GET" {
			handlers.BulkDeletePromptsPageHandler(w, r)
		} else if r.Method == "POST" {
			handlers.BulkDeletePromptsHandler(w, r)
		}
	case "/prompts/suites/new":
		pageName = "New Prompt Suite"
		handlers.NewPromptSuiteHandler(w, r)
	case "/prompts/suites/edit":
		pageName = "Edit Prompt Suite"
		handlers.EditPromptSuiteHandler(w, r)
	case "/prompts/suites/delete":
		pageName = "Delete Prompt Suite"
		handlers.DeletePromptSuiteHandler(w, r)
	case "/prompts/suites/select":
		pageName = "Select Prompt Suite"
		handlers.SelectPromptSuiteHandler(w, r)
	case "/results":
		pageName = "Results"
		handlers.ResultsHandler(w, r)
	case "/update_result":
		pageName = "Update Result"
		handlers.UpdateResultHandler(w, r)
	case "/reset_results":
		pageName = "Reset Results"
		handlers.ResetResultsHandler(w, r)
	case "/confirm_refresh_results":
		pageName = "Confirm Refresh Results"
		handlers.ConfirmRefreshResultsHandler(w, r)
	case "/refresh_results":
		pageName = "Refresh Results"
		handlers.RefreshResultsHandler(w, r)
	case "/export_results":
		pageName = "Export Results"
		handlers.ExportResultsHandler(w, r)
	case "/profiles":
		pageName = "Profiles"
		handlers.ProfilesHandler(w, r)
	case "/add_profile":
		pageName = "Add Profile"
		handlers.AddProfileHandler(w, r)
	case "/edit_profile":
		pageName = "Edit Profile"
		handlers.EditProfileHandler(w, r)
	case "/delete_profile":
		pageName = "Delete Profile"
		handlers.DeleteProfileHandler(w, r)
	case "/reset_profiles":
		pageName = "Reset Profiles"
		handlers.ResetProfilesHandler(w, r)
	default:
		log.Printf("Redirecting to /prompts from %s", r.URL.Path)
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}
