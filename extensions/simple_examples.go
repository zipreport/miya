package extensions

import (
	"fmt"
	"time"

	"github.com/zipreport/miya/lexer"
	"github.com/zipreport/miya/parser"
	"github.com/zipreport/miya/runtime"
)

// Type alias to make function signatures cleaner
type ExtensionNode = parser.ExtensionNode

// SimpleTimestampExtension provides basic timestamp functionality
type SimpleTimestampExtension struct {
	*BaseExtension
}

// NewSimpleTimestampExtension creates a new simple timestamp extension
func NewSimpleTimestampExtension() *SimpleTimestampExtension {
	return &SimpleTimestampExtension{
		BaseExtension: NewBaseExtension("timestamp", []string{"now", "timestamp"}),
	}
}

// ParseTag handles parsing of timestamp tags
func (ste *SimpleTimestampExtension) ParseTag(tagName string, parser ExtensionParser) (parser.Node, error) {
	startToken := parser.Current()
	node := parser.NewExtensionNode("timestamp", tagName, startToken.Line, startToken.Column)

	// Note: tag name is already consumed by ExtensionAwareParser

	switch tagName {
	case "now":
		// {% now %}
		node.SetEvaluateFunc(func(n *ExtensionNode, ctx interface{}) (interface{}, error) {
			return time.Now().Format("2006-01-02 15:04:05"), nil
		})

	case "timestamp":
		// {% timestamp %}
		node.SetEvaluateFunc(func(n *ExtensionNode, ctx interface{}) (interface{}, error) {
			return time.Now().Unix(), nil
		})
	}

	return node, parser.ExpectBlockEnd()
}

// HelloExtension provides a simple hello tag
type HelloExtension struct {
	*BaseExtension
}

// NewHelloExtension creates a new hello extension
func NewHelloExtension() *HelloExtension {
	return &HelloExtension{
		BaseExtension: NewBaseExtension("hello", []string{"hello"}),
	}
}

// ParseTag handles parsing of hello tags
func (he *HelloExtension) ParseTag(tagName string, parser ExtensionParser) (parser.Node, error) {
	startToken := parser.Current()
	node := parser.NewExtensionNode("hello", tagName, startToken.Line, startToken.Column)

	// Note: tag name is already consumed by ExtensionAwareParser

	// {% hello %}
	node.SetEvaluateFunc(func(n *ExtensionNode, ctx interface{}) (interface{}, error) {
		return "Hello from extension!", nil
	})

	return node, parser.ExpectBlockEnd()
}

// VersionExtension provides version information
type VersionExtension struct {
	*BaseExtension
	version string
}

// NewVersionExtension creates a new version extension
func NewVersionExtension(version string) *VersionExtension {
	return &VersionExtension{
		BaseExtension: NewBaseExtension("version", []string{"version"}),
		version:       version,
	}
}

// ParseTag handles parsing of version tags
func (ve *VersionExtension) ParseTag(tagName string, parser ExtensionParser) (parser.Node, error) {
	startToken := parser.Current()
	node := parser.NewExtensionNode("version", tagName, startToken.Line, startToken.Column)

	// Note: tag name is already consumed by ExtensionAwareParser

	// Capture the version in the closure
	version := ve.version

	// {% version %}
	node.SetEvaluateFunc(func(n *ExtensionNode, ctx interface{}) (interface{}, error) {
		return fmt.Sprintf("Version: %s", version), nil
	})

	return node, parser.ExpectBlockEnd()
}

// HighlightExtension provides syntax highlighting functionality
type HighlightExtension struct {
	*BaseExtension
}

// NewHighlightExtension creates a new highlight extension
func NewHighlightExtension() *HighlightExtension {
	blockTags := map[string]string{
		"highlight": "endhighlight",
	}
	return &HighlightExtension{
		BaseExtension: NewBlockExtension("highlight", blockTags),
	}
}

