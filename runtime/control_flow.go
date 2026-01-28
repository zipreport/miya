package runtime

import (
	"fmt"
	"strings"

	"github.com/zipreport/miya/parser"
)

// LoopControlError represents break and continue flow control
type LoopControlError struct {
	Type string // "break" or "continue"
}

func (e *LoopControlError) Error() string {
	return fmt.Sprintf("%s statement", e.Type)
}

// IsBreak returns true if this is a break control error
func (e *LoopControlError) IsBreak() bool {
	return e.Type == "break"
}

// IsContinue returns true if this is a continue control error
func (e *LoopControlError) IsContinue() bool {
	return e.Type == "continue"
}

// NewBreakError creates a break control error
func NewBreakError() *LoopControlError {
	return &LoopControlError{Type: "break"}
}

// NewContinueError creates a continue control error
func NewContinueError() *LoopControlError {
	return &LoopControlError{Type: "continue"}
}

// ControlFlowEvaluator handles control flow statements like if, for, while
type ControlFlowEvaluator struct {
	evaluator *DefaultEvaluator
}

func NewControlFlowEvaluator(evaluator *DefaultEvaluator) *ControlFlowEvaluator {
	return &ControlFlowEvaluator{
		evaluator: evaluator,
	}
}

// EvalIfStatement evaluates if/elif/else constructs
func (cf *ControlFlowEvaluator) EvalIfStatement(node *parser.IfNode, ctx Context) (string, error) {
	// Evaluate main condition
	condition, err := cf.evaluator.EvalNode(node.Condition, ctx)
	if err != nil {
		return "", fmt.Errorf("error evaluating if condition: %w", err)
	}

	if cf.evaluator.isTruthy(condition) {
		return cf.evalBodyNodes(node.Body, ctx)
	}

	// Check elif conditions
	for _, elif := range node.ElseIfs {
		condition, err := cf.evaluator.EvalNode(elif.Condition, ctx)
		if err != nil {
			return "", fmt.Errorf("error evaluating elif condition: %w", err)
		}

		if cf.evaluator.isTruthy(condition) {
			return cf.evalBodyNodes(elif.Body, ctx)
		}
	}

	// Else clause
	if len(node.Else) > 0 {
		return cf.evalBodyNodes(node.Else, ctx)
	}

	return "", nil
}

// EvalForLoop evaluates for loops with proper loop context and break/continue support
func (cf *ControlFlowEvaluator) EvalForLoop(node *parser.ForNode, ctx Context) (string, error) {
	// Evaluate iterable
	iterable, err := cf.evaluator.EvalNode(node.Iterable, ctx)
	if err != nil {
		return "", fmt.Errorf("error evaluating iterable: %w", err)
	}

	// Convert to slice
	items, err := cf.evaluator.makeIterable(iterable)
	if err != nil {
		return "", fmt.Errorf("error making iterable: %w", err)
	}

	// Handle empty iteration - execute else clause if present
	if len(items) == 0 && len(node.Else) > 0 {
		return cf.evalBodyNodes(node.Else, ctx)
	}

	// Create loop context
	loopCtx := ctx.Clone()

	// Use strings.Builder for efficient string concatenation
	var result strings.Builder
	loopBroken := false
	length := len(items)

	// Get pooled loop info map to reduce allocations
	loopInfoMap, _ := loopInfoPool.Get().(map[string]interface{})
	if loopInfoMap == nil {
		loopInfoMap = make(map[string]interface{}, 8)
	}
	// Pre-set the length which doesn't change
	loopInfoMap["length"] = length
	defer func() {
		// Clear and return to pool
		for k := range loopInfoMap {
			delete(loopInfoMap, k)
		}
		loopInfoPool.Put(loopInfoMap)
	}()

	for i, item := range items {
		// Set loop variable(s)
		if len(node.Variables) == 1 {
			// Single variable assignment
			loopCtx.SetVariable(node.Variables[0], item)
		} else {
			// Multiple variable assignment - unpack the item
			itemSlice, err := cf.evaluator.makeIterable(item)
			if err != nil {
				return "", fmt.Errorf("cannot unpack non-iterable %T for loop variables", item)
			}

			if len(itemSlice) != len(node.Variables) {
				return "", fmt.Errorf("cannot unpack %d values into %d variables", len(itemSlice), len(node.Variables))
			}

			// Assign each value to its corresponding variable
			for j, variable := range node.Variables {
				loopCtx.SetVariable(variable, itemSlice[j])
			}
		}

		// Update loop info map (reuse the same map)
		loopInfoMap["index"] = i + 1
		loopInfoMap["index0"] = i
		loopInfoMap["revindex"] = length - i
		loopInfoMap["revindex0"] = length - i - 1
		loopInfoMap["first"] = i == 0
		loopInfoMap["last"] = i == length-1
		loopInfoMap["length"] = length

		// Set loop info for template access
		loopCtx.SetVariable("loop", loopInfoMap)

		// Evaluate loop body
		bodyResult, err := cf.evalBodyNodesWithLoopControlBuilder(node.Body, loopCtx, &result)
		if err != nil {
			// Check if it's a loop control error
			if loopErr, ok := err.(*LoopControlError); ok {
				if loopErr.IsBreak() {
					loopBroken = true
					break
				} else if loopErr.IsContinue() {
					continue
				}
			}
			return "", fmt.Errorf("error in loop body at index %d: %w", i, err)
		}
		_ = bodyResult // Result is written directly to builder
	}

	// If loop was broken and there are no results, execute else clause
	if loopBroken && result.Len() == 0 && len(node.Else) > 0 {
		return cf.evalBodyNodes(node.Else, ctx)
	}

	return result.String(), nil
}

