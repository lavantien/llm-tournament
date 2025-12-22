package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"

	"llm-tournament/middleware"
	"llm-tournament/templates"
)

var readAll = io.ReadAll

// PromptListHandler handles the prompt list page (backward compatible wrapper)
func PromptListHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.PromptList(w, r)
}

// UpdatePromptsOrderHandler handles updating prompts order (backward compatible wrapper)
func UpdatePromptsOrderHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.UpdatePromptsOrder(w, r)
}

// AddPromptHandler handles adding a prompt (backward compatible wrapper)
func AddPromptHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.AddPrompt(w, r)
}

// ExportPromptsHandler handles exporting prompts (backward compatible wrapper)
func ExportPromptsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ExportPrompts(w, r)
}

// ImportPromptsHandler handles importing prompts (backward compatible wrapper)
func ImportPromptsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ImportPrompts(w, r)
}

// ImportResultsHandler handles importing results (backward compatible wrapper)
func ImportResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ImportResults(w, r)
}

// EditPromptHandler handles editing a prompt (backward compatible wrapper)
func EditPromptHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.EditPrompt(w, r)
}

// BulkDeletePromptsPageHandler handles bulk delete prompts page (backward compatible wrapper)
func BulkDeletePromptsPageHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.BulkDeletePromptsPage(w, r)
}

// BulkDeletePromptsHandler handles bulk delete prompts (backward compatible wrapper)
func BulkDeletePromptsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.BulkDeletePrompts(w, r)
}

// DeletePromptHandler handles deleting a prompt (backward compatible wrapper)
func DeletePromptHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.DeletePrompt(w, r)
}

// MovePromptHandler handles moving a prompt (backward compatible wrapper)
func MovePromptHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.MovePrompt(w, r)
}

// ResetPromptsHandler handles resetting prompts (backward compatible wrapper)
func ResetPromptsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ResetPrompts(w, r)
}

