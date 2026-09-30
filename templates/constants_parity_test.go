package templates

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func mustReadJS(t *testing.T, name string) string {
	t.Helper()
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	return string(raw)
}

// parseJSObject parses "key: \"value\"" entries from a const object body and
// fails on any content beyond those entries, so stray or malformed members
// cannot slip through.
func parseJSObject(t *testing.T, js, constName string) map[string]string {
	t.Helper()

	m := regexp.MustCompile(`(?s)const ` + constName + ` = \{(.*?)\};`).FindStringSubmatch(js)
	if m == nil {
		t.Fatalf("no const %s object found", constName)
	}
	body := m[1]

	entryRe := regexp.MustCompile(`(\d+)\s*:\s*"([^"]*)"`)
	entries := entryRe.FindAllStringSubmatch(body, -1)
	if entries == nil {
		t.Fatalf("no entries found in %s", constName)
	}

	parsed := make(map[string]string, len(entries))
	for _, e := range entries {
		parsed[e[1]] = e[2]
	}

	remainder := entryRe.ReplaceAllString(body, "")
	remainder = strings.NewReplacer(",", "", " ", "", "\n", "", "\r", "", "\t", "").Replace(remainder)
	if remainder != "" {
		t.Errorf("unexpected content in %s: %q", constName, remainder)
	}
	return parsed
}

func goScoreReference(t *testing.T) (values []int, labels map[int]string) {
	t.Helper()

	labels = make(map[int]string, len(ScoreOptions))
	for label, value := range ScoreOptions {
		prefix, ok := strings.CutSuffix(label, fmt.Sprintf(" (%d)", value))
		if !ok {
			t.Fatalf("ScoreOptions key %q does not follow the %q format for value %d", label, fmt.Sprintf("label (%d)", value), value)
		}
		labels[value] = prefix
		values = append(values, value)
	}
	slices.Sort(values)
	return values, labels
}

func TestConstantsJS_MirrorsGoScoreConstants(t *testing.T) {
	js := mustReadJS(t, "constants.js")

	values, labels := goScoreReference(t)

	// SCORE_VALUES must list the ScoreOptions values in ascending order
	m := regexp.MustCompile(`(?s)const SCORE_VALUES = \[([0-9,\s]*)\];`).FindStringSubmatch(js)
	if m == nil {
		t.Fatal("no const SCORE_VALUES array found")
	}
	var jsValues []int
	for _, part := range strings.Split(m[1], ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		v, err := strconv.Atoi(part)
		if err != nil {
			t.Fatalf("SCORE_VALUES contains non-numeric entry %q", part)
		}
		jsValues = append(jsValues, v)
	}
	if !slices.Equal(jsValues, values) {
		t.Errorf("SCORE_VALUES = %v, want %v (ascending ScoreOptions values)", jsValues, values)
	}

	// SCORE_LABELS must map every score value to the label half of its
	// ScoreOptions key, with no extra members
	wantLabels := make(map[string]string, len(labels))
	for v, l := range labels {
		wantLabels[strconv.Itoa(v)] = l
	}
	jsLabels := parseJSObject(t, js, "SCORE_LABELS")
	for key, want := range wantLabels {
		if got, ok := jsLabels[key]; !ok {
			t.Errorf("SCORE_LABELS is missing entry for score %s", key)
		} else if got != want {
			t.Errorf("SCORE_LABELS[%s] = %q, want %q", key, got, want)
		}
	}
	for key := range jsLabels {
		if _, ok := wantLabels[key]; !ok {
			t.Errorf("SCORE_LABELS has entry for unknown score %s", key)
		}
	}

	// SCORE_COLORS must equal ScoreColors entry for entry
	wantColors := make(map[string]string, len(ScoreColors))
	for v, c := range ScoreColors {
		wantColors[strconv.Itoa(v)] = c
	}
	jsColors := parseJSObject(t, js, "SCORE_COLORS")
	for key, want := range wantColors {
		if got, ok := jsColors[key]; !ok {
			t.Errorf("SCORE_COLORS is missing entry for score %s", key)
		} else if got != want {
			t.Errorf("SCORE_COLORS[%s] = %q, want %q", key, got, want)
		}
	}
	for key := range jsColors {
		if _, ok := wantColors[key]; !ok {
			t.Errorf("SCORE_COLORS has entry for unknown score %s", key)
		}
	}

	// The palette must cover exactly the selectable score values
	if len(jsColors) != len(jsValues) {
		t.Errorf("SCORE_COLORS has %d entries, SCORE_VALUES has %d", len(jsColors), len(jsValues))
	}
}

func TestScoreUtilsJS_FallbackMatchesGoScoreColorFallback(t *testing.T) {
	js := mustReadJS(t, "score-utils.js")

	m := regexp.MustCompile(`SCORE_COLORS\[\w+\]\s*\|\|\s*SCORE_COLORS\[(\d+)\]`).FindStringSubmatch(js)
	if m == nil {
		t.Fatal("getScoreColor fallback expression not found in score-utils.js")
	}
	fallback, err := strconv.Atoi(m[1])
	if err != nil {
		t.Fatalf("getScoreColor fallback index %q is not numeric", m[1])
	}

	goFallback, ok := ScoreColors[0]
	if !ok {
		t.Fatal("Go ScoreColors has no entry for score 0")
	}
	if fallback != 0 {
		t.Errorf("getScoreColor falls back to SCORE_COLORS[%d], want SCORE_COLORS[0] (scoreColor falls back to ScoreColors[0])", fallback)
	}

	scoreColor, ok := FuncMap["scoreColor"].(func(int) string)
	if !ok {
		t.Fatal("FuncMap[scoreColor] does not have type func(int) string")
	}
	for _, unknown := range []int{-1, 7, 55} {
		if got := scoreColor(unknown); got != goFallback {
			t.Errorf("scoreColor(%d) = %q, want fallback %q", unknown, got, goFallback)
		}
	}

	// The fallback color the JS serves must be the same hex as the Go one
	jsColors := parseJSObject(t, mustReadJS(t, "constants.js"), "SCORE_COLORS")
	if jsColors[strconv.Itoa(fallback)] != goFallback {
		t.Errorf("SCORE_COLORS[%d] = %q, want Go fallback color %q", fallback, jsColors[strconv.Itoa(fallback)], goFallback)
	}
}
