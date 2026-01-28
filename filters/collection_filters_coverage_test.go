package filters

import (
	"reflect"
	"testing"
)

// TestFirstFilter tests the FirstFilter function
func TestFirstFilter(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result, err := FirstFilter(nil)
		if err != nil {
			t.Fatalf("FirstFilter(nil) returned error: %v", err)
		}
		if result != nil {
			t.Errorf("FirstFilter(nil) = %v, want nil", result)
		}
	})

	t.Run("interface slice", func(t *testing.T) {
		result, err := FirstFilter([]interface{}{1, 2, 3})
		if err != nil {
			t.Fatalf("FirstFilter returned error: %v", err)
		}
		if result != 1 {
			t.Errorf("FirstFilter = %v, want 1", result)
		}
	})

	t.Run("empty interface slice", func(t *testing.T) {
		result, err := FirstFilter([]interface{}{})
		if err != nil {
			t.Fatalf("FirstFilter returned error: %v", err)
		}
		if result != nil {
			t.Errorf("FirstFilter([]) = %v, want nil", result)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		result, err := FirstFilter([]string{"a", "b", "c"})
		if err != nil {
			t.Fatalf("FirstFilter returned error: %v", err)
		}
		if result != "a" {
			t.Errorf("FirstFilter = %v, want 'a'", result)
		}
	})

	t.Run("string", func(t *testing.T) {
		result, err := FirstFilter("hello")
		if err != nil {
			t.Fatalf("FirstFilter returned error: %v", err)
		}
		if result != "h" {
			t.Errorf("FirstFilter('hello') = %v, want 'h'", result)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		result, err := FirstFilter("")
		if err != nil {
			t.Fatalf("FirstFilter returned error: %v", err)
		}
		if result != nil {
			t.Errorf("FirstFilter('') = %v, want nil", result)
		}
	})

	t.Run("int slice via reflection", func(t *testing.T) {
		result, err := FirstFilter([]int{10, 20, 30})
		if err != nil {
			t.Fatalf("FirstFilter returned error: %v", err)
		}
		if result != 10 {
			t.Errorf("FirstFilter = %v, want 10", result)
		}
	})

	t.Run("array via reflection", func(t *testing.T) {
		result, err := FirstFilter([3]int{100, 200, 300})
		if err != nil {
			t.Fatalf("FirstFilter returned error: %v", err)
		}
		if result != 100 {
			t.Errorf("FirstFilter = %v, want 100", result)
		}
	})
}

// TestLastFilter tests the LastFilter function
func TestLastFilter(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result, err := LastFilter(nil)
		if err != nil {
			t.Fatalf("LastFilter(nil) returned error: %v", err)
		}
		if result != nil {
			t.Errorf("LastFilter(nil) = %v, want nil", result)
		}
	})

	t.Run("interface slice", func(t *testing.T) {
		result, err := LastFilter([]interface{}{1, 2, 3})
		if err != nil {
			t.Fatalf("LastFilter returned error: %v", err)
		}
		if result != 3 {
			t.Errorf("LastFilter = %v, want 3", result)
		}
	})

	t.Run("empty interface slice", func(t *testing.T) {
		result, err := LastFilter([]interface{}{})
		if err != nil {
			t.Fatalf("LastFilter returned error: %v", err)
		}
		if result != nil {
			t.Errorf("LastFilter([]) = %v, want nil", result)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		result, err := LastFilter([]string{"a", "b", "c"})
		if err != nil {
			t.Fatalf("LastFilter returned error: %v", err)
		}
		if result != "c" {
			t.Errorf("LastFilter = %v, want 'c'", result)
		}
	})

	t.Run("string", func(t *testing.T) {
		result, err := LastFilter("hello")
		if err != nil {
			t.Fatalf("LastFilter returned error: %v", err)
		}
		if result != "o" {
			t.Errorf("LastFilter('hello') = %v, want 'o'", result)
		}
	})

	t.Run("int slice via reflection", func(t *testing.T) {
		result, err := LastFilter([]int{10, 20, 30})
		if err != nil {
			t.Fatalf("LastFilter returned error: %v", err)
		}
		if result != 30 {
			t.Errorf("LastFilter = %v, want 30", result)
		}
	})

	t.Run("array via reflection", func(t *testing.T) {
		result, err := LastFilter([3]int{100, 200, 300})
		if err != nil {
			t.Fatalf("LastFilter returned error: %v", err)
		}
		if result != 300 {
			t.Errorf("LastFilter = %v, want 300", result)
		}
	})
}

