package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool)
var clientsMutex sync.Mutex

func main() {
	log.Println("Starting the server...")
	http.HandleFunc("/", router)
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/templates/", http.StripPrefix("/templates/", http.FileServer(http.Dir("templates"))))
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	log.Println("Server is listening on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// Handle move prompt
func movePromptHandler(w http.ResponseWriter, r *http.Request) {
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
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			funcMap := template.FuncMap{
				"inc": func(i int) int {
					return i + 1
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
				Prompts []Prompt
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
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) && newIndex >= 0 && newIndex <= len(prompts) {
			prompt := prompts[index]
			prompts = append(prompts[:index], prompts[index+1:]...)
			if newIndex > index {
				newIndex--
			}
			prompts = append(prompts[:newIndex], append([]Prompt{prompt}, prompts[newIndex:]...)...)
		}
		err = writePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt moved successfully")
		broadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

func router(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received: %s %s", r.Method, r.URL.Path)
	switch r.URL.Path {
	case "/prompts":
		promptListHandler(w, r)
	case "/results":
		resultsHandler(w, r)
	case "/update_result":
		updateResultHandler(w, r)
	case "/add_model":
		addModelHandler(w, r)
	case "/edit_model":
		editModelHandler(w, r)
	case "/delete_model":
		deleteModelHandler(w, r)
	case "/add_prompt":
		addPromptHandler(w, r)
	case "/edit_prompt":
		editPromptHandler(w, r)
	case "/delete_prompt":
		deletePromptHandler(w, r)
	case "/move_prompt":
		movePromptHandler(w, r)
	case "/reset_results":
		resetResultsHandler(w, r)
	case "/confirm_refresh_results":
		confirmRefreshResultsHandler(w, r)
	case "/refresh_results":
		refreshResultsHandler(w, r)
	case "/export_results":
		exportResultsHandler(w, r)
	case "/import_results":
		importResultsHandler(w, r)
	case "/export_prompts":
		exportPromptsHandler(w, r)
	case "/import_prompts":
		importPromptsHandler(w, r)
	case "/update_prompts_order":
		updatePromptsOrderHandler(w, r)
	case "/reset_prompts":
		resetPromptsHandler(w, r)
	case "/import_error":
		importErrorHandler(w, r)
	default:
		log.Printf("Redirecting to /prompts from %s", r.URL.Path)
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling websocket connection")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	defer func() {
		clientsMutex.Lock()
		delete(clients, conn)
		clientsMutex.Unlock()
		log.Println("Closing websocket connection")
		conn.Close()
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Websocket error: %v", err)
			}
			break
		}

		var message struct {
			Type  string `json:"type"`
			Order []int  `json:"order"`
		}
		if err := json.Unmarshal(msg, &message); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		switch message.Type {
		case "update_prompts_order":
			updatePromptsOrder(message.Order)
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}
	}
}

func broadcastResults() {
	prompts := readPrompts()
	results := readResults()

	modelPassPercentages := make(map[string]float64)
	modelTotalScores := make(map[string]int)
	for model, result := range results {
		score := 0
		for _, pass := range result.Passes {
			if pass {
				score++
			}
		}
		modelPassPercentages[model] = float64(score) / float64(len(prompts)) * 100
		modelTotalScores[model] = score * 100
	}

	models := make([]string, 0, len(results))
	for model := range results {
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool {
		return modelTotalScores[models[i]] > modelTotalScores[models[j]]
	})

	payload := struct {
		Results         map[string][]bool
		Models          []string
		PassPercentages map[string]float64
		TotalScores     map[string]int
		Prompts         []string
	}{
		Results:         resultsToBoolMap(results),
		Models:          models,
		PassPercentages: modelPassPercentages,
		TotalScores:     modelTotalScores,
		Prompts:         promptsToStringArray(prompts),
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		err := client.WriteJSON(payload)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func resultsToBoolMap(results map[string]Result) map[string][]bool {
	resultsForTemplate := make(map[string][]bool)
	for model, result := range results {
		resultsForTemplate[model] = result.Passes
	}
	return resultsForTemplate
}

func promptsToStringArray(prompts []Prompt) []string {
	promptsTexts := make([]string, len(prompts))
	for i, prompt := range prompts {
		promptsTexts[i] = prompt.Text
	}
	return promptsTexts
}

// Handle update prompts order
func updatePromptsOrderHandler(w http.ResponseWriter, r *http.Request) {
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
	updatePromptsOrder(order)
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

func updatePromptsOrder(order []int) {
	prompts := readPrompts()
	if len(order) != len(prompts) {
		log.Println("Invalid order length")
		return
	}
	orderedPrompts := make([]Prompt, len(prompts))
	for i, index := range order {
		if index < 0 || index >= len(prompts) {
			log.Println("Invalid index in order")
			return
		}
		orderedPrompts[i] = prompts[index]
	}
	err := writePrompts(orderedPrompts)
	if err != nil {
		log.Printf("Error writing prompts: %v", err)
		return
	}
	log.Println("Prompts order updated successfully")
	broadcastResults()
}

// Handle add prompt
func addPromptHandler(w http.ResponseWriter, r *http.Request) {
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
	prompts := readPrompts()
	prompts = append(prompts, Prompt{Text: promptText})
	err = writePrompts(prompts)
	if err != nil {
		log.Printf("Error writing prompts: %v", err)
		http.Error(w, "Error writing prompts", http.StatusInternalServerError)
		return
	}
	log.Println("Prompt added successfully")
	broadcastResults()
	http.Redirect(w, r, "/prompts", http.StatusSeeOther)
}

// Handle add model
func addModelHandler(w http.ResponseWriter, r *http.Request) {
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
	results := readResults()
	if results == nil {
		results = make(map[string]Result)
	}
	if _, ok := results[modelName]; !ok {
		results[modelName] = Result{Passes: make([]bool, len(readPrompts()))}
	}
	err = writeResults(results)
	if err != nil {
		log.Printf("Error writing results: %v", err)
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}
	log.Println("Model added successfully")
	broadcastResults()
	http.Redirect(w, r, "/results", http.StatusSeeOther)
}

// Handle edit model
func editModelHandler(w http.ResponseWriter, r *http.Request) {
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

		results := readResults()
		if _, exists := results[newModelName]; exists {
			http.Error(w, "Model with this name already exists", http.StatusBadRequest)
			return
		}

		results[newModelName] = results[modelName]
		delete(results, modelName)
		writeResults(results)

		broadcastResults()
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
func deleteModelHandler(w http.ResponseWriter, r *http.Request) {
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

		results := readResults()
		delete(results, modelName)
		writeResults(results)

		broadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}

// Handle export prompts
func exportPromptsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling export prompts")
	prompts := readPrompts()

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
func importPromptsHandler(w http.ResponseWriter, r *http.Request) {
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

		var prompts []Prompt
		for _, line := range lines {
			if line == "" || line == "Prompt" {
				continue
			}
			prompts = append(prompts, Prompt{Text: line})
		}
		writePrompts(prompts)
		log.Println("Prompts imported successfully")
		broadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	} else {
		t, _ := template.ParseFiles("templates/import_prompts.html")
		t.Execute(w, nil)
	}
}

// Handle import results
func importResultsHandler(w http.ResponseWriter, r *http.Request) {
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

		results := make(map[string]Result)
		prompts := readPrompts()
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
			results[model] = Result{Passes: passes}
		}
		err = writeResults(results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results imported successfully")
		broadcastResults()
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
func editPromptHandler(w http.ResponseWriter, r *http.Request) {
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
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			funcMap := template.FuncMap{
				"markdown": func(text string) template.HTML {
					unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
					html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
					return template.HTML(html)
				},
			}
			t, err := template.New("edit_prompt.html").Funcs(funcMap).ParseFiles("templates/edit_prompt.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index  int
				Prompt string
			}{
				Index:  index,
				Prompt: prompts[index].Text,
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
		if editedPrompt == "" {
			log.Println("Prompt text cannot be empty")
			http.Error(w, "Prompt text cannot be empty", http.StatusBadRequest)
			return
		}
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			prompts[index].Text = editedPrompt
		}
		err = writePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt edited successfully")
		broadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle delete prompt
func deletePromptHandler(w http.ResponseWriter, r *http.Request) {
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
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			t, err := template.ParseFiles("templates/delete_prompt.html")
			if err != nil {
				log.Printf("Error parsing template: %v", err)
				http.Error(w, "Error parsing template", http.StatusInternalServerError)
				return
			}
			err = t.Execute(w, struct {
				Index  int
				Prompt string
			}{
				Index:  index,
				Prompt: prompts[index].Text,
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
		prompts := readPrompts()
		if index >= 0 && index < len(prompts) {
			prompts = append(prompts[:index], prompts[index+1:]...)
		}
		err = writePrompts(prompts)
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompt deleted successfully")
		broadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle prompt list page
func promptListHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling prompt list page")
	prompts := readPrompts()
	orderFilter := r.FormValue("order_filter")

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

	promptTexts := make([]string, len(prompts))
	promptIndices := make([]int, len(prompts))
	for i, prompt := range prompts {
		promptTexts[i] = prompt.Text
		promptIndices[i] = i + 1
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
	}

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

	err = t.Execute(w, struct {
		Prompts       []string
		PromptIndices []int
		OrderFilter   int
	}{
		Prompts:       promptTexts,
		PromptIndices: promptIndices,
		OrderFilter:   orderFilterInt,
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		// http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Println("Prompt list page rendered successfully")
}

// Handle results page
func resultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling results page")
	prompts := readPrompts()
	results := readResults()

	log.Println("Calculating total scores for each model")
	// Calculate total scores for each model
	modelScores := make(map[string]int)
	for model, result := range results {
		score := 0
		for _, pass := range result.Passes {
			if pass {
				score++
			}
		}
		modelScores[model] = score
	}

	log.Println("Sorting models by score in descending order")
	// Sort models by score in descending order
	models := make([]string, 0, len(results))
	for model := range results {
		models = append(models, model)
	}
	sort.Slice(models, func(i, j int) bool {
		return modelScores[models[i]] > modelScores[models[j]]
	})
	log.Printf("Sorted models: %v", models)

	modelFilter := r.FormValue("model_filter")
	filteredResults := make(map[string]Result)
	if modelFilter != "" {
		if filteredModel, ok := results[modelFilter]; ok {
			filteredResults[modelFilter] = filteredModel
		}
	} else {
		filteredResults = results
	}

	funcMap := template.FuncMap{
		"inc": func(i int) int {
			// slog.Info(strconv.Itoa(i))
			return i + 1
		},
	}
	t, err := template.New("results.html").Funcs(funcMap).ParseFiles("templates/results.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if t == nil {
		log.Println("Error parsing template")
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}
	promptTexts := make([]string, len(prompts))
	for i, prompt := range prompts {
		promptTexts[i] = prompt.Text
	}
	resultsForTemplate := make(map[string][]bool)
	for model, result := range filteredResults {
		resultsForTemplate[model] = result.Passes
	}
	modelPassPercentages := make(map[string]float64)
	modelTotalScores := make(map[string]int)
	promptIndices := make([]int, len(prompts))
	for i := range prompts {
		promptIndices[i] = i + 1
	}
	for model, result := range filteredResults {
		score := 0
		for _, pass := range result.Passes {
			if pass {
				score++
			}
		}
		modelPassPercentages[model] = float64(score) / float64(len(prompts)) * 100
		modelTotalScores[model] = score
	}

	err = t.Execute(w, struct {
		Prompts         []string
		Results         map[string][]bool
		Models          []string
		PassPercentages map[string]float64
		ModelFilter     string
		TotalScores     map[string]int
		PromptIndices   []int
	}{
		Prompts:         promptTexts,
		Results:         resultsForTemplate,
		Models:          models,
		PassPercentages: modelPassPercentages,
		ModelFilter:     modelFilter,
		TotalScores:     modelTotalScores,
		PromptIndices:   promptIndices,
	})
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
	log.Println("Results page rendered successfully")
}

// Handle AJAX requests to update results
func updateResultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling update result")
	r.ParseForm()
	model := r.Form.Get("model")
	promptIndexStr := r.Form.Get("promptIndex")
	passStr := r.Form.Get("pass")
	promptIndex, _ := strconv.Atoi(promptIndexStr)
	pass, err := strconv.ParseBool(passStr)
	if err != nil {
		log.Printf("Invalid pass value: %v", err)
		http.Error(w, "Invalid pass value", http.StatusBadRequest)
		return
	}

	results := readResults()
	if results == nil {
		results = make(map[string]Result)
	}
	if _, ok := results[model]; !ok {
		results[model] = Result{Passes: make([]bool, len(readPrompts()))}
	}

	prompts := readPrompts()
	result := results[model]
	if len(result.Passes) < len(prompts) {
		result.Passes = append(result.Passes, make([]bool, len(prompts)-len(result.Passes))...)
	}

	if promptIndex >= 0 && promptIndex < len(result.Passes) {
		result.Passes[promptIndex] = pass
	}
	results[model] = result
	err = writeResults(results)
	if err != nil {
		log.Printf("Error writing results: %v", err)
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}

	broadcastResults()

	_, err = w.Write([]byte("OK"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("Result updated successfully")
}

// Handle reset results
func resetResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling reset results")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/reset_results.html")
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
		emptyResults := make(map[string]Result)
		err := writeResults(emptyResults)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results reset successfully")
		broadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}

// Handle confirm refresh results
func confirmRefreshResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling confirm refresh results")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/confirm_refresh_results.html")
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
		results := readResults()
		for model := range results {
			results[model] = Result{Passes: make([]bool, len(readPrompts()))}
		}
		err := writeResults(results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results refreshed successfully")
		broadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}

// Handle refresh results
func refreshResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling refresh results")
	if r.Method == "GET" {
		t, err := template.ParseFiles("templates/confirm_refresh_results.html")
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
		results := readResults()
		for model := range results {
			results[model] = Result{Passes: make([]bool, len(readPrompts()))}
		}
		err := writeResults(results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results refreshed successfully")
		broadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	}
}

// Handle export results
func exportResultsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling export results")
	results := readResults()
	prompts := readPrompts()

	// Create CSV string
	csvString := "Model,"
	for i := range prompts {
		csvString += "Prompt " + strconv.Itoa(i+1) + ","
	}
	csvString += "\n"

	for model, result := range results {
		csvString += model + ","
		for _, pass := range result.Passes {
			csvString += strconv.FormatBool(pass) + ","
		}
		csvString += "\n"
	}

	// Set headers for CSV download
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=results.csv")

	// Write CSV to response
	_, err := w.Write([]byte(csvString))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("Results exported successfully")
}

// Handle reset prompts
func resetPromptsHandler(w http.ResponseWriter, r *http.Request) {
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
		err := writePrompts([]Prompt{})
		if err != nil {
			log.Printf("Error writing prompts: %v", err)
			http.Error(w, "Error writing prompts", http.StatusInternalServerError)
			return
		}
		log.Println("Prompts reset successfully")
		broadcastResults()
		http.Redirect(w, r, "/prompts", http.StatusSeeOther)
	}
}

// Handle import error
func importErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling import error")
	t, err := template.ParseFiles("templates/import_error.html")
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
