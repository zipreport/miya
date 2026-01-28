# Dictionary Iteration Implementation Guide

## Overview

This document outlines the steps required to implement dictionary iteration with key-value unpacking in Miya Engine to achieve full compatibility with Python Jinja2's `{% for key, value in dict.items() %}` syntax.

## Current Limitation

**Python Jinja2:**
```jinja2
{% for key, value in user.preferences.items() %}
    <li><strong>{{ key }}:</strong> {{ value }}</li>
{% endfor %}
```

**Miya Engine (Current):**
```jinja2
<!-- Not supported - causes "cannot unpack N values into 2 variables" error -->
{% for key, value in user.preferences %}
    <li><strong>{{ key }}:</strong> {{ value }}</li>
{% endfor %}
```

## Implementation Steps

### 1. Lexer/Parser Modifications

#### 1.1 Update AST Node Structure
**File:** `parser/ast.go`

```go
// Update ForNode to support multiple target variables
type ForNode struct {
    Targets   []string // Multiple variables for unpacking: ["key", "value"]
    Iter      Node     // The iterable expression
    Body      []Node   // Loop body
    Else      []Node   // Optional else clause
    Recursive bool     // Recursive loop flag
}
```

#### 1.2 Modify Parser Logic
**File:** `parser/parser.go`

```go
func (p *Parser) parseForStatement() (*ForNode, error) {
    // Parse multiple target variables
    targets := []string{}
    
    // Parse first target
    target, err := p.parseIdentifier()
    if err != nil {
        return nil, err
    }
    targets = append(targets, target)
    
    // Check for comma-separated additional targets
    for p.current().Type == lexer.COMMA {
        p.advance() // consume comma
        nextTarget, err := p.parseIdentifier()
        if err != nil {
            return nil, err
        }
        targets = append(targets, nextTarget)
    }
    
    // Expect 'in' keyword
    if p.current().Type != lexer.IN {
        return nil, fmt.Errorf("expected 'in', got %s", p.current().Type)
    }
    p.advance()
    
    // Parse iterable expression
    iter, err := p.parseExpression()
    if err != nil {
        return nil, err
    }
    
    return &ForNode{
        Targets: targets,
        Iter:    iter,
        // ... parse body and else clauses
    }, nil
}
```

### 2. Runtime Evaluation Updates

#### 2.1 Update ForNode Evaluation
**File:** `runtime/evaluator.go`

```go
func (e *Evaluator) evaluateForNode(node *ForNode, ctx Context) (interface{}, error) {
    // Evaluate the iterable expression
    iter, err := e.evaluateNode(node.Iter, ctx)
    if err != nil {
        return nil, err
    }
    
    // Handle different types of iterables
    switch iterable := iter.(type) {
    case map[string]interface{}:
        return e.evaluateMapIteration(node, iterable, ctx)
    case []interface{}:
        return e.evaluateSliceIteration(node, iterable, ctx)
    case []string:
        return e.evaluateStringSliceIteration(node, iterable, ctx)
    default:
        return nil, fmt.Errorf("object is not iterable: %T", iter)
    }
}

func (e *Evaluator) evaluateMapIteration(node *ForNode, m map[string]interface{}, ctx Context) (interface{}, error) {
    var result strings.Builder
    loopLength := len(m)
    
    if loopLength == 0 {
        // Execute else clause if present
        if node.Else != nil {
            for _, elseNode := range node.Else {
                elseResult, err := e.evaluateNode(elseNode, ctx)
                if err != nil {
                    return nil, err
                }
                result.WriteString(fmt.Sprintf("%v", elseResult))
            }
        }
        return result.String(), nil
    }
    
    // Create loop context
    loopCtx := e.createLoopContext(ctx)
    index := 0
    
    for key, value := range m {
        // Create iteration context
        iterCtx := e.createIterationContext(loopCtx, index, loopLength)
        
        // Assign target variables based on number of targets
        if len(node.Targets) == 1 {
            // Single target gets the value
            iterCtx.Set(node.Targets[0], value)
        } else if len(node.Targets) == 2 {
            // Two targets get key and value
            iterCtx.Set(node.Targets[0], key)
            iterCtx.Set(node.Targets[1], value)
        } else {
            return nil, fmt.Errorf("cannot unpack map into %d variables", len(node.Targets))
        }
        
        // Execute loop body
        for _, bodyNode := range node.Body {
            bodyResult, err := e.evaluateNode(bodyNode, iterCtx)
            if err != nil {
                return nil, err
            }
            result.WriteString(fmt.Sprintf("%v", bodyResult))
        }
        
        index++
    }
    
    return result.String(), nil
}
```

