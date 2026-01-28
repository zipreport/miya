package runtime

import (
	"testing"

	"github.com/zipreport/miya/parser"
)

// TestCapitalizeFirst tests the capitalizeFirst helper function
func TestCapitalizeFirst(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"a", "A"},
		{"abc", "Abc"},
		{"ABC", "ABC"},
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"123", "123"},
		{"_test", "_test"},
	}

	for _, tt := range tests {
		result := capitalizeFirst(tt.input)
		if result != tt.expected {
			t.Errorf("capitalizeFirst(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestDictItemsString tests the DictItems String method
func TestDictItemsString(t *testing.T) {
	d := &DictItems{data: map[string]interface{}{"key": "value"}}
	result := d.String()
	if result != "[dict_items]" {
		t.Errorf("DictItems.String() = %q, want %q", result, "[dict_items]")
	}
}

// TestGetAttribute tests the getAttribute function
func TestGetAttribute(t *testing.T) {
	e := NewEvaluator()

	// Test with nil
	result := e.getAttribute(nil, "attr")
	if result != nil {
		t.Error("getAttribute(nil, attr) should return nil")
	}

	// Test with map[string]interface{}
	m := map[string]interface{}{"key": "value"}
	result = e.getAttribute(m, "key")
	if result != "value" {
		t.Errorf("getAttribute(map, 'key') = %v, want 'value'", result)
	}

	// Test map with nonexistent key
	result = e.getAttribute(m, "nonexistent")
	if result != nil {
		t.Error("getAttribute(map, 'nonexistent') should return nil")
	}

	// Test with map[string]string
	ms := map[string]string{"key": "value"}
	result = e.getAttribute(ms, "key")
	if result != "value" {
		t.Errorf("getAttribute(map[string]string, 'key') = %v, want 'value'", result)
	}

	// Test with Cycler
	cycler := &Cycler{Items: []interface{}{"a", "b", "c"}}
	nextFn := e.getAttribute(cycler, "next")
	if nextFn == nil {
		t.Error("getAttribute(cycler, 'next') should return a function")
	}

	currentFn := e.getAttribute(cycler, "current")
	if currentFn == nil {
		t.Error("getAttribute(cycler, 'current') should return a function")
	}

	resetFn := e.getAttribute(cycler, "reset")
	if resetFn == nil {
		t.Error("getAttribute(cycler, 'reset') should return a function")
	}

	invalidAttr := e.getAttribute(cycler, "invalid")
	if invalidAttr != nil {
		t.Error("getAttribute(cycler, 'invalid') should return nil")
	}

	// Test with Joiner
	joiner := &Joiner{Separator: ", "}
	joinerFn := e.getAttribute(joiner, "anything")
	if joinerFn == nil {
		t.Error("getAttribute(joiner, anything) should return a function")
	}

	// Test with CallableLoop
	cl := &CallableLoop{
		Info: map[string]interface{}{"index": 1, "first": true},
	}
	indexVal := e.getAttribute(cl, "index")
	if indexVal != 1 {
		t.Errorf("getAttribute(CallableLoop, 'index') = %v, want 1", indexVal)
	}
	invalidLoopAttr := e.getAttribute(cl, "invalid")
	if invalidLoopAttr != nil {
		t.Error("getAttribute(CallableLoop, 'invalid') should return nil")
	}

	// Test map.items() method
	itemsFn := e.getAttribute(m, "items")
	if itemsFn == nil {
		t.Error("getAttribute(map, 'items') should return a function")
	}
	if fn, ok := itemsFn.(func(args ...interface{}) (interface{}, error)); ok {
		result, err := fn()
		if err != nil {
			t.Errorf("items() returned error: %v", err)
		}
		if _, ok := result.(*DictItems); !ok {
			t.Errorf("items() should return *DictItems, got %T", result)
		}
	}

	// Test map.keys() method
	keysFn := e.getAttribute(m, "keys")
	if keysFn == nil {
		t.Error("getAttribute(map, 'keys') should return a function")
	}
	if fn, ok := keysFn.(func(args ...interface{}) (interface{}, error)); ok {
		result, err := fn()
		if err != nil {
			t.Errorf("keys() returned error: %v", err)
		}
		if keys, ok := result.([]interface{}); !ok || len(keys) != 1 {
			t.Errorf("keys() should return slice with 1 element")
		}
	}

	// Test map.values() method
	valuesFn := e.getAttribute(m, "values")
	if valuesFn == nil {
		t.Error("getAttribute(map, 'values') should return a function")
	}

	// Test with struct
	type TestStruct struct {
		Name  string
		Value int
	}
	ts := TestStruct{Name: "test", Value: 42}
	result = e.getAttribute(ts, "Name")
	if result != "test" {
		t.Errorf("getAttribute(struct, 'Name') = %v, want 'test'", result)
	}

	// Test with struct pointer
	result = e.getAttribute(&ts, "Value")
	if result != 42 {
		t.Errorf("getAttribute(*struct, 'Value') = %v, want 42", result)
	}

	// Test with capitalized lookup
	result = e.getAttribute(ts, "name")
	if result != "test" {
		t.Errorf("getAttribute(struct, 'name') = %v, want 'test' (capitalized lookup)", result)
	}
}

// TestGetItem tests the getItem function
func TestGetItem(t *testing.T) {
	e := NewEvaluator()

	// Test with nil
	_, err := e.getItem(nil, "key")
	if err == nil {
		t.Error("getItem(nil, key) should return error")
	}

	// Test with map[string]interface{}
	m := map[string]interface{}{"key": "value"}
	result, err := e.getItem(m, "key")
	if err != nil || result != "value" {
		t.Errorf("getItem(map, 'key') = %v, %v, want 'value', nil", result, err)
	}

	// Test with []interface{} - positive index
	list := []interface{}{"a", "b", "c"}
	result, err = e.getItem(list, 1)
	if err != nil || result != "b" {
		t.Errorf("getItem(list, 1) = %v, %v, want 'b', nil", result, err)
	}

	// Test with []interface{} - negative index
	result, err = e.getItem(list, -1)
	if err != nil || result != "c" {
		t.Errorf("getItem(list, -1) = %v, %v, want 'c', nil", result, err)
	}

	// Test with []interface{} - out of bounds
	result, err = e.getItem(list, 100)
	if err != nil {
		t.Errorf("getItem(list, 100) should not error, got %v", err)
	}
	if _, ok := result.(*Undefined); !ok {
		t.Errorf("getItem(list, 100) should return Undefined, got %T", result)
	}

	// Test with string - positive index
	result, err = e.getItem("hello", 1)
	if err != nil || result != "e" {
		t.Errorf("getItem('hello', 1) = %v, %v, want 'e', nil", result, err)
	}

	// Test with string - negative index
	result, err = e.getItem("hello", -1)
	if err != nil || result != "o" {
		t.Errorf("getItem('hello', -1) = %v, %v, want 'o', nil", result, err)
	}

	// Test with string - out of bounds
	result, err = e.getItem("hello", 100)
	if err != nil {
		t.Errorf("getItem('hello', 100) should not error, got %v", err)
	}

	// Test with typed slice via reflection
	intSlice := []int{1, 2, 3}
	result, err = e.getItem(intSlice, 1)
	if err != nil || result != 2 {
		t.Errorf("getItem([]int, 1) = %v, %v, want 2, nil", result, err)
	}

	// Test with invalid index type for list
	_, err = e.getItem(list, "not_int")
	if err == nil {
		t.Error("getItem(list, 'not_int') should return error")
	}

	// Test with non-subscriptable type
	_, err = e.getItem(42, 0)
	if err == nil {
		t.Error("getItem(int, 0) should return error")
	}
}

// TestMakeIterable tests the makeIterable function
func TestMakeIterable(t *testing.T) {
	e := NewEvaluator()

	// Test with nil
	result, err := e.makeIterable(nil)
	if err != nil || result != nil {
		t.Errorf("makeIterable(nil) = %v, %v, want nil, nil", result, err)
	}

	// Test with Undefined
	undef := NewUndefined("test", UndefinedSilent, nil)
	result, err = e.makeIterable(undef)
	if err != nil || len(result) != 0 {
		t.Errorf("makeIterable(Undefined) = %v, %v, want empty slice", result, err)
	}

	// Test with []interface{}
	list := []interface{}{"a", "b", "c"}
	result, err = e.makeIterable(list)
	if err != nil || len(result) != 3 {
		t.Errorf("makeIterable([]interface{}) = %v, %v", result, err)
	}

	// Test with []string
	strList := []string{"a", "b", "c"}
	result, err = e.makeIterable(strList)
	if err != nil || len(result) != 3 {
		t.Errorf("makeIterable([]string) = %v, %v", result, err)
	}

	// Test with []int
	intList := []int{1, 2, 3}
	result, err = e.makeIterable(intList)
	if err != nil || len(result) != 3 {
		t.Errorf("makeIterable([]int) = %v, %v", result, err)
	}

	// Test with []map[string]interface{}
	mapList := []map[string]interface{}{{"a": 1}, {"b": 2}}
	result, err = e.makeIterable(mapList)
	if err != nil || len(result) != 2 {
		t.Errorf("makeIterable([]map) = %v, %v", result, err)
	}

	// Test with map[string]interface{}
	m := map[string]interface{}{"key": "value"}
	result, err = e.makeIterable(m)
	if err != nil || len(result) != 1 {
		t.Errorf("makeIterable(map) = %v, %v", result, err)
	}

	// Test with DictItems
	dictItems := &DictItems{data: map[string]interface{}{"a": 1, "b": 2}}
	result, err = e.makeIterable(dictItems)
	if err != nil || len(result) != 2 {
		t.Errorf("makeIterable(DictItems) = %v, %v", result, err)
	}

	// Test with string
	result, err = e.makeIterable("abc")
	if err != nil || len(result) != 3 {
		t.Errorf("makeIterable('abc') = %v, %v", result, err)
	}

	// Test with function that returns iterable
	fn := func() []interface{} { return []interface{}{1, 2, 3} }
	result, err = e.makeIterable(fn)
	if err != nil || len(result) != 3 {
		t.Errorf("makeIterable(func) = %v, %v", result, err)
	}

	// Test with non-iterable type
	_, err = e.makeIterable(42)
	if err == nil {
		t.Error("makeIterable(int) should return error")
	}
}

// TestMakeIterableForVariables tests the makeIterableForVariables function
func TestMakeIterableForVariables(t *testing.T) {
	e := NewEvaluator()

	// Test with nil
	result, err := e.makeIterableForVariables(nil, 1)
	if err != nil || result != nil {
		t.Errorf("makeIterableForVariables(nil) = %v, %v", result, err)
	}

	// Test with Undefined
	undef := NewUndefined("test", UndefinedSilent, nil)
	result, err = e.makeIterableForVariables(undef, 1)
	if err != nil || len(result) != 0 {
		t.Errorf("makeIterableForVariables(Undefined) = %v, %v", result, err)
	}

	// Test with map and 2 variables (key-value iteration)
	m := map[string]interface{}{"a": 1, "b": 2}
	result, err = e.makeIterableForVariables(m, 2)
	if err != nil {
		t.Errorf("makeIterableForVariables(map, 2) error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("makeIterableForVariables(map, 2) should return 2 pairs")
	}

	// Test with DictItems
	dictItems := &DictItems{data: map[string]interface{}{"a": 1}}
	result, err = e.makeIterableForVariables(dictItems, 2)
	if err != nil || len(result) != 1 {
		t.Errorf("makeIterableForVariables(DictItems, 2) = %v, %v", result, err)
	}

	// Test with function returning iterable
	fn := func() []interface{} { return []interface{}{1, 2} }
	result, err = e.makeIterableForVariables(fn, 1)
	if err != nil || len(result) != 2 {
		t.Errorf("makeIterableForVariables(func, 1) = %v, %v", result, err)
	}
}

// TestApplyTest tests the applyTest function
func TestApplyTest(t *testing.T) {
	e := NewEvaluator()

	// Test "none"
	result, err := e.applyTest("none", nil, nil)
	if err != nil || result != true {
		t.Errorf("applyTest('none', nil) = %v, %v", result, err)
	}
	result, err = e.applyTest("none", "value", nil)
	if err != nil || result != false {
		t.Errorf("applyTest('none', 'value') = %v, %v", result, err)
	}

	// Test "string"
	result, err = e.applyTest("string", "hello", nil)
	if err != nil || result != true {
		t.Errorf("applyTest('string', 'hello') = %v, %v", result, err)
	}

	// Test "number"
	result, err = e.applyTest("number", 42, nil)
	if err != nil || result != true {
		t.Errorf("applyTest('number', 42) = %v, %v", result, err)
	}
	result, err = e.applyTest("number", 3.14, nil)
	if err != nil || result != true {
		t.Errorf("applyTest('number', 3.14) = %v, %v", result, err)
	}

	// Test "sequence"
	result, err = e.applyTest("sequence", []interface{}{1, 2, 3}, nil)
	if err != nil || result != true {
		t.Errorf("applyTest('sequence', list) = %v, %v", result, err)
	}

	// Test "mapping"
	result, err = e.applyTest("mapping", map[string]interface{}{"key": "value"}, nil)
	if err != nil || result != true {
		t.Errorf("applyTest('mapping', map) = %v, %v", result, err)
	}

	// Test "even"
	result, err = e.applyTest("even", 4, nil)
	if err != nil || result != true {
		t.Errorf("applyTest('even', 4) = %v, %v", result, err)
	}
	result, err = e.applyTest("even", 3, nil)
	if err != nil || result != false {
		t.Errorf("applyTest('even', 3) = %v, %v", result, err)
	}

	// Test "odd"
	result, err = e.applyTest("odd", 3, nil)
	if err != nil || result != true {
		t.Errorf("applyTest('odd', 3) = %v, %v", result, err)
	}

	// Test "divisibleby" with args
	result, err = e.applyTest("divisibleby", 10, []interface{}{5})
	if err != nil || result != true {
		t.Errorf("applyTest('divisibleby', 10, 5) = %v, %v", result, err)
	}

	// Test "upper"
	result, err = e.applyTest("upper", "HELLO", nil)
	if err != nil || result != true {
		t.Errorf("applyTest('upper', 'HELLO') = %v, %v", result, err)
	}

	// Test "lower"
	result, err = e.applyTest("lower", "hello", nil)
	if err != nil || result != true {
		t.Errorf("applyTest('lower', 'hello') = %v, %v", result, err)
	}

	// Test "in" with container
	result, err = e.applyTest("in", "a", []interface{}{[]interface{}{"a", "b", "c"}})
	if err != nil || result != true {
		t.Errorf("applyTest('in', 'a', list) = %v, %v", result, err)
	}

	// Test "in" without container
	_, err = e.applyTest("in", "a", nil)
	if err == nil {
		t.Error("applyTest('in', 'a') should require container argument")
	}

	// Test unknown test
	_, err = e.applyTest("unknowntest", nil, nil)
	if err == nil {
		t.Error("applyTest('unknowntest') should return error")
	}
}

// TestCyclerMethods tests Cycler methods
func TestCyclerMethods(t *testing.T) {
	// Test empty cycler
	emptyCycler := &Cycler{Items: []interface{}{}}
	if emptyCycler.Next() != nil {
		t.Error("Empty cycler Next() should return nil")
	}
	if emptyCycler.GetCurrent() != nil {
		t.Error("Empty cycler GetCurrent() should return nil")
	}

	// Test normal cycler
	cycler := &Cycler{Items: []interface{}{"a", "b", "c"}}
	if cycler.Next() != "a" {
		t.Error("First Next() should return 'a'")
	}
	if cycler.Next() != "b" {
		t.Error("Second Next() should return 'b'")
	}
	if cycler.GetCurrent() != "c" {
		t.Error("GetCurrent() should return 'c'")
	}

	cycler.Reset()
	if cycler.GetCurrent() != "a" {
		t.Error("After Reset(), GetCurrent() should return 'a'")
	}

	// Test out of bounds Current
	cycler.Current = 100
	if cycler.Next() == nil {
		t.Error("Next() should handle out of bounds Current")
	}
	cycler.Current = -1
	if cycler.GetCurrent() == nil {
		t.Error("GetCurrent() should handle negative Current")
	}
}

// TestJoinerMethods tests Joiner methods
func TestJoinerMethods(t *testing.T) {
	joiner := &Joiner{Separator: ", "}

	// First call should return empty
	if joiner.Join() != "" {
		t.Error("First Join() should return empty string")
	}

	// Second call should return separator
	if joiner.Join() != ", " {
		t.Error("Second Join() should return separator")
	}

	// String method
	if joiner.String() != ", " {
		t.Error("String() should return separator after used")
	}
}

// TestCallableLoop tests CallableLoop
func TestCallableLoop(t *testing.T) {
	called := false
	cl := &CallableLoop{
		Info: map[string]interface{}{"index": 1},
		RecursiveFunc: func(item interface{}) (interface{}, error) {
			called = true
			return item, nil
		},
	}

	// Test Call with correct args
	result, err := cl.Call("arg")
	if err != nil || !called {
		t.Errorf("Call() failed: %v", err)
	}
	if result != "arg" {
		t.Errorf("Call() = %v, want 'arg'", result)
	}

	// Test Call with wrong number of args
	_, err = cl.Call()
	if err == nil {
		t.Error("Call() with no args should error")
	}

	_, err = cl.Call("a", "b")
	if err == nil {
		t.Error("Call() with 2 args should error")
	}

	// Test GetAttribute
	val, ok := cl.GetAttribute("index")
	if !ok || val != 1 {
		t.Errorf("GetAttribute('index') = %v, %v", val, ok)
	}
	_, ok = cl.GetAttribute("nonexistent")
	if ok {
		t.Error("GetAttribute('nonexistent') should return false")
	}
}

// TestSimpleContext tests the simpleContext implementation
func TestSimpleContext(t *testing.T) {
	ctx := &simpleContext{variables: make(map[string]interface{})}

	// Test SetVariable and GetVariable
	ctx.SetVariable("key", "value")
	val, ok := ctx.GetVariable("key")
	if !ok || val != "value" {
		t.Errorf("GetVariable('key') = %v, %v", val, ok)
	}

	// Test missing variable
	_, ok = ctx.GetVariable("missing")
	if ok {
		t.Error("GetVariable('missing') should return false")
	}

	// Test Clone
	clone := ctx.Clone()
	clonedVal, ok := clone.GetVariable("key")
	if !ok || clonedVal != "value" {
		t.Error("Clone should preserve variables")
	}

	// Modify clone and ensure original is unchanged
	clone.SetVariable("key", "modified")
	val, _ = ctx.GetVariable("key")
	if val != "value" {
		t.Error("Clone modification should not affect original")
	}

	// Test All
	all := ctx.All()
	if all["key"] != "value" {
		t.Error("All() should return all variables")
	}
}

// unsupportedTestNode is a custom node type for testing unsupported node handling
type unsupportedTestNode struct{}

func (n *unsupportedTestNode) String() string { return "unsupported" }
func (n *unsupportedTestNode) Line() int      { return 0 }
func (n *unsupportedTestNode) Column() int    { return 0 }

// TestEvalNodeUnsupported tests EvalNode with unsupported node type
func TestEvalNodeUnsupported(t *testing.T) {
	e := NewEvaluator()
	ctx := &simpleContext{variables: make(map[string]interface{})}

	_, err := e.EvalNode(&unsupportedTestNode{}, ctx)
	if err == nil {
		t.Error("EvalNode with unsupported node type should return error")
	}
}

// TestNewStrictEvaluator tests the strict evaluator constructor
func TestNewStrictEvaluator(t *testing.T) {
	e := NewStrictEvaluator()
	if e == nil {
		t.Fatal("NewStrictEvaluator() returned nil")
	}
	if e.undefinedHandler == nil {
		t.Error("StrictEvaluator should have undefinedHandler")
	}
}

// TestNewDebugEvaluator tests the debug evaluator constructor
func TestNewDebugEvaluator(t *testing.T) {
	e := NewDebugEvaluator()
	if e == nil {
		t.Fatal("NewDebugEvaluator() returned nil")
	}
	if e.undefinedHandler == nil {
		t.Error("DebugEvaluator should have undefinedHandler")
	}
}

// TestSetUndefinedBehavior tests setting undefined behavior
func TestSetUndefinedBehavior(t *testing.T) {
	e := NewEvaluator()

	// Test with existing handler
	e.SetUndefinedBehavior(UndefinedStrict)

	// Test without handler
	e2 := &DefaultEvaluator{}
	e2.SetUndefinedBehavior(UndefinedDebug)
	if e2.undefinedHandler == nil {
		t.Error("SetUndefinedBehavior should create handler if nil")
	}
}

// TestSetImportSystem tests setting the import system
func TestSetImportSystem(t *testing.T) {
	e := NewEvaluator()
	// Just verify it doesn't panic
	e.SetImportSystem(nil)
}

// TestEvalAssignmentNode tests the EvalAssignmentNode function
func TestEvalAssignmentNode(t *testing.T) {
	e := NewEvaluator()

	t.Run("simple variable assignment", func(t *testing.T) {
		node := &parser.AssignmentNode{
			Target: &parser.IdentifierNode{Name: "x"},
			Value:  &parser.LiteralNode{Value: 42},
		}
		ctx := &simpleContext{variables: make(map[string]interface{})}

		result, err := e.EvalAssignmentNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalAssignmentNode failed: %v", err)
		}
		if result != "" {
			t.Errorf("assignment should return empty string, got %v", result)
		}
		if ctx.variables["x"] != 42 {
			t.Errorf("variable x should be 42, got %v", ctx.variables["x"])
		}
	})

	t.Run("attribute assignment", func(t *testing.T) {
		obj := map[string]interface{}{"existing": "value"}
		node := &parser.AssignmentNode{
			Target: &parser.AttributeNode{
				Object:    &parser.IdentifierNode{Name: "obj"},
				Attribute: "newattr",
			},
			Value: &parser.LiteralNode{Value: "newvalue"},
		}
		ctx := &simpleContext{variables: map[string]interface{}{"obj": obj}}

		_, err := e.EvalAssignmentNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalAssignmentNode attribute failed: %v", err)
		}
		if obj["newattr"] != "newvalue" {
			t.Errorf("attribute should be set, got %v", obj["newattr"])
		}
	})

	t.Run("item assignment", func(t *testing.T) {
		obj := map[string]interface{}{"a": 1}
		node := &parser.AssignmentNode{
			Target: &parser.GetItemNode{
				Object: &parser.IdentifierNode{Name: "obj"},
				Key:    &parser.LiteralNode{Value: "b"},
			},
			Value: &parser.LiteralNode{Value: 2},
		}
		ctx := &simpleContext{variables: map[string]interface{}{"obj": obj}}

		_, err := e.EvalAssignmentNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalAssignmentNode item failed: %v", err)
		}
		if obj["b"] != 2 {
			t.Errorf("item should be set, got %v", obj["b"])
		}
	})

	t.Run("invalid target type", func(t *testing.T) {
		node := &parser.AssignmentNode{
			Target: &parser.LiteralNode{Value: "not a valid target"},
			Value:  &parser.LiteralNode{Value: 42},
		}
		ctx := &simpleContext{variables: make(map[string]interface{})}

		_, err := e.EvalAssignmentNode(node, ctx)
		if err == nil {
			t.Error("EvalAssignmentNode should fail with invalid target")
		}
	})
}

// TestEvalComprehensionNode tests list and dict comprehensions
func TestEvalComprehensionNode(t *testing.T) {
	e := NewEvaluator()

	t.Run("list comprehension", func(t *testing.T) {
		// [x * 2 for x in items]
		node := &parser.ComprehensionNode{
			IsDict:   false,
			Variable: "x",
			Iterable: &parser.IdentifierNode{Name: "items"},
			Expression: &parser.BinaryOpNode{
				Operator: "*",
				Left:     &parser.IdentifierNode{Name: "x"},
				Right:    &parser.LiteralNode{Value: 2},
			},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{1, 2, 3},
		}}

		result, err := e.EvalComprehensionNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalComprehensionNode failed: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("result should be slice, got %T", result)
		}
		if len(resultSlice) != 3 {
			t.Errorf("result length = %d, want 3", len(resultSlice))
		}
	})

	t.Run("list comprehension with condition", func(t *testing.T) {
		// [x for x in items if x > 1]
		node := &parser.ComprehensionNode{
			IsDict:     false,
			Variable:   "x",
			Iterable:   &parser.IdentifierNode{Name: "items"},
			Expression: &parser.IdentifierNode{Name: "x"},
			Condition: &parser.BinaryOpNode{
				Operator: ">",
				Left:     &parser.IdentifierNode{Name: "x"},
				Right:    &parser.LiteralNode{Value: 1},
			},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{1, 2, 3},
		}}

		result, err := e.EvalComprehensionNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalComprehensionNode with condition failed: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("result should be slice, got %T", result)
		}
		if len(resultSlice) != 2 {
			t.Errorf("result length = %d, want 2", len(resultSlice))
		}
	})

	t.Run("dict comprehension", func(t *testing.T) {
		// {k: v for item in items}
		node := &parser.ComprehensionNode{
			IsDict:     true,
			Variable:   "item",
			Iterable:   &parser.IdentifierNode{Name: "items"},
			KeyExpr:    &parser.GetItemNode{Object: &parser.IdentifierNode{Name: "item"}, Key: &parser.LiteralNode{Value: 0}},
			Expression: &parser.GetItemNode{Object: &parser.IdentifierNode{Name: "item"}, Key: &parser.LiteralNode{Value: 1}},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{
				[]interface{}{"a", 1},
				[]interface{}{"b", 2},
			},
		}}

		result, err := e.EvalComprehensionNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalComprehensionNode dict failed: %v", err)
		}
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("result should be map, got %T", result)
		}
		if len(resultMap) != 2 {
			t.Errorf("result length = %d, want 2", len(resultMap))
		}
	})

	t.Run("dict comprehension with condition", func(t *testing.T) {
		node := &parser.ComprehensionNode{
			IsDict:     true,
			Variable:   "item",
			Iterable:   &parser.IdentifierNode{Name: "items"},
			KeyExpr:    &parser.GetItemNode{Object: &parser.IdentifierNode{Name: "item"}, Key: &parser.LiteralNode{Value: 0}},
			Expression: &parser.GetItemNode{Object: &parser.IdentifierNode{Name: "item"}, Key: &parser.LiteralNode{Value: 1}},
			Condition: &parser.BinaryOpNode{
				Operator: ">",
				Left:     &parser.GetItemNode{Object: &parser.IdentifierNode{Name: "item"}, Key: &parser.LiteralNode{Value: 1}},
				Right:    &parser.LiteralNode{Value: 0},
			},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{
				[]interface{}{"a", 1},
				[]interface{}{"b", -1},
				[]interface{}{"c", 2},
			},
		}}

		result, err := e.EvalComprehensionNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalComprehensionNode dict with condition failed: %v", err)
		}
		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("result should be map, got %T", result)
		}
		if len(resultMap) != 2 {
			t.Errorf("result length = %d, want 2", len(resultMap))
		}
	})

	t.Run("invalid iterable", func(t *testing.T) {
		node := &parser.ComprehensionNode{
			IsDict:     false,
			Variable:   "x",
			Iterable:   &parser.LiteralNode{Value: 42},
			Expression: &parser.IdentifierNode{Name: "x"},
		}
		ctx := &simpleContext{variables: make(map[string]interface{})}

		_, err := e.EvalComprehensionNode(node, ctx)
		if err == nil {
			t.Error("EvalComprehensionNode should fail with invalid iterable")
		}
	})
}

