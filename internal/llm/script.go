package llm

import (
	"strings"
)

// ExtractScriptContent extracts the content between <script> tags from an LLM response.
// If no <script> tags are found, returns the entire response.
func ExtractScriptContent(response string) string {
	// Find the start and end tags
	start := strings.Index(response, "<script>")
	end := strings.LastIndex(response, "</script>")

	// If no tags found, return the entire response
	if start == -1 || end == -1 {
		return strings.TrimSpace(response)
	}

	// Extract content between tags and trim whitespace
	content := response[start+8 : end]
	return strings.TrimSpace(content)
}
