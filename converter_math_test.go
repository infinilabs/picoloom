package picoloom

import (
	"strings"
	"testing"
)

func TestInjectKaTeX_WithHead(t *testing.T) {
	html := `<!DOCTYPE html><html><head><title>X</title></head><body></body></html>`
	result := injectKaTeX(html, "./katex")

	if !strings.Contains(result, `href="file://`) {
		t.Error("expected file:// URL in result")
	}
	if !strings.Contains(result, "katex.min.css") {
		t.Error("expected katex.min.css link")
	}
	if !strings.Contains(result, "auto-render.min.js") {
		t.Error("expected auto-render.min.js script")
	}
}

func TestInjectKaTeX_SpacedPath(t *testing.T) {
	html := `<!DOCTYPE html><html><head></head><body></body></html>`
	result := injectKaTeX(html, "./my katex")

	// Spaces in the path must be percent-encoded in the URL.
	if strings.Contains(result, `file://./my katex`) {
		t.Error("spaces in file:// URL must be percent-encoded")
	}
	if !strings.Contains(result, `%20`) {
		t.Error("expected %20 encoding for space in path")
	}
}

func TestInjectKaTeX_MissingHead(t *testing.T) {
	html := `<!DOCTYPE html><html><body></body></html>`
	result := injectKaTeX(html, "./katex")

	// When </head> is missing, it should fall back to injecting before <body>.
	if !strings.Contains(result, "<body>") {
		t.Error("expected <body> to still be present")
	}
	if !strings.Contains(result, "katex.min.css") {
		t.Error("expected KaTeX CSS to be injected even without </head>")
	}
}
