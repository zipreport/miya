package runtime

import (
	"fmt"

	"github.com/zipreport/miya/parser"
)

// UndefinedBehavior defines how undefined variables should be handled
type UndefinedBehavior int

const (
	// UndefinedSilent returns empty string for undefined variables (default Jinja2 behavior)
	UndefinedSilent UndefinedBehavior = iota
	// UndefinedStrict raises an error when undefined variables are accessed
	UndefinedStrict
	// UndefinedDebug returns a debug string showing the undefined variable name
	UndefinedDebug
	// UndefinedChainFail fails on chained undefined access (e.g., undefined.attr)
	UndefinedChainFail
)

// Undefined represents an undefined value in templates
type Undefined struct {
	Name     string
	Behavior UndefinedBehavior
	Hint     string
	Node     parser.Node
}

// NewUndefined creates a new undefined value
func NewUndefined(name string, behavior UndefinedBehavior, node parser.Node) *Undefined {
	return &Undefined{
		Name:     name,
		Behavior: behavior,
		Node:     node,
	}
}

// NewStrictUndefined creates a new strict undefined value
func NewStrictUndefined(name string, node parser.Node) *Undefined {
	return &Undefined{
		Name:     name,
		Behavior: UndefinedStrict,
		Node:     node,
	}
}

// NewDebugUndefined creates a debug undefined value with hint
func NewDebugUndefined(name string, hint string, node parser.Node) *Undefined {
	return &Undefined{
		Name:     name,
		Behavior: UndefinedDebug,
		Hint:     hint,
		Node:     node,
	}
}

// String returns the string representation based on behavior
func (u *Undefined) String() string {
	switch u.Behavior {
	case UndefinedSilent:
		return ""
	case UndefinedDebug:
		if u.Hint != "" {
			return fmt.Sprintf("{{ undefined variable: %s (%s) }}", u.Name, u.Hint)
		}
		return fmt.Sprintf("{{ undefined variable: %s }}", u.Name)
	case UndefinedStrict:
		// This should not be reached as strict undefined should error before string conversion
		return fmt.Sprintf("StrictUndefined: %s", u.Name)
	case UndefinedChainFail:
		return ""
	default:
		return ""
	}
}

// Error returns an error for strict undefined behavior
func (u *Undefined) Error() error {
	switch u.Behavior {
	case UndefinedStrict:
		return NewUndefinedVariableError(u.Name, u.Node)
	case UndefinedChainFail:
		return NewUndefinedVariableError(u.Name, u.Node)
	default:
		return nil
	}
}

// IsUndefined checks if value is undefined
func IsUndefined(value interface{}) bool {
	_, ok := value.(*Undefined)
	return ok
}

// StrictUndefinedFactory creates strict undefined values
type StrictUndefinedFactory struct{}

// NewStrictUndefinedFactory creates a factory for strict undefined values
func NewStrictUndefinedFactory() *StrictUndefinedFactory {
	return &StrictUndefinedFactory{}
}

// Create creates a new strict undefined value
func (f *StrictUndefinedFactory) Create(name string, node parser.Node) *Undefined {
	return NewStrictUndefined(name, node)
}

// UndefinedHandler handles undefined variable access based on configuration
type UndefinedHandler struct {
	behavior         UndefinedBehavior
	undefinedFactory func(string, parser.Node) *Undefined
}

// NewUndefinedHandler creates a new undefined handler
func NewUndefinedHandler(behavior UndefinedBehavior) *UndefinedHandler {
	handler := &UndefinedHandler{
		behavior: behavior,
	}

	switch behavior {
	case UndefinedStrict:
		handler.undefinedFactory = NewStrictUndefined
	case UndefinedDebug:
		handler.undefinedFactory = func(name string, node parser.Node) *Undefined {
			return NewDebugUndefined(name, "variable not found in context", node)
		}
	case UndefinedChainFail:
		handler.undefinedFactory = func(name string, node parser.Node) *Undefined {
			return NewUndefined(name, UndefinedChainFail, node)
		}
	default:
		handler.undefinedFactory = func(name string, node parser.Node) *Undefined {
			return NewUndefined(name, UndefinedSilent, node)
		}
	}

	return handler
}

// Handle handles an undefined variable access
func (h *UndefinedHandler) Handle(name string, node parser.Node) (interface{}, error) {
	undefined := h.undefinedFactory(name, node)

	if h.behavior == UndefinedStrict {
		return nil, undefined.Error()
	}

	return undefined, nil
}

// HandleAttributeAccess handles attribute access on undefined values
func (h *UndefinedHandler) HandleAttributeAccess(undefined *Undefined, attrName string, node parser.Node) (interface{}, error) {
	switch undefined.Behavior {
	case UndefinedStrict, UndefinedChainFail:
		chainedName := fmt.Sprintf("%s.%s", undefined.Name, attrName)
		return nil, NewUndefinedVariableError(chainedName, node)
	case UndefinedDebug:
		chainedName := fmt.Sprintf("%s.%s", undefined.Name, attrName)
		return NewDebugUndefined(chainedName, "chained attribute access on undefined", node), nil
	default:
		// Silent behavior - return another undefined
		chainedName := fmt.Sprintf("%s.%s", undefined.Name, attrName)
		return NewUndefined(chainedName, UndefinedSilent, node), nil
	}
}

// HandleItemAccess handles item access on undefined values
func (h *UndefinedHandler) HandleItemAccess(undefined *Undefined, key interface{}, node parser.Node) (interface{}, error) {
	switch undefined.Behavior {
	case UndefinedStrict, UndefinedChainFail:
		chainedName := fmt.Sprintf("%s[%v]", undefined.Name, key)
		return nil, NewUndefinedVariableError(chainedName, node)
	case UndefinedDebug:
		chainedName := fmt.Sprintf("%s[%v]", undefined.Name, key)
		return NewDebugUndefined(chainedName, "item access on undefined", node), nil
	default:
		// Silent behavior
		chainedName := fmt.Sprintf("%s[%v]", undefined.Name, key)
		return NewUndefined(chainedName, UndefinedSilent, node), nil
	}
}

// HandleFunctionCall handles function calls on undefined values
func (h *UndefinedHandler) HandleFunctionCall(undefined *Undefined, args []interface{}, node parser.Node) (interface{}, error) {
	switch undefined.Behavior {
	case UndefinedStrict, UndefinedChainFail:
		return nil, NewUndefinedVariableError(fmt.Sprintf("%s()", undefined.Name), node)
	case UndefinedDebug:
		return NewDebugUndefined(fmt.Sprintf("%s()", undefined.Name), "function call on undefined", node), nil
	default:
		return NewUndefined(fmt.Sprintf("%s()", undefined.Name), UndefinedSilent, node), nil
	}
}

// IsCallableUndefined checks if an undefined value should allow function calls
func (h *UndefinedHandler) IsCallableUndefined(value interface{}) bool {
	if undefined, ok := value.(*Undefined); ok {
		return undefined.Behavior == UndefinedSilent || undefined.Behavior == UndefinedDebug
	}
	return false
}

// GetUndefinedBehavior returns the current undefined behavior
func (h *UndefinedHandler) GetUndefinedBehavior() UndefinedBehavior {
	return h.behavior
}

// SetUndefinedBehavior sets the undefined behavior
func (h *UndefinedHandler) SetUndefinedBehavior(behavior UndefinedBehavior) {
	h.behavior = behavior
	h.undefinedFactory = NewUndefinedHandler(behavior).undefinedFactory
}
