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
		"transcendental": {},
		"cosmic":         {},
		"divine":         {},
		"celestial":      {},
		"ascendant":      {},
		"ethereal":       {},
		"mystic":         {},
		"astral":         {},
		"spiritual":      {},
		"primal":         {},
		"mortal":         {},
		"primordial":     {},
	}

	tierRanges := map[string]string{
		"transcendental": "3780+",
		"cosmic":         "3360-3779",
		"divine":         "2700-3359",
		"celestial":      "2400-2699",
		"ascendant":      "2100-2399",
		"ethereal":       "1800-2099",
		"mystic":         "1500-1799",
		"astral":         "1200-1499",
		"spiritual":      "900-1199",
		"primal":         "600-899",
		"mortal":         "300-599",
		"primordial":     "0-299",
	}

	for model, score := range totalScores {
		switch {
		case score >= 3780:
			tiers["transcendental"] = append(tiers["transcendental"], model)
		case score >= 3360:
			tiers["cosmic"] = append(tiers["cosmic"], model)
		case score >= 2700:
			tiers["divine"] = append(tiers["divine"], model)
		case score >= 2400:
			tiers["celestial"] = append(tiers["celestial"], model)
		case score >= 2100:
			tiers["ascendant"] = append(tiers["ascendant"], model)
		case score >= 1800:
			tiers["ethereal"] = append(tiers["ethereal"], model)
		case score >= 1500:
			tiers["mystic"] = append(tiers["mystic"], model)
		case score >= 1200:
			tiers["astral"] = append(tiers["astral"], model)
		case score >= 900:
			tiers["spiritual"] = append(tiers["spiritual"], model)
		case score >= 600:
			tiers["primal"] = append(tiers["primal"], model)
		case score >= 300:
			tiers["mortal"] = append(tiers["mortal"], model)
		default:
			tiers["primordial"] = append(tiers["primordial"], model)
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

		// Double-check total score calculation to ensure consistency
		calculatedTotal := stats.Count20*20 + stats.Count40*40 + stats.Count60*60 + stats.Count80*80 + stats.Count100*100
		if calculatedTotal != stats.TotalScore {
			log.Printf("Warning: Score mismatch for %s: calculated=%d, summed=%d", model, calculatedTotal, stats.TotalScore)
			// Fix the total score if there's a discrepancy
			stats.TotalScore = calculatedTotal
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
			"transcendental",
			"cosmic",
			"divine",
			"celestial",
			"ascendant",
			"ethereal",
			"mystic",
			"astral",
			"spiritual",
			"primal",
			"mortal",
			"primordial",
		},
	}

	// Parse and execute template
	t, err := template.New("stats.html").Funcs(template.FuncMap{
		"json": func(v interface{}) template.JS { // Updated to return template.JS
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
		"tierClass": func(tier string) string {
			return tier
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