// PromptList handles the prompt list page
func (h *Handler) PromptList(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling prompt list page")
	orderFilter := r.FormValue("order_filter")
	profileFilter := r.FormValue("profile_filter")
	searchQuery := r.FormValue("search_query")

	orderFilterInt := 0
	if orderFilter != "" {
		var err error
		orderFilterInt, err = strconv.Atoi(orderFilter)
		if err != nil {
			log.Printf("Invalid order filter: %v", err)
			http.Error(w, "Invalid order filter", http.StatusBadRequest)
			return
		}
	}

	funcMap := templates.FuncMap

	suites, err := h.DataStore.ListPromptSuites()
	if err != nil {
		log.Printf("Error listing prompt suites: %v", err)
		http.Error(w, "Error listing prompt suites", http.StatusInternalServerError)
		return
	}

	currentSuite := h.DataStore.GetCurrentSuiteName()
	var prompts []middleware.Prompt
	if currentSuite == "" {
		currentSuite = "default"
	}
	prompts, err = h.DataStore.ReadPromptSuite(currentSuite)
	if err != nil {
		log.Printf("Error reading prompt suite: %v", err)
		http.Error(w, "Error reading prompt suite", http.StatusInternalServerError)
		return
	}
	if len(prompts) == 0 && currentSuite == "default" {
		prompts, err = h.DataStore.ReadPromptSuite("default")
		if err != nil {
			log.Printf("Error reading default prompt suite: %v", err)
			http.Error(w, "Error reading default prompt suite", http.StatusInternalServerError)
			return
		}
	}
	promptTexts := make([]middleware.Prompt, len(prompts))
	promptIndices := make([]int, len(prompts))
	for i, prompt := range prompts {
		promptTexts[i] = prompt
		promptIndices[i] = i + 1
	}

	profiles := h.DataStore.ReadProfiles()
	pageName := "Prompts"

	err = h.Renderer.Render(w, "prompt_list.html", funcMap, struct {
		PageName      string
		Prompts       []middleware.Prompt
		PromptIndices []int
		Profiles      []middleware.Profile
		OrderFilter   int
		ProfileFilter string
		SearchQuery   string
		Suites        []string
		CurrentSuite  string
	}{
		PageName:      pageName,
		Prompts:       promptTexts,
		PromptIndices: promptIndices,
		Profiles:      profiles,
		OrderFilter:   orderFilterInt,
		ProfileFilter: profileFilter,
		SearchQuery:   searchQuery,
		Suites:        suites,
		CurrentSuite:  currentSuite,
	}, "templates/prompt_list.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	log.Println("Prompt list page rendered successfully")
}

// UpdatePromptsOrder handles updating prompts order
func (h *Handler) UpdatePromptsOrder(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling update prompts order")
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
	orderStr := r.Form.Get("order")
	if orderStr == "" {
		log.Println("Order cannot be empty")
		http.Error(w, "Order cannot be empty", http.StatusBadRequest)
		return
	}
	var order []int
	err = json.Unmarshal([]byte(orderStr), &order)
	if err != nil {
		log.Printf("Error parsing order: %v", err)
		http.Error(w, "Error parsing order", http.StatusBadRequest)
		return
	}
	h.DataStore.UpdatePromptsOrder(order)
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// AddPrompt handles adding a prompt
func (h *Handler) AddPrompt(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling add prompt")
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
	promptText := r.Form.Get("prompt")
	if promptText == "" {
		log.Println("Prompt text cannot be empty")
		http.Error(w, "Prompt text cannot be empty", http.StatusBadRequest)
		return
	}
	solutionText := r.Form.Get("solution")
	profile := r.Form.Get("profile")

	currentSuite := h.DataStore.GetCurrentSuiteName()
	if currentSuite == "" {
		currentSuite = "default"
	}

	prompts, err := h.DataStore.ReadPromptSuite(currentSuite)
	if err != nil {
		log.Printf("Error reading prompt suite: %v", err)
		http.Error(w, "Error reading prompt suite", http.StatusInternalServerError)
		return
	}

	prompts = append(prompts, middleware.Prompt{Text: promptText, Solution: solutionText, Profile: profile})
	err = h.DataStore.WritePromptSuite(currentSuite, prompts)
	if err != nil {
		log.Printf("Error writing prompts: %v", err)
		http.Error(w, "Error writing prompts", http.StatusInternalServerError)
		return
	}
	log.Println("Prompt added successfully")
	h.DataStore.BroadcastResults()
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// ExportPrompts handles exporting prompts
func (h *Handler) ExportPrompts(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling export prompts")
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	prompts := h.DataStore.ReadPrompts()

	// Convert prompts to JSON
	jsonData, _ := json.MarshalIndent(prompts, "", "  ")

	// Set headers for JSON download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment;filename=prompts.json")

	// Write JSON to response
	_, err := w.Write(jsonData)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("Prompts exported successfully as JSON")
}

// ImportPrompts handles importing prompts
func (h *Handler) ImportPrompts(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling import prompts")
	switch r.Method {
	case http.MethodPost:
		file, _, err := r.FormFile("prompts_file")
		if err != nil {
			log.Printf("Error uploading file: %v", err)
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}
		defer func() { _ = file.Close() }()

		// Read the file content
		data, err := readAll(file)
		if err != nil {
			log.Printf("Error reading file: %v", err)
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		// Parse JSON data
		var prompts []middleware.Prompt
		err = json.Unmarshal(data, &prompts)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}

		// Validate imported prompts
		if len(prompts) == 0 {
			log.Println("No prompts found in JSON file")
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}

		// Write the imported prompts
		err = h.DataStore.WritePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}

		log.Println("Prompts imported successfully from JSON")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	case http.MethodGet:
		if err := h.Renderer.RenderTemplateSimple(w, "import_prompts.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ImportResults handles importing results
func (h *Handler) ImportResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling import results")
	switch r.Method {
	case http.MethodPost:
		file, _, err := r.FormFile("results_file")
		if err != nil {
			log.Printf("Error uploading file: %v", err)
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}
		defer func() { _ = file.Close() }()

		// Read the file content
		data, err := readAll(file)
		if err != nil {
			log.Printf("Error reading file: %v", err)
			http.Error(w, "Error reading file", http.StatusInternalServerError)
			return
		}

		// Parse JSON data
		var results map[string]middleware.Result
		err = json.Unmarshal(data, &results)
		if err != nil {
			log.Printf("Error parsing JSON: %v", err)
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}

		// Validate imported results
		if len(results) == 0 {
			log.Println("No results found in JSON file")
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}

		// Ensure scores arrays match prompts length
		prompts := h.DataStore.ReadPrompts()
		for model, result := range results {
			if len(result.Scores) < len(prompts) {
				newScores := make([]int, len(prompts))
				copy(newScores, result.Scores)
				result.Scores = newScores
				results[model] = result
			}
		}

		// Write the imported results
		suiteName := h.DataStore.GetCurrentSuiteName()
		err = h.DataStore.WriteResults(suiteName, results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}

		log.Println("Results imported successfully from JSON")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	case http.MethodGet:
		if err := h.Renderer.RenderTemplateSimple(w, "import_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// EditPrompt handles editing a prompt
func (h *Handler) EditPrompt(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling edit prompt")
	switch r.Method {
	case "GET":
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		prompts := h.DataStore.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			funcMap := templates.FuncMap
			profiles := h.DataStore.ReadProfiles()
			err := h.Renderer.Render(w, "edit_prompt.html", funcMap, struct {
				Index    int
				Prompt   middleware.Prompt
				Profiles []middleware.Profile
			}{
				Index:    index,
				Prompt:   prompts[index],
				Profiles: profiles,
			}, "templates/edit_prompt.html")
			if err != nil {
				log.Printf("Error rendering template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
				return
			}
		}
	case "POST":
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		editedPrompt := r.Form.Get("prompt")
		editedSolution := r.Form.Get("solution")
		editedProfile := r.Form.Get("profile")
		if editedPrompt == "" {
			log.Println("Prompt text cannot be empty")
			http.Error(w, "Prompt text cannot be empty", http.StatusBadRequest)
			return
		}
		prompts := h.DataStore.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			prompts[index].Text = editedPrompt
			prompts[index].Solution = editedSolution
			prompts[index].Profile = editedProfile
		}
		err = h.DataStore.WritePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt edited successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// BulkDeletePromptsPage handles bulk delete prompts page
func (h *Handler) BulkDeletePromptsPage(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling bulk delete prompts page")
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	indicesStr := r.URL.Query().Get("indices")
	if indicesStr == "" {
		log.Println("No indices provided for deletion")
		http.Error(w, "No indices provided for deletion", http.StatusBadRequest)
		return
	}

	var indices []int
	err := json.Unmarshal([]byte(indicesStr), &indices)
	if err != nil {
		log.Printf("Error unmarshalling indices: %v", err)
		http.Error(w, "Error unmarshalling indices", http.StatusBadRequest)
		return
	}

	prompts := h.DataStore.ReadPrompts()
	var selectedPrompts []middleware.Prompt
	for _, index := range indices {
		if index >= 0 && index < len(prompts) {
			selectedPrompts = append(selectedPrompts, prompts[index])
		}
	}

	funcMap := templates.FuncMap

	err = h.Renderer.Render(w, "bulk_delete_prompts.html", funcMap, struct {
		Indices string
		Prompts []middleware.Prompt
	}{
		Indices: indicesStr,
		Prompts: selectedPrompts,
	}, "templates/bulk_delete_prompts.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// BulkDeletePrompts handles bulk delete prompts
func (h *Handler) BulkDeletePrompts(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling bulk delete prompts")
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Indices []int `json:"indices"`
	}

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	indices := request.Indices

	prompts := h.DataStore.ReadPrompts()
	if len(prompts) == 0 {
		log.Println("No prompts to delete")
		http.Error(w, "No prompts to delete", http.StatusBadRequest)
		return
	}

	if len(indices) == 0 {
		log.Println("No indices provided for deletion")
		http.Error(w, "No indices provided for deletion", http.StatusBadRequest)
		return
	}

	var filteredPrompts []middleware.Prompt
	for i, prompt := range prompts {
		found := false
		for _, index := range indices {
			if i == index {
				found = true
				break
			}
		}
		if !found {
			filteredPrompts = append(filteredPrompts, prompt)
		}
	}

	err = h.DataStore.WritePrompts(filteredPrompts)
	if err != nil {
		log.Printf("Error writing prompts: %v", err)
		http.Error(w, "Error writing prompts", http.StatusInternalServerError)
		return
	}

	log.Println("Prompts deleted successfully")
	h.DataStore.BroadcastResults()
	w.WriteHeader(http.StatusOK)
}

// DeletePrompt handles deleting a prompt
func (h *Handler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling delete prompt")
	switch r.Method {
	case "GET":
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		prompts := h.DataStore.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			funcMap := templates.FuncMap
			err := h.Renderer.Render(w, "delete_prompt.html", funcMap, struct {
				Index  int
				Prompt middleware.Prompt
			}{
				Index:  index,
				Prompt: prompts[index],
			}, "templates/delete_prompt.html")
			if err != nil {
				log.Printf("Error rendering template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
				return
			}
		}
	case "POST":
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		prompts := h.DataStore.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			prompts = append(prompts[:index], prompts[index+1:]...)
		}
		err = h.DataStore.WritePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt deleted successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// MovePrompt handles moving a prompt
func (h *Handler) MovePrompt(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling move prompt")
	switch r.Method {
	case "GET":
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		prompts := h.DataStore.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			funcMap := templates.FuncMap
			err := h.Renderer.Render(w, "move_prompt.html", funcMap, struct {
				Index   int
				Prompt  string
				Prompts []middleware.Prompt
			}{
				Index:   index,
				Prompt:  prompts[index].Text,
				Prompts: prompts,
			}, "templates/move_prompt.html")
			if err != nil {
				log.Printf("Error rendering template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
				return
			}
		}
	case "POST":
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}
		indexStr := r.Form.Get("index")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			log.Printf("Invalid index: %v", err)
			http.Error(w, "Invalid index", http.StatusBadRequest)
			return
		}
		newIndexStr := r.Form.Get("new_index")
		newIndex, err := strconv.Atoi(newIndexStr)
		if err != nil {
			log.Printf("Invalid new index: %v", err)
			http.Error(w, "Invalid new index", http.StatusBadRequest)
			return
		}
		prompts := h.DataStore.ReadPrompts()
		if index >= 0 && index < len(prompts) && newIndex >= 0 && newIndex <= len(prompts) {
			prompt := prompts[index]
			prompts = append(prompts[:index], prompts[index+1:]...)
			if newIndex > index {
				newIndex--
			}
			prompts = append(prompts[:newIndex], append([]middleware.Prompt{prompt}, prompts[newIndex:]...)...)
		}
		err = h.DataStore.WritePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt moved successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ResetPrompts handles resetting prompts
func (h *Handler) ResetPrompts(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling reset prompts")
	if r.Method == "GET" {
		if err := h.Renderer.RenderTemplateSimple(w, "reset_prompts.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	} else if r.Method == "POST" {
		err := h.DataStore.WritePrompts([]middleware.Prompt{})
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompts reset successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