#### 2.2 Add Helper Methods
```go
func (e *Evaluator) createLoopContext(parent Context) Context {
    // Create new context with loop variables
    loopCtx := NewContext()
    // Copy parent context
    for key, value := range parent.(*ContextImpl).data {
        loopCtx.Set(key, value)
    }
    return loopCtx
}

func (e *Evaluator) createIterationContext(parent Context, index, length int) Context {
    iterCtx := NewContext()
    // Copy parent context
    for key, value := range parent.(*ContextImpl).data {
        iterCtx.Set(key, value)
    }
    
    // Add loop variables
    iterCtx.Set("loop", map[string]interface{}{
        "index":   index + 1,
        "index0":  index,
        "first":   index == 0,
        "last":    index == length-1,
        "length":  length,
    })
    
    return iterCtx
}
```

### 3. Type System Enhancements

#### 3.1 Add Dictionary Item Access
**File:** `runtime/evaluator.go`

```go
// Add support for .items() method on maps
func (e *Evaluator) evaluateAttributeAccess(obj interface{}, attr string) (interface{}, error) {
    switch v := obj.(type) {
    case map[string]interface{}:
        if attr == "items" {
            // Return a special items iterator
            return &DictItems{data: v}, nil
        }
        if value, exists := v[attr]; exists {
            return value, nil
        }
        return nil, fmt.Errorf("'%T' object has no attribute '%s'", obj, attr)
    // ... other cases
    }
}

// Special type for dictionary items iteration
type DictItems struct {
    data map[string]interface{}
}

func (d *DictItems) Iterator() []interface{} {
    items := make([]interface{}, 0, len(d.data))
    for key, value := range d.data {
        items = append(items, []interface{}{key, value})
    }
    return items
}
```

#### 3.2 Handle Items Iterator
```go
func (e *Evaluator) evaluateItemsIteration(node *ForNode, items *DictItems, ctx Context) (interface{}, error) {
    var result strings.Builder
    iterator := items.Iterator()
    loopLength := len(iterator)
    
    if loopLength == 0 {
        // Handle else clause
        return e.handleElseClause(node, ctx)
    }
    
    loopCtx := e.createLoopContext(ctx)
    
    for index, item := range iterator {
        iterCtx := e.createIterationContext(loopCtx, index, loopLength)
        
        // Unpack key-value pair
        if pair, ok := item.([]interface{}); ok && len(pair) == 2 {
            if len(node.Targets) == 2 {
                iterCtx.Set(node.Targets[0], pair[0]) // key
                iterCtx.Set(node.Targets[1], pair[1]) // value
            } else if len(node.Targets) == 1 {
                iterCtx.Set(node.Targets[0], pair) // whole pair
            } else {
                return nil, fmt.Errorf("cannot unpack %d values into %d variables", 2, len(node.Targets))
            }
        } else {
            return nil, fmt.Errorf("invalid items format")
        }
        
        // Execute body
        for _, bodyNode := range node.Body {
            bodyResult, err := e.evaluateNode(bodyNode, iterCtx)
            if err != nil {
                return nil, err
            }
            result.WriteString(fmt.Sprintf("%v", bodyResult))
        }
    }
    
    return result.String(), nil
}
```

