package handlers

import (
	"encoding/json"
	"io"
	"llm-tournament/middleware"
	"llm-tournament/templates"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// min returns the smaller of x or y
func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// max returns the larger of x or y
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// initRand returns a new random number generator seeded with the current time
func initRand() *rand.Rand {
	source := rand.NewSource(time.Now().UnixNano())
	return rand.New(source)
}

// GroupedPrompt represents a prompt with its profile information
type GroupedPrompt struct {
	Index       int
	Text        string
	ProfileID   string
	ProfileName string
}

// ResultsHandler handles the results page (backward compatible wrapper)
func ResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.Results(w, r)
}

// UpdateResultHandler handles updating results (backward compatible wrapper)
func UpdateResultHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.UpdateResult(w, r)
}

// ResetResultsHandler handles resetting results (backward compatible wrapper)
func ResetResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ResetResults(w, r)
}

// ConfirmRefreshResultsHandler handles confirm refresh results (backward compatible wrapper)
func ConfirmRefreshResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ConfirmRefreshResults(w, r)
}

// RefreshResultsHandler handles refresh results (backward compatible wrapper)
func RefreshResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.RefreshResults(w, r)
}

// EvaluateResult handles evaluating individual results (backward compatible wrapper)
func EvaluateResult(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.EvaluateResultHandler(w, r)
}

// ExportResultsHandler handles exporting results (backward compatible wrapper)
func ExportResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.ExportResults(w, r)
}

// UpdateMockResultsHandler handles updating mock results (backward compatible wrapper)
func UpdateMockResultsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.UpdateMockResults(w, r)
}

