# List/Dict Comprehensions - Practical Examples

This document provides practical, ready-to-use examples of list and dictionary comprehensions in various real-world scenarios.

## E-commerce Examples

### Product Filtering and Display

```jinja2
<!-- Filter and display available products -->
<div class="product-grid">
  {% for product in products %}
    {% if product.in_stock %}
      <div class="product-card">
        <h3>{{ product.name }}</h3>
        <p class="price">${{ product.price }}</p>
        
        <!-- Available sizes using list comprehension -->
        <div class="sizes">
          {% for size in [variant.size for variant in product.variants if variant.stock > 0] %}
            <span class="size">{{ size }}</span>
          {% endfor %}
        </div>
        
        <!-- Color options using dict comprehension -->
        {% set color_stock = {variant.color: variant.stock for variant in product.variants if variant.stock > 0} %}
        <div class="colors">
          {% for color, stock in color_stock.items() %}
            <span class="color" data-stock="{{ stock }}">{{ color }}</span>
          {% endfor %}
        </div>
      </div>
    {% endif %}
  {% endfor %}
</div>
```

### Shopping Cart Summary

```jinja2
<!-- Cart totals using comprehensions -->
{% set cart_items = [{"name": item.product.name, "price": item.price, "qty": item.quantity, "total": item.price * item.quantity} for item in cart.items] %}
{% set subtotal = [item.total for item in cart_items]|sum %}
{% set tax = subtotal * 0.08 %}

<div class="cart-summary">
  <h3>Order Summary</h3>
  
  {% for item in cart_items %}
    <div class="cart-item">
      <span>{{ item.name }} ({{ item.qty }}x)</span>
      <span>${{ item.total }}</span>
    </div>
  {% endfor %}
  
  <div class="totals">
    <div>Subtotal: ${{ subtotal }}</div>
    <div>Tax: ${{ tax|round(2) }}</div>
    <div class="total">Total: ${{ (subtotal + tax)|round(2) }}</div>
  </div>
</div>
```

## Data Dashboard Examples

### Analytics Dashboard

```jinja2
<!-- User engagement metrics -->
{% set active_users = [user for user in users if user.last_login > thirty_days_ago] %}
{% set user_stats = {
  "total": users|length,
  "active": active_users|length,
  "inactive": users|length - active_users|length,
  "premium": [user for user in active_users if user.subscription_type == "premium"]|length
} %}

<div class="dashboard">
  <div class="metrics-grid">
    {% for metric, value in user_stats.items() %}
      <div class="metric-card">
        <h3>{{ metric|title }}</h3>
        <div class="value">{{ value }}</div>
      </div>
    {% endfor %}
  </div>
  
  <!-- Recent activity using comprehensions -->
  {% set recent_actions = [{"user": action.user.name, "action": action.type, "time": action.created_at} for action in actions if action.created_at > last_24_hours][:10] %}
  
  <div class="recent-activity">
    <h3>Recent Activity</h3>
    {% for action in recent_actions %}
      <div class="activity-item">
        <strong>{{ action.user }}</strong> {{ action.action }} 
        <span class="time">{{ action.time|timeago }}</span>
      </div>
    {% endfor %}
  </div>
</div>
```

### Sales Report

```jinja2
<!-- Monthly sales breakdown -->
{% set sales_by_month = {} %}
{% for sale in sales %}
  {% set month = sale.date.strftime("%Y-%m") %}
  {% if month not in sales_by_month %}
    {% set _ = sales_by_month.update({month: []}) %}
  {% endif %}
  {% set _ = sales_by_month[month].append(sale.amount) %}
{% endfor %}

<!-- Convert to monthly totals using comprehensions -->
{% set monthly_totals = {month: amounts|sum for month, amounts in sales_by_month.items()} %}
{% set best_month = monthly_totals.items()|max(attribute=1) %}

<div class="sales-report">
  <h2>Sales Performance</h2>
  
  <div class="highlight">
    <strong>Best Month:</strong> {{ best_month[0] }} ({{ best_month[1] }})
  </div>
  
  <div class="monthly-breakdown">
    {% for month, total in monthly_totals.items()|sort %}
      <div class="month-bar">
        <span class="month">{{ month }}</span>
        <div class="bar" style="width: {{ (total / monthly_totals.values()|max * 100)|round }}%"></div>
        <span class="amount">${{ total }}</span>
      </div>
    {% endfor %}
  </div>
</div>
```