// ParseTag handles parsing of highlight tags
func (he *HighlightExtension) ParseTag(tagName string, parser ExtensionParser) (parser.Node, error) {
	startToken := parser.Current()
	node := parser.NewExtensionNode("highlight", tagName, startToken.Line, startToken.Column)

	// Note: tag name is already consumed by ExtensionAwareParser

	switch tagName {
	case "highlight":
		// Parse optional language argument: {% highlight "python" %}

		// Check if there's a language argument
		if !parser.Check(lexer.TokenBlockEnd) && !parser.Check(lexer.TokenBlockEndTrim) {
			langExpr, err := parser.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.AddArgument(langExpr)
		}

		// Expect block end %}
		err := parser.ExpectBlockEnd()
		if err != nil {
			return nil, err
		}

		// Parse block content until {% endhighlight %}
		bodyNodes, err := parser.ParseBlock("endhighlight")
		if err != nil {
			return nil, err
		}

		// Add body nodes
		for _, bodyNode := range bodyNodes {
			node.AddBodyNode(bodyNode)
		}

		// Set evaluation function
		node.SetEvaluateFunc(func(n *ExtensionNode, ctx interface{}) (interface{}, error) {
			// Get language from arguments if provided
			language := "text"
			if len(n.Arguments) > 0 {
				evaluator := runtime.NewEvaluator()
				runtimeCtx, ok := ctx.(runtime.Context)
				if !ok {
					return nil, fmt.Errorf("invalid context type")
				}

				langResult, err := evaluator.EvalNode(n.Arguments[0], runtimeCtx)
				if err != nil {
					return nil, err
				}
				if langStr, ok := langResult.(string); ok {
					language = langStr
				}
			}

			// Render the body content
			evaluator := runtime.NewEvaluator()
			runtimeCtx, ok := ctx.(runtime.Context)
			if !ok {
				return nil, fmt.Errorf("invalid context type")
			}

			var bodyContent string
			for _, bodyNode := range n.Body {
				result, err := evaluator.EvalNode(bodyNode, runtimeCtx)
				if err != nil {
					return nil, err
				}
				if result != nil {
					bodyContent += fmt.Sprintf("%v", result)
				}
			}

			// Simple syntax highlighting (just wrap in a div with class)
			return fmt.Sprintf(`<div class="highlight-%s"><pre>%s</pre></div>`, language, bodyContent), nil
		})

		return node, nil

	case "endhighlight":
		// End tags are handled automatically by the parser
		return node, nil
	}

	return nil, parser.Error(fmt.Sprintf("unknown highlight tag: %s", tagName))
}

// CacheExtension provides template fragment caching
type CacheExtension struct {
	*BaseExtension
}

// NewCacheExtension creates a new cache extension
func NewCacheExtension() *CacheExtension {
	blockTags := map[string]string{
		"cache": "endcache",
	}
	return &CacheExtension{
		BaseExtension: NewBlockExtension("cache", blockTags),
	}
}

// ParseTag handles parsing of cache tags
func (ce *CacheExtension) ParseTag(tagName string, parser ExtensionParser) (parser.Node, error) {
	startToken := parser.Current()
	node := parser.NewExtensionNode("cache", tagName, startToken.Line, startToken.Column)

	// Note: tag name is already consumed by ExtensionAwareParser

	switch tagName {
	case "cache":
		// Parse required timeout argument: {% cache 300 %}
		timeoutExpr, err := parser.ParseExpression()
		if err != nil {
			return nil, fmt.Errorf("cache tag requires timeout argument")
		}
		node.AddArgument(timeoutExpr)

		// Parse optional cache key
		if !parser.Check(lexer.TokenBlockEnd) && !parser.Check(lexer.TokenBlockEndTrim) {
			keyExpr, err := parser.ParseExpression()
			if err != nil {
				return nil, err
			}
			node.AddArgument(keyExpr)
		}

		// Expect block end %}
		err = parser.ExpectBlockEnd()
		if err != nil {
			return nil, err
		}

		// Parse block content until {% endcache %}
		bodyNodes, err := parser.ParseBlock("endcache")
		if err != nil {
			return nil, err
		}

		// Add body nodes
		for _, bodyNode := range bodyNodes {
			node.AddBodyNode(bodyNode)
		}

		// Set evaluation function
		node.SetEvaluateFunc(func(n *ExtensionNode, ctx interface{}) (interface{}, error) {
			// Simple cache implementation (in real implementation, you'd use Redis, etc.)
			// For now, just render the content without caching
			evaluator := runtime.NewEvaluator()
			runtimeCtx, ok := ctx.(runtime.Context)
			if !ok {
				return nil, fmt.Errorf("invalid context type")
			}

			var bodyContent string
			for _, bodyNode := range n.Body {
				result, err := evaluator.EvalNode(bodyNode, runtimeCtx)
				if err != nil {
					return nil, err
				}
				if result != nil {
					bodyContent += fmt.Sprintf("%v", result)
				}
			}

			return bodyContent, nil
		})

		return node, nil

	case "endcache":
		// End tags are handled automatically by the parser
		return node, nil
	}

	return nil, parser.Error(fmt.Sprintf("unknown cache tag: %s", tagName))
}