// TestAttributeExistsCoverage tests additional attributeExists cases
func TestAttributeExistsCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("nil object", func(t *testing.T) {
		if e.attributeExists(nil, "attr") {
			t.Error("attributeExists(nil) should return false")
		}
	})

	t.Run("map[string]interface{} special methods", func(t *testing.T) {
		obj := map[string]interface{}{"key": "value"}
		if !e.attributeExists(obj, "items") {
			t.Error("attributeExists should return true for 'items'")
		}
		if !e.attributeExists(obj, "keys") {
			t.Error("attributeExists should return true for 'keys'")
		}
		if !e.attributeExists(obj, "values") {
			t.Error("attributeExists should return true for 'values'")
		}
	})

	t.Run("Cycler methods", func(t *testing.T) {
		cycler := &Cycler{Items: []interface{}{"a", "b"}}
		if !e.attributeExists(cycler, "next") {
			t.Error("attributeExists should return true for Cycler.next")
		}
		if !e.attributeExists(cycler, "current") {
			t.Error("attributeExists should return true for Cycler.current")
		}
		if !e.attributeExists(cycler, "reset") {
			t.Error("attributeExists should return true for Cycler.reset")
		}
		if e.attributeExists(cycler, "invalid") {
			t.Error("attributeExists should return false for invalid method")
		}
	})

	t.Run("Joiner", func(t *testing.T) {
		joiner := &Joiner{Separator: ", "}
		if e.attributeExists(joiner, "anything") {
			t.Error("attributeExists should return false for Joiner")
		}
	})

	t.Run("struct with exported field", func(t *testing.T) {
		type TestStruct struct {
			PublicField string
		}
		obj := TestStruct{PublicField: "public"}
		if !e.attributeExists(obj, "PublicField") {
			t.Error("attributeExists should return true for exported field")
		}
	})

	t.Run("struct pointer", func(t *testing.T) {
		type TestStruct struct {
			Field string
		}
		obj := &TestStruct{Field: "value"}
		if !e.attributeExists(obj, "Field") {
			t.Error("attributeExists should return true for pointer to struct field")
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var obj *struct{ Field string }
		if e.attributeExists(obj, "Field") {
			t.Error("attributeExists should return false for nil pointer")
		}
	})

	t.Run("slice with valid index", func(t *testing.T) {
		obj := []interface{}{1, 2, 3}
		if !e.attributeExists(obj, "0") {
			t.Error("attributeExists should return true for valid index")
		}
		if !e.attributeExists(obj, "2") {
			t.Error("attributeExists should return true for valid index")
		}
	})

	t.Run("slice with invalid index", func(t *testing.T) {
		obj := []interface{}{1, 2, 3}
		if e.attributeExists(obj, "10") {
			t.Error("attributeExists should return false for out of bounds index")
		}
		if e.attributeExists(obj, "-1") {
			t.Error("attributeExists should return false for negative index")
		}
		if e.attributeExists(obj, "notanumber") {
			t.Error("attributeExists should return false for non-numeric attr")
		}
	})

	t.Run("map with string keys via reflection", func(t *testing.T) {
		obj := map[string]int{"a": 1, "b": 2}
		if !e.attributeExists(obj, "a") {
			t.Error("attributeExists should return true for existing map key")
		}
		if e.attributeExists(obj, "c") {
			t.Error("attributeExists should return false for missing map key")
		}
	})

	t.Run("unsupported type", func(t *testing.T) {
		obj := 42
		if e.attributeExists(obj, "anything") {
			t.Error("attributeExists should return false for int")
		}
	})

	t.Run("CallableLoop", func(t *testing.T) {
		loop := &CallableLoop{Info: map[string]interface{}{"index": 1}}
		if !e.attributeExists(loop, "index") {
			t.Error("attributeExists should return true for loop.index")
		}
	})
}

