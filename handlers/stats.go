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
		"cosmic-unity":           {},
		"divine-enlightenment":   {},
		"celestial-ascension":    {},
		"transcendent-awakening": {},
		"ethereal-harmony":       {},
		"mystic-resonance":       {},
		"astral-perception":      {},
		"spiritual-awakening":    {},
		"primal-awareness":       {},
		"mortal-consciousness":   {},
		"primordial-essence":     {},
	}

	tierRanges := map[string]string{
		"cosmic-unity":           "3000+",
		"divine-enlightenment":   "2700-2999",
		"celestial-ascension":    "2400-2699",
		"transcendent-awakening": "2100-2399",
		"ethereal-harmony":       "1800-2099",
		"mystic-resonance":       "1500-1799",
		"astral-perception":      "1200-1499",
		"spiritual-awakening":    "900-1199",
		"primal-awareness":       "600-899",
		"mortal-consciousness":   "300-599",
		"primordial-essence":     "0-299",
	}

	for model, score := range totalScores {
		switch {
		case score >= 3000:
			tiers["cosmic-unity"] = append(tiers["cosmic-unity"], model)
		case score >= 2700:
			tiers["divine-enlightenment"] = append(tiers["divine-enlightenment"], model)
		case score >= 2400:
			tiers["celestial-ascension"] = append(tiers["celestial-ascension"], model)
		case score >= 2100:
			tiers["transcendent-awakening"] = append(tiers["transcendent-awakening"], model)
		case score >= 1800:
			tiers["ethereal-harmony"] = append(tiers["ethereal-harmony"], model)
		case score >= 1500:
			tiers["mystic-resonance"] = append(tiers["mystic-resonance"], model)
		case score >= 1200:
			tiers["astral-perception"] = append(tiers["astral-perception"], model)
		case score >= 900:
			tiers["spiritual-awakening"] = append(tiers["spiritual-awakening"], model)
		case score >= 600:
			tiers["primal-awareness"] = append(tiers["primal-awareness"], model)
		case score >= 300:
			tiers["mortal-consciousness"] = append(tiers["mortal-consciousness"], model)
		default:
			tiers["primordial-essence"] = append(tiers["primordial-essence"], model)
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
			"cosmic-unity",
			"divine-enlightenment",
			"celestial-ascension",
			"transcendent-awakening",
			"ethereal-harmony",
			"mystic-resonance",
			"astral-perception",
			"spiritual-awakening",
			"primal-awareness",
			"mortal-consciousness",
			"primordial-essence",
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
