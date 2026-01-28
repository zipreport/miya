#!/usr/bin/env python3
"""
Performance comparison between Miya (Go) and Python Jinja2
"""

from jinja2 import Environment, Template
import time
import statistics

def benchmark_simple_template():
    """Benchmark simple variable interpolation"""
    template_str = "Hello {{ name }}!"
    env = Environment()

    # Warmup
    tmpl = env.from_string(template_str)
    for _ in range(100):
        tmpl.render(name="World")

    # Benchmark template compilation (cold start)
    compile_times = []
    for _ in range(100):
        start = time.perf_counter()
        env.from_string(template_str)
        compile_times.append((time.perf_counter() - start) * 1_000_000)  # Convert to microseconds

    # Benchmark cached rendering
    tmpl = env.from_string(template_str)
    render_times = []
    for _ in range(10000):
        start = time.perf_counter()
        tmpl.render(name="World")
        render_times.append((time.perf_counter() - start) * 1_000_000)  # Convert to microseconds

    return {
        'compile_avg': statistics.mean(compile_times),
        'compile_median': statistics.median(compile_times),
        'render_avg': statistics.mean(render_times),
        'render_median': statistics.median(render_times),
    }

def benchmark_loop_template():
    """Benchmark template with for loop"""
    template_str = "{% for item in items %}{{ item }} {% endfor %}"
    env = Environment()

    # Warmup
    tmpl = env.from_string(template_str)
    items = ["apple", "banana", "cherry", "date", "elderberry"]
    for _ in range(100):
        tmpl.render(items=items)

    # Benchmark rendering
    tmpl = env.from_string(template_str)
    render_times = []
    for _ in range(10000):
        start = time.perf_counter()
        tmpl.render(items=items)
        render_times.append((time.perf_counter() - start) * 1_000_000)  # Convert to microseconds

    return {
        'render_avg': statistics.mean(render_times),
        'render_median': statistics.median(render_times),
    }

def benchmark_complex_template():
    """Benchmark complex template with nested loops and filters"""
    template_str = """
{% for user in users %}
  <div class="user">
    <h2>{{ user.name|upper }}</h2>
    <p>Email: {{ user.email }}</p>
    <p>Age: {{ user.age }}</p>
    {% if user.active %}Active{% else %}Inactive{% endif %}
  </div>
{% endfor %}
"""
    env = Environment()

    # Warmup
    tmpl = env.from_string(template_str)
    users = [
        {"name": "Alice", "email": "alice@example.com", "age": 30, "active": True},
        {"name": "Bob", "email": "bob@example.com", "age": 25, "active": False},
        {"name": "Charlie", "email": "charlie@example.com", "age": 35, "active": True},
    ]
    for _ in range(100):
        tmpl.render(users=users)

    # Benchmark rendering
    tmpl = env.from_string(template_str)
    render_times = []
    for _ in range(10000):
        start = time.perf_counter()
        tmpl.render(users=users)
        render_times.append((time.perf_counter() - start) * 1_000_000)  # Convert to microseconds

    return {
        'render_avg': statistics.mean(render_times),
        'render_median': statistics.median(render_times),
    }

if __name__ == "__main__":
    print("=" * 60)
    print("Python Jinja2 Performance Benchmark")
    print("=" * 60)
    print()

    print("1. Simple Template (Hello {{ name }}!)")
    print("-" * 60)
    simple_results = benchmark_simple_template()
    print(f"   Compilation (cold start):")
    print(f"     Average:  {simple_results['compile_avg']:,.2f} μs")
    print(f"     Median:   {simple_results['compile_median']:,.2f} μs")
    print(f"   Rendering (cached template):")
    print(f"     Average:  {simple_results['render_avg']:,.2f} μs")
    print(f"     Median:   {simple_results['render_median']:,.2f} μs")
    print()

    print("2. Loop Template ({% for item in items %})")
    print("-" * 60)
    loop_results = benchmark_loop_template()
    print(f"   Rendering (5 items):")
    print(f"     Average:  {loop_results['render_avg']:,.2f} μs")
    print(f"     Median:   {loop_results['render_median']:,.2f} μs")
    print()

    print("3. Complex Template (nested loops, filters, conditionals)")
    print("-" * 60)
    complex_results = benchmark_complex_template()
    print(f"   Rendering (3 users):")
    print(f"     Average:  {complex_results['render_avg']:,.2f} μs")
    print(f"     Median:   {complex_results['render_median']:,.2f} μs")
    print()

    print("=" * 60)
    print("Summary")
    print("=" * 60)
    print(f"Simple template rendering:  {simple_results['render_avg']:,.2f} μs")
    print(f"Loop template rendering:    {loop_results['render_avg']:,.2f} μs")
    print(f"Complex template rendering: {complex_results['render_avg']:,.2f} μs")
