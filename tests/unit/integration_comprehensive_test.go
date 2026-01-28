package miya_test

import (
	miya "github.com/zipreport/miya"
	"github.com/zipreport/miya/loader"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mustParseTime parses a date string and panics on error (for test data)
func mustParseTime(dateStr string) time.Time {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		panic(err)
	}
	return t
}

// =============================================================================
// COMPREHENSIVE INTEGRATION TESTS
// =============================================================================
// This file provides integration tests for real-world usage patterns
// to improve overall test coverage and validate end-to-end functionality
// =============================================================================

// Test Complete Web Application Template System
func TestWebApplicationTemplateSystem(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jinja2_webapp_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a realistic web application template structure
	templates := map[string]string{
		"layouts/base.html": `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{% block title %}{{ site_name|default('My Website') }}{% endblock %}</title>
    {% block extra_head %}{% endblock %}
    <link rel="stylesheet" href="/static/css/main.css">
</head>
<body class="{% block body_class %}{% endblock %}">
    {% include "partials/header.html" %}
    
    <main class="container">
        {% if messages %}
            {% for message in messages %}
                <div class="alert alert-{{ message.type }}">
                    {{ message.text|escape }}
                    {% if message.dismissible %}
                        <button type="button" class="btn-close" data-dismiss="alert"></button>
                    {% endif %}
                </div>
            {% endfor %}
        {% endif %}
        
        {% block breadcrumb %}{% endblock %}
        
        <div class="content">
            {% block content %}{% endblock %}
        </div>
    </main>
    
    {% include "partials/footer.html" %}
    
    {% block extra_js %}{% endblock %}
</body>
</html>`,

		"partials/header.html": `<header class="navbar">
    <div class="navbar-brand">
        <a href="/">{{ site_name|default('My Website') }}</a>
    </div>
    <nav class="navbar-nav">
        {% for item in navigation %}
            <a href="{{ item.url }}" class="nav-link{% if item.active %} active{% endif %}">
                {{ item.title }}
            </a>
        {% endfor %}
    </nav>
    <div class="navbar-user">
        {% if user %}
            <span>Welcome, {{ user.name }}!</span>
            <a href="/logout">Logout</a>
        {% else %}
            <a href="/login">Login</a>
        {% endif %}
    </div>
</header>`,

		"partials/footer.html": `<footer class="footer">
    <div class="container">
        <div class="row">
            <div class="col-md-6">
                <p>&copy; {{ current_year }} {{ site_name|default('My Website') }}. All rights reserved.</p>
            </div>
            <div class="col-md-6">
                <ul class="footer-links">
                    {% for link in footer_links %}
                        <li><a href="{{ link.url }}">{{ link.title }}</a></li>
                    {% endfor %}
                </ul>
            </div>
        </div>
    </div>
</footer>`,

		"pages/home.html": `{% extends "layouts/base.html" %}

{% block title %}Home - {{ super() }}{% endblock %}

{% block body_class %}home-page{% endblock %}

{% block content %}
    <div class="hero-section">
        <h1>{{ hero.title|default('Welcome to Our Site') }}</h1>
        <p class="lead">{{ hero.subtitle|default('Discover amazing content') }}</p>
        {% if hero.cta %}
            <a href="{{ hero.cta.url }}" class="btn btn-primary">{{ hero.cta.text }}</a>
        {% endif %}
    </div>
    
    {% if featured_posts %}
        <section class="featured-posts">
            <h2>Featured Posts</h2>
            <div class="post-grid">
                {% for post in featured_posts %}
                    {% include "components/post_card.html" %}
                {% endfor %}
            </div>
        </section>
    {% endif %}
    
    {% if stats %}
        <section class="stats-section">
            <h2>Our Impact</h2>
            <div class="stats-grid">
                {% for stat in stats %}
                    <div class="stat-item">
                        <div class="stat-number">{{ stat.value|default(0) }}</div>
                        <div class="stat-label">{{ stat.label }}</div>
                    </div>
                {% endfor %}
            </div>
        </section>
    {% endif %}
{% endblock %}`,

		"components/post_card.html": `<article class="post-card">
    {% if post.featured_image %}
        <div class="post-image">
            <img src="{{ post.featured_image }}" alt="{{ post.title|escape }}">
        </div>
    {% endif %}
    <div class="post-content">
        <h3><a href="{{ post.url }}">{{ post.title }}</a></h3>
        <p class="post-excerpt">{{ post.excerpt|truncate(150) }}</p>
        <div class="post-meta">
            <span class="post-author">By {{ post.author.name }}</span>
            <span class="post-date">{{ post.published_at|date('%B %d, %Y') }}</span>
            {% if post.tags %}
                <div class="post-tags">
                    {% for tag in post.tags %}
                        <span class="tag">{{ tag.name }}</span>
                    {% endfor %}
                </div>
            {% endif %}
        </div>
    </div>
</article>`,

		"pages/blog_post.html": `{% extends "layouts/base.html" %}

{% block title %}{{ post.title }} - {{ super() }}{% endblock %}

{% block extra_head %}
    <meta name="description" content="{{ post.excerpt|striptags|truncate(160) }}">
    <meta property="og:title" content="{{ post.title }}">
    <meta property="og:description" content="{{ post.excerpt|striptags|truncate(160) }}">
    {% if post.featured_image %}
        <meta property="og:image" content="{{ post.featured_image }}">
    {% endif %}
{% endblock %}

{% block body_class %}blog-post-page{% endblock %}

{% block breadcrumb %}
    <nav class="breadcrumb">
        <a href="/">Home</a> &gt; 
        <a href="/blog">Blog</a> &gt; 
        <span>{{ post.title|truncate(50) }}</span>
    </nav>
{% endblock %}

{% block content %}
    <article class="blog-post">
        <header class="post-header">
            <h1>{{ post.title }}</h1>
            <div class="post-meta">
                <span class="author">
                    By <a href="{{ post.author.profile_url }}">{{ post.author.name }}</a>
                </span>
                <span class="date">{{ post.published_at|date('%B %d, %Y') }}</span>
                <span class="reading-time">{{ post.reading_time }} min read</span>
            </div>
            {% if post.tags %}
                <div class="post-tags">
                    {% for tag in post.tags %}
                        <a href="/blog/tags/{{ tag.slug }}" class="tag">{{ tag.name }}</a>
                    {% endfor %}
                </div>
            {% endif %}
        </header>
        
        {% if post.featured_image %}
            <div class="post-featured-image">
                <img src="{{ post.featured_image }}" alt="{{ post.title|escape }}">
            </div>
        {% endif %}
        
        <div class="post-content">
            {{ post.content|safe }}
        </div>
        
        {% if post.author %}
            <div class="author-bio">
                <h3>About {{ post.author.name }}</h3>
                <div class="author-info">
                    {% if post.author.avatar %}
                        <img src="{{ post.author.avatar }}" alt="{{ post.author.name|escape }}" class="author-avatar">
                    {% endif %}
                    <div class="author-details">
                        <p>{{ post.author.bio }}</p>
                        {% if post.author.social_links %}
                            <div class="author-social">
                                {% for link in post.author.social_links %}
                                    <a href="{{ link.url }}" class="social-link">{{ link.platform }}</a>
                                {% endfor %}
                            </div>
                        {% endif %}
                    </div>
                </div>
            </div>
        {% endif %}
    </article>
    
    {% if related_posts %}
        <section class="related-posts">
            <h2>Related Posts</h2>
            <div class="post-grid">
                {% for post in related_posts %}
                    {% include "components/post_card.html" %}
                {% endfor %}
            </div>
        </section>
    {% endif %}
{% endblock %}`,

		"forms/contact.html": `{% extends "layouts/base.html" %}
{% from "macros/forms.html" import input_field, textarea_field, select_field %}

{% block title %}Contact Us - {{ super() }}{% endblock %}

{% block content %}
    <div class="contact-page">
        <div class="row">
            <div class="col-md-8">
                <h1>Get in Touch</h1>
                
                <form method="post" action="/contact" class="contact-form">
                    {% if form_errors %}
                        <div class="alert alert-danger">
                            <ul>
                                {% for field, errors in form_errors.items() %}
                                    {% for error in errors %}
                                        <li>{{ field|title }}: {{ error }}</li>
                                    {% endfor %}
                                {% endfor %}
                            </ul>
                        </div>
                    {% endif %}
                    
                    {{ input_field('name', 'text', form_data.name|default(''), 'Full Name', required=true) }}
                    {{ input_field('email', 'email', form_data.email|default(''), 'Email Address', required=true) }}
                    {{ select_field('subject', form_data.subject|default(''), 'Subject', subject_options, required=true) }}
                    {{ textarea_field('message', form_data.message|default(''), 'Your Message', rows=5, required=true) }}
                    
                    <button type="submit" class="btn btn-primary">Send Message</button>
                </form>
            </div>
            
            <div class="col-md-4">
                <div class="contact-info">
                    <h3>Contact Information</h3>
                    {% if contact_info %}
                        {% if contact_info.email %}
                            <p><strong>Email:</strong> <a href="mailto:{{ contact_info.email }}">{{ contact_info.email }}</a></p>
                        {% endif %}
                        {% if contact_info.phone %}
                            <p><strong>Phone:</strong> <a href="tel:{{ contact_info.phone }}">{{ contact_info.phone }}</a></p>
                        {% endif %}
                        {% if contact_info.address %}
                            <p><strong>Address:</strong><br>{{ contact_info.address|nl2br|safe }}</p>
                        {% endif %}
                    {% endif %}
                </div>
            </div>
        </div>
    </div>
{% endblock %}`,

		"macros/forms.html": `{% macro input_field(name, type, value, label, required=false, placeholder='', class='form-control') %}
    <div class="form-group">
        {% if label %}
            <label for="{{ name }}">
                {{ label }}
                {% if required %}<span class="required">*</span>{% endif %}
            </label>
        {% endif %}
        <input 
            type="{{ type }}" 
            name="{{ name }}" 
            id="{{ name }}"
            value="{{ value }}" 
            class="{{ class }}"
            {% if placeholder %}placeholder="{{ placeholder }}"{% endif %}
            {% if required %}required{% endif %}
        >
    </div>
{% endmacro %}

{% macro textarea_field(name, value, label, rows=3, required=false, placeholder='', class='form-control') %}
    <div class="form-group">
        {% if label %}
            <label for="{{ name }}">
                {{ label }}
                {% if required %}<span class="required">*</span>{% endif %}
            </label>
        {% endif %}
        <textarea 
            name="{{ name }}" 
            id="{{ name }}"
            rows="{{ rows }}"
            class="{{ class }}"
            {% if placeholder %}placeholder="{{ placeholder }}"{% endif %}
            {% if required %}required{% endif %}
        >{{ value }}</textarea>
    </div>
{% endmacro %}

{% macro select_field(name, value, label, options, required=false, class='form-control') %}
    <div class="form-group">
        {% if label %}
            <label for="{{ name }}">
                {{ label }}
                {% if required %}<span class="required">*</span>{% endif %}
            </label>
        {% endif %}
        <select 
            name="{{ name }}" 
            id="{{ name }}"
            class="{{ class }}"
            {% if required %}required{% endif %}
        >
            <option value="">Choose...</option>
            {% for option_value, option_label in options %}
                <option value="{{ option_value }}"{% if option_value == value %} selected{% endif %}>
                    {{ option_label }}
                </option>
            {% endfor %}
        </select>
    </div>
{% endmacro %}`,
	}

	// Write all templates to files
	for templatePath, content := range templates {
		fullPath := filepath.Join(tmpDir, templatePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template %s: %v", templatePath, err)
		}
	}

	// Create environment with filesystem loader
	directParser := loader.NewDirectTemplateParser()
	fsLoader := loader.NewFileSystemLoader([]string{tmpDir}, directParser)
	env := miya.NewEnvironment(miya.WithLoader(fsLoader), miya.WithAutoEscape(true))

	// Test complete web application scenarios
	tests := []struct {
		name            string
		templateName    string
		data            map[string]interface{}
		expectedContent []string
	}{
		{
			name:         "home page rendering",
			templateName: "pages/home.html",
			data: map[string]interface{}{
				"site_name":    "Test Website",
				"current_year": 2024,
				"user": map[string]interface{}{
					"name": "John Doe",
				},
				"navigation": []interface{}{
					map[string]interface{}{"url": "/", "title": "Home", "active": true},
					map[string]interface{}{"url": "/blog", "title": "Blog", "active": false},
					map[string]interface{}{"url": "/contact", "title": "Contact", "active": false},
				},
				"hero": map[string]interface{}{
					"title":    "Welcome to Test Website",
					"subtitle": "Your amazing journey starts here",
					"cta": map[string]interface{}{
						"url":  "/get-started",
						"text": "Get Started",
					},
				},
				"featured_posts": []interface{}{
					map[string]interface{}{
						"title":        "First Post",
						"excerpt":      "This is the excerpt of the first post.",
						"url":          "/blog/first-post",
						"author":       map[string]interface{}{"name": "Jane Author"},
						"published_at": mustParseTime("2024-01-15"),
						"tags": []interface{}{
							map[string]interface{}{"name": "Tech"},
							map[string]interface{}{"name": "Tutorial"},
						},
					},
				},
				"stats": []interface{}{
					map[string]interface{}{"value": "10,000", "label": "Happy Users"},
					map[string]interface{}{"value": "500", "label": "Projects Completed"},
				},
				"messages": []interface{}{}, // Add empty messages array
				"footer_links": []interface{}{ // Add empty footer_links array
					map[string]interface{}{"url": "/privacy", "title": "Privacy Policy"},
					map[string]interface{}{"url": "/terms", "title": "Terms of Service"},
				},
			},
			expectedContent: []string{
				"Test Website",
				"Welcome, John Doe!",
				"Welcome to Test Website",
				"Your amazing journey starts here",
				"Get Started",
				"First Post",
				"Jane Author",
				"Happy Users",
				"Projects Completed",
				"&copy; 2024 Test Website",
			},
		},
		{
			name:         "blog post rendering",
			templateName: "pages/blog_post.html",
			data: map[string]interface{}{
				"site_name":    "Test Blog",
				"current_year": 2024,
				"navigation": []interface{}{
					map[string]interface{}{"url": "/", "title": "Home", "active": false},
					map[string]interface{}{"url": "/blog", "title": "Blog", "active": true},
				},
				"messages": []interface{}{},
				"footer_links": []interface{}{
					map[string]interface{}{"url": "/privacy", "title": "Privacy Policy"},
				},
				"post": map[string]interface{}{
					"title":          "How to Use Go Templates",
					"excerpt":        "Learn the fundamentals of Go template system",
					"content":        "<p>This is the full content of the blog post with <strong>HTML</strong>.</p>",
					"url":            "/blog/go-templates",
					"reading_time":   5,
					"published_at":   mustParseTime("2024-01-20"),
					"featured_image": "/images/go-templates.jpg",
					"author": map[string]interface{}{
						"name":        "Tech Writer",
						"profile_url": "/authors/tech-writer",
						"bio":         "A passionate technical writer with 10 years of experience.",
						"avatar":      "/avatars/tech-writer.jpg",
						"social_links": []interface{}{
							map[string]interface{}{"platform": "Twitter", "url": "https://twitter.com/techwriter"},
						},
					},
					"tags": []interface{}{
						map[string]interface{}{"name": "Go", "slug": "go"},
						map[string]interface{}{"name": "Programming", "slug": "programming"},
					},
				},
				"related_posts": []interface{}{
					map[string]interface{}{
						"title":        "Advanced Go Techniques",
						"url":          "/blog/advanced-go",
						"author":       map[string]interface{}{"name": "Go Expert"},
						"published_at": mustParseTime("2024-01-10"),
						"excerpt":      "Deep dive into advanced Go programming patterns",
					},
				},
			},
			expectedContent: []string{
				"How to Use Go Templates - Test Blog",
				"By Tech Writer",
				"5 min read",
				"This is the full content of the blog post with <strong>HTML</strong>",
				"About Tech Writer",
				"A passionate technical writer",
				"Related Posts",
				"Advanced Go Techniques",
			},
		},
		{
			name:         "contact form rendering",
			templateName: "forms/contact.html",
			data: map[string]interface{}{
				"site_name": "Test Site",
				"navigation": []interface{}{
					map[string]interface{}{"url": "/", "title": "Home", "active": false},
					map[string]interface{}{"url": "/contact", "title": "Contact", "active": true},
				},
				"form_data": map[string]interface{}{
					"name":    "John",
					"email":   "john@example.com",
					"subject": "general",
				},
				"form_errors": map[string]interface{}{
					"message": []interface{}{"Message is required"},
				},
				"contact_info": map[string]interface{}{
					"email":   "contact@testsite.com",
					"phone":   "+1-555-0123",
					"address": "123 Main St\nAnytown, ST 12345",
				},
				"subject_options": []interface{}{
					map[string]interface{}{"value": "general", "label": "General Inquiry"},
					map[string]interface{}{"value": "support", "label": "Technical Support"},
					map[string]interface{}{"value": "business", "label": "Business Inquiry"},
					map[string]interface{}{"value": "other", "label": "Other"},
				},
			},
			expectedContent: []string{
				"Contact Us - Test Site",
				"Get in Touch",
				"Message is required",
				"Full Name",
				"john@example.com",
				"Contact Information",
				"contact@testsite.com",
				"+1-555-0123",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(test.templateName)
			if err != nil {
				t.Fatalf("Failed to load template %s: %v", test.templateName, err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expectedContent {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result[:500]+"...")
				}
			}

			// Verify HTML structure is present
			if !strings.Contains(result, "<!DOCTYPE html>") {
				t.Error("Expected HTML doctype in result")
			}
			if !strings.Contains(result, "</html>") {
				t.Error("Expected closing HTML tag in result")
			}
		})
	}
}

