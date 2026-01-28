package miya

// ContextInterface defines the shared context interface used across packages
type ContextInterface interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Push() ContextInterface
	Pop() ContextInterface
	All() map[string]interface{}
}
