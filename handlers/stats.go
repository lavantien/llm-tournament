package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"llm-tournament/middleware"
	"log"
	"net/http"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

// StatsHandler handles the stats page (backward compatible wrapper)
func StatsHandler(w http.ResponseWriter, r *http.Request) {
	DefaultHandler.Stats(w, r)
}

// Stats handles the stats page
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling stats page")
	results := h.DataStore.ReadResults()

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

	// Get prompt count to calculate dynamic max score
	db := middleware.GetDB()
	var promptCount int
	err := db.QueryRow("SELECT COUNT(*) FROM prompts").Scan(&promptCount)
	if err != nil {
		log.Printf("Warning: failed to get prompt count: %v, using default 50", err)
		promptCount = 50
	}
	maxScore := promptCount * 100

	// Calculate tiers with dynamic max score
	tiers, tierRanges := calculateTiersWithMaxScore(totalScores, maxScore)

	// Prepare template data
	templateData := struct {
		PageName     string
		MaxScore     int
		TotalScores  map[string]ScoreStats
		Tiers        map[string][]string
		TierRanges   map[string]string
		OrderedTiers []string
		CurrentPath  string
	}{
		PageName:    "Statistics",
		MaxScore:    maxScore,
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
		CurrentPath: "/stats",
	}

	funcMap := template.FuncMap{
		"json": func(v interface{}) template.JS {
			a, _ := json.Marshal(v)
			return template.JS(a)
		},
		"eqs": func(a, b string) bool {
			return a == b
		},
		"tierClass": func(tier string) string {
			return tier
		},
		"formatTierName": func(tier string) string {
			return cases.Title(language.English).String(strings.ReplaceAll(tier, "-", " "))
		},
		"join": strings.Join,
	}

	err = h.Renderer.Render(w, "stats.html", funcMap, templateData, "templates/stats.html", "templates/nav.html")
	if err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}

// calculateTiersWithMaxScore calculates tiers based on total scores with dynamic max score
func calculateTiersWithMaxScore(totalScores map[string]int, maxScore int) (map[string][]string, map[string]string) {
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

	// Calculate even distribution across 12 tiers
	thresholds := map[string]int{
		"transcendental": (maxScore * 11) / 12,
		"cosmic":         (maxScore * 10) / 12,
		"divine":         (maxScore * 9) / 12,
		"celestial":      (maxScore * 8) / 12,
		"ascendant":      (maxScore * 7) / 12,
		"ethereal":       (maxScore * 6) / 12,
		"mystic":         (maxScore * 5) / 12,
		"astral":         (maxScore * 4) / 12,
		"spiritual":      (maxScore * 3) / 12,
		"primal":         (maxScore * 2) / 12,
		"mortal":         (maxScore * 1) / 12,
	}

	// Calculate percentages for display
	percentages := map[string]float64{
		"transcendental": float64(thresholds["transcendental"]) * 100 / float64(maxScore),
		"cosmic":         float64(thresholds["cosmic"]) * 100 / float64(maxScore),
		"divine":         float64(thresholds["divine"]) * 100 / float64(maxScore),
		"celestial":      float64(thresholds["celestial"]) * 100 / float64(maxScore),
		"ascendant":      float64(thresholds["ascendant"]) * 100 / float64(maxScore),
		"ethereal":       float64(thresholds["ethereal"]) * 100 / float64(maxScore),
		"mystic":         float64(thresholds["mystic"]) * 100 / float64(maxScore),
		"astral":         float64(thresholds["astral"]) * 100 / float64(maxScore),
		"spiritual":      float64(thresholds["spiritual"]) * 100 / float64(maxScore),
		"primal":         float64(thresholds["primal"]) * 100 / float64(maxScore),
		"mortal":         float64(thresholds["mortal"]) * 100 / float64(maxScore),
		"primordial":     0.0,
	}

	tierRanges := map[string]string{
		"transcendental": fmt.Sprintf("%d+ (%.1f%%+)", thresholds["transcendental"], percentages["transcendental"]),
		"cosmic":         fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["cosmic"], thresholds["transcendental"]-1, percentages["cosmic"], percentages["transcendental"]),
		"divine":         fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["divine"], thresholds["cosmic"]-1, percentages["divine"], percentages["cosmic"]),
		"celestial":      fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["celestial"], thresholds["divine"]-1, percentages["celestial"], percentages["divine"]),
		"ascendant":      fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["ascendant"], thresholds["celestial"]-1, percentages["ascendant"], percentages["celestial"]),
		"ethereal":       fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["ethereal"], thresholds["ascendant"]-1, percentages["ethereal"], percentages["ascendant"]),
		"mystic":         fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["mystic"], thresholds["ethereal"]-1, percentages["mystic"], percentages["ethereal"]),
		"astral":         fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["astral"], thresholds["mystic"]-1, percentages["astral"], percentages["mystic"]),
		"spiritual":      fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["spiritual"], thresholds["astral"]-1, percentages["spiritual"], percentages["astral"]),
		"primal":         fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["primal"], thresholds["spiritual"]-1, percentages["primal"], percentages["spiritual"]),
		"mortal":         fmt.Sprintf("%d-%d (%.1f%%-%.1f%%)", thresholds["mortal"], thresholds["primal"]-1, percentages["mortal"], percentages["primal"]),
		"primordial":     fmt.Sprintf("0-%d (0-%.1f%%)", thresholds["mortal"]-1, percentages["mortal"]),
	}

	for model, score := range totalScores {
		switch {
		case score >= thresholds["transcendental"]:
			tiers["transcendental"] = append(tiers["transcendental"], model)
		case score >= thresholds["cosmic"]:
			tiers["cosmic"] = append(tiers["cosmic"], model)
		case score >= thresholds["divine"]:
			tiers["divine"] = append(tiers["divine"], model)
		case score >= thresholds["celestial"]:
			tiers["celestial"] = append(tiers["celestial"], model)
		case score >= thresholds["ascendant"]:
			tiers["ascendant"] = append(tiers["ascendant"], model)
		case score >= thresholds["ethereal"]:
			tiers["ethereal"] = append(tiers["ethereal"], model)
		case score >= thresholds["mystic"]:
			tiers["mystic"] = append(tiers["mystic"], model)
		case score >= thresholds["astral"]:
			tiers["astral"] = append(tiers["astral"], model)
		case score >= thresholds["spiritual"]:
			tiers["spiritual"] = append(tiers["spiritual"], model)
		case score >= thresholds["primal"]:
			tiers["primal"] = append(tiers["primal"], model)
		case score >= thresholds["mortal"]:
			tiers["mortal"] = append(tiers["mortal"], model)
		default:
			tiers["primordial"] = append(tiers["primordial"], model)
		}
	}

	return tiers, tierRanges
}
