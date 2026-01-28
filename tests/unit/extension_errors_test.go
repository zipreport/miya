package miya_test

import (
	"errors"
	miya "github.com/zipreport/miya"
	"strings"
	"testing"

	"github.com/zipreport/miya/extensions"
	"github.com/zipreport/miya/parser"
)

// ErrorTestExtension is designed to trigger various error conditions
type ErrorTestExtension struct {
	*extensions.BaseExtension
	OnLoadError       error
	BeforeRenderError error
	AfterRenderError  error
	OnUnloadError     error
	ParseError        error
}

func NewErrorTestExtension() *ErrorTestExtension {
	return &ErrorTestExtension{
		BaseExtension: extensions.NewBaseExtension("error_test", []string{"error"}),
	}
}

func (ete *ErrorTestExtension) OnLoad(env extensions.ExtensionEnvironment) error {
	return ete.OnLoadError
}

func (ete *ErrorTestExtension) BeforeRender(ctx extensions.ExtensionContext, templateName string) error {
	return ete.BeforeRenderError
}

func (ete *ErrorTestExtension) AfterRender(ctx extensions.ExtensionContext, templateName string, result interface{}, err error) error {
	return ete.AfterRenderError
}

func (ete *ErrorTestExtension) OnUnload() error {
	return ete.OnUnloadError
}

func (ete *ErrorTestExtension) ParseTag(tagName string, parser extensions.ExtensionParser) (parser.Node, error) {
	if ete.ParseError != nil {
		return nil, ete.ParseError
	}

	startToken := parser.Current()
	node := parser.NewExtensionNode("error_test", tagName, startToken.Line, startToken.Column)

	node.SetEvaluateFunc(func(node *ExtensionNode, ctx interface{}) (interface{}, error) {
		return "error test", nil
	})

	return node, parser.ExpectBlockEnd()
}

func TestExtensionErrorMessages(t *testing.T) {
	tests := []struct {
		name            string
		setupError      func(*ErrorTestExtension)
		expectedMessage string
		errorType       string
	}{
		{
			name: "OnLoad error",
			setupError: func(ext *ErrorTestExtension) {
				ext.OnLoadError = errors.New("failed to initialize")
			},
			expectedMessage: "extension 'error_test' error: OnLoad lifecycle hook failed",
			errorType:       "registration",
		},
		{
			name: "BeforeRender error",
			setupError: func(ext *ErrorTestExtension) {
				ext.BeforeRenderError = errors.New("failed to prepare")
			},
			expectedMessage: "extension 'error_test' error in template 'test-template': BeforeRender lifecycle hook failed",
			errorType:       "render",
		},
		{
			name: "AfterRender error",
			setupError: func(ext *ErrorTestExtension) {
				ext.AfterRenderError = errors.New("failed to cleanup")
			},
			expectedMessage: "extension 'error_test' error in template 'test-template': AfterRender lifecycle hook failed",
			errorType:       "render",
		},
		{
			name: "OnUnload error",
			setupError: func(ext *ErrorTestExtension) {
				ext.OnUnloadError = errors.New("failed to cleanup")
			},
			expectedMessage: "extension 'error_test' error: OnUnload lifecycle hook failed",
			errorType:       "unregistration",
		},
		{
			name: "Parse error",
			setupError: func(ext *ErrorTestExtension) {
				ext.ParseError = errors.New("invalid tag syntax")
			},
			expectedMessage: "extension 'error_test' error in tag 'error' at line",
			errorType:       "parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := extensions.NewRegistry()
			env := miya.NewEnvironment()
			registry.SetEnvironment(env)

			ext := NewErrorTestExtension()
			tt.setupError(ext)

			switch tt.errorType {
			case "registration":
				err := registry.Register(ext)
				if err == nil {
					t.Fatal("Expected error during registration")
				}
				if !strings.Contains(err.Error(), tt.expectedMessage) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.expectedMessage, err.Error())
				}

				// Verify it's an ExtensionError
				if extErr, ok := err.(*extensions.ExtensionError); ok {
					if extErr.ExtensionName != "error_test" {
						t.Errorf("Expected extension name 'error_test', got '%s'", extErr.ExtensionName)
					}
				} else {
					t.Error("Expected ExtensionError type")
				}

			case "render":
				// First register without error
				ext.OnLoadError = nil
				err := registry.Register(ext)
				if err != nil {
					t.Fatalf("Failed to register extension: %v", err)
				}

				// Create test context
				ctx := &extensionContextAdapter{ctx: miya.NewTemplateContextAdapter(miya.NewContext(), env)}

				if tt.name == "BeforeRender error" {
					err = registry.BeforeRender(ctx, "test-template")
				} else {
					err = registry.AfterRender(ctx, "test-template", "result", nil)
				}

				if err == nil {
					t.Fatal("Expected error during render hook")
				}
				if !strings.Contains(err.Error(), tt.expectedMessage) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.expectedMessage, err.Error())
				}

			case "unregistration":
				// First register without error
				ext.OnLoadError = nil
				err := registry.Register(ext)
				if err != nil {
					t.Fatalf("Failed to register extension: %v", err)
				}

				// Now unregister with error
				err = registry.Unregister("error_test")
				if err == nil {
					t.Fatal("Expected error during unregistration")
				}
				if !strings.Contains(err.Error(), tt.expectedMessage) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.expectedMessage, err.Error())
				}

			case "parse":
				// First register without error
				ext.OnLoadError = nil
				err := registry.Register(ext)
				if err != nil {
					t.Fatalf("Failed to register extension: %v", err)
				}

				// Try to parse template with error
				template := `{% error %}`
				tokens, err := extensions.CreateTokensFromString(template)
				if err != nil {
					t.Fatalf("Failed to create tokens: %v", err)
				}

				extParser := extensions.NewExtensionAwareParser(tokens, registry)
				_, err = extParser.ParseTopLevelPublic()
				if err == nil {
					t.Fatal("Expected error during parsing")
				}
				if !strings.Contains(err.Error(), tt.expectedMessage) {
					t.Errorf("Expected error message to contain '%s', got: %s", tt.expectedMessage, err.Error())
				}
			}
		})
	}
}