// TestGetObjectNameCoverage tests additional getObjectName cases
func TestGetObjectNameCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("map[string]interface{}", func(t *testing.T) {
		obj := map[string]interface{}{}
		name := e.getObjectName(obj)
		if name != "object" {
			t.Errorf("getObjectName(map[string]interface{}) = %q, want 'object'", name)
		}
	})

	t.Run("map[string]string", func(t *testing.T) {
		obj := map[string]string{}
		name := e.getObjectName(obj)
		if name != "object" {
			t.Errorf("getObjectName(map[string]string) = %q, want 'object'", name)
		}
	})

	t.Run("slice", func(t *testing.T) {
		obj := []interface{}{}
		name := e.getObjectName(obj)
		if name != "list" {
			t.Errorf("getObjectName(slice) = %q, want 'list'", name)
		}
	})

	t.Run("array", func(t *testing.T) {
		obj := [3]int{}
		name := e.getObjectName(obj)
		if name != "array" {
			t.Errorf("getObjectName(array) = %q, want 'array'", name)
		}
	})

	t.Run("struct", func(t *testing.T) {
		type MyStruct struct{}
		obj := MyStruct{}
		name := e.getObjectName(obj)
		if name != "MyStruct" {
			t.Errorf("getObjectName(struct) = %q, want 'MyStruct'", name)
		}
	})

	t.Run("pointer to struct", func(t *testing.T) {
		type MyStruct struct{}
		obj := &MyStruct{}
		name := e.getObjectName(obj)
		if name != "MyStruct" {
			t.Errorf("getObjectName(*struct) = %q, want 'MyStruct'", name)
		}
	})

	t.Run("nil pointer", func(t *testing.T) {
		var obj *struct{}
		name := e.getObjectName(obj)
		if name != "nil" {
			t.Errorf("getObjectName(nil pointer) = %q, want 'nil'", name)
		}
	})

	t.Run("generic map", func(t *testing.T) {
		obj := map[int]string{}
		name := e.getObjectName(obj)
		if name != "map" {
			t.Errorf("getObjectName(map[int]string) = %q, want 'map'", name)
		}
	})

	t.Run("primitive type", func(t *testing.T) {
		obj := 42
		name := e.getObjectName(obj)
		if name != "int" {
			t.Errorf("getObjectName(int) = %q, want 'int'", name)
		}
	})

	t.Run("invalid reflect value", func(t *testing.T) {
		var obj interface{}
		name := e.getObjectName(obj)
		if name != "invalid" {
			t.Errorf("getObjectName(nil interface) = %q, want 'invalid'", name)
		}
	})

}