// TestLengthFilter tests the LengthFilter function
func TestLengthFilter(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result, err := LengthFilter(nil)
		if err != nil {
			t.Fatalf("LengthFilter(nil) returned error: %v", err)
		}
		if result != 0 {
			t.Errorf("LengthFilter(nil) = %v, want 0", result)
		}
	})

	t.Run("string", func(t *testing.T) {
		result, err := LengthFilter("hello")
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 5 {
			t.Errorf("LengthFilter('hello') = %v, want 5", result)
		}
	})

	t.Run("unicode string", func(t *testing.T) {
		result, err := LengthFilter("日本語")
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 3 {
			t.Errorf("LengthFilter('日本語') = %v, want 3", result)
		}
	})

	t.Run("interface slice", func(t *testing.T) {
		result, err := LengthFilter([]interface{}{1, 2, 3})
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 3 {
			t.Errorf("LengthFilter = %v, want 3", result)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		result, err := LengthFilter([]string{"a", "b"})
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 2 {
			t.Errorf("LengthFilter = %v, want 2", result)
		}
	})

	t.Run("map[string]interface{}", func(t *testing.T) {
		result, err := LengthFilter(map[string]interface{}{"a": 1, "b": 2})
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 2 {
			t.Errorf("LengthFilter = %v, want 2", result)
		}
	})

	t.Run("map[string]string", func(t *testing.T) {
		result, err := LengthFilter(map[string]string{"x": "1"})
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 1 {
			t.Errorf("LengthFilter = %v, want 1", result)
		}
	})

	t.Run("int slice via reflection", func(t *testing.T) {
		result, err := LengthFilter([]int{1, 2, 3, 4})
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 4 {
			t.Errorf("LengthFilter = %v, want 4", result)
		}
	})

	t.Run("array via reflection", func(t *testing.T) {
		result, err := LengthFilter([5]int{1, 2, 3, 4, 5})
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 5 {
			t.Errorf("LengthFilter = %v, want 5", result)
		}
	})

	t.Run("map via reflection", func(t *testing.T) {
		result, err := LengthFilter(map[int]string{1: "a", 2: "b"})
		if err != nil {
			t.Fatalf("LengthFilter returned error: %v", err)
		}
		if result != 2 {
			t.Errorf("LengthFilter = %v, want 2", result)
		}
	})
}

// TestReverseFilter tests the ReverseFilter function
func TestReverseFilter(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		// ReverseFilter returns error for nil
		_, err := ReverseFilter(nil)
		if err == nil {
			t.Error("ReverseFilter(nil) should return error")
		}
	})

	t.Run("interface slice", func(t *testing.T) {
		result, err := ReverseFilter([]interface{}{1, 2, 3})
		if err != nil {
			t.Fatalf("ReverseFilter returned error: %v", err)
		}
		expected := []interface{}{3, 2, 1}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("ReverseFilter = %v, want %v", result, expected)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		result, err := ReverseFilter([]string{"a", "b", "c"})
		if err != nil {
			t.Fatalf("ReverseFilter returned error: %v", err)
		}
		expected := []string{"c", "b", "a"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("ReverseFilter = %v, want %v", result, expected)
		}
	})

	t.Run("string", func(t *testing.T) {
		result, err := ReverseFilter("hello")
		if err != nil {
			t.Fatalf("ReverseFilter returned error: %v", err)
		}
		if result != "olleh" {
			t.Errorf("ReverseFilter('hello') = %v, want 'olleh'", result)
		}
	})

	t.Run("int slice not supported", func(t *testing.T) {
		// ReverseFilter doesn't support typed slices via reflection
		_, err := ReverseFilter([]int{1, 2, 3})
		if err == nil {
			t.Error("ReverseFilter should return error for []int")
		}
	})
}

