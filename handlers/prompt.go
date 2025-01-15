package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"

	"llm-tournament/middleware"
)

// Handle prompt list page
func PromptListHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling prompt list page")
	orderFilter := r.FormValue("order_filter")
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

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"markdown": func(text string) template.HTML {
			unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
			html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
			return template.HTML(html)
		},
		"string": func(i int) string {
			return strconv.Itoa(i)
		},
		"tolower":  strings.ToLower,
		"contains": strings.Contains,
	}
	funcMap["json"] = func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	}
	pageName := "Prompts"
	t, err := template.New("prompt_list.html").Funcs(funcMap).ParseFiles("templates/prompt_list.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		return
	}
	if t == nil {
		log.Println("Error parsing template")
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	suites, err := middleware.ListPromptSuites()
	if err != nil {
		log.Printf("Error listing prompt suites: %v", err)
		http.Error(w, "Error listing prompt suites", http.StatusInternalServerError)
		return
	}

	currentSuite := middleware.GetCurrentSuiteName()
	var prompts []middleware.Prompt
	if currentSuite == "" {
		currentSuite = "default"
	}
	prompts, err = middleware.ReadPromptSuite(currentSuite)
	if err != nil {
		log.Printf("Error reading prompt suite: %v", err)
		http.Error(w, "Error reading prompt suite", http.StatusInternalServerError)
		return
	}
	if len(prompts) == 0 && currentSuite == "default" {
		prompts, err = middleware.ReadPromptSuite("default")
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

	profiles := middleware.ReadProfiles()
	err = t.Execute(w, struct {
		PageName      string
		Prompts       []middleware.Prompt
		PromptIndices []int
		Profiles      []middleware.Profile
		OrderFilter   int
		SearchQuery   string
		Suites        []string
		CurrentSuite  string
	}{
		PageName:      pageName,
		Prompts:       promptTexts,
		PromptIndices: promptIndices,
		Profiles:      profiles,
		OrderFilter:   orderFilterInt,
		SearchQuery:   searchQuery,
		Suites:        suites,
		CurrentSuite:  currentSuite,
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		return
	}
	log.Println("Prompt list page rendered successfully")
}

// Handle update prompts order
func UpdatePromptsOrderHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling update prompts order")
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
	middleware.UpdatePromptsOrder(order)
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// Handle add prompt
func AddPromptHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling add prompt")
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

	currentSuite := middleware.GetCurrentSuiteName()
	if currentSuite == "" {
		currentSuite = "default"
	}

	prompts, err := middleware.ReadPromptSuite(currentSuite)
	if err != nil {
		log.Printf("Error reading prompt suite: %v", err)
		http.Error(w, "Error reading prompt suite", http.StatusInternalServerError)
		return
	}

	prompts = append(prompts, middleware.Prompt{Text: promptText, Solution: solutionText, Profile: profile})
	err = middleware.WritePromptSuite(currentSuite, prompts)
	if err != nil {
		log.Printf("Error writing prompts: %v", err)
		http.Error(w, "Error writing prompts", http.StatusInternalServerError)
		return
	}
	log.Println("Prompt added successfully")
	middleware.BroadcastResults()
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// Handle export prompts
func ExportPromptsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling export prompts")
	prompts := middleware.ReadPrompts()

	// Create CSV string
	csvString := "Prompt\n"
	for _, prompt := range prompts {
		csvString += prompt.Text + "\n"
	}

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=prompts.csv")

	// Write CSV to response
	_, err := w.Write([]byte(csvString))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("Prompts exported successfully")
}

// Handle import prompts
func ImportPromptsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling import prompts")
	if r.Method == "POST" {
		file, _, err := r.FormFile("prompts_file")
		if err != nil {
			log.Printf("Error uploading file: %v", err)
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}
		defer file.Close()

		if file == nil {
			log.Println("No file provided")
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}

		// Read the file content
		data := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if err != nil && err.Error() != "EOF" {
				log.Printf("Error reading file: %v", err)
				http.Error(w, "Error reading file", http.StatusInternalServerError)
				return
			}
			if n > 0 {
				data = append(data, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Parse CSV data
		lines := strings.Split(string(data), "\n")
		if len(lines) <= 1 {
			log.Println("Invalid CSV format: No data found")
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}

		var prompts []middleware.Prompt
		for _, line := range lines {
			if line == "" || line == "Prompt" {
				continue
			}
			prompts = append(prompts, middleware.Prompt{Text: line})
		}
		middleware.WritePrompts(prompts)
		log.Println("Prompts imported successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	} else {
		t, _ := template.ParseFiles("templates/import_prompts.html")
		t.Execute(w, nil)
	}
}

// Handle import results
func ImportResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling import results")
	if r.Method == "POST" {
		file, _, err := r.FormFile("results_file")
		if err != nil {
			log.Printf("Error uploading file: %v", err)
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}
		defer file.Close()

		if file == nil {
			log.Println("No file provided")
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}

		// Read the file content
		data := make([]byte, 0)
		buf := make([]byte, 1024)
		for {
			n, err := file.Read(buf)
			if n > 0 {
				data = append(data, buf[:n]...)
			}
			if err != nil {
				break
			}
		}

		// Parse CSV data
		lines := strings.Split(string(data), "\n")
		if len(lines) <= 1 {
			log.Println("Invalid CSV format: No data found")
			http.Redirect(w, r, "/import_error", http.StatusSeeOther)
			return
		}

		suiteName := middleware.GetCurrentSuiteName()
		results := make(map[string]middleware.Result)
		prompts := middleware.ReadPrompts()
		for i, line := range lines {
			if i == 0 || line == "" {
				continue
			}
			parts := strings.Split(line, ",")
			if len(parts) < 2 {
				continue
			}
			model := parts[0]
			var passes []bool
			for _, passStr := range parts[1:] {
				pass, _ := strconv.ParseBool(passStr)
				passes = append(passes, pass)
			}
			if len(passes) < len(prompts) {
				passes = append(passes, make([]bool, len(prompts)-len(passes))...)
			}
			results[model] = middleware.Result{Passes: passes}
		}
		suiteName = middleware.GetCurrentSuiteName()
		err = middleware.WriteResults(suiteName, results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results imported successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	} else {
		t, err := template.ParseFiles("templates/import_results.html")
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
	}
}

// Handle edit prompt
func EditPromptHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling edit prompt")
	if r.Method == "GET" {
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
		prompts := middleware.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			funcMap := template.FuncMap{
				"markdown": func(text string) template.HTML {
					unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
					html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
					return template.HTML(html)
				},
			}
			profiles := middleware.ReadProfiles()
			t, err := template.New("edit_prompt.html").Funcs(funcMap).ParseFiles("templates/edit_prompt.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index    int
				Prompt   middleware.Prompt
				Profiles []middleware.Profile
			}{
				Index:    index,
				Prompt:   prompts[index],
				Profiles: profiles,
			})
			if err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Error executing template", http.StatusInternalServerError)
				return
			}
		}
	} else if r.Method == "POST" {
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
		prompts := middleware.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			prompts[index].Text = editedPrompt
			prompts[index].Solution = editedSolution
			prompts[index].Profile = editedProfile
		}
		err = middleware.WritePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt edited successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle bulk delete prompts page