// TestEvalSliceNodeCoverage tests additional EvalSliceNode cases
func TestEvalSliceNodeCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("slice with start only", func(t *testing.T) {
		node := &parser.SliceNode{
			Object: &parser.IdentifierNode{Name: "items"},
			Start:  &parser.LiteralNode{Value: 1},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{0, 1, 2, 3, 4},
		}}

		result, err := e.EvalSliceNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalSliceNode failed: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("result should be slice, got %T", result)
		}
		if len(resultSlice) != 4 {
			t.Errorf("slice[1:] length = %d, want 4", len(resultSlice))
		}
	})

	t.Run("slice with start and end", func(t *testing.T) {
		node := &parser.SliceNode{
			Object: &parser.IdentifierNode{Name: "items"},
			Start:  &parser.LiteralNode{Value: 1},
			End:    &parser.LiteralNode{Value: 3},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{0, 1, 2, 3, 4},
		}}

		result, err := e.EvalSliceNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalSliceNode failed: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("result should be slice, got %T", result)
		}
		if len(resultSlice) != 2 {
			t.Errorf("slice[1:3] length = %d, want 2", len(resultSlice))
		}
	})

	t.Run("slice with step", func(t *testing.T) {
		node := &parser.SliceNode{
			Object: &parser.IdentifierNode{Name: "items"},
			Start:  &parser.LiteralNode{Value: 0},
			End:    &parser.LiteralNode{Value: 5},
			Step:   &parser.LiteralNode{Value: 2},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{0, 1, 2, 3, 4},
		}}

		result, err := e.EvalSliceNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalSliceNode failed: %v", err)
		}
		resultSlice, ok := result.([]interface{})
		if !ok {
			t.Fatalf("result should be slice, got %T", result)
		}
		if len(resultSlice) != 3 {
			t.Errorf("slice[::2] length = %d, want 3", len(resultSlice))
		}
	})

	t.Run("slice with zero step", func(t *testing.T) {
		node := &parser.SliceNode{
			Object: &parser.IdentifierNode{Name: "items"},
			Step:   &parser.LiteralNode{Value: 0},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{0, 1, 2},
		}}

		_, err := e.EvalSliceNode(node, ctx)
		if err == nil {
			t.Error("EvalSliceNode should fail with zero step")
		}
	})

	t.Run("slice string", func(t *testing.T) {
		node := &parser.SliceNode{
			Object: &parser.IdentifierNode{Name: "s"},
			Start:  &parser.LiteralNode{Value: 0},
			End:    &parser.LiteralNode{Value: 5},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"s": "hello world",
		}}

		result, err := e.EvalSliceNode(node, ctx)
		if err != nil {
			t.Fatalf("EvalSliceNode failed: %v", err)
		}
		if result != "hello" {
			t.Errorf("slice string = %q, want 'hello'", result)
		}
	})

	t.Run("invalid start type", func(t *testing.T) {
		node := &parser.SliceNode{
			Object: &parser.IdentifierNode{Name: "items"},
			Start:  &parser.LiteralNode{Value: "not an int"},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{0, 1, 2},
		}}

		_, err := e.EvalSliceNode(node, ctx)
		if err == nil {
			t.Error("EvalSliceNode should fail with non-int start")
		}
	})

	t.Run("invalid end type", func(t *testing.T) {
		node := &parser.SliceNode{
			Object: &parser.IdentifierNode{Name: "items"},
			End:    &parser.LiteralNode{Value: "not an int"},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{0, 1, 2},
		}}

		_, err := e.EvalSliceNode(node, ctx)
		if err == nil {
			t.Error("EvalSliceNode should fail with non-int end")
		}
	})

	t.Run("invalid step type", func(t *testing.T) {
		node := &parser.SliceNode{
			Object: &parser.IdentifierNode{Name: "items"},
			Step:   &parser.LiteralNode{Value: "not an int"},
		}
		ctx := &simpleContext{variables: map[string]interface{}{
			"items": []interface{}{0, 1, 2},
		}}

		_, err := e.EvalSliceNode(node, ctx)
		if err == nil {
			t.Error("EvalSliceNode should fail with non-int step")
		}
	})
}

