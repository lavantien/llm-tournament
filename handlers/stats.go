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
		"divine":               {},
		"legendary":            {},
		"mythical":             {},
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
		"divine":               "3000+",
		"legendary":            "2800-2999",
		"mythical":             "2600-2799",
		"transcendent":         "2400-2599",
		"super-grandmaster":    "2200-2399",
		"grandmaster":          "2000-2199",
		"international-master": "1800-1999",
		"master":               "1600-1799",
		"expert":               "1400-1599",
		"pro-player":           "1200-1399",
		"advanced-player":      "1000-1199",
		"intermediate-player":  "800-999",
		"veteran":              "600-799",
		"beginner":             "0-599",
	}

	for model, score := range totalScores {
		switch {
		case score >= 3000:
			tiers["divine"] = append(tiers["divine"], model)
		case score >= 2800:
			tiers["legendary"] = append(tiers["legendary"], model)
		case score >= 2600:
			tiers["mythical"] = append(tiers["mythical"], model)
		case score >= 2400:
			tiers["transcendent"] = append(tiers["transcendent"], model)
		case score >= 2200:
			tiers["super-grandmaster"] = append(tiers["super-grandmaster"], model)
		case score >= 2000:
			tiers["grandmaster"] = append(tiers["grandmaster"], model)
		case score >= 1800:
			tiers["international-master"] = append(tiers["international-master"], model)
		case score >= 1600:
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
		Count40    int `json:"Count40"`
		Count60    int `json:"Count60"`
		Count80    int `json:"Count80"`
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
			case 40:
				stats.Count40++
			case 60:
				stats.Count60++
			case 80:
				stats.Count80++
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
			"divine",
			"legendary",
			"mythical",
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