// Results handles the results page
func (h *Handler) Results(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling results page")
	prompts := h.DataStore.ReadPrompts()
	results := h.DataStore.ReadResults()

	// Group prompts by profile
	var orderedPrompts []GroupedPrompt

	// Get all profiles first (to include empty ones)
	profiles := h.DataStore.ReadProfiles()

	// Get profile groups using the utility function
	profileGroups, profileMap := middleware.GetProfileGroups(prompts, profiles)

	// Check if we have any uncategorized prompts
	hasUncategorized := false
	for _, prompt := range prompts {
		if prompt.Profile == "" {
			hasUncategorized = true
			break
		}
	}

	// Add a group for prompts with no profile only if needed
	if hasUncategorized {
		noProfileGroup := &middleware.ProfileGroup{
			ID:       "none",
			Name:     "Uncategorized",
			Color:    "hsl(0, 0%, 50%)",
			StartCol: -1,
			EndCol:   -1,
		}
		profileGroups = append(profileGroups, noProfileGroup)
		profileMap[""] = noProfileGroup
	}

	// Process prompts and assign them to profile groups
	currentCol := 0
	for i, prompt := range prompts {
		profileName := prompt.Profile

		group := profileMap[profileName]

		if group.StartCol == -1 {
			group.StartCol = currentCol
		}
		group.EndCol = currentCol

		orderedPrompts = append(orderedPrompts, GroupedPrompt{
			Index:       i,
			Text:        prompt.Text,
			ProfileID:   group.ID,
			ProfileName: profileName,
		})

		currentCol++
	}

	log.Println("Calculating total scores for each model")
	// Calculate total scores for each model
	modelScores := make(map[string]int)
	for model, result := range results {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		modelScores[model] = totalScore
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
	searchQuery := strings.ToLower(r.FormValue("search"))

	filteredResults := make(map[string]middleware.Result)
	for model, result := range results {
		// Apply model filter if specified
		if modelFilter != "" && model != modelFilter {
			continue
		}
		// Apply search filter if specified
		if searchQuery != "" && !strings.Contains(strings.ToLower(model), searchQuery) {
			continue
		}
		filteredResults[model] = result
	}

	pageName := templates.PageNameResults
	promptTexts := make([]string, len(prompts))
	for i, prompt := range prompts {
		promptTexts[i] = prompt.Text
	}
	resultsForTemplate := make(map[string]middleware.Result)
	for model, result := range filteredResults {
		// Initialize scores array if nil
		if result.Scores == nil {
			result.Scores = make([]int, len(prompts))
		}

		// Ensure scores array matches prompts length
		if len(result.Scores) != len(prompts) {
			newScores := make([]int, len(prompts))
			copy(newScores, result.Scores)
			result.Scores = newScores
		}

		// Ensure all scores are valid (0-100)
		for i, score := range result.Scores {
			if score < 0 || score > 100 {
				result.Scores[i] = 0
			}
		}

		// Create a new Result struct to ensure proper initialization
		resultsForTemplate[model] = middleware.Result{
			Scores: result.Scores,
		}
	}
	modelPassPercentages := make(map[string]float64)
	modelTotalScores := make(map[string]int)
	promptIndices := make([]int, len(prompts))
	for i := range prompts {
		promptIndices[i] = i + 1
	}
	for model, result := range filteredResults {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		// Avoid division by zero when there are no prompts
		if len(prompts) > 0 {
			modelPassPercentages[model] = float64(totalScore) / float64(len(prompts)*100) * 100
		} else {
			modelPassPercentages[model] = 0
		}
		modelTotalScores[model] = totalScore
	}

	// Log the data we're about to send to the template for debugging
	if len(models) > 0 && len(promptTexts) > 0 {
		log.Printf("First model: %s, scores: %v", models[0], resultsForTemplate[models[0]].Scores)
	}

	templateData := struct {
		PageName        string
		Prompts         []string
		Results         map[string]middleware.Result
		Models          []string
		PassPercentages map[string]float64
		ModelFilter     string
		TotalScores     map[string]int
		PromptIndices   []int
		SearchQuery     string
		ProfileGroups   []*middleware.ProfileGroup
		OrderedPrompts  []GroupedPrompt
		CurrentPath     string
	}{
		PageName:        pageName,
		Prompts:         promptTexts,
		Results:         resultsForTemplate,
		Models:          models,
		PassPercentages: modelPassPercentages,
		ModelFilter:     modelFilter,
		TotalScores:     modelTotalScores,
		PromptIndices:   promptIndices,
		SearchQuery:     searchQuery,
		ProfileGroups:   profileGroups,
		OrderedPrompts:  orderedPrompts,
		CurrentPath:     "/results",
	}

	err := h.Renderer.Render(w, "results.html", templates.FuncMap, templateData, "templates/results.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	log.Println("Results page rendered successfully")
}

// UpdateResult handles AJAX requests to update results
func (h *Handler) UpdateResult(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling update result")
	_ = r.ParseForm()
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

	suiteName := h.DataStore.GetCurrentSuiteName()
	results := h.DataStore.ReadResults()
	if results == nil {
		results = make(map[string]middleware.Result)
	}
	if _, ok := results[model]; !ok {
		results[model] = middleware.Result{
			Scores: make([]int, len(h.DataStore.ReadPrompts())),
		}
	}

	prompts := h.DataStore.ReadPrompts()
	result := results[model]
	if len(result.Scores) < len(prompts) {
		result.Scores = append(result.Scores, make([]int, len(prompts)-len(result.Scores))...)
	}
	if promptIndex >= 0 && promptIndex < len(result.Scores) {
		if pass {
			result.Scores[promptIndex] = 100
		} else {
			result.Scores[promptIndex] = 0
		}
	}
	results[model] = result
	err = h.DataStore.WriteResults(suiteName, results)
	if err != nil {
		log.Printf("Error writing results: %v", err)
		http.Error(w, "Error writing results", http.StatusInternalServerError)
		return
	}

	h.DataStore.BroadcastResults()

	_, err = w.Write([]byte("OK"))
	if err != nil {
		log.Printf("Error writing response: %v", err)
		http.Error(w, "Error writing response", http.StatusInternalServerError)
		return
	}
	log.Println("protocols.Result updated successfully")
}

// ResetResults handles resetting results
func (h *Handler) ResetResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling reset results")
	switch r.Method {
	case "GET":
		if err := h.Renderer.RenderTemplateSimple(w, "reset_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	case "POST":
		emptyResults := make(map[string]middleware.Result)
		suiteName := h.DataStore.GetCurrentSuiteName()
		err := h.DataStore.WriteResults(suiteName, emptyResults)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results reset successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ConfirmRefreshResults handles confirm refresh results
func (h *Handler) ConfirmRefreshResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling confirm refresh results")
	switch r.Method {
	case "GET":
		if err := h.Renderer.RenderTemplateSimple(w, "confirm_refresh_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	case "POST":
		results := h.DataStore.ReadResults()
		for model := range results {
			results[model] = middleware.Result{
				Scores: make([]int, len(h.DataStore.ReadPrompts())),
			}
		}
		suiteName := h.DataStore.GetCurrentSuiteName()
		err := h.DataStore.WriteResults(suiteName, results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results refreshed successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// RefreshResults handles refresh results
func (h *Handler) RefreshResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling refresh results")
	switch r.Method {
	case "GET":
		if err := h.Renderer.RenderTemplateSimple(w, "confirm_refresh_results.html", nil); err != nil {
			log.Printf("Error rendering template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	case "POST":
		results := h.DataStore.ReadResults()
		for model := range results {
			results[model] = middleware.Result{Scores: make([]int, len(h.DataStore.ReadPrompts()))}
		}
		suiteName := h.DataStore.GetCurrentSuiteName()
		err := h.DataStore.WriteResults(suiteName, results)
		if err != nil {
			log.Printf("Error writing results: %v", err)
			http.Error(w, "Error writing results", http.StatusInternalServerError)
			return
		}
		log.Println("Results refreshed successfully")
		h.DataStore.BroadcastResults()
		http.Redirect(w, r, "/results", http.StatusSeeOther)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// EvaluateResultHandler handles evaluation of individual results
func (h *Handler) EvaluateResultHandler(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")
	promptIndexStr := r.URL.Query().Get("prompt")

	// Validate required query parameters
	if model == "" || promptIndexStr == "" {
		http.Redirect(w, r, "/results", http.StatusSeeOther)
		return
	}
	if r.Method == "POST" {
		scoreStr := r.FormValue("score")
		score, err := strconv.Atoi(scoreStr)
		if err != nil {
			http.Error(w, "Invalid score value", http.StatusBadRequest)
			return
		}

		results := h.DataStore.ReadResults()
		if results == nil {
			results = make(map[string]middleware.Result)
		}

		result, exists := results[model]
		if !exists {
			// Initialize new result with scores array matching prompts length
			prompts := h.DataStore.ReadPrompts()
			result = middleware.Result{
				Scores: make([]int, len(prompts)),
			}
		}

		index, err := strconv.Atoi(promptIndexStr)
		if err != nil || index < 0 || index >= len(result.Scores) {
			http.Error(w, "Invalid prompt index", http.StatusBadRequest)
			return
		}

		// Update the score (ensure it's within 0-100 range)
		if score < 0 {
			score = 0
		} else if score > 100 {
			score = 100
		}
		result.Scores[index] = score
		results[model] = result

		// Write updated results
		err = h.DataStore.WriteResults(h.DataStore.GetCurrentSuiteName(), results)
		if err != nil {
			http.Error(w, "Failed to save results", http.StatusInternalServerError)
			return
		}

		// Broadcast updated results to all clients
		h.DataStore.BroadcastResults()

		// Add debug logging
		log.Printf("Updated score for model %s, prompt %d: %d", model, index, score)
		log.Printf("Current results for model %s: %v", model, result.Scores)

		// Redirect back to results page
		http.Redirect(w, r, "/results", http.StatusSeeOther)
		return
	}

	// Get current score for this model/prompt
	results := h.DataStore.ReadResults()
	currentScore := 0
	if result, exists := results[model]; exists {
		if index, err := strconv.Atoi(promptIndexStr); err == nil && index < len(result.Scores) {
			currentScore = result.Scores[index]
		}
	}

	// Get the prompt text and solution for display
	prompts := h.DataStore.ReadPrompts()
	var promptText, solution string
	promptIndex, err := strconv.Atoi(promptIndexStr)
	if err == nil && promptIndex >= 0 && promptIndex < len(prompts) {
		promptText = prompts[promptIndex].Text
		solution = prompts[promptIndex].Solution
	}

	// Get model response if available
	var modelResponse string
	db := middleware.GetDB()
	var modelID int
	var promptID int

	// Get model_id from model name
	err = db.QueryRow("SELECT id FROM models WHERE name = ?", model).Scan(&modelID)
	if err == nil {
		// Get prompt_id from database using prompt index (1-indexed)
		err = db.QueryRow("SELECT id FROM prompts WHERE suite_id = 1 ORDER BY display_order LIMIT 1 OFFSET ?", promptIndex).Scan(&promptID)
		if err == nil {
			// Get the response for this model/prompt pair
			err = db.QueryRow("SELECT response_text FROM model_responses WHERE model_id = ? AND prompt_id = ?", modelID, promptID).Scan(&modelResponse)
			if err != nil {
				// No response found, leave empty
				modelResponse = ""
			}
		}
	}

	data := struct {
		PageName      string
		Model         string
		PromptIndex   string
		ScoreOptions  map[string]int
		CurrentScore  int
		PromptText    string
		Solution      string
		TotalPrompts  int
		ModelResponse string
		ModelID       int
		PromptID      int
		CurrentPath   string
	}{
		PageName:      templates.PageNameEvaluate,
		Model:         model,
		PromptIndex:   promptIndexStr,
		ScoreOptions:  templates.ScoreOptions,
		CurrentScore:  currentScore,
		PromptText:    promptText,
		Solution:      solution,
		TotalPrompts:  len(prompts),
		ModelResponse: modelResponse,
		ModelID:       modelID,
		PromptID:      promptID,
		CurrentPath:   "/evaluate",
	}

	err = h.Renderer.Render(w, "evaluate.html", templates.FuncMap, data, "templates/evaluate.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

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

// UpdateMockResults handles updating results with randomly generated mock data
// that ensures even distribution across all tier levels
func (h *Handler) UpdateMockResults(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling update mock results")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body
	var mockData struct {
		Results         map[string]middleware.Result `json:"results"`
		Models          []string                     `json:"models"`
		PassPercentages map[string]float64           `json:"passPercentages"`
		TotalScores     map[string]int               `json:"totalScores"`
	}

	log.Println("Received mock data request")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &mockData)
	if err != nil {
		log.Printf("Error decoding mock data: %v", err)
		http.Error(w, "Invalid JSON data", http.StatusBadRequest)
		return
	}

	// Use client-provided scores instead of generating new ones
	log.Println("Using client-provided scores for mock data")

	prompts := h.DataStore.ReadPrompts()

	// If no prompts exist, create mock prompts with profiles
	if len(prompts) == 0 {
		db := middleware.GetDB()
		suiteID, err := middleware.GetCurrentSuiteID()
		if err != nil || suiteID == 0 {
			suiteID = 1
		}

		// First create the 5 profiles
		profileNames := []string{"Math", "Philosophy", "Programming", "Science", "Writing"}
		profileDescriptions := map[string]string{
			"Math":        "Mathematics and logic problems",
			"Philosophy":  "Philosophical and ethical questions",
			"Programming": "Programming and software development",
			"Science":     "Scientific and natural world questions",
			"Writing":     "Creative and technical writing",
		}
		profileIDs := make(map[string]int64)
		for _, name := range profileNames {
			result, err := db.Exec("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)",
				name, profileDescriptions[name], suiteID)
			if err != nil {
				log.Printf("Error inserting mock profile: %v", err)
				continue
			}
			profileID, err := result.LastInsertId()
			if err != nil {
				log.Printf("Error getting profile ID: %v", err)
				continue
			}
			profileIDs[name] = profileID
		}
		log.Printf("Created %d mock profiles", len(profileNames))

		// Define 10 prompts with solutions for each profile
		type PromptWithSolution struct {
			Text     string
			Solution string
		}
		profilePrompts := map[string][]PromptWithSolution{
			"Math": {
				{"What is 2 + 2?", "The answer is **4**."},
				{"Solve for x: 2x = 10", "Divide both sides by 2: x = **5**"},
				{"What is the derivative of x^2?", "Using the power rule: **2x**"},
				{"Calculate the area of a circle with radius 5", "Area = πr² = π × 5² = **25π** or approximately **78.54** square units"},
				{"What is the square root of 144?", "**12**"},
				{"If f(x) = 3x + 1, what is f(5)?", "f(5) = 3(5) + 1 = **16**"},
				{"What is the Pythagorean theorem?", "a² + b² = c², where c is the hypotenuse of a right triangle"},
				{"Simplify: (x + 2)(x - 3)", "x² - 3x + 2x - 6 = **x² - x - 6**"},
				{"What is 15% of 200?", "0.15 × 200 = **30**"},
				{"What is the sum of angles in a triangle?", "**180 degrees**"},
			},
			"Philosophy": {
				{"What is the meaning of life?", "A profound philosophical question with many perspectives including: finding purpose, creating meaning, seeking happiness, or serving others."},
				{"Explain Plato's allegory of the cave", "Plato's allegory describes prisoners chained in a cave seeing only shadows. One escapes to see reality (the sun), representing enlightenment and the philosopher's journey from ignorance to knowledge."},
				{"Is free will compatible with determinism?", "A central debate in philosophy. Compatibilists argue free will and determinism can coexist, while incompatibilists believe they cannot."},
				{"What is ethics?", "The branch of philosophy studying morality, including principles of right and wrong conduct, moral values, and how we should live."},
				{"Describe utilitarianism", "An ethical theory stating the best action maximizes overall happiness or utility - 'the greatest good for the greatest number.'"},
				{"What is consciousness?", "The subjective experience of awareness, qualia, and self-reflective thought - one of philosophy's 'hard problems.'"},
				{"Does objective morality exist?", "Moral realists argue yes; moral relativists disagree. This debate examines whether moral truths are independent of human opinion."},
				{"Explain the trolley problem", "A thought experiment: do you pull a lever to kill one person and save five? Explores utilitarianism vs. deontological ethics."},
				{"What is epistemology?", "The philosophical study of knowledge - its nature, origin, limits, and justification."},
				{"Can we truly know anything?", "Skeptical questioning that challenges certainty. Responses range from radical skepticism to pragmatic acceptance of justified true belief."},
			},
			"Programming": {
				{"Write a function to reverse a string", "```python\ndef reverse_string(s):\n    return s[::-1]\n```"},
				{"What is the time complexity of binary search?", "**O(log n)** - the search space is halved each iteration."},
				{"Explain recursion", "A function that calls itself to solve smaller instances of the same problem. Requires base case(s) and recursive case(s)."},
				{"What is a closure in JavaScript?", "A function bundled with its lexical environment. Closures remember variables from their outer scope even after the outer function returns."},
				{"Write a function to check if a number is prime", "```python\ndef is_prime(n):\n    if n < 2:\n        return False\n    for i in range(2, int(n**0.5) + 1):\n        if n % i == 0:\n            return False\n    return True\n```"},
				{"What is the difference between == and ===?", "In JavaScript, `==` checks equality with type coercion, while `===` checks strict equality without type coercion."},
				{"Explain the concept of Big O notation", "A mathematical notation describing algorithm efficiency as input size grows, focusing on worst-case time and space complexity."},
				{"What is a race condition?", "A bug where output depends on the timing of uncontrollable events, often occurring in concurrent programming when multiple threads access shared data."},
				{"Write a function to merge two sorted arrays", "```python\ndef merge_sorted_arrays(arr1, arr2):\n    result = []\n    i = j = 0\n    while i < len(arr1) and j < len(arr2):\n        if arr1[i] <= arr2[j]:\n            result.append(arr1[i])\n            i += 1\n        else:\n            result.append(arr2[j])\n            j += 1\n    result.extend(arr1[i:])\n    result.extend(arr2[j:])\n    return result\n```"},
				{"What is dependency injection?", "A design pattern where dependencies are provided to a class rather than created within it, improving testability and loose coupling."},
			},
			"Science": {
				{"What is photosynthesis?", "The process by which plants convert light energy (sunlight), CO₂, and water into glucose and oxygen. Equation: 6CO₂ + 6H₂O + light → C₆H₁₂O₆ + 6O₂"},
				{"Explain the theory of evolution", "The scientific theory that species change over time through natural selection, where organisms with advantageous traits are more likely to survive and reproduce."},
				{"What is the speed of light?", "**299,792,458 meters per second** in a vacuum, denoted as **c**."},
				{"Describe the structure of an atom", "An atom consists of a nucleus containing protons and neutrons, surrounded by electrons in electron shells/orbitals."},
				{"What is Newton's first law of motion?", "An object at rest stays at rest, and an object in motion stays in motion at constant velocity, unless acted upon by an external force (law of inertia)."},
				{"Explain the water cycle", "The continuous cycle of water evaporation, condensation (forming clouds), precipitation (rain/snow), and collection (oceans, lakes, groundwater)."},
				{"What is DNA?", "Deoxyribonucleic acid - the molecule carrying genetic instructions. A double helix of nucleotides (A, T, C, G) that encodes genetic information."},
				{"Describe the process of mitosis", "Cell division producing two genetically identical daughter cells. Phases: prophase, metaphase, anaphase, telophase, followed by cytokinesis."},
				{"What is the greenhouse effect?", "Gases in Earth's atmosphere trap heat from the sun, warming the planet. Key greenhouse gases include CO₂, methane, and water vapor."},
				{"Explain the concept of entropy", "A measure of disorder or randomness in a system. The second law of thermodynamics states entropy in an isolated system always increases."},
			},
			"Writing": {
				{"Write a haiku about nature", "Morning dew glistens,\nLeaves dance in gentle breeze,\nLife awakes anew."},
				{"Describe your perfect day", "Waking to golden sunlight, a warm cup of coffee, meaningful conversations with loved ones, time for creativity, and ending with gratitude under starlight."},
				{"Write a short story about adventure", "The ancient map crinkled in her hands. X marked the hidden temple. She stepped into the jungle, heart racing, ready for whatever lay beyond the veil of leaves."},
				{"What makes a good character?", "Depth, flaws, clear motivation, growth arc, authentic voice, and relatable struggles that resonate with readers."},
				{"Write a persuasive paragraph about climate change", "Climate change is an urgent crisis demanding immediate action. Rising temperatures, extreme weather, and ecosystem collapse threaten our future. We must transition to renewable energy, reduce emissions, and protect our planet for generations to come."},
				{"Describe the taste of chocolate", "Rich, velvety sweetness melting across the tongue - hints of vanilla, earthy cocoa, and a lingering embrace of comfort."},
				{"Write a metaphor for time", "Time is a river - always flowing, never stopping, carving memories into the canyon of our lives."},
				{"What is the difference between fiction and non-fiction?", "Fiction presents invented stories, characters, and events. Non-fiction presents factual information about real people, places, and events."},
				{"Write a dialogue between two strangers", "\\\"Excuse me, is this seat taken?\\\" \\\"No, please.\\\" \\\"Thanks. Long day?\\\" \\\"The longest. You?\\\" \\\"Same. At least we're in this together.\\\" A small smile, shared understanding."},
				{"Describe the feeling of nostalgia", "A bittersweet ache - warm memories gilded by time's golden filter, yet tinged with longing for moments that can never return."},
			},
		}

		// Create all prompts with their associated profile_id
		displayOrder := 0
		for _, profileName := range profileNames {
			profileID, exists := profileIDs[profileName]
			if !exists {
				log.Printf("Profile ID not found for %s", profileName)
				continue
			}

			promptsForProfile := profilePrompts[profileName]
			for _, prompt := range promptsForProfile {
				_, err = db.Exec("INSERT INTO prompts (text, solution, suite_id, display_order, type, profile_id) VALUES (?, ?, ?, ?, 'objective', ?)",
					prompt.Text, prompt.Solution, suiteID, displayOrder, profileID)
				if err != nil {
					log.Printf("Error inserting mock prompt for profile %s: %v", profileName, err)
				}
				displayOrder++
			}
		}

		prompts = h.DataStore.ReadPrompts()
		log.Printf("Created %d mock prompts with profiles", len(prompts))

		// Create a second suite with 4 profiles and 5 prompts each
		secondSuiteName := "Alternative Suite"
		result, err := db.Exec("INSERT INTO suites (name, is_current) VALUES (?, 0)", secondSuiteName)
		if err != nil {
			log.Printf("Error creating second suite: %v", err)
		} else {
			secondSuiteID, err := result.LastInsertId()
			if err != nil {
				log.Printf("Error getting second suite ID: %v", err)
			} else {
				log.Printf("Created second suite: %s (ID: %d)", secondSuiteName, secondSuiteID)

				// Create 4 profiles for the second suite
				secondSuiteProfileNames := []string{"History", "Geography", "Literature", "Art"}
				secondSuiteProfileDescriptions := map[string]string{
					"History":    "Historical events and figures",
					"Geography":  "Geographical and environmental topics",
					"Literature": "Literary analysis and creative writing",
					"Art":        "Visual arts, music, and aesthetics",
				}
				secondSuiteProfileIDs := make(map[string]int64)
				for _, name := range secondSuiteProfileNames {
					result, err := db.Exec("INSERT INTO profiles (name, description, suite_id) VALUES (?, ?, ?)",
						name, secondSuiteProfileDescriptions[name], secondSuiteID)
					if err != nil {
						log.Printf("Error inserting second suite profile: %v", err)
						continue
					}
					profileID, err := result.LastInsertId()
					if err != nil {
						log.Printf("Error getting second suite profile ID: %v", err)
						continue
					}
					secondSuiteProfileIDs[name] = profileID
				}
				log.Printf("Created %d mock profiles for second suite", len(secondSuiteProfileNames))

				// Define 5 prompts with solutions for each profile in the second suite
				secondSuitePrompts := map[string][]PromptWithSolution{
					"History": {
						{"When did World War II end?", "World War II ended in **1945**, with Germany surrendering in May and Japan in September."},
						{"Who was the first President of the United States?", "**George Washington** served as the first U.S. President from 1789 to 1797."},
						{"What was the Renaissance?", "A cultural movement (14th-17th century) marking the transition from medieval to modern times, characterized by renewed interest in classical art, literature, and learning."},
						{"Explain the causes of the French Revolution", "Key causes included: social inequality (three estates), financial crisis from wars, Enlightenment ideas, food shortages, and resentment toward the monarchy."},
						{"What was the Silk Road?", "A network of trade routes connecting East Asia to the Mediterranean, facilitating exchange of goods, ideas, and cultures from around 130 BCE to 1453 CE."},
					},
					"Geography": {
						{"What is the capital of Australia?", "**Canberra** is the capital of Australia, not Sydney or Melbourne as commonly assumed."},
						{"What is the longest river in the world?", "The **Nile River** in Africa is generally considered the longest at approximately 6,650 km (4,130 miles)."},
						{"Explain the water cycle", "The continuous movement of water: evaporation from surfaces, condensation into clouds, precipitation as rain/snow, and collection in bodies of water."},
						{"What are the seven continents?", "Africa, Antarctica, Asia, Australia/Oceania, Europe, North America, and South America."},
						{"What is a tectonic plate?", "A large, rigid slab of Earth's lithosphere that moves and interacts with other plates, causing earthquakes, volcanic activity, and mountain formation."},
					},
					"Literature": {
						{"Who wrote 'Romeo and Juliet'?", "**William Shakespeare** wrote this tragic play around 1595-1597."},
						{"What is a haiku?", "A Japanese poetic form with three lines and a 5-7-5 syllable structure, traditionally focusing on nature and seasonal imagery."},
						{"Explain the concept of foreshadowing", "A literary device where hints or clues suggest future events, building anticipation and creating dramatic tension."},
						{"Who wrote '1984'?", "**George Orwell** published this dystopian novel in 1949, exploring themes of totalitarianism and surveillance."},
						{"What is magical realism?", "A literary genre where magical elements blend realistically into ordinary settings, common in Latin American literature (e.g., Gabriel García Márquez)."},
					},
					"Art": {
						{"Who painted the Mona Lisa?", "**Leonardo da Vinci** painted this masterpiece between 1503 and 1519."},
						{"What is Impressionism?", "A 19th-century art movement characterized by visible brush strokes, emphasis on light, and ordinary subject matter (e.g., Monet, Renoir)."},
						{"Explain the golden ratio in art", "A mathematical proportion (approximately 1.618) believed to create aesthetically pleasing compositions, used by artists like da Vinci and architects throughout history."},
						{"Who sculpted David?", "**Michelangelo** carved this marble statue of the biblical hero between 1501 and 1504."},
						{"What is abstract art?", "Art that doesn't represent visual reality accurately, using colors, forms, and gestures to achieve its effect (e.g., Kandinsky, Pollock)."},
					},
				}

				// Create all prompts for the second suite with their associated profile_id
				displayOrder := 0
				for _, profileName := range secondSuiteProfileNames {
					profileID, exists := secondSuiteProfileIDs[profileName]
					if !exists {
						log.Printf("Profile ID not found for %s in second suite", profileName)
						continue
					}

					promptsForProfile := secondSuitePrompts[profileName]
					for _, prompt := range promptsForProfile {
						_, err = db.Exec("INSERT INTO prompts (text, solution, suite_id, display_order, type, profile_id) VALUES (?, ?, ?, ?, 'objective', ?)",
							prompt.Text, prompt.Solution, secondSuiteID, displayOrder, profileID)
						if err != nil {
							log.Printf("Error inserting second suite prompt for profile %s: %v", profileName, err)
						}
						displayOrder++
					}
				}
				log.Printf("Created %d mock prompts for second suite", displayOrder)

				// Create mock models for the second suite (same tiers as first suite)
				secondSuiteTiers := []string{
					"Cosmic", "Transcendent", "Ethereal", "Celestial", "Infinite",
					"Quantum", "Nebular", "Stellar", "Galactic", "Universal", "Dimensional",
				}
				var secondSuiteModels []string
				for i := 0; i < 12; i++ {
					tier := secondSuiteTiers[i%len(secondSuiteTiers)]
					num := i/len(secondSuiteTiers) + 1
					modelName := tier + "-" + strconv.Itoa(num)
					_, err = db.Exec("INSERT INTO models (name, suite_id) VALUES (?, ?)",
						modelName, secondSuiteID)
					if err != nil {
						log.Printf("Error inserting second suite model: %v", err)
						continue
					}
					secondSuiteModels = append(secondSuiteModels, modelName)
				}
				log.Printf("Created %d mock models for second suite", len(secondSuiteModels))

				// Create mock scores for the second suite models
				// Get all prompts for the second suite
				promptRows, err := db.Query("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order", secondSuiteID)
				if err != nil {
					log.Printf("Error querying second suite prompts: %v", err)
				} else {
					defer func() {
						if err := promptRows.Close(); err != nil {
							log.Printf("Error closing prompt rows: %v", err)
						}
					}()
					var promptIDs []int
					for promptRows.Next() {
						var promptID int
						if err := promptRows.Scan(&promptID); err != nil {
							log.Printf("Error scanning prompt ID: %v", err)
							continue
						}
						promptIDs = append(promptIDs, promptID)
					}

					// Get model IDs and create scores
					for _, modelName := range secondSuiteModels {
						var modelID int
						err = db.QueryRow("SELECT id FROM models WHERE name = ? AND suite_id = ?", modelName, secondSuiteID).Scan(&modelID)
						if err != nil {
							log.Printf("Error getting model ID for %s: %v", modelName, err)
							continue
						}

						// Create scores with tier-based distribution (similar to first suite)
						tierIndex := (len(secondSuiteModels) - 1) / len(secondSuiteTiers)
						for _, promptID := range promptIDs {
							score := getRandomScoreForTierWrapper(tierIndex)
							_, err = db.Exec("INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?)",
								modelID, promptID, score)
							if err != nil {
								log.Printf("Error inserting score for model %s prompt %d: %v", modelName, promptID, err)
							}
						}
					}
					log.Printf("Created mock scores for second suite models")
				}
			}
		}
	}

	// Get all model names
	models := mockData.Models
	if len(models) == 0 {
		// If no models passed, use models from existing results
		for model := range mockData.Results {
			models = append(models, model)
		}
	}

	// Use the client's results directly
	results := mockData.Results

	// Generate mock models if both models and results are empty
	if len(models) == 0 && len(results) == 0 {
		results = make(map[string]middleware.Result)
		tiers := []string{
			"Cosmic", "Transcendent", "Ethereal", "Celestial", "Infinite",
			"Quantum", "Nebular", "Stellar", "Galactic", "Universal", "Dimensional",
		}
		for i := 0; i < 24; i++ {
			tier := tiers[i%len(tiers)]
			num := i/len(tiers) + 1
			modelName := tier + "-" + strconv.Itoa(num)
			models = append(models, modelName)
			results[modelName] = middleware.Result{Scores: make([]int, len(prompts))}
		}
	}

	// Validate that all scores are legitimate values: 0, 20, 40, 60, 80, 100
	for model, result := range results {
		for i, score := range result.Scores {
			// Only allow valid score values
			switch score {
			case 0, 20, 40, 60, 80, 100:
				// Valid score, keep it
			default:
				// Invalid score, set to 0
				log.Printf("Correcting invalid score %d for model %s prompt %d", score, model, i)
				result.Scores[i] = 0
			}
		}
		results[model] = result
	}

	// Skip the evenly distributed tier generation since we're using client scores

	// Save the evenly distributed mock results
	suiteName := h.DataStore.GetCurrentSuiteName()
	err = h.DataStore.WriteResults(suiteName, results)
	if err != nil {
		log.Printf("Error writing mock results: %v", err)
		http.Error(w, "Error saving mock results", http.StatusInternalServerError)
		return
	}

	// Generate mock responses for each model and prompt combination
	// Get database for inserting mock responses
	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		suiteID = 1 // fallback to default suite
	}

	// Get all prompts for response generation
	for _, modelName := range models {
		// Get model ID
		var modelID int
		err = db.QueryRow("SELECT id FROM models WHERE name = ? AND suite_id = ?", modelName, suiteID).Scan(&modelID)
		if err != nil {
			continue // model might not exist yet
		}

		// Generate mock response for each prompt
		for promptIdx := range prompts {
			var promptID int
			err = db.QueryRow("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order LIMIT 1 OFFSET ?",
				suiteID, promptIdx).Scan(&promptID)
			if err != nil {
				continue
			}

			// Generate Lorem ipsum mock response
			loremPhrases := []string{
				"Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
				"Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
				"Ut enim ad minim veniam, quis nostrud exercitation ullamco.",
				"Duis aute irure dolor in reprehenderit in voluptate velit esse.",
				"Excepteur sint occaecat cupidatat non proident sunt in culpa.",
			}

			// Build 3-5 random sentences
			numSentences := 3 + rand.Intn(3)
			var responseParts []string
			for i := 0; i < numSentences; i++ {
				responseParts = append(responseParts, loremPhrases[rand.Intn(len(loremPhrases))])
			}
			mockResponse := strings.Join(responseParts, " ")

			// Insert or update mock response
			_, err = db.Exec(
				"INSERT INTO model_responses (model_id, prompt_id, response_text, response_source) "+
					"VALUES (?, ?, ?, 'mock') "+
					"ON CONFLICT(model_id, prompt_id) DO UPDATE SET "+
					"response_text = excluded.response_text, response_source = 'mock', updated_at = CURRENT_TIMESTAMP",
				modelID, promptID, mockResponse)
			if err != nil {
				log.Printf("Error inserting mock response for model %s prompt %d: %v", modelName, promptIdx, err)
			}
		}
	}

	// Broadcast the updated results to all connected clients
	h.DataStore.BroadcastResults()

	// Calculate totalScores and passPercentages for the response
	totalScores := make(map[string]int)
	passPercentages := make(map[string]float64)

	log.Println("Calculating total scores for each model:")
	for model, result := range results {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		totalScores[model] = totalScore
		// Avoid division by zero when there are no prompts
		if len(prompts) > 0 {
			passPercentages[model] = float64(totalScore) / float64(len(prompts)*100) * 100
		} else {
			passPercentages[model] = 0
		}

		log.Printf("Model %s: total score = %d, pass percentage = %.2f%%",
			model, totalScore, passPercentages[model])
	}

	// Sort models by total score in descending order
	sort.Slice(models, func(i, j int) bool {
		return totalScores[models[i]] > totalScores[models[j]]
	})

	log.Printf("Sorted models after mock generation: %v", models[:min(5, len(models))])

	// Return success response with the generated data
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":          "success",
		"results":         results,
		"models":          models, // Now sorted by score
		"totalScores":     totalScores,
		"passPercentages": passPercentages,
	}

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding response: %v", err)
	}

	log.Println("Mock results with even tier distribution updated successfully")
}

// getRandomScoreForTierWrapper wraps the score generation for the second suite
func getRandomScoreForTierWrapper(tierIndex int) int {
	// Use the same logic as getRandomScoreForTier but accessible outside UpdateMockResults
	tierWeights := []map[int]int{
		{0: 1, 20: 1, 40: 8, 60: 15, 80: 25, 100: 50},   // cosmic (highest tier)
		{0: 1, 20: 2, 40: 10, 60: 20, 80: 40, 100: 27},  // divine
		{0: 2, 20: 5, 40: 15, 60: 30, 80: 35, 100: 13},  // celestial
		{0: 5, 20: 10, 40: 25, 60: 30, 80: 20, 100: 10}, // ascendant
		{0: 7, 20: 15, 40: 33, 60: 25, 80: 15, 100: 5},  // ethereal
		{0: 10, 20: 20, 40: 35, 60: 20, 80: 10, 100: 5}, // mystic
		{0: 15, 20: 30, 40: 30, 60: 15, 80: 8, 100: 2},  // astral
		{0: 20, 20: 35, 40: 25, 60: 15, 80: 4, 100: 1},  // spiritual
		{0: 30, 20: 35, 40: 20, 60: 12, 80: 2, 100: 1},  // primal
		{0: 40, 20: 35, 40: 15, 60: 8, 80: 2, 100: 0},   // mortal
		{0: 55, 20: 30, 40: 10, 60: 5, 80: 0, 100: 0},   // primordial (lowest tier)
	}

	if tierIndex >= len(tierWeights) {
		tierIndex = len(tierWeights) - 1
	}

	weightsMap := tierWeights[tierIndex]
	weightValues := []int{0, 20, 40, 60, 80, 100}
	weights := []int{
		weightsMap[0],
		weightsMap[20],
		weightsMap[40],
		weightsMap[60],
		weightsMap[80],
		weightsMap[100],
	}

	// Simple weighted random selection
	totalWeight := 0
	for _, w := range weights {
		totalWeight += w
	}

	random := rand.Intn(totalWeight)
	runningTotal := 0
	for i, w := range weights {
		runningTotal += w
		if random < runningTotal {
			return weightValues[i]
		}
	}

	return 0 // fallback
}

// RandomizeScoresHandler handles randomizing scores (backward compatible wrapper)
func RandomizeScoresHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.RandomizeScores(w, r)
}

// RandomizeScores randomizes existing scores in the database without creating new models or prompts
func (h *Handler) RandomizeScores(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling randomize scores")

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	db := middleware.GetDB()
	suiteID, err := middleware.GetCurrentSuiteID()
	if err != nil {
		log.Printf("Error getting suite ID: %v", err)
		http.Error(w, "Error getting suite ID", http.StatusInternalServerError)
		return
	}

	modelRows, err := db.Query("SELECT id, name FROM models WHERE suite_id = ?", suiteID)
	if err != nil {
		log.Printf("Error querying models: %v", err)
		http.Error(w, "Error querying models", http.StatusInternalServerError)
		return
	}
	defer func() { _ = modelRows.Close() }()

	var models []struct {
		ID   int
		Name string
	}

	for modelRows.Next() {
		var m struct {
			ID   int
			Name string
		}
		if err := modelRows.Scan(&m.ID, &m.Name); err != nil {
			continue
		}
		models = append(models, m)
	}

	promptRows, err := db.Query("SELECT id FROM prompts WHERE suite_id = ? ORDER BY display_order", suiteID)
	if err != nil {
		http.Error(w, "Error querying prompts", http.StatusInternalServerError)
		return
	}
	defer func() { _ = promptRows.Close() }()

	var promptIDs []int
	for promptRows.Next() {
		var id int
		if err := promptRows.Scan(&id); err != nil {
			continue
		}
		promptIDs = append(promptIDs, id)
	}

	rng := initRand()
	numPrompts := len(promptIDs)
	maxScore := numPrompts * 100
	numTiers := 12

	for i, model := range models {
		tierIndex := (i * numTiers) / len(models)
		if tierIndex >= numTiers {
			tierIndex = numTiers - 1
		}

		tierMinPercent := float64(tierIndex) / float64(numTiers)
		tierMaxPercent := float64(tierIndex+1) / float64(numTiers)
		tierMidPercent := (tierMinPercent + tierMaxPercent) / 2
		targetTotal := int(float64(maxScore) * tierMidPercent)

		remaining := targetTotal
		for j, promptID := range promptIDs {
			var score int
			if j == len(promptIDs)-1 {
				score = clampToValidScore(remaining)
			} else {
				promptsRemaining := len(promptIDs) - j - 1
				minScore := max(0, remaining-promptsRemaining*100)
				maxScoreVal := min(100, remaining)

				midScore := (minScore + maxScoreVal) / 2
				randomOffset := rng.Intn(21) - 10
				score = clampToValidScore(midScore + randomOffset)

				if score > maxScoreVal {
					score = clampToValidScore(maxScoreVal)
				}
				if score < minScore {
					score = clampToValidScore(minScore)
				}
			}

			_, _ = db.Exec(
				"INSERT INTO scores (model_id, prompt_id, score) VALUES (?, ?, ?) "+
					"ON CONFLICT(model_id, prompt_id) DO UPDATE SET score = excluded.score",
				model.ID, promptID, score)
			remaining -= score
		}
	}

	h.DataStore.BroadcastResults()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func clampToValidScore(score int) int {
	minDiff := 1000
	result := 0
	for _, vs := range []int{0, 20, 40, 60, 80, 100} {
		diff := score - vs
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			result = vs
		}
	}
	return result
}