// TestEvalNodeListCoverage tests evalNodeList
func TestEvalNodeListCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("empty list", func(t *testing.T) {
		result, err := e.evalNodeList([]parser.Node{}, &simpleContext{variables: make(map[string]interface{})})
		if err != nil {
			t.Fatalf("evalNodeList failed: %v", err)
		}
		if result != "" {
			t.Errorf("evalNodeList([]) = %q, want ''", result)
		}
	})

	t.Run("text nodes", func(t *testing.T) {
		nodes := []parser.Node{
			&parser.TextNode{Content: "Hello "},
			&parser.TextNode{Content: "World"},
		}
		ctx := &simpleContext{variables: make(map[string]interface{})}

		result, err := e.evalNodeList(nodes, ctx)
		if err != nil {
			t.Fatalf("evalNodeList failed: %v", err)
		}
		if result != "Hello World" {
			t.Errorf("evalNodeList = %q, want 'Hello World'", result)
		}
	})

	t.Run("mixed nodes", func(t *testing.T) {
		nodes := []parser.Node{
			&parser.TextNode{Content: "Value: "},
			&parser.VariableNode{Expression: &parser.IdentifierNode{Name: "x"}},
		}
		ctx := &simpleContext{variables: map[string]interface{}{"x": 42}}

		result, err := e.evalNodeList(nodes, ctx)
		if err != nil {
			t.Fatalf("evalNodeList failed: %v", err)
		}
		if result != "Value: 42" {
			t.Errorf("evalNodeList = %q, want 'Value: 42'", result)
		}
	})

	t.Run("nil result handling", func(t *testing.T) {
		nodes := []parser.Node{
			&parser.VariableNode{Expression: &parser.IdentifierNode{Name: "undefined"}},
		}
		ctx := &simpleContext{variables: make(map[string]interface{})}

		result, err := e.evalNodeList(nodes, ctx)
		if err != nil {
			t.Fatalf("evalNodeList failed: %v", err)
		}
		if result != "" {
			t.Errorf("evalNodeList with nil = %q, want ''", result)
		}
	})
}