func BulkDeletePromptsPageHandler(w http.ResponseWriter, r *http.Request) {
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

	prompts := middleware.ReadPrompts()
	var selectedPrompts []middleware.Prompt
	for _, index := range indices {
		if index >= 0 && index < len(prompts) {
			selectedPrompts = append(selectedPrompts, prompts[index])
		}
	}

	funcMap := template.FuncMap{
		"markdown": func(text string) template.HTML {
			unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
			html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
			return template.HTML(html)
		},
	}

	t, err := template.New("bulk_delete_prompts.html").Funcs(funcMap).ParseFiles("templates/bulk_delete_prompts.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, struct {
		Indices string
		Prompts []middleware.Prompt
	}{
		Indices: indicesStr,
		Prompts: selectedPrompts,
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

// Handle bulk delete prompts
func BulkDeletePromptsHandler(w http.ResponseWriter, r *http.Request) {
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

	prompts := middleware.ReadPrompts()
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

	err = middleware.WritePrompts(filteredPrompts)
	if err != nil {
		log.Printf("Error writing prompts: %v", err)
		http.Error(w, "Error writing prompts", http.StatusInternalServerError)
		return
	}

	log.Println("Prompts deleted successfully")
	middleware.BroadcastResults()
	w.WriteHeader(http.StatusOK)
}

// Handle delete prompt
func DeletePromptHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling delete prompt")
	if r.Method == "GET" {
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
		prompts := middleware.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			funcMap := template.FuncMap{
				"markdown": func(text string) template.HTML {
					unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
					html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
					return template.HTML(html)
				},
			}
			t, err := template.New("delete_prompt.html").Funcs(funcMap).ParseFiles("templates/delete_prompt.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index  int
				Prompt middleware.Prompt
			}{
				Index:  index,
				Prompt: prompts[index],
			})
			if err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Error executing template", http.StatusInternalServerError)
				return
			}
		}
	} else if r.Method == "POST" {
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
		prompts := middleware.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			prompts = append(prompts[:index], prompts[index+1:]...)
		}
		err = middleware.WritePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt deleted successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle move prompt
func MovePromptHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling move prompt")
	if r.Method == "GET" {
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
		prompts := middleware.ReadPrompts()
		if index >= 0 && index < len(prompts) {
			funcMap := template.FuncMap{
				"inc": func(i int) int {
					return i + 1
				},
				"markdown": func(text string) template.HTML {
					unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
					html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
					return template.HTML(html)
				},
			}
			t, err := template.New("move_prompt.html").Funcs(funcMap).ParseFiles("templates/move_prompt.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index   int
				Prompt  string
				Prompts []middleware.Prompt
			}{
				Index:   index,
				Prompt:  prompts[index].Text,
				Prompts: prompts,
			})
			if err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Error executing template", http.StatusInternalServerError)
				return
			}
		}
	} else if r.Method == "POST" {
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
		prompts := middleware.ReadPrompts()
		if index >= 0 && index < len(prompts) && newIndex >= 0 && newIndex <= len(prompts) {
			prompt := prompts[index]
			prompts = append(prompts[:index], prompts[index+1:]...)
			if newIndex > index {
				newIndex--
			}
			prompts = append(prompts[:newIndex], append([]middleware.Prompt{prompt}, prompts[newIndex:]...)...)
		}
		err = middleware.WritePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt moved successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle reset prompts
func ResetPromptsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling reset prompts")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/reset_prompts.html")
		if err != nil {
			http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = t.Execute(w, nil)
		if err != nil {
			http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if r.Method == "POST" {
		err := middleware.WritePrompts([]middleware.Prompt{})
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompts reset successfully")
		middleware.BroadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}
