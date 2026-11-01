package runtime

import "testing"

// scopeCountingContext records whether loops asked it for a child scope or a
// full copy.
type scopeCountingContext struct {
	*simpleContext
	children, clones *int
}

func (c *scopeCountingContext) Child() Context {
	*c.children++
	return c
}

func (c *scopeCountingContext) Clone() Context {
	*c.clones++
	return c
}

// A for-loop inside {% autoescape %} or a ContextWrapper must take a child
// scope of the inner context, not copy every variable per iteration.
func TestEscapeWrappersTakeChildScope(t *testing.T) {
	var children, clones int
	inner := &scopeCountingContext{&simpleContext{variables: map[string]interface{}{}}, &children, &clones}
	wrappers := []Context{
		&autoescapeContext{Context: inner, autoescape: true},
		NewContextWrapper(inner, NewAutoEscaper(nil), EscapeContextJS),
	}
	for _, w := range wrappers {
		children, clones = 0, 0
		c, ok := w.(childScoper)
		if !ok {
			t.Fatalf("%T has no Child", w)
		}
		child := c.Child()
		if children != 1 || clones != 0 {
			t.Fatalf("%T.Child: inner Child %d times, Clone %d times", w, children, clones)
		}
		if on, ec, _ := resolveEscape(child); !on || ec != escapeContextOf(w) {
			t.Fatalf("%T.Child lost the escape setting: on=%v ec=%v", w, on, ec)
		}
	}
}

func escapeContextOf(ctx Context) EscapeContext {
	_, ec, _ := resolveEscape(ctx)
	return ec
}