// Test E-commerce Template System
func TestEcommerceTemplateSystem(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "jinja2_ecommerce_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	templates := map[string]string{
		"shop/product_list.html": `{% extends "layouts/base.html" %}

{% block content %}
    <div class="product-catalog">
        <div class="filters">
            <h3>Filter Products</h3>
            {% for category in categories %}
                <label>
                    <input type="checkbox" name="category" value="{{ category.id }}">
                    {{ category.name }} ({{ category.count }})
                </label>
            {% endfor %}
        </div>
        
        <div class="products">
            <div class="products-header">
                <h1>Products</h1>
                <div class="sort-options">
                    <select name="sort">
                        <option value="name">Sort by Name</option>
                        <option value="price_low">Price: Low to High</option>
                        <option value="price_high">Price: High to Low</option>
                    </select>
                </div>
            </div>
            
            <div class="product-grid">
                {% for product in products %}
                    <div class="product-card">
                        <div class="product-image">
                            <img src="{{ product.image|default('/images/no-image.jpg') }}" 
                                 alt="{{ product.name|escape }}">
                            {% if product.discount %}
                                <div class="discount-badge">-{{ product.discount }}%</div>
                            {% endif %}
                        </div>
                        <div class="product-info">
                            <h3><a href="/products/{{ product.slug }}">{{ product.name }}</a></h3>
                            <p class="product-description">{{ product.short_description|truncate(100) }}</p>
                            <div class="product-price">
                                {% if product.sale_price %}
                                    <span class="original-price">${{ product.price }}</span>
                                    <span class="sale-price">${{ product.sale_price }}</span>
                                {% else %}
                                    <span class="price">${{ product.price }}</span>
                                {% endif %}
                            </div>
                            <div class="product-rating">
                                {% for i in range(5) %}
                                    <span class="star{% if i < product.rating %} filled{% endif %}">★</span>
                                {% endfor %}
                                <span class="rating-count">({{ product.review_count }})</span>
                            </div>
                            <button class="btn btn-primary add-to-cart" data-product-id="{{ product.id }}">
                                Add to Cart
                            </button>
                        </div>
                    </div>
                {% else %}
                    <div class="no-products">
                        <p>No products found matching your criteria.</p>
                    </div>
                {% endfor %}
            </div>
            
            {% if pagination %}
                <nav class="pagination">
                    {% if pagination.has_previous %}
                        <a href="?page={{ pagination.previous_page }}" class="page-link">Previous</a>
                    {% endif %}
                    
                    {% for page in pagination.pages %}
                        {% if page == pagination.current_page %}
                            <span class="page-link current">{{ page }}</span>
                        {% else %}
                            <a href="?page={{ page }}" class="page-link">{{ page }}</a>
                        {% endif %}
                    {% endfor %}
                    
                    {% if pagination.has_next %}
                        <a href="?page={{ pagination.next_page }}" class="page-link">Next</a>
                    {% endif %}
                </nav>
            {% endif %}
        </div>
    </div>
{% endblock %}`,

		"shop/product_detail.html": `{% extends "layouts/base.html" %}

{% block title %}{{ product.name }} - {{ super() }}{% endblock %}

{% block content %}
    <div class="product-detail">
        <nav class="breadcrumb">
            <a href="/">Home</a> &gt; 
            <a href="/shop">Shop</a> &gt; 
            {% if product.category %}
                <a href="/shop/category/{{ product.category.slug }}">{{ product.category.name }}</a> &gt; 
            {% endif %}
            <span>{{ product.name }}</span>
        </nav>
        
        <div class="product-main">
            <div class="product-images">
                <div class="main-image">
                    <img src="{{ product.main_image }}" alt="{{ product.name|escape }}" id="main-product-image">
                </div>
                {% if product.gallery %}
                    <div class="thumbnail-images">
                        {% for image in product.gallery %}
                            <img src="{{ image.thumbnail }}" alt="{{ product.name|escape }}" 
                                 class="thumbnail" data-full="{{ image.full }}">
                        {% endfor %}
                    </div>
                {% endif %}
            </div>
            
            <div class="product-details">
                <h1>{{ product.name }}</h1>
                <div class="product-meta">
                    <div class="product-rating">
                        {% for i in range(5) %}
                            <span class="star{% if i < product.rating %} filled{% endif %}">★</span>
                        {% endfor %}
                        <span class="rating-text">{{ product.rating }}/5 ({{ product.review_count }} reviews)</span>
                    </div>
                    <div class="product-sku">SKU: {{ product.sku }}</div>
                </div>
                
                <div class="product-price">
                    {% if product.sale_price %}
                        <span class="original-price">${{ product.price }}</span>
                        <span class="sale-price">${{ product.sale_price }}</span>
                        <span class="savings">You save ${{ (product.price - product.sale_price)|round(2) }}</span>
                    {% else %}
                        <span class="price">${{ product.price }}</span>
                    {% endif %}
                </div>
                
                <div class="product-description">
                    {{ product.description|safe }}
                </div>
                
                {% if product.variants %}
                    <div class="product-options">
                        {% for variant_type, options in product.variants.items() %}
                            <div class="option-group">
                                <label>{{ variant_type|title }}:</label>
                                <select name="{{ variant_type }}" class="variant-select">
                                    {% for option in options %}
                                        <option value="{{ option.value }}" 
                                                {% if option.stock == 0 %}disabled{% endif %}>
                                            {{ option.label }}
                                            {% if option.price_modifier %}
                                                (+${{ option.price_modifier }})
                                            {% endif %}
                                            {% if option.stock == 0 %}(Out of Stock){% endif %}
                                        </option>
                                    {% endfor %}
                                </select>
                            </div>
                        {% endfor %}
                    </div>
                {% endif %}
                
                <div class="product-quantity">
                    <label>Quantity:</label>
                    <input type="number" name="quantity" value="1" min="1" max="{{ product.max_quantity|default(10) }}">
                </div>
                
                <div class="product-actions">
                    {% if product.in_stock %}
                        <button class="btn btn-primary btn-large add-to-cart" data-product-id="{{ product.id }}">
                            Add to Cart
                        </button>
                        <button class="btn btn-secondary add-to-wishlist" data-product-id="{{ product.id }}">
                            Add to Wishlist
                        </button>
                    {% else %}
                        <button class="btn btn-disabled" disabled>Out of Stock</button>
                        <button class="btn btn-secondary notify-when-available" data-product-id="{{ product.id }}">
                            Notify When Available
                        </button>
                    {% endif %}
                </div>
                
                {% if product.features %}
                    <div class="product-features">
                        <h3>Features</h3>
                        <ul>
                            {% for feature in product.features %}
                                <li>{{ feature }}</li>
                            {% endfor %}
                        </ul>
                    </div>
                {% endif %}
            </div>
        </div>
        
        {% if related_products %}
            <section class="related-products">
                <h2>You Might Also Like</h2>
                <div class="product-grid">
                    {% for product in related_products %}
                        <div class="product-card">
                            <img src="{{ product.image }}" alt="{{ product.name|escape }}">
                            <h4><a href="/products/{{ product.slug }}">{{ product.name }}</a></h4>
                            <div class="price">${{ product.price }}</div>
                        </div>
                    {% endfor %}
                </div>
            </section>
        {% endif %}
    </div>
{% endblock %}`,

		"shop/cart.html": `{% extends "layouts/base.html" %}

{% block title %}Shopping Cart - {{ super() }}{% endblock %}

{% block content %}
    <div class="shopping-cart">
        <h1>Your Shopping Cart</h1>
        
        {% if cart.items %}
            <div class="cart-content">
                <div class="cart-items">
                    {% for item in cart.items %}
                        <div class="cart-item" data-item-id="{{ item.id }}">
                            <div class="item-image">
                                <img src="{{ item.product.image }}" alt="{{ item.product.name|escape }}">
                            </div>
                            <div class="item-details">
                                <h3><a href="/products/{{ item.product.slug }}">{{ item.product.name }}</a></h3>
                                {% if item.variant_info %}
                                    <div class="variant-info">
                                        {% for key, value in item.variant_info.items() %}
                                            <span>{{ key|title }}: {{ value }}</span>
                                        {% endfor %}
                                    </div>
                                {% endif %}
                                <div class="item-price">
                                    {% if item.sale_price %}
                                        <span class="original-price">${{ item.original_price }}</span>
                                        <span class="sale-price">${{ item.sale_price }}</span>
                                    {% else %}
                                        <span class="price">${{ item.price }}</span>
                                    {% endif %}
                                </div>
                            </div>
                            <div class="item-quantity">
                                <label>Quantity:</label>
                                <input type="number" name="quantity" value="{{ item.quantity }}" 
                                       min="1" max="{{ item.max_quantity|default(10) }}"
                                       class="quantity-input" data-item-id="{{ item.id }}">
                            </div>
                            <div class="item-total">
                                ${{ (item.price * item.quantity)|round(2) }}
                            </div>
                            <div class="item-actions">
                                <button class="btn btn-small remove-item" data-item-id="{{ item.id }}">Remove</button>
                            </div>
                        </div>
                    {% endfor %}
                </div>
                
                <div class="cart-summary">
                    <h2>Order Summary</h2>
                    <div class="summary-line">
                        <span>Subtotal ({{ cart.total_items }} items):</span>
                        <span>${{ cart.subtotal|round(2) }}</span>
                    </div>
                    {% if cart.discount %}
                        <div class="summary-line discount">
                            <span>Discount ({{ cart.discount.code }}):</span>
                            <span>-${{ cart.discount.amount|round(2) }}</span>
                        </div>
                    {% endif %}
                    <div class="summary-line">
                        <span>Shipping:</span>
                        <span>
                            {% if cart.free_shipping %}
                                Free
                            {% else %}
                                ${{ cart.shipping_cost|round(2) }}
                            {% endif %}
                        </span>
                    </div>
                    <div class="summary-line">
                        <span>Tax:</span>
                        <span>${{ cart.tax|round(2) }}</span>
                    </div>
                    <div class="summary-line total">
                        <span><strong>Total:</strong></span>
                        <span><strong>${{ cart.total|round(2) }}</strong></span>
                    </div>
                    
                    {% if not cart.discount %}
                        <div class="discount-code">
                            <input type="text" placeholder="Discount code" name="discount_code">
                            <button class="btn btn-secondary apply-discount">Apply</button>
                        </div>
                    {% endif %}
                    
                    <div class="checkout-actions">
                        <a href="/checkout" class="btn btn-primary btn-large">Proceed to Checkout</a>
                        <a href="/shop" class="btn btn-secondary">Continue Shopping</a>
                    </div>
                </div>
            </div>
        {% else %}
            <div class="empty-cart">
                <h2>Your cart is empty</h2>
                <p>Add some items to get started!</p>
                <a href="/shop" class="btn btn-primary">Start Shopping</a>
            </div>
        {% endif %}
    </div>
{% endblock %}`,

		"layouts/base.html": `<!DOCTYPE html>
<html>
<head><title>{% block title %}E-commerce Site{% endblock %}</title></head>
<body>
    <header>
        <nav>
            <a href="/">Home</a> | 
            <a href="/shop">Shop</a> | 
            <a href="/cart">Cart{% if cart.total_items %} ({{ cart.total_items }}){% endif %}</a>
        </nav>
    </header>
    <main>{% block content %}{% endblock %}</main>
    <footer>&copy; 2024 E-commerce Site</footer>
</body>
</html>`,
	}

	// Write templates to files
	for templatePath, content := range templates {
		fullPath := filepath.Join(tmpDir, templatePath)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write template %s: %v", templatePath, err)
		}
	}

	// Create environment
	directParser := loader.NewDirectTemplateParser()
	fsLoader := loader.NewFileSystemLoader([]string{tmpDir}, directParser)
	env := miya.NewEnvironment(miya.WithLoader(fsLoader), miya.WithAutoEscape(true), miya.WithTrimBlocks(true), miya.WithLstripBlocks(true))

	// Test e-commerce scenarios
	tests := []struct {
		name            string
		templateName    string
		data            map[string]interface{}
		expectedContent []string
	}{
		{
			name:         "product catalog with filters",
			templateName: "shop/product_list.html",
			data: map[string]interface{}{
				"categories": []interface{}{
					map[string]interface{}{"id": 1, "name": "Electronics", "count": 25},
					map[string]interface{}{"id": 2, "name": "Clothing", "count": 40},
				},
				"products": []interface{}{
					map[string]interface{}{
						"id":                1,
						"name":              "Wireless Headphones",
						"slug":              "wireless-headphones",
						"short_description": "High-quality wireless headphones with noise cancellation",
						"price":             99.99,
						"sale_price":        79.99,
						"discount":          20,
						"rating":            4.5,
						"review_count":      128,
						"image":             "/images/headphones.jpg",
					},
				},
				"pagination": map[string]interface{}{
					"current_page": 1,
					"has_next":     true,
					"next_page":    2,
					"has_previous": false,
					"pages":        []interface{}{1, 2, 3},
				},
			},
			expectedContent: []string{
				"Filter Products",
				"Electronics (25)",
				"Clothing (40)",
				"Wireless Headphones",
				"High-quality wireless headphones",
				"$99.99",
				"$79.99",
				"-20%",
				"★",
				"(128)",
				"Add to Cart",
				"Next",
			},
		},
		{
			name:         "detailed product page",
			templateName: "shop/product_detail.html",
			data: map[string]interface{}{
				"product": map[string]interface{}{
					"id":           1,
					"name":         "Premium Laptop",
					"slug":         "premium-laptop",
					"sku":          "LAP-001",
					"description":  "<p>High-performance laptop perfect for professionals.</p>",
					"price":        1499.99,
					"sale_price":   1299.99,
					"rating":       4.8,
					"review_count": 89,
					"main_image":   "/images/laptop-main.jpg",
					"in_stock":     true,
					"max_quantity": 5,
					"category": map[string]interface{}{
						"name": "Computers",
						"slug": "computers",
					},
					"variants": map[string]interface{}{
						"color": []interface{}{
							map[string]interface{}{"value": "black", "label": "Black", "stock": 10},
							map[string]interface{}{"value": "silver", "label": "Silver", "stock": 5},
						},
						"storage": []interface{}{
							map[string]interface{}{"value": "256gb", "label": "256GB", "price_modifier": 0, "stock": 8},
							map[string]interface{}{"value": "512gb", "label": "512GB", "price_modifier": 200, "stock": 3},
						},
					},
					"features": []interface{}{
						"Intel Core i7 processor",
						"16GB RAM",
						"Dedicated graphics card",
						"All-day battery life",
					},
					"gallery": []interface{}{
						map[string]interface{}{"thumbnail": "/images/laptop-thumb1.jpg", "full": "/images/laptop-1.jpg"},
						map[string]interface{}{"thumbnail": "/images/laptop-thumb2.jpg", "full": "/images/laptop-2.jpg"},
					},
				},
				"related_products": []interface{}{
					map[string]interface{}{
						"name":  "Laptop Bag",
						"slug":  "laptop-bag",
						"price": 49.99,
						"image": "/images/laptop-bag.jpg",
					},
				},
			},
			expectedContent: []string{
				"Premium Laptop",
				"SKU: LAP-001",
				"$1,499.99",
				"$1,299.99",
				"You save $200.00",
				"4.8/5 (89 reviews)",
				"High-performance laptop",
				"Color:",
				"Black",
				"Silver",
				"Storage:",
				"256GB",
				"512GB (+$200)",
				"Intel Core i7 processor",
				"16GB RAM",
				"Add to Cart",
				"Add to Wishlist",
				"You Might Also Like",
				"Laptop Bag",
			},
		},
		{
			name:         "shopping cart with items",
			templateName: "shop/cart.html",
			data: map[string]interface{}{
				"cart": map[string]interface{}{
					"total_items":   3,
					"subtotal":      299.97,
					"tax":           24.00,
					"shipping_cost": 9.99,
					"total":         333.96,
					"free_shipping": false,
					"items": []interface{}{
						map[string]interface{}{
							"id":       1,
							"quantity": 2,
							"price":    99.99,
							"product": map[string]interface{}{
								"name":  "Wireless Mouse",
								"slug":  "wireless-mouse",
								"image": "/images/mouse.jpg",
							},
							"variant_info": map[string]interface{}{
								"color": "Black",
							},
						},
						map[string]interface{}{
							"id":       2,
							"quantity": 1,
							"price":    99.99,
							"product": map[string]interface{}{
								"name":  "Keyboard",
								"slug":  "keyboard",
								"image": "/images/keyboard.jpg",
							},
						},
					},
				},
			},
			expectedContent: []string{
				"Your Shopping Cart",
				"Wireless Mouse",
				"Keyboard",
				"Color: Black",
				"Subtotal (3 items):",
				"$299.97",
				"Shipping:",
				"$9.99",
				"Tax:",
				"$24.00",
				"Total:",
				"$333.96",
				"Proceed to Checkout",
				"Continue Shopping",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(test.templateName)
			if err != nil {
				t.Fatalf("Failed to load template %s: %v", test.templateName, err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expectedContent {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result[:500]+"...")
				}
			}
		})
	}
}

