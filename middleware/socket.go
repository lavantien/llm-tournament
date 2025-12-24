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
		_ = conn.Close()
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
			_ = client.Close()
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

// BroadcastEvaluationProgress broadcasts evaluation progress to all clients
func BroadcastEvaluationProgress(jobID, current, total int, cost float64) {
	payload := struct {
		Type string `json:"type"`
		Data struct {
			JobID   int     `json:"job_id"`
			Current int     `json:"current"`
			Total   int     `json:"total"`
			Cost    float64 `json:"cost"`
		} `json:"data"`
	}{
		Type: "evaluation_progress",
	}
	payload.Data.JobID = jobID
	payload.Data.Current = current
	payload.Data.Total = total
	payload.Data.Cost = cost

	broadcastMessage(payload)
}

// BroadcastEvaluationCompleted broadcasts evaluation completion
func BroadcastEvaluationCompleted(jobID int, finalCost float64) {
	payload := struct {
		Type string `json:"type"`
		Data struct {
			JobID     int     `json:"job_id"`
			FinalCost float64 `json:"final_cost"`
		} `json:"data"`
	}{
		Type: "evaluation_completed",
	}
	payload.Data.JobID = jobID
	payload.Data.FinalCost = finalCost

	broadcastMessage(payload)
	// Also refresh results
	BroadcastResults()
}

// BroadcastEvaluationFailed broadcasts evaluation failure
func BroadcastEvaluationFailed(jobID int, errorMsg string) {
	payload := struct {
		Type string `json:"type"`
		Data struct {
			JobID int    `json:"job_id"`
			Error string `json:"error"`
		} `json:"data"`
	}{
		Type: "evaluation_failed",
	}
	payload.Data.JobID = jobID
	payload.Data.Error = errorMsg

	broadcastMessage(payload)
}

// BroadcastCostAlert broadcasts cost threshold alert
func BroadcastCostAlert(suiteID int, currentCost, threshold float64) {
	payload := struct {
		Type string `json:"type"`
		Data struct {
			SuiteID     int     `json:"suite_id"`
			CurrentCost float64 `json:"current_cost"`
			Threshold   float64 `json:"threshold"`
		} `json:"data"`
	}{
		Type: "cost_alert",
	}
	payload.Data.SuiteID = suiteID
	payload.Data.CurrentCost = currentCost
	payload.Data.Threshold = threshold

	broadcastMessage(payload)
}

// broadcastMessage sends a JSON message to all connected clients
func broadcastMessage(payload interface{}) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for client := range clients {
		err := client.WriteJSON(payload)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}
