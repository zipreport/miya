package runtime

import (
	"fmt"
	"html"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// Pre-compiled regex pattern for control character escaping (performance optimization)
var reControlChars = regexp.MustCompile(`[\x00-\x1f\x7f]`)

// EscapeContext represents different auto-escape contexts
type EscapeContext string

const (
	// HTML context - escape HTML entities
	EscapeContextHTML EscapeContext = "html"
	// XHTML context - like HTML but stricter
	EscapeContextXHTML EscapeContext = "xhtml"
	// XML context - escape XML entities
	EscapeContextXML EscapeContext = "xml"
	// JavaScript context - escape for JavaScript strings
	EscapeContextJS EscapeContext = "js"
	// CSS context - escape for CSS strings
	EscapeContextCSS EscapeContext = "css"
	// URL context - URL encode values
	EscapeContextURL EscapeContext = "url"
	// JSON context - escape for JSON strings
	EscapeContextJSON EscapeContext = "json"
	// None/disabled - no escaping
	EscapeContextNone EscapeContext = "none"
)

// AutoEscapeConfig holds auto-escape configuration
type AutoEscapeConfig struct {
	Enabled    bool
	Context    EscapeContext
	DetectFn   func(templateName string) EscapeContext // Function to auto-detect context from filename
	Extensions map[string]EscapeContext                // File extension mappings
	ContextMap map[string]EscapeContext                // Template name to context mapping
}

// DefaultAutoEscapeConfig returns the default auto-escape configuration
func DefaultAutoEscapeConfig() *AutoEscapeConfig {
	return &AutoEscapeConfig{
		Enabled: true,
		Context: EscapeContextHTML,
		Extensions: map[string]EscapeContext{
			".html":  EscapeContextHTML,
			".htm":   EscapeContextHTML,
			".xhtml": EscapeContextXHTML,
			".xml":   EscapeContextXML,
			".js":    EscapeContextJS,
			".css":   EscapeContextCSS,
			".json":  EscapeContextJSON,
		},
		ContextMap: make(map[string]EscapeContext),
	}
}

// AutoEscaper handles auto-escaping for different contexts
type AutoEscaper struct {
	config *AutoEscapeConfig
}

// NewAutoEscaper creates a new auto-escaper with the given configuration
func NewAutoEscaper(config *AutoEscapeConfig) *AutoEscaper {
	if config == nil {
		config = DefaultAutoEscapeConfig()
	}
	return &AutoEscaper{config: config}
}

// DetectContext detects the appropriate escape context for a template
func (ae *AutoEscaper) DetectContext(templateName string) EscapeContext {
	// Check explicit context mapping first
	if context, exists := ae.config.ContextMap[templateName]; exists {
		return context
	}

	// Use custom detection function if provided
	if ae.config.DetectFn != nil {
		return ae.config.DetectFn(templateName)
	}

	// Check file extension
	for ext, context := range ae.config.Extensions {
		if strings.HasSuffix(strings.ToLower(templateName), ext) {
			return context
		}
	}

	// Default to configured context
	return ae.config.Context
}

// Escape escapes a value according to the specified context
func (ae *AutoEscaper) Escape(value interface{}, context EscapeContext) string {
	if !ae.config.Enabled || context == EscapeContextNone {
		return ToString(value)
	}

	// If already a SafeValue, don't escape
	if safeVal, ok := value.(SafeValue); ok {
		return ToString(safeVal.Value)
	}

	str := ToString(value)

	switch context {
	case EscapeContextHTML:
		return ae.escapeHTML(str)
	case EscapeContextXHTML:
		return ae.escapeXHTML(str)
	case EscapeContextXML:
		return ae.escapeXML(str)
	case EscapeContextJS:
		return ae.escapeJS(str)
	case EscapeContextCSS:
		return ae.escapeCSS(str)
	case EscapeContextURL:
		return ae.escapeURL(str)
	case EscapeContextJSON:
		return ae.escapeJSON(str)
	default:
		return ae.escapeHTML(str) // Default to HTML escaping
	}
}

// escapeHTML escapes HTML entities
func (ae *AutoEscaper) escapeHTML(s string) string {
	return html.EscapeString(s)
}

// escapeXHTML escapes XHTML entities (stricter than HTML)
func (ae *AutoEscaper) escapeXHTML(s string) string {
	s = html.EscapeString(s)
	// XHTML requires additional escaping
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// escapeXML escapes XML entities
func (ae *AutoEscaper) escapeXML(s string) string {
	// XML escaping is similar to HTML but more strict
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// escapeJS escapes JavaScript string literals
func (ae *AutoEscaper) escapeJS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\b", "\\b")
	s = strings.ReplaceAll(s, "\f", "\\f")
	s = strings.ReplaceAll(s, "\v", "\\v")
	s = strings.ReplaceAll(s, "\u0000", "\\u0000")
	// Escape HTML entities that could break out of script tags
	s = strings.ReplaceAll(s, "<", "\\u003c")
	s = strings.ReplaceAll(s, ">", "\\u003e")
	s = strings.ReplaceAll(s, "&", "\\u0026")
	return s
}

// escapeCSS escapes CSS string literals
func (ae *AutoEscaper) escapeCSS(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\A ")
	s = strings.ReplaceAll(s, "\r", "\\D ")
	s = strings.ReplaceAll(s, "\t", "\\9 ")
	// Escape potential CSS injection characters (using pre-compiled regex)
	s = reControlChars.ReplaceAllStringFunc(s, func(match string) string {
		return "\\x" + strings.ToUpper(string(rune(match[0])))
	})
	return s
}

// escapeURL escapes URL components
func (ae *AutoEscaper) escapeURL(s string) string {
	return url.QueryEscape(s)
}

// escapeJSON escapes JSON string literals
func (ae *AutoEscaper) escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	s = strings.ReplaceAll(s, "\b", "\\b")
	s = strings.ReplaceAll(s, "\f", "\\f")
	// Escape control characters (using pre-compiled regex)
	s = reControlChars.ReplaceAllStringFunc(s, func(match string) string {
		return "\\u" + strings.ToUpper(string(rune(match[0])))
	})
	return s
}

// SafeValue represents a value that should not be auto-escaped
type SafeValue struct {
	Value interface{}
}

// ToString converts SafeValue to string
func (sv SafeValue) String() string {
	return ToString(sv.Value)
}

// ToString safely converts any value to string
func ToString(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	if sv, ok := value.(SafeValue); ok {
		return ToString(sv.Value)
	}

	// Special formatting for float numbers
	switch v := value.(type) {
	case float64:
		// If the number looks like a currency amount (has decimals or is > 999),
		// format with comma separators
		if v >= 999.95 || (v > 0 && v != float64(int64(v))) {
			return formatCurrencyNumber(v)
		}
		// For other floats, if they're whole numbers, format as integers
		if v == float64(int64(v)) {
			return fmt.Sprintf("%.0f", v)
		}
	case float32:
		f64 := float64(v)
		if f64 >= 999.95 || (f64 > 0 && f64 != float64(int64(f64))) {
			return formatCurrencyNumber(f64)
		}
		// For other floats, if they're whole numbers, format as integers
		if f64 == float64(int64(f64)) {
			return fmt.Sprintf("%.0f", f64)
		}
	}

	return fmt.Sprintf("%v", value)
}

// formatCurrencyNumber formats a float with comma thousands separators
func formatCurrencyNumber(value float64) string {
	// Check if it's a whole number
	if value == float64(int64(value)) {
		// Format as integer with commas
		intVal := int64(value)
		return formatIntegerWithCommas(intVal)
	}

	// Format as decimal with commas
	str := fmt.Sprintf("%.2f", value)
	parts := strings.Split(str, ".")
	integerPart := parts[0]
	decimalPart := parts[1]

	// Add commas to integer part
	formattedInteger := formatIntegerPartWithCommas(integerPart)

	// Remove trailing zeros from decimal part, but keep at least original precision
	decimalPart = strings.TrimRight(decimalPart, "0")
	if decimalPart == "" {
		return formattedInteger
	}

	return formattedInteger + "." + decimalPart
}

// formatIntegerWithCommas formats an integer with comma separators
func formatIntegerWithCommas(value int64) string {
	str := fmt.Sprintf("%d", value)
	return formatIntegerPartWithCommas(str)
}

// formatIntegerPartWithCommas adds commas to an integer string
func formatIntegerPartWithCommas(str string) string {
	if len(str) <= 3 {
		return str
	}

	var result strings.Builder
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(digit)
	}
	return result.String()
}