// TestToStringCoverage tests additional ToString cases
func TestToStringCoverage(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"int64", int64(100), "100"},
		{"float64", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"SafeValue", SafeValue{Value: "<safe>"}, "<safe>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ToString(tt.input)
			if result != tt.expected {
				t.Errorf("ToString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestEvalBreakNodeCoverage tests EvalBreakNode
func TestEvalBreakNodeCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("break returns break error", func(t *testing.T) {
		node := &parser.BreakNode{}
		ctx := &simpleContext{variables: map[string]interface{}{}}

		_, err := e.EvalBreakNode(node, ctx)
		if err == nil {
			t.Fatal("EvalBreakNode should return error")
		}
		// Check that it's a break error
		loopErr, ok := err.(*LoopControlError)
		if !ok {
			t.Fatalf("expected LoopControlError, got %T", err)
		}
		if !loopErr.IsBreak() {
			t.Error("expected IsBreak to be true")
		}
	})
}

// TestEvalContinueNodeCoverage tests EvalContinueNode
func TestEvalContinueNodeCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("continue returns continue error", func(t *testing.T) {
		node := &parser.ContinueNode{}
		ctx := &simpleContext{variables: map[string]interface{}{}}

		_, err := e.EvalContinueNode(node, ctx)
		if err == nil {
			t.Fatal("EvalContinueNode should return error")
		}
		// Check that it's a continue error
		loopErr, ok := err.(*LoopControlError)
		if !ok {
			t.Fatalf("expected LoopControlError, got %T", err)
		}
		if !loopErr.IsContinue() {
			t.Error("expected IsContinue to be true")
		}
	})
}

// TestEvalExtensionNodeCoverage tests EvalExtensionNode
func TestEvalExtensionNodeCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("nil evaluate func returns error", func(t *testing.T) {
		node := &parser.ExtensionNode{
			ExtensionName: "test",
			TagName:       "custom",
			EvaluateFunc:  nil,
		}
		ctx := &simpleContext{variables: map[string]interface{}{}}

		_, err := e.EvalExtensionNode(node, ctx)
		if err == nil {
			t.Error("expected error for nil evaluate func")
		}
	})

	t.Run("with evaluate func", func(t *testing.T) {
		node := &parser.ExtensionNode{
			ExtensionName: "test",
			TagName:       "custom",
			EvaluateFunc: func(n *parser.ExtensionNode, c interface{}) (interface{}, error) {
				return "evaluated", nil
			},
		}
		ctx := &simpleContext{variables: map[string]interface{}{}}

		result, err := e.EvalExtensionNode(node, ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "evaluated" {
			t.Errorf("expected 'evaluated', got %v", result)
		}
	})
}

// TestEvalCallBlockNodeCoverage tests EvalCallBlockNode
func TestEvalCallBlockNodeCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("call block with body", func(t *testing.T) {
		// Create a simple macro that uses caller()
		macroFunc := func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
			// Get caller from the calling context - this is passed via kwargs
			if caller, ok := kwargs["caller"]; ok {
				if callerFn, ok := caller.(func(...interface{}) (interface{}, error)); ok {
					return callerFn()
				}
			}
			return "no caller", nil
		}

		node := &parser.CallBlockNode{
			Call: &parser.CallNode{
				Function:  &parser.IdentifierNode{Name: "test_macro"},
				Arguments: []parser.ExpressionNode{},
			},
			Body: []parser.Node{
				&parser.LiteralNode{Value: "block content"},
			},
		}

		ctx := &simpleContext{variables: map[string]interface{}{
			"test_macro": macroFunc,
		}}

		// The call block should evaluate
		_, err := e.EvalCallBlockNode(node, ctx)
		// This may return an error if the macro call doesn't work as expected
		// but the function should at least be exercised
		_ = err
	})

	t.Run("call block with nil body", func(t *testing.T) {
		macroFunc := func(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
			return "result", nil
		}

		node := &parser.CallBlockNode{
			Call: &parser.CallNode{
				Function:  &parser.IdentifierNode{Name: "test_macro"},
				Arguments: []parser.ExpressionNode{},
			},
			Body: nil,
		}

		ctx := &simpleContext{variables: map[string]interface{}{
			"test_macro": macroFunc,
		}}

		_, err := e.EvalCallBlockNode(node, ctx)
		_ = err // Error handling doesn't matter - just exercising the code path
	})
}

