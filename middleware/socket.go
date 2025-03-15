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

// ProfileGroup represents a group of prompts with the same profile
type ProfileGroup struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	StartCol int    `json:"startCol"` // Column index where this profile starts
	EndCol   int    `json:"endCol"`   // Column index where this profile ends
	Color    string `json:"color"`    // Generated color for this profile
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
	log.Println("BroadcastResults results:", results)

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

	// Group prompts by profile
	var orderedPrompts []struct {
		Index       int    `json:"index"`
		Text        string `json:"text"`
		ProfileID   string `json:"profileId"`
		ProfileName string `json:"profileName"`
	}

	// Get all profiles first (to include empty ones)
	profiles := ReadProfiles()

	// Get profile groups using the utility function
	profileGroups, profileMap := GetProfileGroups(prompts, profiles)

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
		noProfileGroup := &ProfileGroup{
			ID:    "none",
			Name:  "Uncategorized",
			Color: "hsl(0, 0%, 50%)",
		}
		profileGroups = append(profileGroups, noProfileGroup)
		profileMap[""] = noProfileGroup
	}

	// Process prompts and assign them to profile groups
	currentCol := 0
	for i, prompt := range prompts {
		profileName := prompt.Profile

		group, exists := profileMap[profileName]
		if !exists {
			// Skip if group doesn't exist and we don't have uncategorized group
			if _, hasUncategorized := profileMap[""]; !hasUncategorized {
				continue
			}
			group = profileMap[""]
		}

		if group.StartCol == -1 {
			group.StartCol = currentCol
		}
		group.EndCol = currentCol

		orderedPrompts = append(orderedPrompts, struct {
			Index       int    `json:"index"`
			Text        string `json:"text"`
			ProfileID   string `json:"profileId"`
			ProfileName string `json:"profileName"`
		}{
			Index:       i,
			Text:        prompt.Text,
			ProfileID:   group.ID,
			ProfileName: profileName,
		})

		currentCol++
	}

	// Log the data we're about to send
	log.Printf("Broadcasting data - Models: %v", models)

	payload := struct {
		Type string `json:"type"`
		Data struct {
			Results         map[string]Result  `json:"results"`
			Models          []string           `json:"models"`
			TotalScores     map[string]int     `json:"totalScores"`
			PassPercentages map[string]float64 `json:"passPercentages"`
			Prompts         []string           `json:"prompts"`
			SuiteName       string             `json:"suiteName"`
			ProfileGroups   []*ProfileGroup    `json:"profileGroups"`
			OrderedPrompts  interface{}        `json:"orderedPrompts"`
		} `json:"data"`
	}{
		Type: "results",
		Data: struct {
			Results         map[string]Result  `json:"results"`
			Models          []string           `json:"models"`
			TotalScores     map[string]int     `json:"totalScores"`
			PassPercentages map[string]float64 `json:"passPercentages"`
			Prompts         []string           `json:"prompts"`
			SuiteName       string             `json:"suiteName"`
			ProfileGroups   []*ProfileGroup    `json:"profileGroups"`
			OrderedPrompts  interface{}        `json:"orderedPrompts"`
		}{
			Results:         results,
			Models:          models,
			TotalScores:     modelTotalScores,
			PassPercentages: calculatePassPercentages(results, len(prompts)),
			Prompts:         promptsToStringArray(prompts),
			SuiteName:       suiteName,
			ProfileGroups:   profileGroups,
			OrderedPrompts:  orderedPrompts,
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
