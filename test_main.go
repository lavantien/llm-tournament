package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestAddPromptHandler(t *testing.T) {
	// Set up a test server
	ts := httptest.NewServer(http.HandlerFunc(router))
	defer ts.Close()

	// Create a new prompt
	newPrompt := "Test prompt"

	// Send a POST request to add the prompt
	resp, err := http.PostForm(ts.URL+"/add_prompt", url.Values{"prompt": {newPrompt}})
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
	}

	// Check if the prompt was added to the file
	prompts := readPrompts()
	if len(prompts) == 0 || prompts[len(prompts)-1].Text != newPrompt {
		t.Errorf("Expected prompt '%s' to be added to the file, got '%v'", newPrompt, prompts)
	}

	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
}

func TestPromptsCRUD(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Test Create (Add)
    newPrompt := "Test prompt"
    resp, err := http.PostForm(ts.URL+"/add_prompt", url.Values{"prompt": {newPrompt}})
    if err != nil {
        t.Fatalf("Failed to send POST request: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusSeeOther {
        t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
    }
    prompts := readPrompts()
    if len(prompts) == 0 || prompts[len(prompts)-1].Text != newPrompt {
        t.Errorf("Expected prompt '%s' to be added to the file, got '%v'", newPrompt, prompts)
    }

    // Test Read (List) - implicitly tested by the above and below tests

    // Test Update (Edit)
    index := 0
    editedPrompt := "Edited prompt"
    resp, err = http.PostForm(ts.URL+"/edit_prompt", url.Values{"index": {string(rune(index))}, "prompt": {editedPrompt}})
    if err != nil {
        t.Fatalf("Failed to send POST request: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusSeeOther {
        t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
    }
    prompts = readPrompts()
    if len(prompts) == 0 || prompts[index].Text != editedPrompt {
        t.Errorf("Expected prompt '%s' to be edited to '%s', got '%v'", newPrompt, editedPrompt, prompts)
    }

    // Test Delete
    resp, err = http.PostForm(ts.URL+"/delete_prompt", url.Values{"index": {string(rune(index))}})
    if err != nil {
        t.Fatalf("Failed to send POST request: %v", err)
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusSeeOther {
        t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
    }
    prompts = readPrompts()
    if len(prompts) != 0 {
        t.Errorf("Expected prompt '%s' to be deleted, got '%v'", editedPrompt, prompts)
    }

    // Clean up the test file
	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
}

func TestDeletePromptHandler(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Create a test prompt
	initialPrompt := "Initial prompt"
	prompts := []Prompt{{Text: initialPrompt}}
	writePrompts(prompts)

    // Get the index of the prompt
    index := 0

    // Send a GET request to delete the prompt
    req, err := http.NewRequest("GET", ts.URL+"/delete_prompt?index="+string(rune(index)), nil)
    if err != nil {
        t.Fatalf("Failed to create GET request: %v", err)
    }
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("Failed to send GET request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response status code
    if resp.StatusCode != http.StatusSeeOther {
        t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
    }

    // Check if the prompt was deleted from the file
	prompts = readPrompts()
    if len(prompts) != 0 {
        t.Errorf("Expected prompt '%s' to be deleted, got '%v'", initialPrompt, prompts)
    }

    // Clean up the test file
	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
}


func TestEditPromptHandler(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Create a test prompt
    initialPrompt := "Initial prompt"
	prompts := []Prompt{{Text: initialPrompt}}
	writePrompts(prompts)

    // Get the index of the prompt
    index := 0

    // Send a POST request to edit the prompt
    editedPrompt := "Edited prompt"
    resp, err := http.PostForm(ts.URL+"/edit_prompt", url.Values{"index": {string(rune(index))}, "prompt": {editedPrompt}})
    if err != nil {
        t.Fatalf("Failed to send POST request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response status code
    if resp.StatusCode != http.StatusSeeOther {
        t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
    }

    // Check if the prompt was edited in the file
	prompts = readPrompts()
    if len(prompts) == 0 || prompts[index].Text != editedPrompt {
        t.Errorf("Expected prompt '%s' to be edited to '%s', got '%v'", initialPrompt, editedPrompt, prompts)
    }

    // Clean up the test file
	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
}

func TestUpdateResultHandler(t *testing.T) {
	// Set up a test server
	ts := httptest.NewServer(http.HandlerFunc(router))
	defer ts.Close()

	// Create a test prompt
	initialPrompt := "Test prompt"
	prompts := []Prompt{{Text: initialPrompt}}
	writePrompts(prompts)

	// Create a test result
	model := "test_model"
	promptIndex := 0
	pass := true

	// Send a POST request to update the result
	resp, err := http.PostForm(ts.URL+"/update_result", url.Values{
		"model":       {model},
		"promptIndex": {string(rune(promptIndex))},
		"pass":        {strconv.FormatBool(pass)},
	})
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check if the result was updated in the file
	results := readResults()
	if _, ok := results[model]; !ok {
		t.Errorf("Expected model '%s' to be added to the results", model)
	}
	if len(results[model].Passes) == 0 || results[model].Passes[promptIndex] != pass {
		t.Errorf("Expected result for model '%s' at index %d to be %t, got %v", model, promptIndex, pass, results[model])
	}

	// Clean up the test file
	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
	os.WriteFile("data/results.json", []byte("{}"), 0644)
}

func TestExportPromptsHandler(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Create test prompts
    initialPrompts := []Prompt{{Text: "Prompt 1"}, {Text: "Prompt 2"}}
    writePrompts(initialPrompts)

    // Send a POST request to export the prompts
    resp, err := http.PostForm(ts.URL+"/export_prompts", nil)
    if err != nil {
        t.Fatalf("Failed to send POST request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response status code
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
    }

    // Check the content type
    if resp.Header.Get("Content-Type") != "text/csv" {
        t.Errorf("Expected Content-Type to be text/csv, got %s", resp.Header.Get("Content-Type"))
    }

    // Check the content disposition
    if resp.Header.Get("Content-Disposition") != "attachment;filename=prompts.csv" {
        t.Errorf("Expected Content-Disposition to be attachment;filename=prompts.csv, got %s", resp.Header.Get("Content-Disposition"))
    }

    // Clean up the test file
	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
}

func TestImportPromptsHandler(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Create a test CSV file
    csvData := `Prompt
Prompt 1
Prompt 2
`
    // Send a POST request to import the prompts
    resp, err := http.PostForm(ts.URL+"/import_prompts", url.Values{
        "prompts_file": {csvData},
    })
    if err != nil {
        t.Fatalf("Failed to send POST request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response status code
    if resp.StatusCode != http.StatusSeeOther {
        t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
    }

    // Check if the prompts were imported correctly
    prompts := readPrompts()
    if len(prompts) != 2 {
        t.Errorf("Expected 2 prompts to be imported, got %d", len(prompts))
    }
    if prompts[0].Text != "Prompt 1" || prompts[1].Text != "Prompt 2" {
        t.Errorf("Expected prompts to be [Prompt 1, Prompt 2], got %v", prompts)
    }

    // Clean up the test file
	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
}

func TestImportResultsHandler(t *testing.T) {
	// Set up a test server
	ts := httptest.NewServer(http.HandlerFunc(router))
	defer ts.Close()

	// Create a test CSV file
	csvData := `Model,Prompt 1,Prompt 2
Model A,true,false
Model B,false,true
`
	// Send a POST request to import the results
	resp, err := http.PostForm(ts.URL+"/import_results", url.Values{
		"results_file": {csvData},
	})
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
	}

	// Check if the results were imported correctly
	results := readResults()
	if len(results) != 2 {
		t.Errorf("Expected 2 models to be imported, got %d", len(results))
	}
	if results["Model A"].Passes[0] != true || results["Model A"].Passes[1] != false {
		t.Errorf("Expected Model A results to be [true, false], got %v", results["Model A"].Passes)
	}
	if results["Model B"].Passes[0] != false || results["Model B"].Passes[1] != true {
		t.Errorf("Expected Model B results to be [false, true], got %v", results["Model B"].Passes)
	}

	// Clean up the test file
	os.WriteFile("data/results.json", []byte("{}"), 0644)
}

func TestAddModelHandler(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Create a new model
    newModel := "Test Model"

    // Send a POST request to add the model
    resp, err := http.PostForm(ts.URL+"/add_model", url.Values{"model": {newModel}})
    if err != nil {
        t.Fatalf("Failed to send POST request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response status code
    if resp.StatusCode != http.StatusSeeOther {
        t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
    }

    // Check if the model was added to the file
    results := readResults()
    if _, ok := results[newModel]; !ok {
        t.Errorf("Expected model '%s' to be added to the results", newModel)
    }

    // Clean up the test file
	os.WriteFile("data/results.json", []byte("{}"), 0644)
}

func TestResultsAdjustsToNewPrompts(t *testing.T) {
	// Set up a test server
	ts := httptest.NewServer(http.HandlerFunc(router))
	defer ts.Close()

	// Create initial prompts
	initialPrompts := []Prompt{{Text: "Prompt 1"}, {Text: "Prompt 2"}}
	writePrompts(initialPrompts)

	// Create initial results
	results := map[string]Result{
		"Model A": {Passes: []bool{true, false}},
	}
	writeResults(results)

	// Add a new prompt
	newPrompt := "Prompt 3"
	resp, err := http.PostForm(ts.URL+"/add_prompt", url.Values{"prompt": {newPrompt}})
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response status code
	if resp.StatusCode != http.StatusSeeOther {
		t.Errorf("Expected status %d, got %d", http.StatusSeeOther, resp.StatusCode)
	}

	// Check if the results were adjusted correctly
	updatedResults := readResults()
	if len(updatedResults["Model A"].Passes) != 3 {
		t.Errorf("Expected 3 passes for Model A, got %d", len(updatedResults["Model A"].Passes))
	}

	// Clean up the test files
	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
	os.WriteFile("data/results.json", []byte("{}"), 0644)
}

func TestResultsHandlerSorting(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Create test prompts
	initialPrompts := []Prompt{{Text: "Prompt 1"}, {Text: "Prompt 2"}}
	writePrompts(initialPrompts)

    // Create test results
    results := map[string]Result{
        "Model A": {Passes: []bool{true, false}},
        "Model B": {Passes: []bool{true, true}},
        "Model C": {Passes: []bool{false, false}},
    }
	writeResults(results)

    // Send a GET request to the results page
    resp, err := http.Get(ts.URL + "/results")
    if err != nil {
        t.Fatalf("Failed to send GET request: %v", err)
    }
    defer resp.Body.Close()

    // Check the response status code
    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status %d, got %d", http.StatusOK, resp.StatusCode)
    }

    // Check if the models are sorted correctly
    expectedOrder := []string{"Model B", "Model A", "Model C"}
    actualOrder := getModelsOrderFromHTML(resp)

    if len(actualOrder) != len(expectedOrder) {
        t.Fatalf("Expected %d models, got %d", len(expectedOrder), len(actualOrder))
    }

    for i, model := range actualOrder {
        if model != expectedOrder[i] {
            t.Errorf("Expected model at index %d to be '%s', got '%s'", i, expectedOrder[i], model)
        }
    }

    // Clean up the test files
	os.WriteFile("data/prompts.json", []byte("[]"), 0644)
	os.WriteFile("data/results.json", []byte("{}"), 0644)
}

func getModelsOrderFromHTML(resp *http.Response) []string {
    // Parse the HTML response to extract the model order
    // This is a simplified implementation and might need adjustments based on the actual HTML structure
    // In a real application, you might want to use a proper HTML parsing library
    var models []string
    data := make([]byte, 1024)
    n, _ := resp.Body.Read(data)
    html := string(data[:n])
    lines := strings.Split(html, "\n")
    for _, line := range lines {
        if strings.Contains(line, "<td>") && !strings.Contains(line, "<input") {
            model := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(line, "<td>", ""), "</td>", ""))
            if model != "Model" && model != "Total" && model != "" {
                models = append(models, model)
            }
        }
    }
    return models
}