// EvalBreak handles break statements
func (cf *ControlFlowEvaluator) EvalBreak() error {
	return NewBreakError()
}

// EvalContinue handles continue statements
func (cf *ControlFlowEvaluator) EvalContinue() error {
	return NewContinueError()
}

// Helper method to evaluate a list of nodes and join their output
func (cf *ControlFlowEvaluator) evalBodyNodes(nodes []parser.Node, ctx Context) (string, error) {
	var result strings.Builder

	for _, node := range nodes {
		nodeResult, err := cf.evaluator.EvalNode(node, ctx)
		if err != nil {
			return "", err
		}

		// Convert result to string and append directly to builder
		if str, ok := nodeResult.(string); ok {
			result.WriteString(str)
		} else {
			result.WriteString(ToString(nodeResult))
		}
	}

	return result.String(), nil
}

// evalBodyNodesWithLoopControl evaluates nodes and propagates loop control errors
func (cf *ControlFlowEvaluator) evalBodyNodesWithLoopControl(nodes []parser.Node, ctx Context) (string, error) {
	var result strings.Builder

	for _, node := range nodes {
		nodeResult, err := cf.evaluator.EvalNode(node, ctx)
		if err != nil {
			// Propagate loop control errors
			if _, ok := err.(*LoopControlError); ok {
				return "", err
			}
			return "", err
		}

		// Convert result to string and append directly to builder
		if str, ok := nodeResult.(string); ok {
			result.WriteString(str)
		} else {
			result.WriteString(ToString(nodeResult))
		}
	}

	return result.String(), nil
}

// evalBodyNodesWithLoopControlBuilder evaluates nodes and writes directly to a shared builder
func (cf *ControlFlowEvaluator) evalBodyNodesWithLoopControlBuilder(nodes []parser.Node, ctx Context, builder *strings.Builder) (string, error) {
	for _, node := range nodes {
		nodeResult, err := cf.evaluator.EvalNode(node, ctx)
		if err != nil {
			// Propagate loop control errors
			if _, ok := err.(*LoopControlError); ok {
				return "", err
			}
			return "", err
		}

		// Convert result to string and append directly to shared builder
		if str, ok := nodeResult.(string); ok {
			builder.WriteString(str)
		} else {
			builder.WriteString(ToString(nodeResult))
		}
	}

	return "", nil
}

// LoopInfo provides information about the current loop iteration
type LoopInfo struct {
	Index     int  // 1-based index
	Index0    int  // 0-based index
	RevIndex  int  // reverse index (1-based)
	RevIndex0 int  // reverse index (0-based)
	First     bool // true if first iteration
	Last      bool // true if last iteration
	Length    int  // total number of items
}

// Additional control flow utilities

// EvalConditionalExpression evaluates ternary-like expressions
func (cf *ControlFlowEvaluator) EvalConditionalExpression(condition, trueExpr, falseExpr parser.ExpressionNode, ctx Context) (interface{}, error) {
	condResult, err := cf.evaluator.EvalNode(condition, ctx)
	if err != nil {
		return nil, err
	}

	if cf.evaluator.isTruthy(condResult) {
		return cf.evaluator.EvalNode(trueExpr, ctx)
	} else {
		return cf.evaluator.EvalNode(falseExpr, ctx)
	}
}

// EvalLogicalAnd evaluates logical AND with short-circuiting
func (cf *ControlFlowEvaluator) EvalLogicalAnd(left, right parser.ExpressionNode, ctx Context) (bool, error) {
	leftResult, err := cf.evaluator.EvalNode(left, ctx)
	if err != nil {
		return false, err
	}

	if !cf.evaluator.isTruthy(leftResult) {
		return false, nil
	}

	rightResult, err := cf.evaluator.EvalNode(right, ctx)
	if err != nil {
		return false, err
	}

	return cf.evaluator.isTruthy(rightResult), nil
}

