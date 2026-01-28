package miya

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// DebugLevel represents the level of debug information
type DebugLevel int

const (
	DebugLevelOff DebugLevel = iota
	DebugLevelBasic
	DebugLevelDetailed
	DebugLevelVerbose
)

// DebugEvent represents a debug event during template execution
type DebugEvent struct {
	Type         string
	TemplateName string
	Line         int
	Column       int
	Message      string
	Variables    map[string]interface{}
	Timestamp    time.Time
}

// DebugTracer provides tracing capabilities for template execution
type DebugTracer struct {
	enabled bool
	level   DebugLevel
	events  []DebugEvent
	filters map[string]bool
	mutex   sync.RWMutex
}

// NewDebugTracer creates a new debug tracer
func NewDebugTracer() *DebugTracer {
	return &DebugTracer{
		enabled: false,
		level:   DebugLevelBasic,
		events:  make([]DebugEvent, 0),
		filters: make(map[string]bool),
	}
}

// Enable enables debug tracing
func (dt *DebugTracer) Enable() {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	dt.enabled = true
}

// Disable disables debug tracing
func (dt *DebugTracer) Disable() {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	dt.enabled = false
}

// SetLevel sets the debug level
func (dt *DebugTracer) SetLevel(level DebugLevel) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	dt.level = level
}

// AddFilter adds an event type filter
func (dt *DebugTracer) AddFilter(eventType string) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	dt.filters[eventType] = true
}

// RemoveFilter removes an event type filter
func (dt *DebugTracer) RemoveFilter(eventType string) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	delete(dt.filters, eventType)
}

// TraceEvent records a debug event
func (dt *DebugTracer) TraceEvent(eventType, templateName string, line, column int, message string, variables map[string]interface{}) {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()

	if !dt.enabled {
		return
	}

	// Check if this event type should be filtered
	if len(dt.filters) > 0 && !dt.filters[eventType] {
		return
	}

	event := DebugEvent{
		Type:         eventType,
		TemplateName: templateName,
		Line:         line,
		Column:       column,
		Message:      message,
		Variables:    variables,
		Timestamp:    time.Now(),
	}

	dt.events = append(dt.events, event)
}

// GetEvents returns all recorded events
func (dt *DebugTracer) GetEvents() []DebugEvent {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()

	// Return a copy to avoid race conditions
	events := make([]DebugEvent, len(dt.events))
	copy(events, dt.events)
	return events
}

// Clear clears all recorded events
func (dt *DebugTracer) Clear() {
	dt.mutex.Lock()
	defer dt.mutex.Unlock()
	dt.events = dt.events[:0]
}

