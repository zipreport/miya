# Changelog

## [v0.1.1]

### Fixed

- `tojson` filter now returns `runtime.SafeValue` instead of a plain string, preventing HTML auto-escaping of JSON output. This matches Jinja2's `tojson` behavior which returns `Markup` (a safe string). Previously, when auto-escaping was enabled, `{{ data | tojson }}` inside `<script>` blocks would produce HTML-escaped entities (e.g., `"` became `&#34;`), resulting in invalid JavaScript since browsers do not decode HTML entities in raw text elements like `<script>`.
