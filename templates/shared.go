package templates

import (
	"encoding/json"
	"html/template"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday/v2"
)

var FuncMap = map[string]interface{}{
	"inc": func(i int) int {
		return i + 1
	},
	"markdown": func(text string) template.HTML {
		unsafe := blackfriday.Run([]byte(text), blackfriday.WithNoExtensions())
		html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)
		return template.HTML(html)
	},
	"tolower":  strings.ToLower,
	"contains": strings.Contains,
	"json": func(v interface{}) (string, error) {
		b, err := json.Marshal(v)
		return string(b), err
	},
}

const (
	PageNameResults  = "Results"
	PageNamePrompts  = "Prompts"
	PageNameProfiles = "Profiles"
	PageNameEvaluate = "Evaluate"
)

var ScoreOptions = map[string]int{
	"0/5 (0)":   0,
	"1/5 (20)":  20,
	"2/5 (40)":  40,
	"3/5 (60)":  60,
	"4/5 (80)":  80,
	"5/5 (100)": 100,
}
