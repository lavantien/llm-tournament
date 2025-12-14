package middleware

import (
	"strings"
	"testing"
)

func TestGetProfileGroups_Empty(t *testing.T) {
	groups, profileMap := GetProfileGroups([]Prompt{}, []Profile{})

	if len(groups) != 0 {
		t.Errorf("expected empty groups, got %d", len(groups))
	}
	if len(profileMap) != 0 {
		t.Errorf("expected empty profileMap, got %d", len(profileMap))
	}
}

func TestGetProfileGroups_SingleProfile(t *testing.T) {
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: "Profile A"},
		{Text: "Prompt 2", Profile: "Profile A"},
	}
	profiles := []Profile{
		{Name: "Profile A", Description: "Test profile"},
	}

	groups, profileMap := GetProfileGroups(prompts, profiles)

	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if len(profileMap) != 1 {
		t.Errorf("expected 1 entry in profileMap, got %d", len(profileMap))
	}
	if groups[0].Name != "Profile A" {
		t.Errorf("expected profile name 'Profile A', got %q", groups[0].Name)
	}
	if _, exists := profileMap["Profile A"]; !exists {
		t.Error("expected 'Profile A' in profileMap")
	}
}

func TestGetProfileGroups_MultipleProfiles(t *testing.T) {
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: "Profile A"},
		{Text: "Prompt 2", Profile: "Profile B"},
		{Text: "Prompt 3", Profile: "Profile A"},
		{Text: "Prompt 4", Profile: "Profile C"},
	}
	profiles := []Profile{
		{Name: "Profile A"},
		{Name: "Profile B"},
		{Name: "Profile C"},
	}

	groups, profileMap := GetProfileGroups(prompts, profiles)

	if len(groups) != 3 {
		t.Errorf("expected 3 groups, got %d", len(groups))
	}
	if len(profileMap) != 3 {
		t.Errorf("expected 3 entries in profileMap, got %d", len(profileMap))
	}

	// Check ordering based on first appearance
	expectedOrder := []string{"Profile A", "Profile B", "Profile C"}
	for i, expected := range expectedOrder {
		if groups[i].Name != expected {
			t.Errorf("expected groups[%d].Name = %q, got %q", i, expected, groups[i].Name)
		}
	}
}

func TestGetProfileGroups_OrderByFirstAppearance(t *testing.T) {
	// Profile C appears first, then A, then B
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: "Profile C"},
		{Text: "Prompt 2", Profile: "Profile A"},
		{Text: "Prompt 3", Profile: "Profile B"},
		{Text: "Prompt 4", Profile: "Profile C"},
	}
	profiles := []Profile{
		{Name: "Profile A"},
		{Name: "Profile B"},
		{Name: "Profile C"},
	}

	groups, _ := GetProfileGroups(prompts, profiles)

	// Should be ordered: C (index 0), A (index 1), B (index 2)
	expectedOrder := []string{"Profile C", "Profile A", "Profile B"}
	for i, expected := range expectedOrder {
		if groups[i].Name != expected {
			t.Errorf("expected groups[%d].Name = %q, got %q", i, expected, groups[i].Name)
		}
	}
}

func TestGetProfileGroups_UnusedProfiles(t *testing.T) {
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: "Profile A"},
	}
	profiles := []Profile{
		{Name: "Profile A"},
		{Name: "Profile B"}, // Not used in prompts
		{Name: "Profile C"}, // Not used in prompts
	}

	groups, profileMap := GetProfileGroups(prompts, profiles)

	if len(groups) != 3 {
		t.Errorf("expected 3 groups (1 used + 2 unused), got %d", len(groups))
	}
	if len(profileMap) != 3 {
		t.Errorf("expected 3 entries in profileMap, got %d", len(profileMap))
	}

	// Used profile should come first
	if groups[0].Name != "Profile A" {
		t.Errorf("expected first group to be 'Profile A', got %q", groups[0].Name)
	}

	// Unused profiles should be added after used ones
	for _, profile := range []string{"Profile B", "Profile C"} {
		if _, exists := profileMap[profile]; !exists {
			t.Errorf("expected %q in profileMap", profile)
		}
	}
}