// Test Email Template System
func TestEmailTemplateSystem(t *testing.T) {
	directParser := loader.NewDirectTemplateParser()
	stringLoader := loader.NewStringLoader(directParser)
	env := miya.NewEnvironment(miya.WithLoader(stringLoader), miya.WithAutoEscape(true))

	// Email templates
	templates := map[string]string{
		"emails/base.html": `<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{% block title %}{{ site_name }}{% endblock %}</title>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .header { background: #f8f9fa; padding: 20px; text-align: center; }
        .content { padding: 20px; }
        .footer { background: #f8f9fa; padding: 15px; text-align: center; font-size: 12px; }
        .btn { display: inline-block; padding: 10px 20px; background: #007bff; color: white; text-decoration: none; border-radius: 4px; }
    </style>
</head>
<body>
    <div class="email-wrapper">
        <div class="header">
            <h1>{{ site_name|default('Our Website') }}</h1>
        </div>
        <div class="content">
            {% block content %}{% endblock %}
        </div>
        <div class="footer">
            <p>&copy; {{ current_year }} {{ site_name|default('Our Website') }}. All rights reserved.</p>
            <p>
                <a href="{{ unsubscribe_url }}">Unsubscribe</a> | 
                <a href="{{ contact_url }}">Contact Us</a>
            </p>
        </div>
    </div>
</body>
</html>`,

		"emails/welcome.html": `{% extends "emails/base.html" %}

{% block title %}Welcome to {{ site_name }}!{% endblock %}

{% block content %}
    <h2>Welcome, {{ user.name }}!</h2>
    
    <p>Thank you for joining {{ site_name }}. We're excited to have you as part of our community!</p>
    
    <p>Here's what you can do next:</p>
    <ul>
        <li><a href="{{ profile_url }}">Complete your profile</a></li>
        <li><a href="{{ browse_url }}">Browse our features</a></li>
        <li><a href="{{ support_url }}">Get help from our support team</a></li>
    </ul>
    
    <p>
        <a href="{{ get_started_url }}" class="btn">Get Started</a>
    </p>
    
    <p>If you have any questions, feel free to reply to this email or contact our support team.</p>
    
    <p>Welcome aboard!</p>
    <p>The {{ site_name }} Team</p>
{% endblock %}`,

		"emails/order_confirmation.html": `{% extends "emails/base.html" %}

{% block title %}Order Confirmation #{{ order.number }}{% endblock %}

{% block content %}
    <h2>Thank you for your order!</h2>
    
    <p>Hi {{ order.customer.name }},</p>
    
    <p>We've received your order and it's being processed. Here are the details:</p>
    
    <div class="order-details">
        <h3>Order #{{ order.number }}</h3>
        <p><strong>Order Date:</strong> {{ order.created_at|date('%B %d, %Y') }}</p>
        <p><strong>Estimated Delivery:</strong> {{ order.estimated_delivery|date('%B %d, %Y') }}</p>
        
        <h4>Items Ordered:</h4>
        <table style="width: 100%; border-collapse: collapse;">
            <thead>
                <tr style="background: #f8f9fa;">
                    <th style="padding: 10px; text-align: left; border: 1px solid #ddd;">Item</th>
                    <th style="padding: 10px; text-align: center; border: 1px solid #ddd;">Quantity</th>
                    <th style="padding: 10px; text-align: right; border: 1px solid #ddd;">Price</th>
                </tr>
            </thead>
            <tbody>
                {% for item in order.items %}
                    <tr>
                        <td style="padding: 10px; border: 1px solid #ddd;">
                            {{ item.product.name }}
                            {% if item.variant_info %}
                                <br><small>{{ item.variant_info|join(', ') }}</small>
                            {% endif %}
                        </td>
                        <td style="padding: 10px; text-align: center; border: 1px solid #ddd;">{{ item.quantity }}</td>
                        <td style="padding: 10px; text-align: right; border: 1px solid #ddd;">${{ (item.price * item.quantity)|round(2) }}</td>
                    </tr>
                {% endfor %}
            </tbody>
            <tfoot>
                <tr>
                    <td colspan="2" style="padding: 10px; text-align: right; border: 1px solid #ddd;"><strong>Subtotal:</strong></td>
                    <td style="padding: 10px; text-align: right; border: 1px solid #ddd;"><strong>${{ order.subtotal|round(2) }}</strong></td>
                </tr>
                <tr>
                    <td colspan="2" style="padding: 10px; text-align: right; border: 1px solid #ddd;">Shipping:</td>
                    <td style="padding: 10px; text-align: right; border: 1px solid #ddd;">${{ order.shipping|round(2) }}</td>
                </tr>
                <tr>
                    <td colspan="2" style="padding: 10px; text-align: right; border: 1px solid #ddd;">Tax:</td>
                    <td style="padding: 10px; text-align: right; border: 1px solid #ddd;">${{ order.tax|round(2) }}</td>
                </tr>
                <tr style="background: #f8f9fa;">
                    <td colspan="2" style="padding: 10px; text-align: right; border: 1px solid #ddd;"><strong>Total:</strong></td>
                    <td style="padding: 10px; text-align: right; border: 1px solid #ddd;"><strong>${{ order.total|round(2) }}</strong></td>
                </tr>
            </tfoot>
        </table>
    </div>
    
    <h4>Shipping Address:</h4>
    <p>
        {{ order.shipping_address.name }}<br>
        {{ order.shipping_address.address_line_1 }}<br>
        {% if order.shipping_address.address_line_2 %}
            {{ order.shipping_address.address_line_2 }}<br>
        {% endif %}
        {{ order.shipping_address.city }}, {{ order.shipping_address.state }} {{ order.shipping_address.postal_code }}<br>
        {{ order.shipping_address.country }}
    </p>
    
    <p>
        <a href="{{ track_order_url }}" class="btn">Track Your Order</a>
    </p>
    
    <p>Questions about your order? <a href="{{ support_url }}">Contact our support team</a>.</p>
    
    <p>Thank you for your business!</p>
{% endblock %}`,
	}

	// Add all templates
	for name, content := range templates {
		stringLoader.AddTemplate(name, content)
	}

	tests := []struct {
		name            string
		templateName    string
		data            map[string]interface{}
		expectedContent []string
	}{
		{
			name:         "welcome email",
			templateName: "emails/welcome.html",
			data: map[string]interface{}{
				"site_name":       "TechCorp",
				"current_year":    2024,
				"user":            map[string]interface{}{"name": "Alice Johnson"},
				"profile_url":     "https://techcorp.com/profile",
				"browse_url":      "https://techcorp.com/browse",
				"support_url":     "https://techcorp.com/support",
				"get_started_url": "https://techcorp.com/get-started",
				"unsubscribe_url": "https://techcorp.com/unsubscribe",
				"contact_url":     "https://techcorp.com/contact",
			},
			expectedContent: []string{
				"Welcome to TechCorp!",
				"Welcome, Alice Johnson!",
				"Thank you for joining TechCorp",
				"Complete your profile",
				"Browse our features",
				"Get help from our support team",
				"Get Started",
				"The TechCorp Team",
				"&copy; 2024 TechCorp",
				"Unsubscribe",
			},
		},
		{
			name:         "order confirmation email",
			templateName: "emails/order_confirmation.html",
			data: map[string]interface{}{
				"site_name":    "ShopCorp",
				"current_year": 2024,
				"order": map[string]interface{}{
					"number":             "ORD-12345",
					"created_at":         mustParseTime("2024-01-15"),
					"estimated_delivery": mustParseTime("2024-01-20"),
					"subtotal":           199.98,
					"shipping":           9.99,
					"tax":                16.00,
					"total":              225.97,
					"customer":           map[string]interface{}{"name": "Bob Smith"},
					"items": []interface{}{
						map[string]interface{}{
							"product":      map[string]interface{}{"name": "Wireless Headphones"},
							"quantity":     1,
							"price":        99.99,
							"variant_info": []interface{}{"Black", "Over-ear"},
						},
						map[string]interface{}{
							"product":  map[string]interface{}{"name": "Phone Case"},
							"quantity": 1,
							"price":    99.99,
						},
					},
					"shipping_address": map[string]interface{}{
						"name":           "Bob Smith",
						"address_line_1": "123 Main Street",
						"address_line_2": "Apt 4B",
						"city":           "Anytown",
						"state":          "ST",
						"postal_code":    "12345",
						"country":        "USA",
					},
				},
				"track_order_url": "https://shopcorp.com/track/ORD-12345",
				"support_url":     "https://shopcorp.com/support",
				"unsubscribe_url": "https://shopcorp.com/unsubscribe",
				"contact_url":     "https://shopcorp.com/contact",
			},
			expectedContent: []string{
				"Order Confirmation #ORD-12345",
				"Thank you for your order!",
				"Hi Bob Smith,",
				"Order #ORD-12345",
				"Wireless Headphones",
				"Black, Over-ear",
				"Phone Case",
				"$199.98",
				"$9.99",
				"$16.00",
				"$225.97",
				"123 Main Street",
				"Apt 4B",
				"Anytown, ST 12345",
				"Track Your Order",
				"Contact our support team",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tmpl, err := env.GetTemplate(test.templateName)
			if err != nil {
				t.Fatalf("Failed to load template %s: %v", test.templateName, err)
			}

			result, err := tmpl.Render(miya.NewContextFrom(test.data))
			if err != nil {
				t.Fatalf("Failed to render template: %v", err)
			}

			for _, expected := range test.expectedContent {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result[:500]+"...")
				}
			}

			// Verify email structure
			if !strings.Contains(result, "<!DOCTYPE html>") {
				t.Error("Expected HTML doctype in email result")
			}
			if !strings.Contains(result, "font-family: Arial") {
				t.Error("Expected CSS styles in email result")
			}
		})
	}
}
