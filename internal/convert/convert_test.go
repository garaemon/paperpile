package convert

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdown_paragraph(t *testing.T) {
	html := "<p>Hello world</p>"
	result, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown() error: %v", err)
	}
	trimmed := strings.TrimSpace(result)
	if trimmed != "Hello world" {
		t.Errorf("got %q, want %q", trimmed, "Hello world")
	}
}

func TestHTMLToMarkdown_bold(t *testing.T) {
	html := "<p>This is <b>bold</b> text</p>"
	result, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown() error: %v", err)
	}
	if !strings.Contains(result, "**bold**") {
		t.Errorf("got %q, want to contain **bold**", result)
	}
}

func TestHTMLToMarkdown_italic(t *testing.T) {
	html := "<p>This is <i>italic</i> text</p>"
	result, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown() error: %v", err)
	}
	if !strings.Contains(result, "*italic*") {
		t.Errorf("got %q, want to contain *italic*", result)
	}
}

func TestHTMLToMarkdown_link(t *testing.T) {
	html := `<p>Visit <a href="https://example.com">here</a></p>`
	result, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown() error: %v", err)
	}
	if !strings.Contains(result, "[here](https://example.com)") {
		t.Errorf("got %q, want to contain markdown link", result)
	}
}

func TestHTMLToMarkdown_unorderedList(t *testing.T) {
	html := "<ul><li>item 1</li><li>item 2</li></ul>"
	result, err := HTMLToMarkdown(html)
	if err != nil {
		t.Fatalf("HTMLToMarkdown() error: %v", err)
	}
	if !strings.Contains(result, "- item 1") || !strings.Contains(result, "- item 2") {
		t.Errorf("got %q, want unordered list items", result)
	}
}

func TestHTMLToMarkdown_empty(t *testing.T) {
	result, err := HTMLToMarkdown("")
	if err != nil {
		t.Fatalf("HTMLToMarkdown() error: %v", err)
	}
	if strings.TrimSpace(result) != "" {
		t.Errorf("got %q, want empty string", result)
	}
}

func TestMarkdownToHTML_paragraph(t *testing.T) {
	md := "Hello world"
	result := MarkdownToHTML(md)
	if !strings.Contains(result, "<p>Hello world</p>") {
		t.Errorf("got %q, want to contain <p>Hello world</p>", result)
	}
}

func TestMarkdownToHTML_bold(t *testing.T) {
	md := "This is **bold** text"
	result := MarkdownToHTML(md)
	if !strings.Contains(result, "<strong>bold</strong>") {
		t.Errorf("got %q, want to contain <strong>bold</strong>", result)
	}
}

func TestMarkdownToHTML_italic(t *testing.T) {
	md := "This is *italic* text"
	result := MarkdownToHTML(md)
	if !strings.Contains(result, "<em>italic</em>") {
		t.Errorf("got %q, want to contain <em>italic</em>", result)
	}
}

func TestMarkdownToHTML_link(t *testing.T) {
	md := "Visit [here](https://example.com)"
	result := MarkdownToHTML(md)
	if !strings.Contains(result, `<a href="https://example.com">here</a>`) {
		t.Errorf("got %q, want to contain html link", result)
	}
}

func TestMarkdownToHTML_list(t *testing.T) {
	md := "- item 1\n- item 2"
	result := MarkdownToHTML(md)
	if !strings.Contains(result, "<li>item 1</li>") || !strings.Contains(result, "<li>item 2</li>") {
		t.Errorf("got %q, want to contain list items", result)
	}
}

func TestMarkdownToHTML_empty(t *testing.T) {
	result := MarkdownToHTML("")
	if strings.TrimSpace(result) != "" {
		t.Errorf("got %q, want empty string", result)
	}
}
