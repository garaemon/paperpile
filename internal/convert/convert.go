// Package convert provides HTML-to-Markdown and Markdown-to-HTML conversion
// for Paperpile note content.
package convert

import (
	"bytes"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/yuin/goldmark"
)

// HTMLToMarkdown converts an HTML string to Markdown (best-effort).
func HTMLToMarkdown(html string) (string, error) {
	if strings.TrimSpace(html) == "" {
		return "", nil
	}
	return htmltomarkdown.ConvertString(html)
}

// MarkdownToHTML converts a Markdown string to HTML.
func MarkdownToHTML(md string) string {
	if strings.TrimSpace(md) == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(md), &buf); err != nil {
		return md
	}
	return strings.TrimSpace(buf.String())
}