## Content Management Examples

### Blog Post Listing

```jinja2
<!-- Blog posts with category filtering -->
{% set published_posts = [post for post in posts if post.status == "published" and post.published_date <= now] %}
{% set posts_by_category = {} %}

{% for post in published_posts %}
  {% for category in post.categories %}
    {% if category.name not in posts_by_category %}
      {% set _ = posts_by_category.update({category.name: []}) %}
    {% endif %}
    {% set _ = posts_by_category[category.name].append(post) %}
  {% endfor %}
{% endfor %}

<!-- Category navigation -->
<nav class="categories">
  {% for category, category_posts in posts_by_category.items() %}
    <a href="#{{ category|slugify }}" class="category-link">
      {{ category }} ({{ category_posts|length }})
    </a>
  {% endfor %}
</nav>

<!-- Posts by category -->
{% for category, category_posts in posts_by_category.items() %}
  <section id="{{ category|slugify }}" class="category-section">
    <h2>{{ category }}</h2>
    
    <div class="posts-grid">
      {% for post in category_posts[:6] %}  <!-- Show max 6 posts per category -->
        <article class="post-card">
          <h3><a href="{{ post.url }}">{{ post.title }}</a></h3>
          <p>{{ post.excerpt }}</p>
          <div class="meta">
            <span>{{ post.author.name }}</span>
            <time>{{ post.published_date|dateformat }}</time>
            <!-- Related tags using comprehension -->
            <div class="tags">
              {% for tag in [t.name for t in post.tags if t.featured] %}
                <span class="tag">{{ tag }}</span>
              {% endfor %}
            </div>
          </div>
        </article>
      {% endfor %}
    </div>
  </section>
{% endfor %}
```

### Tag Cloud with Weights

```jinja2
<!-- Dynamic tag cloud -->
{% set all_tags = [] %}
{% for post in posts %}
  {% set _ = all_tags.extend([tag.name for tag in post.tags]) %}
{% endfor %}

<!-- Count tag occurrences -->
{% set tag_counts = {} %}
{% for tag in all_tags %}
  {% set _ = tag_counts.update({tag: tag_counts.get(tag, 0) + 1}) %}
{% endfor %}

<!-- Filter and weight tags -->
{% set popular_tags = {tag: count for tag, count in tag_counts.items() if count >= 3} %}
{% set max_count = popular_tags.values()|max %}

<div class="tag-cloud">
  {% for tag, count in popular_tags.items()|sort(attribute=1, reverse=true) %}
    {% set weight = (count / max_count * 4 + 1)|round %}  <!-- 1-5 scale -->
    <a href="/tags/{{ tag|slugify }}" 
       class="tag weight-{{ weight }}" 
       title="{{ count }} posts">
      {{ tag }}
    </a>
  {% endfor %}
</div>

<style>
.tag-cloud .weight-1 { font-size: 0.8em; opacity: 0.6; }
.tag-cloud .weight-2 { font-size: 0.9em; opacity: 0.7; }
.tag-cloud .weight-3 { font-size: 1.0em; opacity: 0.8; }
.tag-cloud .weight-4 { font-size: 1.2em; opacity: 0.9; }
.tag-cloud .weight-5 { font-size: 1.4em; opacity: 1.0; font-weight: bold; }
</style>
```

## Form Processing Examples

### Dynamic Form Generation