func TestExtensionConflictErrors(t *testing.T) {
	registry := extensions.NewRegistry()
	env := miya.NewEnvironment()
	registry.SetEnvironment(env)

	// Register first extension
	ext1 := NewErrorTestExtension()
	err := registry.Register(ext1)
	if err != nil {
		t.Fatalf("Failed to register first extension: %v", err)
	}

	// Try to register another extension with same name
	ext2 := NewErrorTestExtension()
	err = registry.Register(ext2)
	if err == nil {
		t.Fatal("Expected error when registering extension with duplicate name")
	}

	expectedMsg := "extension 'error_test' error: extension is already registered"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}

	// Verify it's an ExtensionError with correct fields
	if extErr, ok := err.(*extensions.ExtensionError); ok {
		if extErr.ExtensionName != "error_test" {
			t.Errorf("Expected extension name 'error_test', got '%s'", extErr.ExtensionName)
		}
	} else {
		t.Error("Expected ExtensionError type")
	}
}

func TestExtensionTagConflictErrors(t *testing.T) {
	registry := extensions.NewRegistry()
	env := miya.NewEnvironment()
	registry.SetEnvironment(env)

	// Register first extension
	ext1 := NewErrorTestExtension()
	err := registry.Register(ext1)
	if err != nil {
		t.Fatalf("Failed to register first extension: %v", err)
	}

	// Create conflicting extension with same tag
	conflictingExt := extensions.NewBaseExtension("conflict", []string{"error"})
	ext2 := &ErrorTestExtension{BaseExtension: conflictingExt}

	err = registry.Register(ext2)
	if err == nil {
		t.Fatal("Expected error when registering extension with conflicting tag")
	}

	expectedMsg := "extension 'conflict' error in tag 'error': tag is already handled by extension 'error_test'"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}

	// Verify it's an ExtensionError with correct fields
	if extErr, ok := err.(*extensions.ExtensionError); ok {
		if extErr.ExtensionName != "conflict" {
			t.Errorf("Expected extension name 'conflict', got '%s'", extErr.ExtensionName)
		}
		if extErr.TagName != "error" {
			t.Errorf("Expected tag name 'error', got '%s'", extErr.TagName)
		}
	} else {
		t.Error("Expected ExtensionError type")
	}
}

func TestExtensionDependencyErrors(t *testing.T) {
	registry := extensions.NewRegistry()
	env := miya.NewEnvironment()
	registry.SetEnvironment(env)

	// Create extension with missing dependency
	dependentExt := NewDependentExtension()
	err := registry.Register(dependentExt)
	if err == nil {
		t.Fatal("Expected error when registering extension with missing dependency")
	}

	expectedMsg := "extension 'dependent' error: depends on 'lifecycle' which is not registered"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}

	// Verify it's an ExtensionError
	if extErr, ok := err.(*extensions.ExtensionError); ok {
		if extErr.ExtensionName != "dependent" {
			t.Errorf("Expected extension name 'dependent', got '%s'", extErr.ExtensionName)
		}
	} else {
		t.Error("Expected ExtensionError type")
	}
}

func TestExtensionUnregisterDependencyErrors(t *testing.T) {
	registry := extensions.NewRegistry()
	env := miya.NewEnvironment()
	registry.SetEnvironment(env)

	// Register base extension
	baseExt := NewLifecycleTestExtension()
	err := registry.Register(baseExt)
	if err != nil {
		t.Fatalf("Failed to register base extension: %v", err)
	}

	// Register dependent extension
	dependentExt := NewDependentExtension()
	err = registry.Register(dependentExt)
	if err != nil {
		t.Fatalf("Failed to register dependent extension: %v", err)
	}

	// Try to unregister base extension while dependent still exists
	err = registry.Unregister("lifecycle")
	if err == nil {
		t.Fatal("Expected error when unregistering extension with dependents")
	}

	expectedMsg := "extension 'lifecycle' error: cannot unregister: extension 'dependent' depends on it"
	if err.Error() != expectedMsg {
		t.Errorf("Expected '%s', got '%s'", expectedMsg, err.Error())
	}

	// Verify it's an ExtensionError
	if extErr, ok := err.(*extensions.ExtensionError); ok {
		if extErr.ExtensionName != "lifecycle" {
			t.Errorf("Expected extension name 'lifecycle', got '%s'", extErr.ExtensionName)
		}
	} else {
		t.Error("Expected ExtensionError type")
	}
}