// TestSliceFilter tests the SliceFilter function
func TestSliceFilter(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		// SliceFilter returns error for nil
		_, err := SliceFilter(nil, 0, 1)
		if err == nil {
			t.Error("SliceFilter(nil) should return error")
		}
	})

	t.Run("interface slice with start and stop", func(t *testing.T) {
		result, err := SliceFilter([]interface{}{1, 2, 3, 4, 5}, 1, 4)
		if err != nil {
			t.Fatalf("SliceFilter returned error: %v", err)
		}
		expected := []interface{}{2, 3, 4}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("SliceFilter = %v, want %v", result, expected)
		}
	})

	t.Run("interface slice with only start", func(t *testing.T) {
		result, err := SliceFilter([]interface{}{1, 2, 3, 4, 5}, 2)
		if err != nil {
			t.Fatalf("SliceFilter returned error: %v", err)
		}
		expected := []interface{}{3, 4, 5}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("SliceFilter = %v, want %v", result, expected)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		result, err := SliceFilter([]string{"a", "b", "c", "d"}, 1, 3)
		if err != nil {
			t.Fatalf("SliceFilter returned error: %v", err)
		}
		expected := []string{"b", "c"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("SliceFilter = %v, want %v", result, expected)
		}
	})

	t.Run("string", func(t *testing.T) {
		result, err := SliceFilter("hello world", 0, 5)
		if err != nil {
			t.Fatalf("SliceFilter returned error: %v", err)
		}
		if result != "hello" {
			t.Errorf("SliceFilter = %v, want 'hello'", result)
		}
	})

	t.Run("negative indices", func(t *testing.T) {
		result, err := SliceFilter([]interface{}{1, 2, 3, 4, 5}, -2)
		if err != nil {
			t.Fatalf("SliceFilter returned error: %v", err)
		}
		expected := []interface{}{4, 5}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("SliceFilter = %v, want %v", result, expected)
		}
	})

	t.Run("int slice not supported", func(t *testing.T) {
		// SliceFilter doesn't support typed slices via reflection
		_, err := SliceFilter([]int{10, 20, 30, 40}, 1, 3)
		if err == nil {
			t.Error("SliceFilter should return error for []int")
		}
	})
}

// TestListFilter tests the ListFilter function
func TestListFilter(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		result, err := ListFilter(nil)
		if err != nil {
			t.Fatalf("ListFilter(nil) returned error: %v", err)
		}
		// nil returns empty list or nil - just verify no error
		_ = result
	})

	t.Run("interface slice", func(t *testing.T) {
		input := []interface{}{1, 2, 3}
		result, err := ListFilter(input)
		if err != nil {
			t.Fatalf("ListFilter returned error: %v", err)
		}
		if !reflect.DeepEqual(result, input) {
			t.Errorf("ListFilter = %v, want %v", result, input)
		}
	})

	t.Run("string slice", func(t *testing.T) {
		result, err := ListFilter([]string{"a", "b"})
		if err != nil {
			t.Fatalf("ListFilter returned error: %v", err)
		}
		expected := []interface{}{"a", "b"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("ListFilter = %v, want %v", result, expected)
		}
	})

	t.Run("string to list", func(t *testing.T) {
		result, err := ListFilter("abc")
		if err != nil {
			t.Fatalf("ListFilter returned error: %v", err)
		}
		expected := []interface{}{"a", "b", "c"}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("ListFilter = %v, want %v", result, expected)
		}
	})

	t.Run("int slice via reflection", func(t *testing.T) {
		result, err := ListFilter([]int{1, 2, 3})
		if err != nil {
			t.Fatalf("ListFilter returned error: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("ListFilter should return []interface{}, got %T", result)
		}
		if len(resultSlice) != 3 {
			t.Errorf("ListFilter result length = %d, want 3", len(resultSlice))
		}
	})

	t.Run("map", func(t *testing.T) {
		result, err := ListFilter(map[string]int{"a": 1, "b": 2})
		if err != nil {
			t.Fatalf("ListFilter returned error: %v", err)
		}
		// Map conversion to list may return keys only
		_ = result
	})
}

