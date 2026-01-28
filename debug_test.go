package miya

import (
	"strings"
	"testing"
)

// TestDebugTracer tests the DebugTracer type
func TestDebugTracer(t *testing.T) {
	t.Run("NewDebugTracer", func(t *testing.T) {
		dt := NewDebugTracer()
		if dt == nil {
			t.Fatal("NewDebugTracer returned nil")
		}
		if dt.enabled {
			t.Error("tracer should be disabled by default")
		}
		if dt.level != DebugLevelBasic {
			t.Error("default level should be DebugLevelBasic")
		}
	})

	t.Run("Enable and Disable", func(t *testing.T) {
		dt := NewDebugTracer()

		dt.Enable()
		if !dt.enabled {
			t.Error("tracer should be enabled after Enable()")
		}

		dt.Disable()
		if dt.enabled {
			t.Error("tracer should be disabled after Disable()")
		}
	})

	t.Run("SetLevel", func(t *testing.T) {
		dt := NewDebugTracer()

		dt.SetLevel(DebugLevelVerbose)
		if dt.level != DebugLevelVerbose {
			t.Error("level should be DebugLevelVerbose")
		}

		dt.SetLevel(DebugLevelDetailed)
		if dt.level != DebugLevelDetailed {
			t.Error("level should be DebugLevelDetailed")
		}
	})

	t.Run("AddFilter and RemoveFilter", func(t *testing.T) {
		dt := NewDebugTracer()

		dt.AddFilter("render")
		if !dt.filters["render"] {
			t.Error("filter 'render' should be added")
		}

		dt.AddFilter("evaluate")
		if len(dt.filters) != 2 {
			t.Error("should have 2 filters")
		}

		dt.RemoveFilter("render")
		if dt.filters["render"] {
			t.Error("filter 'render' should be removed")
		}
	})

	t.Run("TraceEvent when disabled", func(t *testing.T) {
		dt := NewDebugTracer()

		dt.TraceEvent("render", "test.html", 1, 1, "test message", nil)

		events := dt.GetEvents()
		if len(events) != 0 {
			t.Error("should not record events when disabled")
		}
	})

	t.Run("TraceEvent when enabled", func(t *testing.T) {
		dt := NewDebugTracer()
		dt.Enable()

		dt.TraceEvent("render", "test.html", 10, 5, "rendering template", map[string]interface{}{"var": "value"})

		events := dt.GetEvents()
		if len(events) != 1 {
			t.Fatalf("should have 1 event, got %d", len(events))
		}
		if events[0].Type != "render" {
			t.Error("event type should be 'render'")
		}
		if events[0].TemplateName != "test.html" {
			t.Error("template name should be 'test.html'")
		}
		if events[0].Line != 10 {
			t.Error("line should be 10")
		}
	})

	t.Run("TraceEvent with filter", func(t *testing.T) {
		dt := NewDebugTracer()
		dt.Enable()
		dt.AddFilter("render")

		dt.TraceEvent("render", "test.html", 1, 1, "allowed", nil)
		dt.TraceEvent("evaluate", "test.html", 2, 1, "filtered out", nil)

		events := dt.GetEvents()
		if len(events) != 1 {
			t.Fatalf("should have 1 event (filtered), got %d", len(events))
		}
		if events[0].Type != "render" {
			t.Error("only 'render' events should be recorded")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		dt := NewDebugTracer()
		dt.Enable()

		dt.TraceEvent("render", "test.html", 1, 1, "event 1", nil)
		dt.TraceEvent("render", "test.html", 2, 1, "event 2", nil)

		if len(dt.GetEvents()) != 2 {
			t.Fatal("should have 2 events before clear")
		}

		dt.Clear()

		if len(dt.GetEvents()) != 0 {
			t.Error("should have 0 events after clear")
		}
	})

	t.Run("GetSummary", func(t *testing.T) {
		dt := NewDebugTracer()
		dt.Enable()

		dt.TraceEvent("render", "test.html", 1, 1, "event 1", nil)
		dt.TraceEvent("evaluate", "test.html", 2, 1, "event 2", nil)
		dt.TraceEvent("render", "other.html", 3, 1, "event 3", nil)

		summary := dt.GetSummary()

		// Summary is a string, check it contains expected info
		if !strings.Contains(summary, "3") {
			t.Error("summary should contain event count")
		}
	})

	t.Run("GetSummary empty", func(t *testing.T) {
		dt := NewDebugTracer()

		summary := dt.GetSummary()

		if !strings.Contains(summary, "No debug events") {
			t.Error("empty summary should indicate no events")
		}
	})

	t.Run("GetDetailedLog", func(t *testing.T) {
		dt := NewDebugTracer()
		dt.Enable()

		dt.TraceEvent("render", "test.html", 1, 1, "rendering", nil)

		log := dt.GetDetailedLog()

		if !strings.Contains(log, "render") {
			t.Error("log should contain event type")
		}
		if !strings.Contains(log, "test.html") {
			t.Error("log should contain template name")
		}
	})
}

// TestPerformanceProfiler tests the PerformanceProfiler type
func TestPerformanceProfiler(t *testing.T) {
	t.Run("NewPerformanceProfiler", func(t *testing.T) {
		pp := NewPerformanceProfiler()
		if pp == nil {
			t.Fatal("NewPerformanceProfiler returned nil")
		}
		if pp.enabled {
			t.Error("profiler should be disabled by default")
		}
	})

	t.Run("Enable and Disable", func(t *testing.T) {
		pp := NewPerformanceProfiler()

		pp.Enable()
		if !pp.enabled {
			t.Error("profiler should be enabled after Enable()")
		}

		pp.Disable()
		if pp.enabled {
			t.Error("profiler should be disabled after Disable()")
		}
	})

	t.Run("StartMeasurement when disabled", func(t *testing.T) {
		pp := NewPerformanceProfiler()

		endFn := pp.StartMeasurement("test")
		endFn() // Should not panic

		measurements := pp.GetMeasurements()
		if len(measurements) != 0 {
			t.Error("should not record measurements when disabled")
		}
	})

	t.Run("StartMeasurement when enabled", func(t *testing.T) {
		pp := NewPerformanceProfiler()
		pp.Enable()

		endFn := pp.StartMeasurement("render")
		// Do some work
		for i := 0; i < 1000; i++ {
			_ = i * 2
		}
		endFn()

		measurements := pp.GetMeasurements()
		if len(measurements) == 0 {
			t.Fatal("should have at least one measurement")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		pp := NewPerformanceProfiler()
		pp.Enable()

		endFn := pp.StartMeasurement("render")
		endFn()

		if len(pp.GetMeasurements()) == 0 {
			t.Fatal("should have measurements before clear")
		}

		pp.Clear()

		if len(pp.GetMeasurements()) != 0 {
			t.Error("should have no measurements after clear")
		}
	})

	t.Run("GetReport", func(t *testing.T) {
		pp := NewPerformanceProfiler()
		pp.Enable()

		endFn := pp.StartMeasurement("render")
		endFn()

		report := pp.GetReport()

		if !strings.Contains(report, "render") {
			t.Error("report should contain operation name")
		}
	})
}

// TestInteractiveDebugger tests the InteractiveDebugger type
func TestInteractiveDebugger(t *testing.T) {
	t.Run("NewInteractiveDebugger", func(t *testing.T) {
		id := NewInteractiveDebugger()
		if id == nil {
			t.Fatal("NewInteractiveDebugger returned nil")
		}
		if id.enabled {
			t.Error("debugger should be disabled by default")
		}
	})

	t.Run("Enable and Disable", func(t *testing.T) {
		id := NewInteractiveDebugger()

		id.Enable()
		if !id.enabled {
			t.Error("debugger should be enabled after Enable()")
		}

		id.Disable()
		if id.enabled {
			t.Error("debugger should be disabled after Disable()")
		}
	})

	t.Run("SetBreakpoint and RemoveBreakpoint", func(t *testing.T) {
		id := NewInteractiveDebugger()

		id.SetBreakpoint("test.html", 10)
		if len(id.breakpoints["test.html"]) != 1 {
			t.Error("should have 1 breakpoint for test.html")
		}

		id.SetBreakpoint("test.html", 20)
		if len(id.breakpoints["test.html"]) != 2 {
			t.Error("should have 2 breakpoints for test.html")
		}

		// Setting same breakpoint again should not duplicate
		id.SetBreakpoint("test.html", 10)
		if len(id.breakpoints["test.html"]) != 2 {
			t.Error("duplicate breakpoint should not be added")
		}

		id.RemoveBreakpoint("test.html", 10)
		if len(id.breakpoints["test.html"]) != 1 {
			t.Error("should have 1 breakpoint after removal")
		}
	})

	t.Run("Watch", func(t *testing.T) {
		id := NewInteractiveDebugger()

		id.Watch("myVar")
		if len(id.watchVars) != 1 {
			t.Error("should have 1 watch variable")
		}

		id.Watch("otherVar")
		if len(id.watchVars) != 2 {
			t.Error("should have 2 watch variables")
		}
	})

	t.Run("ShouldBreak", func(t *testing.T) {
		id := NewInteractiveDebugger()
		id.Enable()

		id.SetBreakpoint("test.html", 10)

		if !id.ShouldBreak("test.html", 10) {
			t.Error("should break at breakpoint")
		}

		if id.ShouldBreak("test.html", 5) {
			t.Error("should not break at non-breakpoint line")
		}

		if id.ShouldBreak("other.html", 10) {
			t.Error("should not break in different template")
		}
	})

	t.Run("ShouldBreak when disabled", func(t *testing.T) {
		id := NewInteractiveDebugger()

		id.SetBreakpoint("test.html", 10)

		if id.ShouldBreak("test.html", 10) {
			t.Error("should not break when disabled")
		}
	})

	t.Run("GetWatchedValues", func(t *testing.T) {
		id := NewInteractiveDebugger()

		id.Watch("name")
		id.Watch("value")

		ctx := NewContext()
		ctx.Set("name", "test")
		ctx.Set("value", 42)
		ctx.Set("other", "ignored")

		values := id.GetWatchedValues(ctx)

		if len(values) != 2 {
			t.Errorf("should have 2 watched values, got %d", len(values))
		}
		if values["name"] != "test" {
			t.Error("watched 'name' should be 'test'")
		}
		if values["value"] != 42 {
			t.Error("watched 'value' should be 42")
		}
	})
}