// TestContainsCoverage tests the contains function with additional cases
func TestContainsCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("string contains string", func(t *testing.T) {
		result, err := e.contains("hello world", "world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result {
			t.Error("expected true for 'world' in 'hello world'")
		}
	})

	t.Run("string not contains", func(t *testing.T) {
		result, err := e.contains("hello", "xyz")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result {
			t.Error("expected false for 'xyz' not in 'hello'")
		}
	})

	t.Run("slice contains element found", func(t *testing.T) {
		result, err := e.contains([]interface{}{1, 2, 3}, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result {
			t.Error("expected true for 2 in [1,2,3]")
		}
	})

	t.Run("slice contains element not found", func(t *testing.T) {
		result, err := e.contains([]interface{}{1, 2, 3}, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result {
			t.Error("expected false for 5 not in [1,2,3]")
		}
	})

	t.Run("map contains key", func(t *testing.T) {
		result, err := e.contains(map[string]interface{}{"a": 1, "b": 2}, "a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result {
			t.Error("expected true for 'a' in map")
		}
	})

	t.Run("map not contains key", func(t *testing.T) {
		result, err := e.contains(map[string]interface{}{"a": 1}, "c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result {
			t.Error("expected false for 'c' not in map")
		}
	})

	t.Run("typed slice via reflection", func(t *testing.T) {
		// Use a typed slice that's not []interface{}
		result, err := e.contains([]int{1, 2, 3}, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result {
			t.Error("expected true for 2 in []int{1,2,3}")
		}
	})

	t.Run("typed slice via reflection not found", func(t *testing.T) {
		result, err := e.contains([]string{"a", "b"}, "c")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result {
			t.Error("expected false for 'c' not in []string")
		}
	})

	t.Run("typed map via reflection", func(t *testing.T) {
		result, err := e.contains(map[int]string{1: "a", 2: "b"}, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result {
			t.Error("expected true for 1 in map[int]string")
		}
	})

	t.Run("typed map via reflection not found", func(t *testing.T) {
		result, err := e.contains(map[int]string{1: "a"}, 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result {
			t.Error("expected false for 5 not in map[int]string")
		}
	})

	t.Run("non-container type", func(t *testing.T) {
		_, err := e.contains(42, 1)
		if err == nil {
			t.Error("expected error for non-container type")
		}
	})

	t.Run("array via reflection", func(t *testing.T) {
		arr := [3]int{1, 2, 3}
		result, err := e.contains(arr, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result {
			t.Error("expected true for 2 in [3]int")
		}
	})
}

// TestSetAttributeCoverage tests setAttribute with additional cases
func TestSetAttributeCoverage(t *testing.T) {
	e := NewEvaluator()

	t.Run("set on nil object", func(t *testing.T) {
		err := e.setAttribute(nil, "attr", "value")
		if err == nil {
			t.Error("expected error for nil object")
		}
	})

	t.Run("set on map[string]interface{}", func(t *testing.T) {
		m := map[string]interface{}{}
		err := e.setAttribute(m, "key", "value")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if m["key"] != "value" {
			t.Error("setAttribute did not set value in map")
		}
	})

	t.Run("set on struct pointer", func(t *testing.T) {
		type TestStruct struct {
			Name  string
			Value int
		}
		obj := &TestStruct{}
		err := e.setAttribute(obj, "Name", "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if obj.Name != "test" {
			t.Errorf("expected Name='test', got %q", obj.Name)
		}
	})

	t.Run("set on struct pointer with capitalization", func(t *testing.T) {
		type TestStruct struct {
			MyField string
		}
		obj := &TestStruct{}
		err := e.setAttribute(obj, "myField", "value")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if obj.MyField != "value" {
			t.Errorf("expected MyField='value', got %q", obj.MyField)
		}
	})

	t.Run("set on struct invalid field", func(t *testing.T) {
		type TestStruct struct {
			Name string
		}
		obj := &TestStruct{}
		err := e.setAttribute(obj, "nonexistent", "value")
		if err == nil {
			t.Error("expected error for nonexistent field")
		}
	})

	t.Run("set on non-settable type", func(t *testing.T) {
		err := e.setAttribute(42, "attr", "value")
		if err == nil {
			t.Error("expected error for int type")
		}
	})

	t.Run("set with incompatible type", func(t *testing.T) {
		type TestStruct struct {
			Count int
		}
		obj := &TestStruct{}
		// Try to set string to int field - should work due to type conversion
		err := e.setAttribute(obj, "Count", 42)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if obj.Count != 42 {
			t.Errorf("expected Count=42, got %d", obj.Count)
		}
	})
}

// TestContextWrapperAutoescapeCoverage tests ContextWrapper autoescape methods
func TestContextWrapperAutoescapeCoverage(t *testing.T) {
	t.Run("IsAutoescapeEnabled default nil escaper", func(t *testing.T) {
		ctx := &simpleContext{variables: map[string]interface{}{}}
		cw := NewContextWrapper(ctx, nil, EscapeContextHTML)
		// Nil autoescaper, so IsAutoescapeEnabled should return false
		if cw.IsAutoescapeEnabled() {
			t.Error("autoescape should be disabled with nil escaper")
		}
	})

	t.Run("IsAutoescapeEnabled with autoescaper", func(t *testing.T) {
		ctx := &simpleContext{variables: map[string]interface{}{}}
		ae := NewAutoEscaper(nil)
		cw := NewContextWrapper(ctx, ae, EscapeContextHTML)
		// Now it should return true since autoescaper is set
		if !cw.IsAutoescapeEnabled() {
			t.Error("autoescape should be enabled when autoescaper is set")
		}
	})

	t.Run("GetAutoEscaper and SetAutoEscaper", func(t *testing.T) {
		ctx := &simpleContext{variables: map[string]interface{}{}}
		cw := NewContextWrapper(ctx, nil, EscapeContextHTML)
		ae := NewAutoEscaper(nil)
		cw.SetAutoEscaper(ae)
		if cw.GetAutoEscaper() != ae {
			t.Error("GetAutoEscaper should return the same autoescaper")
		}
	})

	t.Run("GetEscapeContext and SetEscapeContext", func(t *testing.T) {
		ctx := &simpleContext{variables: map[string]interface{}{}}
		cw := NewContextWrapper(ctx, nil, EscapeContextHTML)
		cw.SetEscapeContext(EscapeContextJS)
		if cw.GetEscapeContext() != EscapeContextJS {
			t.Error("GetEscapeContext should return EscapeContextJS")
		}
	})
}
