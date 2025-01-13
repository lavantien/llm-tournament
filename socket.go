package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"sync"

	"github.com/gorilla/websocket"
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
