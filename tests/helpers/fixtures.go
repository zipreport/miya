package helpers

import (
	"time"

	jinja2 "github.com/zipreport/miya"
)

// StandardUser returns a standard user object for testing
func StandardUser() map[string]interface{} {
	return map[string]interface{}{
		"id":         12345,
		"name":       "john doe",
		"first_name": "John",
		"last_name":  "Doe",
		"email":      "john.doe@example.com",
		"age":        25,
		"status":     "active",
		"is_premium": true,
		"is_admin":   false,
		"preferences": map[string]interface{}{
			"theme":         "dark",
			"language":      "english",
			"notifications": true,
		},
	}
}

// StandardProducts returns a list of test products
func StandardProducts() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":     "Laptop",
			"price":    999.99,
			"in_stock": true,
			"category": "Electronics",
		},
		{
			"name":     "Mouse",
			"price":    25.50,
			"in_stock": true,
			"category": "Accessories",
		},
		{
			"name":     "Keyboard",
			"price":    75.00,
			"in_stock": false,
			"category": "Accessories",
		},
		{
			"name":     "Monitor",
			"price":    299.99,
			"in_stock": true,
			"category": "Electronics",
		},
	}
}

// StandardContext returns a context with common test data
func StandardContext() map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"title":            "Test Template",
		"formatted_date":   now.Format("January 02, 2006"),
		"current_datetime": now,
		"description":      "This is a comprehensive test of template features",
		"html_content":     "<strong>Bold HTML</strong>",
		"price":            29.99,
		"tags":             []string{"python", "jinja2", "templates", "web", "html"},
		"empty_list":       []interface{}{},
		"none_value":       nil,
		"user":             StandardUser(),
		"products":         StandardProducts(),
		"categories": []map[string]interface{}{
			{
				"name": "Electronics",
				"items": []map[string]interface{}{
					{"name": "Laptop"},
					{"name": "Phone"},
				},
			},
			{
				"name": "Books",
				"items": []map[string]interface{}{
					{"name": "Python Guide"},
					{"name": "Web Development"},
				},
			},
		},
	}
}

// SimpleList returns a simple list for basic tests
func SimpleList() []int {
	return []int{1, 2, 3, 4, 5}
}

// SimpleDict returns a simple dictionary for basic tests
func SimpleDict() map[string]interface{} {
	return map[string]interface{}{
		"a": 1,
		"b": 2,
		"c": 3,
	}
}

// CreateContextFrom creates a context from provided data, with standard fallbacks
func CreateContextFrom(data map[string]interface{}) jinja2.Context {
	ctx := jinja2.NewContext()

	// Set provided data
	for key, value := range data {
		ctx.Set(key, value)
	}

	// Add standard test data if not provided
	standard := StandardContext()
	for key, value := range standard {
		if _, exists := data[key]; !exists {
			ctx.Set(key, value)
		}
	}

	return ctx
}