// TestItemsFilter tests the ItemsFilter function
func TestItemsFilter(t *testing.T) {
	t.Run("map[string]interface{}", func(t *testing.T) {
		result, err := ItemsFilter(map[string]interface{}{"a": 1, "b": 2})
		if err != nil {
			t.Fatalf("ItemsFilter returned error: %v", err)
		}
		// ItemsFilter returns []interface{} where each element is a []interface{}{key, value}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("ItemsFilter should return []interface{}, got %T", result)
		}
		if len(resultSlice) != 2 {
			t.Errorf("ItemsFilter result length = %d, want 2", len(resultSlice))
		}
	})

	t.Run("map[string]string", func(t *testing.T) {
		result, err := ItemsFilter(map[string]string{"x": "1"})
		if err != nil {
			t.Fatalf("ItemsFilter returned error: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("ItemsFilter should return []interface{}, got %T", result)
		}
		if len(resultSlice) != 1 {
			t.Errorf("ItemsFilter result length = %d, want 1", len(resultSlice))
		}
	})

	t.Run("generic map not supported", func(t *testing.T) {
		// ItemsFilter doesn't support maps with non-string keys via reflection
		_, err := ItemsFilter(map[int]string{1: "a", 2: "b"})
		if err == nil {
			t.Error("ItemsFilter should return error for map[int]string")
		}
	})

	t.Run("non-map value", func(t *testing.T) {
		_, err := ItemsFilter("not a map")
		if err == nil {
			t.Error("ItemsFilter should return error for non-map")
		}
	})
}

