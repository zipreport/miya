package filters

import (
	"reflect"
	"testing"
)

func TestFilterRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test that built-in filters are registered
	builtins := []string{
		"upper", "lower", "capitalize", "title", "trim",
		"escape", "safe", "length", "first", "last",
		"join", "sort", "reverse", "unique", "abs",
		"round", "int", "float", "sum", "default",
	}

	for _, name := range builtins {
		if _, ok := registry.Get(name); !ok {
			t.Errorf("built-in filter %q not registered", name)
		}
	}

	// Test custom filter registration
	testFilter := func(value interface{}, args ...interface{}) (interface{}, error) {
		return "test", nil
	}

	err := registry.Register("test_filter", testFilter)
	if err != nil {
		t.Errorf("failed to register custom filter: %v", err)
	}

	// Test duplicate registration
	err = registry.Register("test_filter", testFilter)
	if err == nil {
		t.Error("expected error when registering duplicate filter")
	}

	// Test filter retrieval
	fn, ok := registry.Get("test_filter")
	if !ok {
		t.Error("failed to retrieve registered filter")
	}

	result, err := fn("input", "arg1")
	if err != nil {
		t.Errorf("filter execution failed: %v", err)
	}
	if result != "test" {
		t.Errorf("expected 'test', got %v", result)
	}

	// Test unregister
	if !registry.Unregister("test_filter") {
		t.Error("failed to unregister filter")
	}

	if _, ok := registry.Get("test_filter"); ok {
		t.Error("filter still exists after unregister")
	}
}

func TestStringFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterFunc
		input    interface{}
		args     []interface{}
		expected interface{}
		hasError bool
	}{
		{"upper", UpperFilter, "hello", nil, "HELLO", false},
		{"lower", LowerFilter, "WORLD", nil, "world", false},
		{"capitalize", CapitalizeFilter, "hello world", nil, "Hello world", false},
		{"title", TitleFilter, "hello world", nil, "Hello World", false},
		{"trim", TrimFilter, "  spaced  ", nil, "spaced", false},
		{"trim with chars", TrimFilter, "...trimmed...", []interface{}{"."}, "trimmed", false},
		{"replace", ReplaceFilter, "hello world", []interface{}{"world", "Go"}, "hello Go", false},
		{"replace count", ReplaceFilter, "foo foo foo", []interface{}{"foo", "bar", 2}, "bar bar foo", false},
		{"truncate", TruncateFilter, "this is a long string", []interface{}{10}, "this is a...", false},
		{"truncate killwords", TruncateFilter, "this is a long string", []interface{}{10, true}, "this is a ...", false},
		{"center", CenterFilter, "test", []interface{}{10}, "   test   ", false},
		{"indent", IndentFilter, "line1\nline2", []interface{}{2}, "line1\n  line2", false},
		{"format", FormatFilter, "Hello %s", []interface{}{"World"}, "Hello World", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.filter(tt.input, tt.args...)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHTMLFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterFunc
		input    interface{}
		args     []interface{}
		expected interface{}
		hasError bool
	}{
		{"escape", EscapeFilter, "<script>alert('xss')</script>", nil, "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;", false},
		{"safe", SafeFilter, "<b>bold</b>", nil, SafeValue{Value: "<b>bold</b>"}, false},
		{"urlencode", URLEncodeFilter, "hello world", nil, "hello+world", false},
		{"striptags", StripTagsFilter, "<p>Hello <b>world</b></p>", nil, "Hello world", false},
		{"filesizeformat", FileSizeFormatFilter, 1024, nil, "1.0 KB", false},
		{"filesizeformat binary", FileSizeFormatFilter, 1024, []interface{}{true}, "1.0 KiB", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.filter(tt.input, tt.args...)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCollectionFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterFunc
		input    interface{}
		args     []interface{}
		expected interface{}
		hasError bool
	}{
		{"first", FirstFilter, []interface{}{"a", "b", "c"}, nil, "a", false},
		{"last", LastFilter, []interface{}{"a", "b", "c"}, nil, "c", false},
		{"length", LengthFilter, []interface{}{"a", "b", "c"}, nil, 3, false},
		{"length string", LengthFilter, "hello", nil, 5, false},
		{"join", JoinFilter, []interface{}{"a", "b", "c"}, []interface{}{", "}, "a, b, c", false},
		{"reverse", ReverseFilter, []interface{}{"a", "b", "c"}, nil, []interface{}{"c", "b", "a"}, false},
		{"unique", UniqueFilter, []interface{}{"a", "b", "a", "c"}, nil, []interface{}{"a", "b", "c"}, false},
		{"slice", SliceFilter, []interface{}{"a", "b", "c", "d"}, []interface{}{1, 3}, []interface{}{"b", "c"}, false},
		{"batch", BatchFilter, []interface{}{"a", "b", "c", "d", "e"}, []interface{}{2}, [][]interface{}{{"a", "b"}, {"c", "d"}, {"e"}}, false},
		{"list", ListFilter, "abc", nil, []interface{}{"a", "b", "c"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.filter(tt.input, tt.args...)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestNumericFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterFunc
		input    interface{}
		args     []interface{}
		expected interface{}
		hasError bool
	}{
		{"abs positive", AbsFilter, 5, nil, 5, false},
		{"abs negative", AbsFilter, -5, nil, 5, false},
		{"abs float", AbsFilter, -3.14, nil, 3.14, false},
		{"round", RoundFilter, 3.14159, []interface{}{2}, "3.14", false},
		{"round ceil", RoundFilter, 3.14159, []interface{}{2, "ceil"}, "3.15", false},
		{"round floor", RoundFilter, 3.14159, []interface{}{2, "floor"}, "3.14", false},
		{"int", IntFilter, "123", nil, 123, false},
		{"int default", IntFilter, "invalid", []interface{}{0}, 0, false},
		{"float", FloatFilter, "3.14", nil, 3.14, false},
		{"sum", SumFilter, []interface{}{1, 2, 3, 4}, nil, 10.0, false},
		{"sum with start", SumFilter, []interface{}{1, 2, 3}, []interface{}{10}, 16.0, false},
		{"min", MinFilter, []interface{}{3, 1, 4, 1, 5}, nil, 1, false},
		{"max", MaxFilter, []interface{}{3, 1, 4, 1, 5}, nil, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.filter(tt.input, tt.args...)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestUtilityFilters(t *testing.T) {
	tests := []struct {
		name     string
		filter   FilterFunc
		input    interface{}
		args     []interface{}
		expected interface{}
		hasError bool
	}{
		{"default with value", DefaultFilter, "exists", []interface{}{"fallback"}, "exists", false},
		{"default without value", DefaultFilter, "", []interface{}{"fallback"}, "fallback", false},
		{"default boolean true", DefaultFilter, true, []interface{}{"fallback", true}, true, false},
		{"default boolean false", DefaultFilter, false, []interface{}{"fallback", true}, "fallback", false},
		{"tojson", ToJSONFilter, map[string]interface{}{"key": "value"}, nil, `{"key":"value"}`, false},
		{"tojson indent", ToJSONFilter, map[string]interface{}{"key": "value"}, []interface{}{2}, "{\n  \"key\": \"value\"\n}", false},
		{"fromjson", FromJSONFilter, `{"key":"value"}`, nil, map[string]interface{}{"key": "value"}, false},
		{"string", StringFilter, 123, nil, "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.filter(tt.input, tt.args...)

			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	// Test ToString
	tests := []struct {
		input    interface{}
		expected string
	}{
		{nil, ""},
		{"hello", "hello"},
		{123, "123"},
		{true, "true"},
		{3.14, "3.14"},
		{[]byte("bytes"), "bytes"},
	}

	for _, tt := range tests {
		result := ToString(tt.input)
		if result != tt.expected {
			t.Errorf("ToString(%v): expected %q, got %q", tt.input, tt.expected, result)
		}
	}

	// Test ToBool
	boolTests := []struct {
		input    interface{}
		expected bool
	}{
		{nil, false},
		{true, true},
		{false, false},
		{"", false},
		{"hello", true},
		{0, false},
		{1, true},
		{0.0, false},
		{3.14, true},
		{[]interface{}{}, false},
		{[]interface{}{1, 2}, true},
		{map[string]interface{}{}, false},
		{map[string]interface{}{"key": "value"}, true},
	}

	for _, tt := range boolTests {
		result := ToBool(tt.input)
		if result != tt.expected {
			t.Errorf("ToBool(%v): expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestXMLAttrFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			"simple attributes",
			map[string]interface{}{"class": "btn", "id": "submit"},
			` class="btn" id="submit"`,
		},
		{
			"boolean attributes",
			map[string]interface{}{"disabled": true, "hidden": false, "readonly": true},
			` disabled readonly`,
		},
		{
			"mixed attributes",
			map[string]interface{}{"class": "btn", "disabled": true, "data-value": "test"},
			` class="btn" data-value="test" disabled`,
		},
		{
			"empty map",
			map[string]interface{}{},
			"",
		},
		{
			"nil values",
			map[string]interface{}{"class": "btn", "empty": nil, "id": "test"},
			` class="btn" id="test"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := XMLAttrFilter(tt.input)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
