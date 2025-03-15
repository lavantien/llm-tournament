package middleware

import (
	"fmt"
	"sort"
	"strconv"
)

// GetProfileGroups organizes profiles based on prompt order appearances
// Returns profile groups and a map of profile names to their group
func GetProfileGroups(prompts []Prompt, profiles []Profile) ([]*ProfileGroup, map[string]*ProfileGroup) {
	var profileGroups []*ProfileGroup
	profileMap := make(map[string]*ProfileGroup)

	// First, find all unique profiles used in prompts and their first occurrence
	profileOrder := make(map[string]int)
	profilesInUse := make(map[string]bool)
	
	// Track the order profiles first appear in prompts
	for i, prompt := range prompts {
		if prompt.Profile != "" {
			if _, exists := profileOrder[prompt.Profile]; !exists {
				profileOrder[prompt.Profile] = i
				profilesInUse[prompt.Profile] = true
			}
		}
	}
	
	// Create a sorted list of profile names based on first appearance
	var orderedProfileNames []string
	for profileName := range profileOrder {
		orderedProfileNames = append(orderedProfileNames, profileName)
	}
	sort.Slice(orderedProfileNames, func(i, j int) bool {
		return profileOrder[orderedProfileNames[i]] < profileOrder[orderedProfileNames[j]]
	})
	
	// Add profiles in order of first appearance in prompts
	for i, profileName := range orderedProfileNames {
		// Generate evenly distributed colors based on index
		colorHue := (i * 137) % 360 
		color := fmt.Sprintf("hsl(%d, 70%%, 50%%)", colorHue)
		
		profileGroups = append(profileGroups, &ProfileGroup{
			ID:       strconv.Itoa(i),
			Name:     profileName,
			Color:    color,
			StartCol: -1, // Will be populated later
			EndCol:   -1,
		})
		profileMap[profileName] = profileGroups[len(profileGroups)-1]
	}
	
	// Add any remaining profiles from the database that aren't used in prompts
	unusedIndex := len(profileGroups)
	for _, profile := range profiles {
		if !profilesInUse[profile.Name] {
			colorHue := (unusedIndex * 137) % 360
			color := fmt.Sprintf("hsl(%d, 70%%, 50%%)", colorHue)
			
			profileGroups = append(profileGroups, &ProfileGroup{
				ID:       strconv.Itoa(unusedIndex),
				Name:     profile.Name,
				Color:    color,
				StartCol: -1, // Will be populated later
				EndCol:   -1,
			})
			profileMap[profile.Name] = profileGroups[len(profileGroups)-1]
			unusedIndex++
		}
	}

	return profileGroups, profileMap
}
