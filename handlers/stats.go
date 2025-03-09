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
		"tier-1":  {},
		"tier-2":  {},
		"tier-3":  {},
		"tier-4":  {},
		"tier-5":  {},
		"tier-6":  {},
		"tier-7":  {},
		"tier-8":  {},
		"tier-9":  {},
		"tier-10": {},
		"tier-11": {},
	}

	tierRanges := map[string]string{
		"tier-1":  "3000+",
		"tier-2":  "2700-2999",
		"tier-3":  "2400-2699",
		"tier-4":  "2100-2399",
		"tier-5":  "1800-2099",
		"tier-6":  "1500-1799",
		"tier-7":  "1200-1499",
		"tier-8":  "900-1199",
		"tier-9":  "600-899",
		"tier-10": "300-599",
		"tier-11": "0-299",
	}

	for model, score := range totalScores {
		switch {
		case score >= 3000:
			tiers["tier-1"] = append(tiers["tier-1"], model)
		case score >= 2700:
			tiers["tier-2"] = append(tiers["tier-2"], model)
		case score >= 2400:
			tiers["tier-3"] = append(tiers["tier-3"], model)
		case score >= 2100:
			tiers["tier-4"] = append(tiers["tier-4"], model)
		case score >= 1800:
			tiers["tier-5"] = append(tiers["tier-5"], model)
		case score >= 1500:
			tiers["tier-6"] = append(tiers["tier-6"], model)
		case score >= 1200:
			tiers["tier-7"] = append(tiers["tier-7"], model)
		case score >= 900:
			tiers["tier-8"] = append(tiers["tier-8"], model)
		case score >= 600:
			tiers["tier-9"] = append(tiers["tier-9"], model)
		case score >= 300:
			tiers["tier-10"] = append(tiers["tier-10"], model)
		default:
			tiers["tier-11"] = append(tiers["tier-11"], model)
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
			"tier-1",
			"tier-2",
			"tier-3",
			"tier-4",
			"tier-5",
			"tier-6",
			"tier-7",
			"tier-8",
			"tier-9",
			"tier-10",
			"tier-11",
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
