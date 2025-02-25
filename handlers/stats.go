package handlers

import (
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"llm-tournament/middleware"
)

// Calculate tiers based on total scores
func calculateTiers(totalScores map[string]int) (map[string][]string, map[string]string) {
	tiers := map[string][]string{
		"transcendent":         {},
		"super-grandmaster":    {},
		"grandmaster":          {},
		"international-master": {},
		"master":               {},
		"expert":               {},
		"pro-player":           {},
		"advanced-player":      {},
		"intermediate-player":  {},
		"veteran":              {},
		"beginner":             {},
	}

	tierRanges := map[string]string{
		"transcendent":         "1900-2000",
		"super-grandmaster":    "1800-1899",
		"grandmaster":          "1700-1799",
		"international-master": "1600-1699",
		"master":               "1500-1599",
		"expert":               "1400-1499",
		"pro-player":           "1200-1399",
		"advanced-player":      "1000-1199",
		"intermediate-player":  "800-999",
		"veteran":              "600-799",
		"beginner":             "0-599",
	}

	for model, score := range totalScores {
		switch {
		case score >= 1900:
			tiers["transcendent"] = append(tiers["transcendent"], model)
		case score >= 1800:
			tiers["super-grandmaster"] = append(tiers["super-grandmaster"], model)
		case score >= 1700:
			tiers["grandmaster"] = append(tiers["grandmaster"], model)
		case score >= 1600:
			tiers["international-master"] = append(tiers["international-master"], model)
		case score >= 1500:
			tiers["master"] = append(tiers["master"], model)
		case score >= 1400:
			tiers["expert"] = append(tiers["expert"], model)
		case score >= 1200:
			tiers["pro-player"] = append(tiers["pro-player"], model)
		case score >= 1000:
			tiers["advanced-player"] = append(tiers["advanced-player"], model)
		case score >= 800:
			tiers["intermediate-player"] = append(tiers["intermediate-player"], model)
		case score >= 600:
			tiers["veteran"] = append(tiers["veteran"], model)
		default:
			tiers["beginner"] = append(tiers["beginner"], model)
		}
	}

	return tiers, tierRanges
}

// Handle stats page
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling stats page")
	results := middleware.ReadResults()

	// Calculate score breakdowns
	type ScoreStats struct {
		TotalScore int `json:"TotalScore"`
		Count20    int `json:"Count20"`
		Count50    int `json:"Count50"`
		Count100   int `json:"Count100"`
	}

	scoreStats := make(map[string]ScoreStats)
	for model, result := range results {
		stats := ScoreStats{}
		for _, score := range result.Scores {
			stats.TotalScore += score
			switch score {
			case 20:
				stats.Count20++
			case 50:
				stats.Count50++
			case 100:
				stats.Count100++
			}
		}
		scoreStats[model] = stats
	}

	// Create total scores map for tier calculations
	totalScores := make(map[string]int)
	for model, stats := range scoreStats {
		totalScores[model] = stats.TotalScore
	}

	// Calculate tiers
	tiers, tierRanges := calculateTiers(totalScores)

	// Prepare template data
	templateData := struct {
		PageName     string
		TotalScores  map[string]ScoreStats
		Tiers        map[string][]string
		TierRanges   map[string]string
		OrderedTiers []string
	}{
		PageName:    "Statistics",
		TotalScores: scoreStats,
		Tiers:       tiers,
		TierRanges:  tierRanges,
		OrderedTiers: []string{
			"transcendent",
			"super-grandmaster",
			"grandmaster",
			"international-master",
			"master",
			"expert",
			"pro-player",
			"advanced-player",
			"intermediate-player",
			"veteran",
			"beginner",
		},
	}

	// Parse and execute template
	t, err := template.New("stats.html").Funcs(template.FuncMap{
		"json": func(v interface{}) template.JS { // Updated to return template.JS
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
		"tierClass": func(tier string) string {
			return strings.ReplaceAll(tier, "-", "")
		},
		"formatTierName": func(tier string) string {
			return strings.Title(strings.ReplaceAll(tier, "-", " "))
		},
		"join": strings.Join,
	}).ParseFiles("templates/stats.html", "templates/nav.html")

	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, templateData)
	if err != nil {
		log.Printf("Template execution error: %v", err)
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Write the buffered template output to response
	_, err = buf.WriteTo(w)
	if err != nil {
		log.Printf("Error writing response: %v", err)
	}
}