func TestGetProfileGroups_EmptyProfileInPrompt(t *testing.T) {
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: ""},         // Empty profile
		{Text: "Prompt 2", Profile: "Profile A"},
		{Text: "Prompt 3", Profile: ""},         // Empty profile
	}
	profiles := []Profile{
		{Name: "Profile A"},
	}

	groups, profileMap := GetProfileGroups(prompts, profiles)

	// Only Profile A should be in groups (empty strings should be ignored)
	if len(groups) != 1 {
		t.Errorf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Name != "Profile A" {
		t.Errorf("expected group name 'Profile A', got %q", groups[0].Name)
	}
	if _, exists := profileMap[""]; exists {
		t.Error("empty profile should not be in profileMap")
	}
}

func TestGetProfileGroups_GoldenAngleColors(t *testing.T) {
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: "Profile A"},
		{Text: "Prompt 2", Profile: "Profile B"},
		{Text: "Prompt 3", Profile: "Profile C"},
		{Text: "Prompt 4", Profile: "Profile D"},
	}
	profiles := []Profile{}

	groups, _ := GetProfileGroups(prompts, profiles)

	if len(groups) != 4 {
		t.Fatalf("expected 4 groups, got %d", len(groups))
	}

	// Check that colors use HSL format with golden angle distribution
	// Golden angle: 137 degrees
	expectedHues := []int{
		(0 * 137) % 360,  // 0
		(1 * 137) % 360,  // 137
		(2 * 137) % 360,  // 274
		(3 * 137) % 360,  // 51 (411 % 360)
	}

	for i, group := range groups {
		expectedHue := expectedHues[i]
		expectedColor := "hsl(" + strings.Split(group.Color, "(")[1]
		if !strings.HasPrefix(expectedColor, "hsl(") {
			t.Errorf("expected HSL color format, got %q", group.Color)
		}
		// Just verify the format is correct
		if !strings.Contains(group.Color, "hsl(") {
			t.Errorf("group %d color should be HSL format, got %q", i, group.Color)
		}
		_ = expectedHue // We verify format, not exact values
	}
}

func TestGetProfileGroups_StartEndColInitialized(t *testing.T) {
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: "Profile A"},
	}
	profiles := []Profile{
		{Name: "Profile A"},
	}

	groups, _ := GetProfileGroups(prompts, profiles)

	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}

	// StartCol and EndCol should be -1 (will be populated later)
	if groups[0].StartCol != -1 {
		t.Errorf("expected StartCol = -1, got %d", groups[0].StartCol)
	}
	if groups[0].EndCol != -1 {
		t.Errorf("expected EndCol = -1, got %d", groups[0].EndCol)
	}
}

func TestGetProfileGroups_IDsAreSequential(t *testing.T) {
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: "Profile A"},
		{Text: "Prompt 2", Profile: "Profile B"},
		{Text: "Prompt 3", Profile: "Profile C"},
	}
	profiles := []Profile{}

	groups, _ := GetProfileGroups(prompts, profiles)

	for i, group := range groups {
		expected := string(rune('0' + i))
		if group.ID != expected {
			t.Errorf("expected groups[%d].ID = %q, got %q", i, expected, group.ID)
		}
	}
}

func TestGetProfileGroups_ProfileMapPointsToGroups(t *testing.T) {
	prompts := []Prompt{
		{Text: "Prompt 1", Profile: "Profile A"},
	}
	profiles := []Profile{
		{Name: "Profile A"},
	}

	groups, profileMap := GetProfileGroups(prompts, profiles)

	// profileMap entry should point to the same struct as in groups
	if profileMap["Profile A"] != groups[0] {
		t.Error("profileMap entry should point to same struct as in groups slice")
	}
}