```jinja2
<!-- Generate form fields with validation -->
{% set required_fields = [field.name for field in form_schema.fields if field.required] %}
{% set field_types = {field.name: field.type for field in form_schema.fields} %}
{% set field_options = {field.name: field.options for field in form_schema.fields if field.options} %}

<form id="dynamic-form" method="post">
  {% for field in form_schema.fields %}
    <div class="form-group {% if field.name in required_fields %}required{% endif %}">
      <label for="{{ field.name }}">
        {{ field.label }}
        {% if field.name in required_fields %}<span class="required">*</span>{% endif %}
      </label>
      
      {% if field_types[field.name] == "select" %}
        <select name="{{ field.name }}" id="{{ field.name }}">
          {% if field.name not in required_fields %}
            <option value="">-- Select {{ field.label }} --</option>
          {% endif %}
          {% for option in field_options.get(field.name, []) %}
            <option value="{{ option.value }}">{{ option.label }}</option>
          {% endfor %}
        </select>
        
      {% elif field_types[field.name] == "checkbox_group" %}
        {% for option in field_options.get(field.name, []) %}
          <div class="checkbox-item">
            <input type="checkbox" 
                   name="{{ field.name }}" 
                   value="{{ option.value }}" 
                   id="{{ field.name }}_{{ option.value }}">
            <label for="{{ field.name }}_{{ option.value }}">{{ option.label }}</label>
          </div>
        {% endfor %}
        
      {% else %}
        <input type="{{ field_types[field.name] }}" 
               name="{{ field.name }}" 
               id="{{ field.name }}"
               {% if field.name in required_fields %}required{% endif %}
               {% if field.placeholder %}placeholder="{{ field.placeholder }}"{% endif %}>
      {% endif %}
      
      {% if field.help_text %}
        <div class="help-text">{{ field.help_text }}</div>
      {% endif %}
    </div>
  {% endfor %}
  
  <button type="submit">Submit</button>
</form>
```

### Form Validation Display

```jinja2
<!-- Display form errors with styling -->
{% set field_errors = {error.field: error.messages for error in form.errors} %}
{% set global_errors = [error.message for error in form.errors if error.field == "__all__"] %}

{% if global_errors %}
  <div class="alert alert-error">
    {% for error in global_errors %}
      <p>{{ error }}</p>
    {% endfor %}
  </div>
{% endif %}

<form>
  {% for field in form.fields %}
    {% set has_error = field.name in field_errors %}
    
    <div class="form-group{% if has_error %} has-error{% endif %}">
      <label for="{{ field.name }}">{{ field.label }}</label>
      
      {{ field.render() }}
      
      {% if has_error %}
        <div class="error-messages">
          {% for error in field_errors[field.name] %}
            <span class="error-message">{{ error }}</span>
          {% endfor %}
        </div>
      {% endif %}
    </div>
  {% endfor %}
</form>
```

## API Response Processing

### RESTful Data Display

```jinja2
<!-- Process API response with nested data -->
{% set user_permissions = {user.id: [perm.name for perm in user.permissions] for user in api_response.users} %}
{% set active_users = [user for user in api_response.users if user.status == "active"] %}
{% set users_by_role = {} %}

{% for user in active_users %}
  {% if user.role not in users_by_role %}
    {% set _ = users_by_role.update({user.role: []}) %}
  {% endif %}
  {% set _ = users_by_role[user.role].append(user) %}
{% endfor %}

<div class="user-management">
  <div class="stats">
    <div class="stat">
      <span class="label">Total Users</span>
      <span class="value">{{ api_response.users|length }}</span>
    </div>
    <div class="stat">
      <span class="label">Active Users</span>
      <span class="value">{{ active_users|length }}</span>
    </div>
    <div class="stat">
      <span class="label">Roles</span>
      <span class="value">{{ users_by_role.keys()|length }}</span>
    </div>
  </div>
  
  {% for role, users in users_by_role.items() %}
    <div class="role-section">
      <h3>{{ role|title }} ({{ users|length }})</h3>
      
      <div class="users-list">
        {% for user in users %}
          <div class="user-card">
            <div class="user-info">
              <strong>{{ user.name }}</strong>
              <span class="email">{{ user.email }}</span>
            </div>
            
            <div class="permissions">
              {% for permission in user_permissions[user.id][:3] %}  <!-- Show first 3 permissions -->
                <span class="permission-tag">{{ permission }}</span>
              {% endfor %}
              {% if user_permissions[user.id]|length > 3 %}
                <span class="more">+{{ user_permissions[user.id]|length - 3 }} more</span>
              {% endif %}
            </div>
          </div>
        {% endfor %}
      </div>
    </div>
  {% endfor %}
</div>
```

