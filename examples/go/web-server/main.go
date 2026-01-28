package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"github.com/zipreport/miya/parser"
)

type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
	InStock     bool
	Category    string
}

type PageData struct {
	Title       string
	CurrentTime time.Time
	User        map[string]interface{}
	Products    []Product
	Message     string
}

var env *miya.Environment

// SimpleTemplateParser implements loader.TemplateParser interface
type SimpleTemplateParser struct {
	env *miya.Environment
}

func NewSimpleTemplateParser(env *miya.Environment) *SimpleTemplateParser {
	return &SimpleTemplateParser{env: env}
}

func (stp *SimpleTemplateParser) ParseTemplate(name, content string) (*parser.TemplateNode, error) {
	template, err := stp.env.FromString(content)
	if err != nil {
		return nil, err
	}

	if templateNode, ok := template.AST().(*parser.TemplateNode); ok {
		templateNode.Name = name
		return templateNode, nil
	}

	return nil, fmt.Errorf("failed to extract template node")
}

func init() {
	// Create templates directory
	os.MkdirAll("templates", 0755)

	// Create base template
	baseTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{% block title %}{{ title }}{% endblock %}</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; }
        header { background: #333; color: white; padding: 1rem; margin: -20px -20px 20px -20px; }
        nav { background: #555; padding: 0.5rem; margin-top: 1rem; }
        nav a { color: white; text-decoration: none; margin: 0 1rem; }
        nav a:hover { text-decoration: underline; }
        .product { border: 1px solid #ddd; padding: 1rem; margin: 1rem 0; border-radius: 8px; }
        .price { font-size: 1.2em; color: #28a745; font-weight: bold; }
        .out-of-stock { color: #dc3545; }
        footer { margin-top: 2rem; padding-top: 1rem; border-top: 1px solid #ddd; text-align: center; }
        .message { padding: 1rem; margin: 1rem 0; border-radius: 4px; }
        .message.success { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
        .message.error { background: #f8d7da; color: #721c24; border: 1px solid #f5c6cb; }
    </style>
</head>
<body>
    <header>
        <h1>{% block header %}Miya Engine Web Example{% endblock %}</h1>
        <nav>
            <a href="/">Home</a>
            <a href="/products">Products</a>
            <a href="/about">About</a>
        </nav>
    </header>
    
    {% if message %}
    <div class="message success">{{ message }}</div>
    {% endif %}
    
    <main>
        {% block content %}
        <!-- Page content goes here -->
        {% endblock %}
    </main>
    
    <footer>
        {% block footer %}
        <p>Generated at {{ current_time.Format("2006-01-02 15:04:05") }} | 
           {% if user %}Logged in as {{ user.name }}{% else %}Not logged in{% endif %}</p>
        <p>&copy; 2024 Miya Engine Example. All rights reserved.</p>
        {% endblock %}
    </footer>
</body>
</html>`

	// Create home template
	homeTemplate := `{% extends "base.html" %}

{% block title %}Home - {{ super() }}{% endblock %}

{% block content %}
<h2>Welcome to Miya Engine Web Server Example!</h2>

<p>This example demonstrates using Jinja2 templates with a Go web server.</p>

<h3>Features Demonstrated:</h3>
<ul>
    <li>Template inheritance (base.html)</li>
    <li>Dynamic content rendering</li>
    <li>Loops and conditionals</li>
    <li>Filters and tests</li>
    <li>Auto-escaping for security</li>
</ul>

{% if user %}
<h3>User Information:</h3>
<p>Welcome back, <strong>{{ user.name|title }}</strong>!</p>
<p>Email: {{ user.email }}</p>
<p>Member since: {{ user.joined }}</p>
{% else %}
<p><em>Please log in to see personalized content.</em></p>
{% endif %}

<h3>Latest Products:</h3>
<div class="products">
{% for product in products[:3] %}
    <div class="product">
        <h4>{{ product.Name }}</h4>
        <p>{{ product.Description }}</p>
        <p class="price">${{ "%.2f"|format(product.Price) }}</p>
    </div>
{% else %}
    <p>No products available.</p>
{% endfor %}
</div>

<p><a href="/products">View all products →</a></p>
{% endblock %}`

	// Create products template
	productsTemplate := `{% extends "base.html" %}

{% block title %}Products - {{ super() }}{% endblock %}

{% block content %}
<h2>Our Products</h2>

<p>Browse our selection of {{ products|length }} products:</p>

{% set categories = products|groupby("Category") %}

{% for category, items in categories %}
<h3>{{ category }}</h3>
<div class="category-products">
    {% for product in items %}
    <div class="product">
        <h4>{{ product.Name }} (#{{ product.ID }})</h4>
        <p>{{ product.Description }}</p>
        <p class="price">${{ "%.2f"|format(product.Price) }}</p>
        {% if product.InStock %}
            <p style="color: green;">✓ In Stock</p>
        {% else %}
            <p class="out-of-stock">✗ Out of Stock</p>
        {% endif %}
    </div>
    {% endfor %}
</div>
{% endfor %}

<h3>Statistics:</h3>
<ul>
    <li>Total products: {{ products|length }}</li>
    <li>In stock: {{ products|selectattr("InStock")|list|length }}</li>
    <li>Out of stock: {{ products|rejectattr("InStock")|list|length }}</li>
    <li>Average price: ${{ "%.2f"|format(products|map(attribute="Price")|sum / products|length) }}</li>
</ul>
{% endblock %}`

	// Create about template
	aboutTemplate := `{% extends "base.html" %}

{% block title %}About - {{ super() }}{% endblock %}

{% block content %}
<h2>About This Example</h2>

<p>This web server example demonstrates the integration of Miya Engine template engine with a web application.</p>

<h3>Technical Details:</h3>
<table border="1" cellpadding="8">
    <tr>
        <th>Component</th>
        <th>Description</th>
    </tr>
    <tr>
        <td>Template Engine</td>
        <td>Miya Engine (Jinja2-compatible)</td>
    </tr>
    <tr>
        <td>Server</td>
        <td>Go net/http</td>
    </tr>
    <tr>
        <td>Template Loading</td>
        <td>FileSystemLoader</td>
    </tr>
    <tr>
        <td>Auto-escaping</td>
        <td>Enabled (XSS protection)</td>
    </tr>
</table>

<h3>Template Features Used:</h3>
<ul>
    <li><strong>Inheritance:</strong> All pages extend base.html</li>
    <li><strong>Blocks:</strong> title, content, footer</li>
    <li><strong>Filters:</strong> length, format, title, selectattr, rejectattr, map, sum</li>
    <li><strong>Tests:</strong> Conditional checks with if/else</li>
    <li><strong>Loops:</strong> Iterating over products and categories</li>
    <li><strong>Variables:</strong> Dynamic content from Go structs</li>
</ul>

<h3>Server Information:</h3>
<ul>
    <li>Current Time: {{ current_time.Format("Jan 02, 2006 3:04:05 PM MST") }}</li>
    <li>Go Version: {{ runtime_version }}</li>
</ul>
{% endblock %}`

	// Save templates
	os.WriteFile("templates/base.html", []byte(baseTemplate), 0644)
	os.WriteFile("templates/home.html", []byte(homeTemplate), 0644)
	os.WriteFile("templates/products.html", []byte(productsTemplate), 0644)
	os.WriteFile("templates/about.html", []byte(aboutTemplate), 0644)

	// Initialize Jinja2 environment
	env = miya.NewEnvironment()
	templateParser := NewSimpleTemplateParser(env)
	fsLoader := loader.NewFileSystemLoader([]string{"templates"}, templateParser)
	env.SetLoader(fsLoader)
}

func getSampleProducts() []Product {
	return []Product{
		{ID: 1, Name: "Laptop Pro", Description: "High-performance laptop for professionals", Price: 1299.99, InStock: true, Category: "Electronics"},
		{ID: 2, Name: "Wireless Mouse", Description: "Ergonomic wireless mouse with long battery life", Price: 29.99, InStock: true, Category: "Accessories"},
		{ID: 3, Name: "Mechanical Keyboard", Description: "RGB mechanical keyboard for gaming", Price: 149.99, InStock: false, Category: "Accessories"},
		{ID: 4, Name: "4K Monitor", Description: "27-inch 4K UHD monitor", Price: 399.99, InStock: true, Category: "Electronics"},
		{ID: 5, Name: "USB-C Hub", Description: "7-in-1 USB-C hub adapter", Price: 49.99, InStock: true, Category: "Accessories"},
		{ID: 6, Name: "Webcam HD", Description: "1080p HD webcam with microphone", Price: 79.99, InStock: false, Category: "Electronics"},
	}
}

func renderTemplate(w http.ResponseWriter, templateName string, data PageData) {
	tmpl, err := env.GetTemplate(templateName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template error: %v", err), http.StatusInternalServerError)
		return
	}

	ctx := miya.NewContext()
	ctx.Set("title", data.Title)
	ctx.Set("current_time", data.CurrentTime)
	ctx.Set("user", data.User)
	ctx.Set("products", data.Products)
	ctx.Set("message", data.Message)

	result, err := tmpl.Render(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Render error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(result))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Home",
		CurrentTime: time.Now(),
		User: map[string]interface{}{
			"name":   "john doe",
			"email":  "john@example.com",
			"joined": "2024-01-15",
		},
		Products: getSampleProducts(),
	}
	renderTemplate(w, "home.html", data)
}

func productsHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "Products",
		CurrentTime: time.Now(),
		Products:    getSampleProducts(),
	}
	renderTemplate(w, "products.html", data)
}

func aboutHandler(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:       "About",
		CurrentTime: time.Now(),
	}
	renderTemplate(w, "about.html", data)
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/products", productsHandler)
	http.HandleFunc("/about", aboutHandler)

	fmt.Println("Server starting on http://localhost:8080")
	fmt.Println("Press Ctrl+C to stop")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