// ContextAwareContext extends Context with escape context awareness
type ContextAwareContext interface {
	Context
	GetEscapeContext() EscapeContext
	SetEscapeContext(context EscapeContext)
	GetAutoEscaper() *AutoEscaper
	SetAutoEscaper(escaper *AutoEscaper)
}

// ContextWrapper wraps a Context to add auto-escape awareness
type ContextWrapper struct {
	Context
	escapeContext EscapeContext
	autoEscaper   *AutoEscaper
}

// NewContextWrapper creates a new context wrapper with auto-escape support
func NewContextWrapper(ctx Context, escaper *AutoEscaper, context EscapeContext) *ContextWrapper {
	return &ContextWrapper{
		Context:       ctx,
		escapeContext: context,
		autoEscaper:   escaper,
	}
}

// NewContextWrapperFromInterface creates a wrapper from any context interface
func NewContextWrapperFromInterface(ctx interface{}, escaper *AutoEscaper, context EscapeContext) *ContextWrapper {
	// Try to adapt the context to our runtime.Context interface
	if runtimeCtx, ok := ctx.(Context); ok {
		return NewContextWrapper(runtimeCtx, escaper, context)
	}

	// Create an adapter for foreign context interfaces
	adapter := &ContextAdapter{ctx: ctx}
	return NewContextWrapper(adapter, escaper, context)
}