// GetSummary returns a summary of recorded events
func (dt *DebugTracer) GetSummary() string {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()

	if len(dt.events) == 0 {
		return "No debug events recorded"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Debug Summary: %d events recorded\n", len(dt.events)))
	sb.WriteString(strings.Repeat("=", 40))
	sb.WriteString("\n\n")

	// Count events by type
	eventCounts := make(map[string]int)
	templateCounts := make(map[string]int)

	for _, event := range dt.events {
		eventCounts[event.Type]++
		if event.TemplateName != "" {
			templateCounts[event.TemplateName]++
		}
	}

	sb.WriteString("Event Types:\n")
	for eventType, count := range eventCounts {
		sb.WriteString(fmt.Sprintf("  %s: %d\n", eventType, count))
	}

	if len(templateCounts) > 0 {
		sb.WriteString("\nTemplates:\n")
		for template, count := range templateCounts {
			sb.WriteString(fmt.Sprintf("  %s: %d events\n", template, count))
		}
	}

	return sb.String()
}

// GetDetailedLog returns a detailed log of all events
func (dt *DebugTracer) GetDetailedLog() string {
	dt.mutex.RLock()
	defer dt.mutex.RUnlock()

	if len(dt.events) == 0 {
		return "No debug events recorded"
	}

	var sb strings.Builder
	sb.WriteString("Detailed Debug Log\n")
	sb.WriteString(strings.Repeat("=", 50))
	sb.WriteString("\n\n")

	for i, event := range dt.events {
		sb.WriteString(fmt.Sprintf("%d. [%s] %s", i+1, event.Type, event.Message))

		if event.TemplateName != "" {
			sb.WriteString(fmt.Sprintf(" (template: %s)", event.TemplateName))
		}

		if event.Line > 0 {
			sb.WriteString(fmt.Sprintf(" at line %d", event.Line))
			if event.Column > 0 {
				sb.WriteString(fmt.Sprintf(", column %d", event.Column))
			}
		}

		sb.WriteString(fmt.Sprintf(" - %s\n", event.Timestamp.Format("15:04:05.000")))

		if dt.level >= DebugLevelDetailed && event.Variables != nil && len(event.Variables) > 0 {
			sb.WriteString("   Variables:\n")
			for key, value := range event.Variables {
				sb.WriteString(fmt.Sprintf("     %s: %v\n", key, value))
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

// PerformanceMeasurement represents a performance measurement
type PerformanceMeasurement struct {
	Count     int64
	TotalTime time.Duration
	MinTime   time.Duration
	MaxTime   time.Duration
}

// PerformanceProfiler provides performance profiling capabilities
type PerformanceProfiler struct {
	enabled      bool
	measurements map[string]*PerformanceMeasurement
	mutex        sync.RWMutex
}

// NewPerformanceProfiler creates a new performance profiler
func NewPerformanceProfiler() *PerformanceProfiler {
	return &PerformanceProfiler{
		enabled:      false,
		measurements: make(map[string]*PerformanceMeasurement),
	}
}

// Enable enables performance profiling
func (pp *PerformanceProfiler) Enable() {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.enabled = true
}

// Disable disables performance profiling
func (pp *PerformanceProfiler) Disable() {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.enabled = false
}

// StartMeasurement starts measuring performance for a given operation
func (pp *PerformanceProfiler) StartMeasurement(operation string) func() {
	if !pp.enabled {
		return func() {}
	}

	start := time.Now()

	return func() {
		duration := time.Since(start)
		pp.recordMeasurement(operation, duration)
	}
}

// recordMeasurement records a performance measurement
func (pp *PerformanceProfiler) recordMeasurement(operation string, duration time.Duration) {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()

	measurement := pp.measurements[operation]
	if measurement == nil {
		measurement = &PerformanceMeasurement{
			MinTime: duration,
			MaxTime: duration,
		}
		pp.measurements[operation] = measurement
	}

	measurement.Count++
	measurement.TotalTime += duration

	if duration < measurement.MinTime {
		measurement.MinTime = duration
	}
	if duration > measurement.MaxTime {
		measurement.MaxTime = duration
	}
}

// GetMeasurements returns all performance measurements
func (pp *PerformanceProfiler) GetMeasurements() map[string]*PerformanceMeasurement {
	pp.mutex.RLock()
	defer pp.mutex.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*PerformanceMeasurement)
	for key, measurement := range pp.measurements {
		result[key] = &PerformanceMeasurement{
			Count:     measurement.Count,
			TotalTime: measurement.TotalTime,
			MinTime:   measurement.MinTime,
			MaxTime:   measurement.MaxTime,
		}
	}
	return result
}

// Clear clears all measurements
func (pp *PerformanceProfiler) Clear() {
	pp.mutex.Lock()
	defer pp.mutex.Unlock()
	pp.measurements = make(map[string]*PerformanceMeasurement)
}

// GetReport returns a performance report
func (pp *PerformanceProfiler) GetReport() string {
	measurements := pp.GetMeasurements()

	if len(measurements) == 0 {
		return "No performance measurements recorded"
	}

	var sb strings.Builder
	sb.WriteString("Performance Report\n")
	sb.WriteString(strings.Repeat("=", 50))
	sb.WriteString("\n\n")

	for operation, measurement := range measurements {
		avgTime := measurement.TotalTime / time.Duration(measurement.Count)

		sb.WriteString(fmt.Sprintf("Operation: %s\n", operation))
		sb.WriteString(fmt.Sprintf("  Count: %d\n", measurement.Count))
		sb.WriteString(fmt.Sprintf("  Total Time: %v\n", measurement.TotalTime))
		sb.WriteString(fmt.Sprintf("  Average Time: %v\n", avgTime))
		sb.WriteString(fmt.Sprintf("  Min Time: %v\n", measurement.MinTime))
		sb.WriteString(fmt.Sprintf("  Max Time: %v\n", measurement.MaxTime))
		sb.WriteString("\n")
	}

	return sb.String()
}

// InteractiveDebugger provides interactive debugging capabilities
type InteractiveDebugger struct {
	enabled     bool
	breakpoints map[string][]int
	watchVars   []string
	mutex       sync.RWMutex
}

// NewInteractiveDebugger creates a new interactive debugger
func NewInteractiveDebugger() *InteractiveDebugger {
	return &InteractiveDebugger{
		enabled:     false,
		breakpoints: make(map[string][]int),
		watchVars:   make([]string, 0),
	}
}

// Enable enables the interactive debugger
func (id *InteractiveDebugger) Enable() {
	id.mutex.Lock()
	defer id.mutex.Unlock()
	id.enabled = true
}

// Disable disables the interactive debugger
func (id *InteractiveDebugger) Disable() {
	id.mutex.Lock()
	defer id.mutex.Unlock()
	id.enabled = false
}

// SetBreakpoint sets a breakpoint at the specified line
func (id *InteractiveDebugger) SetBreakpoint(templateName string, line int) {
	id.mutex.Lock()
	defer id.mutex.Unlock()

	if id.breakpoints[templateName] == nil {
		id.breakpoints[templateName] = make([]int, 0)
	}

	// Check if breakpoint already exists
	for _, existingLine := range id.breakpoints[templateName] {
		if existingLine == line {
			return
		}
	}

	id.breakpoints[templateName] = append(id.breakpoints[templateName], line)
}

// RemoveBreakpoint removes a breakpoint
func (id *InteractiveDebugger) RemoveBreakpoint(templateName string, line int) {
	id.mutex.Lock()
	defer id.mutex.Unlock()

	lines := id.breakpoints[templateName]
	for i, existingLine := range lines {
		if existingLine == line {
			id.breakpoints[templateName] = append(lines[:i], lines[i+1:]...)
			break
		}
	}
}

// Watch adds a variable to watch
func (id *InteractiveDebugger) Watch(varName string) {
	id.mutex.Lock()
	defer id.mutex.Unlock()

	// Check if already watching
	for _, existing := range id.watchVars {
		if existing == varName {
			return
		}
	}

	id.watchVars = append(id.watchVars, varName)
}

// GetWatchedValues returns the values of watched variables
func (id *InteractiveDebugger) GetWatchedValues(ctx Context) map[string]interface{} {
	id.mutex.RLock()
	defer id.mutex.RUnlock()

	result := make(map[string]interface{})
	for _, varName := range id.watchVars {
		if value, ok := ctx.Get(varName); ok {
			result[varName] = value
		} else {
			result[varName] = "<undefined>"
		}
	}

	return result
}

// ShouldBreak returns true if execution should break at this location
func (id *InteractiveDebugger) ShouldBreak(templateName string, line int) bool {
	id.mutex.RLock()
	defer id.mutex.RUnlock()

	if !id.enabled {
		return false
	}

	lines := id.breakpoints[templateName]
	for _, breakpointLine := range lines {
		if breakpointLine == line {
			return true
		}
	}

	return false
}
