package runtime

import (
	"fmt"
	"sync"

	"github.com/zipreport/miya/parser"
)

// FilterChainOptimizer optimizes the evaluation of chained filters
type FilterChainOptimizer struct {
	evaluator  *DefaultEvaluator
	chainCache map[string]*compiledFilterChain
	cacheMutex sync.RWMutex
}

// compiledFilterChain represents a pre-analyzed filter chain
type compiledFilterChain struct {
	filters []filterCall
}

type filterCall struct {
	name string
	args []parser.ExpressionNode
}

// NewFilterChainOptimizer creates a new filter chain optimizer
func NewFilterChainOptimizer(evaluator *DefaultEvaluator) *FilterChainOptimizer {
	return &FilterChainOptimizer{
		evaluator:  evaluator,
		chainCache: make(map[string]*compiledFilterChain),
	}
}

// EvalFilterChain evaluates a chain of filters more efficiently
func (fco *FilterChainOptimizer) EvalFilterChain(node parser.ExpressionNode, ctx Context) (interface{}, error) {
	// Extract the filter chain
	chain := fco.extractFilterChain(node)
	if len(chain.filters) == 0 {
		// No filters, just evaluate the expression
		return fco.evaluator.EvalNode(node, ctx)
	}

	// Get the base expression (the innermost expression)
	baseExpr := fco.getBaseExpression(node)

	// Evaluate the base expression once
	value, err := fco.evaluator.EvalNode(baseExpr, ctx)
	if err != nil {
		return nil, err
	}

	// Apply filters in sequence
	for _, filter := range chain.filters {
		// Evaluate filter arguments once
		var args []interface{}
		for _, arg := range filter.args {
			argValue, err := fco.evaluator.EvalNode(arg, ctx)
			if err != nil {
				return nil, err
			}
			args = append(args, argValue)
		}

		// Apply the filter
		if envCtx, ok := ctx.(EnvironmentContext); ok {
			value, err = envCtx.ApplyFilter(filter.name, value, args...)
		} else {
			value, err = fco.evaluator.applyFilter(filter.name, value, args)
		}

		if err != nil {
			return nil, err
		}
	}

	return value, nil
}

// extractFilterChain extracts all filters from a nested filter expression
func (fco *FilterChainOptimizer) extractFilterChain(node parser.ExpressionNode) *compiledFilterChain {
	chain := &compiledFilterChain{
		filters: make([]filterCall, 0),
	}

	// Collect filters in reverse order (as we traverse from outer to inner)
	current := node
	for {
		if filterNode, ok := current.(*parser.FilterNode); ok {
			// Append filter (will reverse later)
			chain.filters = append(chain.filters, filterCall{
				name: filterNode.FilterName,
				args: filterNode.Arguments,
			})

			// Move to the inner expression
			current = filterNode.Expression
		} else {
			// We've reached the base expression
			break
		}
	}

	// Reverse the slice in-place to get correct order (avoids allocations from prepend)
	for i, j := 0, len(chain.filters)-1; i < j; i, j = i+1, j-1 {
		chain.filters[i], chain.filters[j] = chain.filters[j], chain.filters[i]
	}

	return chain
}

// getBaseExpression returns the innermost expression in a filter chain
func (fco *FilterChainOptimizer) getBaseExpression(node parser.ExpressionNode) parser.ExpressionNode {
	current := node
	for {
		if filterNode, ok := current.(*parser.FilterNode); ok {
			current = filterNode.Expression
		} else {
			return current
		}
	}
}

// OptimizedEvaluator wraps DefaultEvaluator with filter chain optimization
type OptimizedFilterEvaluator struct {
	*DefaultEvaluator
	optimizer *FilterChainOptimizer
}

// NewOptimizedFilterEvaluator creates an evaluator with filter chain optimization
func NewOptimizedFilterEvaluator() *OptimizedFilterEvaluator {
	defaultEval := NewEvaluator()
	return &OptimizedFilterEvaluator{
		DefaultEvaluator: defaultEval,
		optimizer:        NewFilterChainOptimizer(defaultEval),
	}
}

// EvalNode evaluates a node with filter chain optimization
func (ofe *OptimizedFilterEvaluator) EvalNode(node parser.Node, ctx Context) (interface{}, error) {
	// Check if this is a filter node that might be part of a chain
	if _, ok := node.(*parser.FilterNode); ok {
		return ofe.optimizer.EvalFilterChain(node.(parser.ExpressionNode), ctx)
	}

	// Otherwise, use the default evaluator
	return ofe.DefaultEvaluator.EvalNode(node, ctx)
}

// BatchFilterEvaluation allows evaluating multiple filter chains in parallel
type BatchFilterEvaluation struct {
	chains []filterChainJob
	wg     sync.WaitGroup
}

type filterChainJob struct {
	node   parser.ExpressionNode
	ctx    Context
	result interface{}
	err    error
}

// NewBatchFilterEvaluation creates a new batch filter evaluation
func NewBatchFilterEvaluation() *BatchFilterEvaluation {
	return &BatchFilterEvaluation{
		chains: make([]filterChainJob, 0),
	}
}

// AddChain adds a filter chain to the batch
func (bfe *BatchFilterEvaluation) AddChain(node parser.ExpressionNode, ctx Context) {
	bfe.chains = append(bfe.chains, filterChainJob{
		node: node,
		ctx:  ctx,
	})
}

// Execute runs all filter chains in parallel
func (bfe *BatchFilterEvaluation) Execute(evaluator *OptimizedFilterEvaluator) []error {
	bfe.wg.Add(len(bfe.chains))

	for i := range bfe.chains {
		go func(idx int) {
			defer bfe.wg.Done()
			result, err := evaluator.EvalNode(bfe.chains[idx].node, bfe.chains[idx].ctx)
			bfe.chains[idx].result = result
			bfe.chains[idx].err = err
		}(i)
	}

	bfe.wg.Wait()

	// Collect errors
	var errors []error
	for _, job := range bfe.chains {
		if job.err != nil {
			errors = append(errors, job.err)
		}
	}

	return errors
}

// GetResults returns the results of all filter chains
func (bfe *BatchFilterEvaluation) GetResults() []interface{} {
	results := make([]interface{}, len(bfe.chains))
	for i, job := range bfe.chains {
		results[i] = job.result
	}
	return results
}

// FilterPipeline represents a reusable filter pipeline
type FilterPipeline struct {
	filters []struct {
		name string
		args []interface{}
	}
}

// NewFilterPipeline creates a new filter pipeline
func NewFilterPipeline() *FilterPipeline {
	return &FilterPipeline{
		filters: make([]struct {
			name string
			args []interface{}
		}, 0),
	}
}

// AddFilter adds a filter to the pipeline
func (fp *FilterPipeline) AddFilter(name string, args ...interface{}) *FilterPipeline {
	fp.filters = append(fp.filters, struct {
		name string
		args []interface{}
	}{name: name, args: args})
	return fp
}

// Apply applies the pipeline to a value
func (fp *FilterPipeline) Apply(value interface{}, ctx Context) (interface{}, error) {
	result := value
	var err error

	for _, filter := range fp.filters {
		if envCtx, ok := ctx.(EnvironmentContext); ok {
			result, err = envCtx.ApplyFilter(filter.name, result, filter.args...)
		} else {
			return nil, fmt.Errorf("context does not support filters")
		}

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
