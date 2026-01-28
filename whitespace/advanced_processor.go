package whitespace

import (
	"regexp"
	"strings"
)

// Pre-compiled regex patterns for whitespace processing (performance optimization)
var (
	// Block tags with both strips: {%- ... -%}
	reBothStripsBlock = regexp.MustCompile(`(\s*)\{%-\s*(.*?)\s*-%\}(\s*)`)
	// Block tags with left strip only: {%- ... %}
	reLeftStripBlock = regexp.MustCompile(`(\s*)\{%-\s*(.*?)\s*%\}`)
	// Block tags with right strip only: {% ... -%}
	reRightStripBlock = regexp.MustCompile(`\{%\s*(.*?)\s*-%\}(\s*)`)
	// Variable tags with both strips: {{- ... -}}
	reBothStripsVar = regexp.MustCompile(`(\s*)\{\{-\s*(.*?)\s*-\}\}(\s*)`)
	// Variable tags with left strip only: {{- ... }}
	reLeftStripVar = regexp.MustCompile(`(\s*)\{\{-\s*(.*?)\s*\}\}`)
	// Variable tags with right strip only: {{ ... -}}
	reRightStripVar = regexp.MustCompile(`\{\{\s*(.*?)\s*-\}\}(\s*)`)
	// Comment tags: {#- ... -#} or {# ... #}
	reCommentTag = regexp.MustCompile(`(\s*)\{#-?\s*(.*?)\s*-?#\}(\s*)`)
	// Trim blocks: remove newline after block statements
	reTrimBlocks = regexp.MustCompile(`(\{%.*?%\})\r?\n`)
	// Lstrip blocks: remove whitespace before block statements
	reLstripBlocks = regexp.MustCompile(`\n[ \t]*(\{%.*?%\})`)
	// Compact whitespace patterns
	reMultipleSpaces   = regexp.MustCompile(`[ \t]+`)
	reMultipleNewlines = regexp.MustCompile(`\n\s*\n`)
)

// AdvancedWhitespaceProcessor provides enhanced whitespace control
type AdvancedWhitespaceProcessor struct {
	trimBlocks          bool
	lstripBlocks        bool
	keepTrailingNewline bool
}

// NewAdvancedWhitespaceProcessor creates an advanced whitespace processor
func NewAdvancedWhitespaceProcessor(trimBlocks, lstripBlocks, keepTrailingNewline bool) *AdvancedWhitespaceProcessor {
	return &AdvancedWhitespaceProcessor{
		trimBlocks:          trimBlocks,
		lstripBlocks:        lstripBlocks,
		keepTrailingNewline: keepTrailingNewline,
	}
}

// ProcessTemplate processes a template string and applies whitespace control
func (a *AdvancedWhitespaceProcessor) ProcessTemplate(template string) string {
	// Process {%- ... -%} syntax for inline whitespace control
	result := a.processInlineWhitespaceControl(template)

	// Apply global whitespace settings
	if a.trimBlocks || a.lstripBlocks {
		result = a.applyGlobalWhitespace(result)
	}

	// Handle trailing newlines
	if !a.keepTrailingNewline {
		result = strings.TrimSuffix(result, "\n")
	}

	return result
}

// replaceWithSubmatch efficiently replaces regex matches using submatch index
// This avoids double regex matching (ReplaceAllStringFunc + FindStringSubmatch)
func replaceWithSubmatch(re *regexp.Regexp, s string, contentGroup int, prefix, suffix string) string {
	matches := re.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return s
	}

	var result strings.Builder
	result.Grow(len(s)) // Pre-allocate approximate size
	lastEnd := 0

	for _, match := range matches {
		// match[0:2] is full match, match[2*n:2*n+2] is group n
		if len(match) < (contentGroup+1)*2 {
			continue
		}

		// Write text before this match
		result.WriteString(s[lastEnd:match[0]])

		// Extract content from the specified capture group
		contentStart := match[contentGroup*2]
		contentEnd := match[contentGroup*2+1]
		if contentStart >= 0 && contentEnd >= 0 {
			content := strings.TrimSpace(s[contentStart:contentEnd])
			result.WriteString(prefix)
			result.WriteString(content)
			result.WriteString(suffix)
		} else {
			// If group didn't match, keep original
			result.WriteString(s[match[0]:match[1]])
		}

		lastEnd = match[1]
	}

	// Write remaining text after last match
	result.WriteString(s[lastEnd:])
	return result.String()
}

// processInlineWhitespaceControl handles {%- and -%} syntax
func (a *AdvancedWhitespaceProcessor) processInlineWhitespaceControl(template string) string {
	result := template

	// Process block tags with both strips: capture group 2 has content
	result = replaceWithSubmatch(reBothStripsBlock, result, 2, "{% ", " %}")

	// Process block tags with left strip only: capture group 2 has content
	result = replaceWithSubmatch(reLeftStripBlock, result, 2, "{% ", " %}")

	// Process block tags with right strip only: capture group 1 has content
	result = replaceWithSubmatch(reRightStripBlock, result, 1, "{% ", " %}")

	// Process variable tags with both strips: capture group 2 has content
	result = replaceWithSubmatch(reBothStripsVar, result, 2, "{{ ", " }}")

	// Process variable tags with left strip only: capture group 2 has content
	result = replaceWithSubmatch(reLeftStripVar, result, 2, "{{ ", " }}")

	// Process variable tags with right strip only: capture group 1 has content
	result = replaceWithSubmatch(reRightStripVar, result, 1, "{{ ", " }}")

	// Process comment tags (they get removed anyway)
	result = reCommentTag.ReplaceAllString(result, "")

	return result
}

// applyGlobalWhitespace applies global trim_blocks and lstrip_blocks settings
func (a *AdvancedWhitespaceProcessor) applyGlobalWhitespace(template string) string {
	result := template

	if a.trimBlocks {
		// Remove newlines after block statements
		result = reTrimBlocks.ReplaceAllString(result, "$1")
	}

	if a.lstripBlocks {
		// Remove whitespace before block statements
		result = reLstripBlocks.ReplaceAllString(result, "\n$1")
	}

	return result
}

// StripWhitespaceAroundTags strips whitespace around template tags based on control modifiers
func StripWhitespaceAroundTags(template string) string {
	processor := NewAdvancedWhitespaceProcessor(false, false, true)
	return processor.processInlineWhitespaceControl(template)
}

// CompactWhitespace removes excessive whitespace while preserving structure
func CompactWhitespace(text string) string {
	// Replace multiple spaces with single space
	result := reMultipleSpaces.ReplaceAllString(text, " ")

	// Replace multiple newlines with single newline
	result = reMultipleNewlines.ReplaceAllString(result, "\n")

	// Trim whitespace at start and end of lines
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	result = strings.Join(lines, "\n")

	return strings.TrimSpace(result)
}
