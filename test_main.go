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
	data, _ := os.ReadFile("data/prompts.txt")
	prompts := strings.Split(string(data), "\n")
	if len(prompts) == 0 || prompts[len(prompts)-1] != newPrompt {
		t.Errorf("Expected prompt '%s' to be added to the file, got '%v'", newPrompt, prompts)
	}

	// Clean up the test file
	os.WriteFile("data/prompts.txt", []byte(""), 0644)
}

func TestEditPromptHandler(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Create a test prompt
    initialPrompt := "Initial prompt"
    os.WriteFile("data/prompts.txt", []byte(initialPrompt), 0644)

    // Get the index of the prompt
    prompts := readPrompts()
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
    data, _ := os.ReadFile("data/prompts.txt")
    prompts = strings.Split(string(data), "\n")
    if len(prompts) == 0 || prompts[len(prompts)-1] != editedPrompt {
        t.Errorf("Expected prompt '%s' to be edited to '%s', got '%v'", initialPrompt, editedPrompt, prompts)
    }

    // Clean up the test file
    os.WriteFile("data/prompts.txt", []byte(""), 0644)
}

func TestUpdateResultHandler(t *testing.T) {
	// Set up a test server
	ts := httptest.NewServer(http.HandlerFunc(router))
	defer ts.Close()

	// Create a test prompt
	initialPrompt := "Test prompt"
	os.WriteFile("data/prompts.txt", []byte(initialPrompt), 0644)

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
	if len(results[model]) == 0 || results[model][promptIndex] != pass {
		t.Errorf("Expected result for model '%s' at index %d to be %t, got %v", model, promptIndex, pass, results[model])
	}

	// Clean up the test file
	os.WriteFile("data/prompts.txt", []byte(""), 0644)
	os.WriteFile("data/results.csv", []byte(""), 0644)
}

func TestDeletePromptHandler(t *testing.T) {
    // Set up a test server
    ts := httptest.NewServer(http.HandlerFunc(router))
    defer ts.Close()

    // Create a test prompt
    initialPrompt := "Initial prompt"
    os.WriteFile("data/prompts.txt", []byte(initialPrompt), 0644)

    // Get the index of the prompt
    prompts := readPrompts()
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
    data, _ := os.ReadFile("data/prompts.txt")
    prompts = strings.Split(string(data), "\n")
    if len(prompts) != 0 {
        t.Errorf("Expected prompt '%s' to be deleted, got '%v'", initialPrompt, prompts)
    }

    // Clean up the test file
    os.WriteFile("data/prompts.txt", []byte(""), 0644)
}
