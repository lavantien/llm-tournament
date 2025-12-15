package templates

import (
	"html/template"
	"strings"
	"testing"
)

// TestMarkdown_XSS_ScriptTag tests that script tags are sanitized
func TestMarkdown_XSS_ScriptTag(t *testing.T) {
	input := "<script>alert('xss')</script>Hello World"
	result := FuncMap["markdown"].(func(string) template.HTML)(input)
	output := string(result)

	if strings.Contains(output, "<script>") {
		t.Error("script tags should be removed by bluemonday")
	}

	// Should preserve safe content
	if !strings.Contains(output, "Hello World") {
		t.Error("expected safe content to be preserved")
	}
}

// TestMarkdown_XSS_EventHandler tests that event handlers are sanitized
func TestMarkdown_XSS_EventHandler(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"img onerror", `<img src="x" onerror="alert(1)">`},
		{"div onclick", `<div onclick="alert(1)">Click me</div>`},
		{"a onload", `<a onload="alert(1)" href="#">Link</a>`},
		{"body onload", `<body onload="alert(1)">Content</body>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuncMap["markdown"].(func(string) template.HTML)(tt.input)
			output := string(result)

			// Event handlers should be stripped
			if strings.Contains(output, "onerror=") || strings.Contains(output, "onclick=") || strings.Contains(output, "onload=") {
				t.Errorf("event handlers should be removed: %s", output)
			}
		})
	}
}

// TestMarkdown_XSS_JavaScriptURL tests that javascript: URLs are sanitized
func TestMarkdown_XSS_JavaScriptURL(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"markdown link", `[Click me](javascript:alert(1))`},
		{"html a tag", `<a href="javascript:alert(1)">Click</a>`},
		{"data url", `<a href="data:text/html,<script>alert(1)</script>">Click</a>`},
		{"vbscript url", `<a href="vbscript:alert(1)">Click</a>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuncMap["markdown"].(func(string) template.HTML)(tt.input)
			output := string(result)

			// Dangerous URL schemes should be removed or neutered
			if strings.Contains(output, "javascript:") || strings.Contains(output, "vbscript:") {
				t.Errorf("dangerous URL scheme should be removed: %s", output)
			}
		})
	}
}

// TestMarkdown_HTMLInjection tests that dangerous HTML elements are sanitized
func TestMarkdown_HTMLInjection(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check string
	}{
		{"iframe", `<iframe src="evil.com"></iframe>`, "<iframe"},
		{"object", `<object data="evil.swf"></object>`, "<object"},
		{"embed", `<embed src="evil.swf">`, "<embed"},
		{"form", `<form action="evil.com"><input type="submit"></form>`, "<form"},
		{"base", `<base href="evil.com">`, "<base"},
		{"meta", `<meta http-equiv="refresh" content="0;url=evil.com">`, "<meta"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuncMap["markdown"].(func(string) template.HTML)(tt.input)
			output := string(result)

			// Dangerous elements should be removed
			if strings.Contains(output, tt.check) {
				t.Errorf("dangerous element %s should be removed: %s", tt.check, output)
			}
		})
	}
}

// TestMarkdown_SpecialCharacters tests handling of HTML special characters
func TestMarkdown_SpecialCharacters(t *testing.T) {
	input := `Test & ampersand, <less than>, "quotes", 'apostrophes'`
	result := FuncMap["markdown"].(func(string) template.HTML)(input)
	output := string(result)

	// Should not break or cause errors
	if len(output) == 0 {
		t.Error("expected non-empty output")
	}

	// Content should be preserved in some form (markdown converts to <p> tags)
	if !strings.Contains(output, "Test") {
		t.Error("expected content to be preserved")
	}
}

// TestMarkdown_EmptyInput tests handling of empty input
func TestMarkdown_EmptyInput(t *testing.T) {
	input := ""
	result := FuncMap["markdown"].(func(string) template.HTML)(input)
	output := string(result)

	// Should return empty or minimal HTML without error
	if strings.Contains(output, "panic") || strings.Contains(output, "error") {
		t.Error("empty input should not cause errors")
	}
}