## Advanced Use Cases

### Matrix Operations

```jinja2
<!-- Create a multiplication table -->
{% set size = 10 %}
{% set multiplication_table = [[x * y for x in range(1, size + 1)] for y in range(1, size + 1)] %}

<table class="multiplication-table">
  <thead>
    <tr>
      <th></th>
      {% for x in range(1, size + 1) %}
        <th>{{ x }}</th>
      {% endfor %}
    </tr>
  </thead>
  <tbody>
    {% for y in range(1, size + 1) %}
      <tr>
        <th>{{ y }}</th>
        {% for result in multiplication_table[y - 1] %}
          <td>{{ result }}</td>
        {% endfor %}
      </tr>
    {% endfor %}
  </tbody>
</table>
```

### Data Transformation Pipeline

```jinja2
<!-- Multi-step data processing -->
{% set raw_data = api_response.transactions %}

<!-- Step 1: Filter valid transactions -->
{% set valid_transactions = [t for t in raw_data if t.status == "completed" and t.amount > 0] %}

<!-- Step 2: Group by date -->
{% set transactions_by_date = {} %}
{% for transaction in valid_transactions %}
  {% set date_key = transaction.date.strftime("%Y-%m-%d") %}
  {% if date_key not in transactions_by_date %}
    {% set _ = transactions_by_date.update({date_key: []}) %}
  {% endif %}
  {% set _ = transactions_by_date[date_key].append(transaction) %}
{% endfor %}

<!-- Step 3: Calculate daily summaries -->
{% set daily_summaries = {
  date: {
    "total": [t.amount for t in transactions]|sum,
    "count": transactions|length,
    "average": ([t.amount for t in transactions]|sum / transactions|length)|round(2),
    "categories": {cat: [t.amount for t in transactions if t.category == cat]|sum for cat in [t.category for t in transactions]|unique}
  }
  for date, transactions in transactions_by_date.items()
} %}

<!-- Step 4: Display processed data -->
<div class="transaction-summary">
  {% for date, summary in daily_summaries.items()|sort(reverse=true) %}
    <div class="daily-summary">
      <h3>{{ date|dateformat }}</h3>
      
      <div class="summary-stats">
        <div class="stat">
          <span class="label">Total</span>
          <span class="value">${{ summary.total }}</span>
        </div>
        <div class="stat">
          <span class="label">Transactions</span>
          <span class="value">{{ summary.count }}</span>
        </div>
        <div class="stat">
          <span class="label">Average</span>
          <span class="value">${{ summary.average }}</span>
        </div>
      </div>
      
      <div class="categories">
        {% for category, amount in summary.categories.items()|sort(attribute=1, reverse=true) %}
          <div class="category-item">
            <span class="category">{{ category|title }}</span>
            <span class="amount">${{ amount }}</span>
            <div class="bar" style="width: {{ (amount / summary.total * 100)|round }}%"></div>
          </div>
        {% endfor %}
      </div>
    </div>
  {% endfor %}
</div>
```

---

These examples demonstrate the power and flexibility of list and dictionary comprehensions in real-world template scenarios. They can significantly reduce template complexity while maintaining readability and performance.