### 4. Lexer Token Updates

#### 4.1 Add Required Tokens
**File:** `lexer/token.go`

```go
const (
    // ... existing tokens
    COMMA   TokenType = "COMMA"     // ","
    ITEMS   TokenType = "ITEMS"     // "items" (for .items() method)
)

// Update keywords map
var keywords = map[string]TokenType{
    // ... existing keywords
    "items": ITEMS,
}
```

### 5. Testing Implementation

#### 5.1 Unit Tests
**File:** `dictionary_iteration_test.go`

```go
func TestDictionaryIteration(t *testing.T) {
    tests := []struct {
        name     string
        template string
        data     map[string]interface{}
        expected string
    }{
        {
            name:     "Simple key-value iteration",
            template: `{% for key, value in dict %}{{ key }}:{{ value }};{% endfor %}`,
            data: map[string]interface{}{
                "dict": map[string]interface{}{
                    "a": "apple",
                    "b": "banana",
                },
            },
            expected: "a:apple;b:banana;",
        },
        {
            name:     "Items method iteration",
            template: `{% for key, value in dict.items() %}{{ key }}={{ value }},{% endfor %}`,
            data: map[string]interface{}{
                "dict": map[string]interface{}{
                    "x": 1,
                    "y": 2,
                },
            },
            expected: "x=1,y=2,",
        },
        {
            name:     "Single variable gets value",
            template: `{% for item in dict %}{{ item }};{% endfor %}`,
            data: map[string]interface{}{
                "dict": map[string]interface{}{
                    "a": "apple",
                    "b": "banana",
                },
            },
            expected: "apple;banana;",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            env := NewEnvironment()
            ctx := NewContext()
            for k, v := range tt.data {
                ctx.Set(k, v)
            }
            
            result, err := env.RenderString(tt.template, ctx)
            if err != nil {
                t.Fatalf("Template render error: %v", err)
            }
            
            if result != tt.expected {
                t.Errorf("Expected %q, got %q", tt.expected, result)
            }
        })
    }
}
```

### 6. Error Handling

#### 6.1 Improved Error Messages
```go
func (e *Evaluator) validateUnpacking(targets []string, valueCount int) error {
    if len(targets) != valueCount {
        return fmt.Errorf("cannot unpack %d values into %d variables (line %d)", 
            valueCount, len(targets), e.currentLine)
    }
    return nil
}

func (e *Evaluator) handleUnpackingError(node *ForNode, actualType interface{}) error {
    return fmt.Errorf("'%T' object is not iterable or doesn't support unpacking into %d variables", 
        actualType, len(node.Targets))
}
```

### 7. Documentation Updates

#### 7.1 Update Template Guide
Add examples showing:
- `{% for key, value in dict %}` syntax
- `{% for key, value in dict.items() %}` method syntax
- Error cases and troubleshooting

#### 7.2 Migration Guide Updates
Document the new functionality and provide migration examples from the current workarounds.

## Implementation Priority

1. **High Priority:**
   - AST node updates for multiple targets
   - Basic map iteration with key-value unpacking
   - Error handling for unpacking mismatches

2. **Medium Priority:**
   - `.items()` method support
   - Comprehensive testing
   - Documentation updates

3. **Low Priority:**
   - Performance optimizations
   - Advanced unpacking patterns
   - Integration with existing filters

## Expected Outcome

After implementation, templates can use:

```jinja2
{% for key, value in user.preferences %}
    <li><strong>{{ key }}:</strong> {{ value }}</li>
{% endfor %}

{% for key, value in data.items() %}
    <li>{{ key }} = {{ value }}</li>
{% endfor %}
```

This would provide full compatibility with Python Jinja2 dictionary iteration patterns.

## Estimated Development Time

- **Core Implementation:** 2-3 days
- **Testing:** 1-2 days  
- **Documentation:** 1 day
- **Integration Testing:** 1 day

**Total:** 5-7 days for complete implementation