// TestMarkdown_WhitespaceOnly tests handling of whitespace-only input
func TestMarkdown_WhitespaceOnly(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"spaces", "   "},
		{"newlines", "\n\n\n"},
		{"tabs", "\t\t\t"},
		{"mixed", "  \n\t  \n  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuncMap["markdown"].(func(string) template.HTML)(tt.input)
			// Should not panic or error
			_ = string(result)
		})
	}
}

// TestMarkdown_VeryLargeInput tests handling of large markdown input
func TestMarkdown_VeryLargeInput(t *testing.T) {
	// Create a 1MB string
	var builder strings.Builder
	for i := 0; i < 10000; i++ {
		builder.WriteString("This is line ")
		builder.WriteString(strings.Repeat("x", 100))
		builder.WriteString("\n")
	}
	input := builder.String()

	result := FuncMap["markdown"].(func(string) template.HTML)(input)
	output := string(result)

	// Should complete without panic
	if len(output) == 0 {
		t.Error("expected non-empty output for large input")
	}
}

// TestMarkdown_InvalidUTF8 tests handling of invalid UTF-8 sequences
func TestMarkdown_InvalidUTF8(t *testing.T) {
	// Invalid UTF-8 byte sequences
	tests := []struct {
		name  string
		input string
	}{
		{"invalid bytes", "Hello \xff\xfe World"},
		{"truncated sequence", "Hello \xc3 World"},
		{"overlong encoding", "Hello \xc0\x80 World"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic even with invalid UTF-8
			result := FuncMap["markdown"].(func(string) template.HTML)(tt.input)
			_ = string(result)
		})
	}
}

// TestMarkdown_NestedXSS tests nested XSS attempts
func TestMarkdown_NestedXSS(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"nested script", `<div><script>alert(1)</script></div>`},
		{"encoded script", `<scr<script>ipt>alert(1)</scr</script>ipt>`},
		{"mixed case", `<ScRiPt>alert(1)</sCrIpT>`},
		{"null byte", "<script\x00>alert(1)</script>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuncMap["markdown"].(func(string) template.HTML)(tt.input)
			output := string(result)

			// No script tags should survive
			if strings.Contains(strings.ToLower(output), "<script") {
				t.Errorf("nested XSS attempt should be blocked: %s", output)
			}
		})
	}
}

// TestMarkdown_MarkdownFeatures tests that legitimate markdown features work
func TestMarkdown_MarkdownFeatures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"bold", "**bold text**", "bold text"},
		{"italic", "*italic text*", "italic text"},
		{"link", "[link](http://example.com)", "link"},
		{"heading", "# Heading", "Heading"},
		{"list", "- item1\n- item2", "item"},
		{"code", "`code`", "code"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuncMap["markdown"].(func(string) template.HTML)(tt.input)
			output := string(result)

			// Should preserve the content
			if !strings.Contains(output, tt.contains) {
				t.Errorf("expected markdown feature to be processed: %s", output)
			}
		})
	}
}

// TestMarkdown_AllowedHTMLElements tests that UGC policy allows safe HTML
func TestMarkdown_AllowedHTMLElements(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"paragraph", "<p>Paragraph</p>", "<p>"},
		{"strong", "<strong>Bold</strong>", "<strong>"},
		{"em", "<em>Italic</em>", "<em>"},
		{"ul list", "<ul><li>Item</li></ul>", "<ul>"},
		{"ol list", "<ol><li>Item</li></ol>", "<ol>"},
		{"blockquote", "<blockquote>Quote</blockquote>", "<blockquote>"},
		{"code block", "<pre><code>code</code></pre>", "<code>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuncMap["markdown"].(func(string) template.HTML)(tt.input)
			output := string(result)

			// UGC policy should allow these safe elements
			if !strings.Contains(output, tt.contains) {
				t.Errorf("expected safe HTML element to be preserved: %s", output)
			}
		})
	}
}
