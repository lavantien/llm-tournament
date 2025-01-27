package middleware

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

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
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
			UpdatePromptsOrder(message.Order)
		default:
			log.Printf("Unknown message type: %s", message.Type)
		}
	}
}

func BroadcastResults() {
    suiteName := GetCurrentSuiteName()
	prompts := ReadPrompts()
	results := ReadResults()

	modelTotalScores := make(map[string]int)
	for model, result := range results {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		modelTotalScores[model] = totalScore
	}

	models := make([]string, 0, len(results))
	for model := range results {
		models = append(models, model)
	}
	if len(models) == 0 {
		models = []string{"Model1", "Model2"} // Example fallback
	}
	sort.Slice(models, func(i, j int) bool {
		return modelTotalScores[models[i]] > modelTotalScores[models[j]]
	})

	payload := struct {
		Type string `json:"type"`
		Data struct {
			Results         map[string]Result `json:"results"`
			Models          []string          `json:"models"`
			TotalScores     map[string]int    `json:"totalScores"`
			PassPercentages map[string]float64 `json:"passPercentages"`
			Prompts         []string          `json:"prompts"`
			SuiteName       string            `json:"suiteName"`
		} `json:"data"`
	}{
		Type: "results",
		Data: struct {
			Results         map[string]Result `json:"results"`
			Models          []string          `json:"models"`
			TotalScores     map[string]int    `json:"totalScores"`
			PassPercentages map[string]float64 `json:"passPercentages"`
			Prompts         []string          `json:"prompts"`
			SuiteName       string            `json:"suiteName"`
		}{
			Results:         results,
			Models:          models,
			TotalScores:     modelTotalScores,
			PassPercentages: calculatePassPercentages(results, len(prompts)),
			Prompts:         promptsToStringArray(prompts),
			SuiteName:       suiteName,
		},
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


func promptsToStringArray(prompts []Prompt) []string {
	promptsTexts := make([]string, len(prompts))
	for i, prompt := range prompts {
		promptsTexts[i] = prompt.Text
	}
	return promptsTexts
}

func calculatePassPercentages(results map[string]Result, promptCount int) map[string]float64 {
	passPercentages := make(map[string]float64)
	for model, result := range results {
		totalScore := 0
		for _, score := range result.Scores {
			totalScore += score
		}
		passPercentages[model] = float64(totalScore) / float64(promptCount*100) * 100
	}
	return passPercentages
}
