package runtime

import (
	"errors"
	"fmt"
	"sort"

	"github.com/zipreport/miya/parser"
)

// Macro is a callable template macro: a {% macro %} definition bound to the
// context it was defined in, or the caller() of a {% call %} block.
type Macro struct {
	Name       string
	Parameters []string
	// Defaults holds parser.ExpressionNode values, evaluated at call time
	// after the earlier parameters are bound; any other value is used as is.
	Defaults map[string]interface{}
	Body     []parser.Node
	parser.MacroUsage

	defCtx    Context
	evaluator *DefaultEvaluator
}

// String returns the Jinja2 representation of a macro.
func (m *Macro) String() string {
	return fmt.Sprintf("<Macro '%s'>", m.Name)
}

// Call binds args and kwargs to the parameters as Jinja2 does and renders
// the body in a child scope of the definition context.
func (m *Macro) Call(args []interface{}, kwargs map[string]interface{}) (interface{}, error) {
	e := m.evaluator
	leave, err := e.enter("macro " + m.Name)
	if err != nil {
		return nil, err
	}
	defer leave()

	if len(args) > len(m.Parameters) && !m.CatchVarargs {
		return nil, fmt.Errorf("macro %q takes not more than %d argument(s)", m.Name, len(m.Parameters))
	}
	rest := make(map[string]interface{}, len(kwargs))
	for k, v := range kwargs {
		rest[k] = v
	}
	behavior := e.undefinedHandler.GetUndefinedBehavior()

	ctx := childScope(m.defCtx)
	for i, name := range m.Parameters {
		if i < len(args) {
			if _, dup := rest[name]; dup {
				return nil, fmt.Errorf("macro %q got multiple values for argument %q", m.Name, name)
			}
			ctx.SetVariable(name, bindArg(args[i]))
		} else if v, ok := rest[name]; ok {
			delete(rest, name)
			ctx.SetVariable(name, bindArg(v))
		} else if def, ok := m.Defaults[name]; ok {
			if expr, isExpr := def.(parser.ExpressionNode); isExpr {
				v, err := e.callArg(expr, ctx)
				if err != nil {
					if errors.Is(err, ErrMaxRecursion) {
						return nil, err
					}
					return nil, fmt.Errorf("error evaluating default value for parameter %s: %w", name, err)
				}
				def = v
			}
			ctx.SetVariable(name, bindArg(def))
		} else {
			ctx.SetVariable(name, NewUndefined(name, behavior, nil))
		}
	}

	if c, ok := rest["caller"]; ok {
		if !m.Caller {
			return nil, fmt.Errorf("macro %q was called with a caller but does not use caller()", m.Name)
		}
		delete(rest, "caller")
		ctx.SetVariable("caller", c)
	} else if m.Caller {
		ctx.SetVariable("caller", &Undefined{Name: "caller", Behavior: behavior, Hint: "no caller defined"})
	}

	if m.CatchVarargs {
		varargs := []interface{}{}
		if len(args) > len(m.Parameters) {
			for _, a := range args[len(m.Parameters):] {
				varargs = append(varargs, unbox(a))
			}
		}
		ctx.SetVariable("varargs", varargs)
	}
	if m.CatchKwargs {
		for k, v := range rest {
			rest[k] = unbox(v)
		}
		ctx.SetVariable("kwargs", rest)
	} else if len(rest) > 0 {
		names := make([]string, 0, len(rest))
		for k := range rest {
			names = append(names, k)
		}
		sort.Strings(names)
		return nil, fmt.Errorf("macro %q takes no keyword argument %q", m.Name, names[0])
	}

	result, err := e.evalNodeList(m.Body, ctx)
	if err != nil {
		if errors.Is(err, ErrMaxRecursion) {
			// Already names the macro; wrapping every level would repeat it 200 times.
			return nil, err
		}
		return nil, fmt.Errorf("error executing macro %s: %w", m.Name, err)
	}
	if result == nil {
		return e.markIfSafe("", m.defCtx), nil
	}
	return e.markIfSafe(ToString(result), m.defCtx), nil
}
