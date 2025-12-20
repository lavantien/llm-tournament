package main

import (
	"log"
	"net/http"

	"llm-tournament/handlers"
	"llm-tournament/middleware"
)

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
	"/update_mock_results":     handlers.UpdateMockResultsHandler,
	"/evaluate":                handlers.EvaluateResult,
	"/profiles":                handlers.ProfilesHandler,
	"/add_profile":             handlers.AddProfileHandler,
	"/edit_profile":            handlers.EditProfileHandler,
	"/delete_profile":          handlers.DeleteProfileHandler,
	"/reset_profiles":          handlers.ResetProfilesHandler,
	"/stats":                   handlers.StatsHandler,
	// New evaluation routes
	"/settings":            handlers.SettingsHandler,
	"/settings/update":     handlers.UpdateSettingsHandler,
	"/settings/test_key":   handlers.TestAPIKeyHandler,
	"/evaluate/all":        handlers.EvaluateAllHandler,
	"/evaluate/model":      handlers.EvaluateModelHandler,
	"/evaluate/prompt":     handlers.EvaluatePromptHandler,
	"/evaluation/progress": handlers.EvaluationProgressHandler,
	"/evaluation/cancel":   handlers.CancelEvaluationHandler,
}

func router(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)

	if handler, exists := routes[r.URL.Path]; exists {
		handler(w, r)
		return
	}

	log.Printf("Redirecting to /prompts from %s", r.URL.Path)
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}