// EvalLogicalOr evaluates logical OR with short-circuiting
func (cf *ControlFlowEvaluator) EvalLogicalOr(left, right parser.ExpressionNode, ctx Context) (interface{}, error) {
	leftResult, err := cf.evaluator.EvalNode(left, ctx)
	if err != nil {
		return nil, err
	}

	if cf.evaluator.isTruthy(leftResult) {
		return leftResult, nil
	}

	return cf.evaluator.EvalNode(right, ctx)
}

// EvalInExpression checks if a value is contained in a collection
func (cf *ControlFlowEvaluator) EvalInExpression(item, container parser.ExpressionNode, ctx Context) (bool, error) {
	itemValue, err := cf.evaluator.EvalNode(item, ctx)
	if err != nil {
		return false, err
	}

	containerValue, err := cf.evaluator.EvalNode(container, ctx)
	if err != nil {
		return false, err
	}

	return cf.evaluator.contains(containerValue, itemValue)
}

// EvalNotInExpression checks if a value is NOT contained in a collection
func (cf *ControlFlowEvaluator) EvalNotInExpression(item, container parser.ExpressionNode, ctx Context) (bool, error) {
	result, err := cf.EvalInExpression(item, container, ctx)
	return !result, err
}

// NestedLoopEvaluator handles nested loops with proper context management
type NestedLoopEvaluator struct {
	cf *ControlFlowEvaluator
}

func NewNestedLoopEvaluator(cf *ControlFlowEvaluator) *NestedLoopEvaluator {
	return &NestedLoopEvaluator{cf: cf}
}

// EvalNestedLoop handles nested for loops with proper variable scoping
func (nl *NestedLoopEvaluator) EvalNestedLoop(outerNode, innerNode *parser.ForNode, ctx Context) (string, error) {
	// Evaluate outer iterable
	outerIterable, err := nl.cf.evaluator.EvalNode(outerNode.Iterable, ctx)
	if err != nil {
		return "", err
	}

	outerItems, err := nl.cf.evaluator.makeIterable(outerIterable)
	if err != nil {
		return "", err
	}

	var results []string
	outerCtx := ctx.Clone()

	for outerIdx, outerItem := range outerItems {
		// Set outer loop variable and context
		if len(outerNode.Variables) == 1 {
			outerCtx.SetVariable(outerNode.Variables[0], outerItem)
		} else {
			// Multiple variable assignment for outer loop
			itemSlice, err := nl.cf.evaluator.makeIterable(outerItem)
			if err != nil {
				return "", fmt.Errorf("cannot unpack non-iterable %T for outer loop variables", outerItem)
			}
			if len(itemSlice) != len(outerNode.Variables) {
				return "", fmt.Errorf("cannot unpack %d values into %d variables", len(itemSlice), len(outerNode.Variables))
			}
			for j, variable := range outerNode.Variables {
				outerCtx.SetVariable(variable, itemSlice[j])
			}
		}
		outerCtx.SetVariable("loop", map[string]interface{}{
			"index":     outerIdx + 1,
			"index0":    outerIdx,
			"revindex":  len(outerItems) - outerIdx,
			"revindex0": len(outerItems) - outerIdx - 1,
			"first":     outerIdx == 0,
			"last":      outerIdx == len(outerItems)-1,
			"length":    len(outerItems),
		})

		// Evaluate inner iterable within outer context
		innerIterable, err := nl.cf.evaluator.EvalNode(innerNode.Iterable, outerCtx)
		if err != nil {
			return "", err
		}

		innerItems, err := nl.cf.evaluator.makeIterable(innerIterable)
		if err != nil {
			return "", err
		}

		innerCtx := outerCtx.Clone()

		for innerIdx, innerItem := range innerItems {
			// Set inner loop variable and context
			if len(innerNode.Variables) == 1 {
				innerCtx.SetVariable(innerNode.Variables[0], innerItem)
			} else {
				// Multiple variable assignment for inner loop
				itemSlice, err := nl.cf.evaluator.makeIterable(innerItem)
				if err != nil {
					return "", fmt.Errorf("cannot unpack non-iterable %T for inner loop variables", innerItem)
				}
				if len(itemSlice) != len(innerNode.Variables) {
					return "", fmt.Errorf("cannot unpack %d values into %d variables", len(itemSlice), len(innerNode.Variables))
				}
				for j, variable := range innerNode.Variables {
					innerCtx.SetVariable(variable, itemSlice[j])
				}
			}
			innerCtx.SetVariable("loop", map[string]interface{}{
				"index":     innerIdx + 1,
				"index0":    innerIdx,
				"revindex":  len(innerItems) - innerIdx,
				"revindex0": len(innerItems) - innerIdx - 1,
				"first":     innerIdx == 0,
				"last":      innerIdx == len(innerItems)-1,
				"length":    len(innerItems),
			})

			// Evaluate inner loop body
			result, err := nl.cf.evalBodyNodes(innerNode.Body, innerCtx)
			if err != nil {
				return "", err
			}

			results = append(results, result)
		}
	}

	return strings.Join(results, ""), nil
}