// GetEscapeContext returns the current escape context
func (cw *ContextWrapper) GetEscapeContext() EscapeContext {
	return cw.escapeContext
}

// SetEscapeContext sets the escape context
func (cw *ContextWrapper) SetEscapeContext(context EscapeContext) {
	cw.escapeContext = context
}

// GetAutoEscaper returns the auto-escaper
func (cw *ContextWrapper) GetAutoEscaper() *AutoEscaper {
	return cw.autoEscaper
}

// SetAutoEscaper sets the auto-escaper
func (cw *ContextWrapper) SetAutoEscaper(escaper *AutoEscaper) {
	cw.autoEscaper = escaper
}

// IsAutoescapeEnabled returns true if auto-escaping is enabled
func (cw *ContextWrapper) IsAutoescapeEnabled() bool {
	return cw.autoEscaper != nil && cw.autoEscaper.config.Enabled
}

// Clone creates a copy of the context wrapper
func (cw *ContextWrapper) Clone() Context {
	return &ContextWrapper{
		Context:       cw.Context.Clone(),
		escapeContext: cw.escapeContext,
		autoEscaper:   cw.autoEscaper,
	}
}

// ContextAdapter adapts foreign context interfaces to runtime.Context
type ContextAdapter struct {
	ctx interface{}
}

// GetVariable gets a variable from the adapted context
func (ca *ContextAdapter) GetVariable(key string) (interface{}, bool) {
	// Use reflection to call GetVariable on the wrapped context
	ctxValue := reflect.ValueOf(ca.ctx)
	method := ctxValue.MethodByName("Get")
	if !method.IsValid() {
		return nil, false
	}

	results := method.Call([]reflect.Value{reflect.ValueOf(key)})
	if len(results) >= 2 {
		value := results[0].Interface()
		exists := results[1].Bool()
		return value, exists
	}
	return nil, false
}

// SetVariable sets a variable in the adapted context
func (ca *ContextAdapter) SetVariable(key string, value interface{}) {
	// Use reflection to call Set on the wrapped context
	ctxValue := reflect.ValueOf(ca.ctx)
	method := ctxValue.MethodByName("Set")
	if !method.IsValid() {
		return
	}

	method.Call([]reflect.Value{reflect.ValueOf(key), reflect.ValueOf(value)})
}

// Clone clones the adapted context
func (ca *ContextAdapter) Clone() Context {
	// Use reflection to call Clone on the wrapped context
	ctxValue := reflect.ValueOf(ca.ctx)
	method := ctxValue.MethodByName("Clone")
	if !method.IsValid() {
		return ca // Return self if Clone not available
	}

	results := method.Call(nil)
	if len(results) > 0 {
		cloned := results[0].Interface()
		return &ContextAdapter{ctx: cloned}
	}
	return ca
}

// All returns all variables (if supported by the wrapped context)
func (ca *ContextAdapter) All() map[string]interface{} {
	// Use reflection to call All on the wrapped context
	ctxValue := reflect.ValueOf(ca.ctx)
	method := ctxValue.MethodByName("All")
	if !method.IsValid() {
		return make(map[string]interface{})
	}

	results := method.Call(nil)
	if len(results) > 0 {
		if m, ok := results[0].Interface().(map[string]interface{}); ok {
			return m
		}
	}
	return make(map[string]interface{})
}