// TestKeysFilter tests the KeysFilter function
func TestKeysFilter(t *testing.T) {
	t.Run("map[string]interface{}", func(t *testing.T) {
		result, err := KeysFilter(map[string]interface{}{"a": 1, "b": 2})
		if err != nil {
			t.Fatalf("KeysFilter returned error: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("KeysFilter should return []interface{}, got %T", result)
		}
		if len(resultSlice) != 2 {
			t.Errorf("KeysFilter result length = %d, want 2", len(resultSlice))
		}
	})

	t.Run("map[string]string", func(t *testing.T) {
		result, err := KeysFilter(map[string]string{"x": "1", "y": "2"})
		if err != nil {
			t.Fatalf("KeysFilter returned error: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("KeysFilter should return []interface{}, got %T", result)
		}
		if len(resultSlice) != 2 {
			t.Errorf("KeysFilter result length = %d, want 2", len(resultSlice))
		}
	})

	t.Run("generic map not supported", func(t *testing.T) {
		// KeysFilter doesn't support maps with non-string keys via reflection
		_, err := KeysFilter(map[int]string{1: "a", 2: "b"})
		if err == nil {
			t.Error("KeysFilter should return error for map[int]string")
		}
	})

	t.Run("non-map value", func(t *testing.T) {
		_, err := KeysFilter("not a map")
		if err == nil {
			t.Error("KeysFilter should return error for non-map")
		}
	})
}

// TestValuesFilter tests the ValuesFilter function
func TestValuesFilter(t *testing.T) {
	t.Run("map[string]interface{}", func(t *testing.T) {
		result, err := ValuesFilter(map[string]interface{}{"a": 1, "b": 2})
		if err != nil {
			t.Fatalf("ValuesFilter returned error: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("ValuesFilter should return []interface{}, got %T", result)
		}
		if len(resultSlice) != 2 {
			t.Errorf("ValuesFilter result length = %d, want 2", len(resultSlice))
		}
	})

	t.Run("map[string]string", func(t *testing.T) {
		result, err := ValuesFilter(map[string]string{"x": "1", "y": "2"})
		if err != nil {
			t.Fatalf("ValuesFilter returned error: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("ValuesFilter should return []interface{}, got %T", result)
		}
		if len(resultSlice) != 2 {
			t.Errorf("ValuesFilter result length = %d, want 2", len(resultSlice))
		}
	})

	t.Run("generic map not supported", func(t *testing.T) {
		// ValuesFilter doesn't support maps with non-string keys via reflection
		_, err := ValuesFilter(map[int]string{1: "a", 2: "b"})
		if err == nil {
			t.Error("ValuesFilter should return error for map[int]string")
		}
	})

	t.Run("non-map value", func(t *testing.T) {
		_, err := ValuesFilter("not a map")
		if err == nil {
			t.Error("ValuesFilter should return error for non-map")
		}
	})
}

// TestToInt tests the ToInt helper function
func TestToInt(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
		hasError bool
	}{
		{"int", 42, 42, false},
		{"int64", int64(100), 100, false},
		{"float64", 3.14, 3, false},
		{"string int", "123", 123, false},
		{"string float", "45.67", 45, false},
		{"invalid string", "abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToInt(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("ToInt(%v) should return error", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("ToInt(%v) returned error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("ToInt(%v) = %d, want %d", tt.input, result, tt.expected)
				}
			}
		})
	}

	// Bool is not supported by ToInt
	t.Run("bool not supported", func(t *testing.T) {
		_, err := ToInt(true)
		if err == nil {
			t.Error("ToInt(bool) should return error")
		}
	})
}

// TestToInterfaceSlice tests the toInterfaceSlice helper function
func TestToInterfaceSlice(t *testing.T) {
	t.Run("int slice", func(t *testing.T) {
		result, err := toInterfaceSlice([]int{1, 2, 3})
		if err != nil {
			t.Fatalf("toInterfaceSlice returned error: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("toInterfaceSlice result length = %d, want 3", len(result))
		}
	})

	t.Run("string slice", func(t *testing.T) {
		result, err := toInterfaceSlice([]string{"a", "b"})
		if err != nil {
			t.Fatalf("toInterfaceSlice returned error: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("toInterfaceSlice result length = %d, want 2", len(result))
		}
	})

	t.Run("interface slice", func(t *testing.T) {
		input := []interface{}{1, "a", true}
		result, err := toInterfaceSlice(input)
		if err != nil {
			t.Fatalf("toInterfaceSlice returned error: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("toInterfaceSlice result length = %d, want 3", len(result))
		}
	})

	t.Run("string converted to char slice", func(t *testing.T) {
		// Strings are specially handled - converted to slice of characters
		result, err := toInterfaceSlice("abc")
		if err != nil {
			t.Fatalf("toInterfaceSlice(string) returned error: %v", err)
		}
		if len(result) != 3 {
			t.Errorf("toInterfaceSlice(string) result length = %d, want 3", len(result))
		}
	})

	t.Run("non-slice", func(t *testing.T) {
		// Use an int which is not a slice/array/string
		_, err := toInterfaceSlice(42)
		if err == nil {
			t.Error("toInterfaceSlice(int) should return error")
		}
	})

	t.Run("nil", func(t *testing.T) {
		result, err := toInterfaceSlice(nil)
		if err != nil {
			t.Fatalf("toInterfaceSlice(nil) returned error: %v", err)
		}
		if len(result) != 0 {
			t.Error("toInterfaceSlice(nil) should return empty slice")
		}
	})
}
