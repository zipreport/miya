package loader

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/zipreport/miya/parser"
)

// Pre-compiled regex patterns for dependency extraction (performance optimization)
var (
	reExtendsPattern = regexp.MustCompile(`{%\s*extends\s+['"]([^'"]+)['"]`)
	reIncludePattern = regexp.MustCompile(`{%\s*include\s+['"]([^'"]+)['"]`)
	reImportPattern  = regexp.MustCompile(`{%\s*import\s+['"]([^'"]+)['"]`)
	reFromPattern    = regexp.MustCompile(`{%\s*from\s+['"]([^'"]+)['"]`)
)

// TemplateSource represents a template source with metadata
type TemplateSource struct {
	Name     string
	Content  string
	ModTime  time.Time
	Checksum string
}

// Base Loader interface (keeping compatibility with existing code)
type Loader interface {
	GetSource(name string) (string, error)
	IsCached(name string) bool
	ListTemplates() ([]string, error)
}

// Enhanced Loader interface for advanced template loading
type AdvancedLoader interface {
	Loader
	LoadTemplate(name string) (*parser.TemplateNode, error)
	ResolveTemplateName(name string) string
	GetSourceWithMetadata(name string) (*TemplateSource, error)
}

// DiscoveryLoader interface for advanced template discovery
type DiscoveryLoader interface {
	AdvancedLoader
	SearchTemplates(pattern string) ([]string, error)
	GetTemplatesByExtension(ext string) ([]string, error)
	GetTemplatesInDirectory(dir string) ([]string, error)
	GetTemplateInfo(name string) (*TemplateInfo, error)
}

// TemplateInfo provides detailed information about a template
type TemplateInfo struct {
	Name         string
	Path         string
	Size         int64
	ModTime      time.Time
	Extension    string
	Directory    string
	Dependencies []string
}

// CachingLoader interface for loaders that support caching
type CachingLoader interface {
	AdvancedLoader
	ClearCache()
	GetCacheStats() CacheStats
}

// TemplateParser interface for parsing template content
type TemplateParser interface {
	ParseTemplate(name, content string) (*parser.TemplateNode, error)
}

// CacheStats provides information about cache performance
type CacheStats struct {
	Hits   int64
	Misses int64
	Size   int
}

// cachedTemplate represents a cached template with metadata
type cachedTemplate struct {
	template *parser.TemplateNode
	source   *TemplateSource
	expires  time.Time
}

// Base implementation (keeping compatibility)
type BaseLoader struct{}

func (b *BaseLoader) IsCached(name string) bool {
	return false
}

func (b *BaseLoader) ListTemplates() ([]string, error) {
	return nil, fmt.Errorf("listing templates not supported")
}

type LoaderFunc func(name string) (string, error)

func (f LoaderFunc) GetSource(name string) (string, error) {
	return f(name)
}

func (f LoaderFunc) IsCached(name string) bool {
	return false
}

func (f LoaderFunc) ListTemplates() ([]string, error) {
	return nil, fmt.Errorf("listing templates not supported")
}

func ReadAll(r io.Reader) (string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FileSystemLoader loads templates from the filesystem
type FileSystemLoader struct {
	searchPaths []string
	extensions  []string
	encoding    string
	followLinks bool
	cache       map[string]*cachedTemplate
	cacheMutex  sync.RWMutex
	parser      TemplateParser
	stats       CacheStats
}

// NewFileSystemLoader creates a new filesystem loader
func NewFileSystemLoader(searchPaths []string, parser TemplateParser) *FileSystemLoader {
	return &FileSystemLoader{
		searchPaths: searchPaths,
		extensions:  []string{".html", ".htm", ".jinja", ".jinja2", ".j2"},
		encoding:    "utf-8",
		followLinks: false,
		cache:       make(map[string]*cachedTemplate),
		parser:      parser,
	}
}

// SetExtensions sets the file extensions to search for
func (f *FileSystemLoader) SetExtensions(extensions []string) {
	f.extensions = extensions
}

// SetFollowLinks enables or disables following symbolic links
func (f *FileSystemLoader) SetFollowLinks(follow bool) {
	f.followLinks = follow
}

// GetSource implements the base Loader interface
func (f *FileSystemLoader) GetSource(name string) (string, error) {
	source, err := f.GetSourceWithMetadata(name)
	if err != nil {
		return "", err
	}
	return source.Content, nil
}

// IsCached checks if a template is cached
func (f *FileSystemLoader) IsCached(name string) bool {
	f.cacheMutex.RLock()
	defer f.cacheMutex.RUnlock()

	cached, ok := f.cache[name]
	return ok && !f.isExpired(cached)
}

// LoadTemplate loads and parses a template by name
func (f *FileSystemLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	f.cacheMutex.RLock()
	if cached, ok := f.cache[name]; ok && !f.isExpired(cached) {
		f.cacheMutex.RUnlock()
		f.stats.Hits++
		return cached.template, nil
	}
	f.cacheMutex.RUnlock()

	f.stats.Misses++

	source, err := f.GetSourceWithMetadata(name)
	if err != nil {
		return nil, err
	}

	template, err := f.parser.ParseTemplate(name, source.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %v", name, err)
	}

	// Cache the template
	f.cacheMutex.Lock()
	f.cache[name] = &cachedTemplate{
		template: template,
		source:   source,
		expires:  time.Now().Add(5 * time.Minute), // 5 minute cache
	}
	f.cacheMutex.Unlock()

	return template, nil
}

// GetSourceWithMetadata retrieves the source content of a template with metadata
func (f *FileSystemLoader) GetSourceWithMetadata(name string) (*TemplateSource, error) {
	resolvedPath, err := f.findTemplate(name)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %v", name, err)
	}

	stat, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat template %s: %v", name, err)
	}

	return &TemplateSource{
		Name:    name,
		Content: string(content),
		ModTime: stat.ModTime(),
	}, nil
}

// ResolveTemplateName resolves a template name to its canonical form
func (f *FileSystemLoader) ResolveTemplateName(name string) string {
	// Clean the path and remove any directory traversal attempts
	name = filepath.Clean(name)
	name = strings.TrimPrefix(name, "/")

	// Ensure no directory traversal
	if strings.Contains(name, "..") {
		return ""
	}

	return name
}

// ListTemplates returns a list of all available templates
func (f *FileSystemLoader) ListTemplates() ([]string, error) {
	var templates []string
	seen := make(map[string]bool)

	for _, searchPath := range f.searchPaths {
		err := filepath.WalkDir(searchPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// Skip directories we can't read
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if d.IsDir() {
				return nil
			}

			// Check if file has a valid extension
			ext := filepath.Ext(path)
			validExt := false
			for _, validExtension := range f.extensions {
				if ext == validExtension {
					validExt = true
					break
				}
			}

			if !validExt {
				return nil
			}

			// Get relative path from search path
			relPath, err := filepath.Rel(searchPath, path)
			if err != nil {
				return nil
			}

			// Convert to forward slashes for consistency
			templateName := filepath.ToSlash(relPath)

			if !seen[templateName] {
				templates = append(templates, templateName)
				seen[templateName] = true
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk directory %s: %v", searchPath, err)
		}
	}

	return templates, nil
}

// SearchTemplates searches for templates matching a pattern (glob-style)
func (f *FileSystemLoader) SearchTemplates(pattern string) ([]string, error) {
	allTemplates, err := f.ListTemplates()
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, templateName := range allTemplates {
		// Direct template name match
		matched, err := filepath.Match(pattern, templateName)
		if err != nil {
			continue // Skip invalid patterns
		}
		if matched {
			matches = append(matches, templateName)
			continue
		}

		// Check if pattern matches just the filename
		baseName := filepath.Base(templateName)
		matched, err = filepath.Match(pattern, baseName)
		if err == nil && matched {
			matches = append(matches, templateName)
			continue
		}

		// Check if pattern matches the full path with different separators
		// Convert pattern to handle directory separators
		if strings.Contains(pattern, "/") {
			matched, err = filepath.Match(pattern, templateName)
			if err == nil && matched {
				matches = append(matches, templateName)
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueMatches []string
	for _, match := range matches {
		if !seen[match] {
			uniqueMatches = append(uniqueMatches, match)
			seen[match] = true
		}
	}

	return uniqueMatches, nil
}

// GetTemplatesByExtension returns all templates with a specific extension
func (f *FileSystemLoader) GetTemplatesByExtension(ext string) ([]string, error) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	allTemplates, err := f.ListTemplates()
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, templateName := range allTemplates {
		if filepath.Ext(templateName) == ext {
			matches = append(matches, templateName)
		}
	}

	return matches, nil
}

// GetTemplatesInDirectory returns all templates in a specific directory
func (f *FileSystemLoader) GetTemplatesInDirectory(dir string) ([]string, error) {
	// Normalize directory path
	dir = strings.TrimSuffix(dir, "/")
	if dir == "" {
		dir = "."
	}

	allTemplates, err := f.ListTemplates()
	if err != nil {
		return nil, err
	}

	var matches []string
	for _, templateName := range allTemplates {
		templateDir := filepath.Dir(templateName)
		if templateDir == dir || (dir == "." && !strings.Contains(templateName, "/")) {
			matches = append(matches, templateName)
		}
	}

	return matches, nil
}

// GetTemplateInfo returns detailed information about a template
func (f *FileSystemLoader) GetTemplateInfo(name string) (*TemplateInfo, error) {
	resolvedPath, err := f.findTemplate(name)
	if err != nil {
		return nil, err
	}

	stat, err := os.Stat(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat template %s: %v", name, err)
	}

	info := &TemplateInfo{
		Name:      name,
		Path:      resolvedPath,
		Size:      stat.Size(),
		ModTime:   stat.ModTime(),
		Extension: filepath.Ext(name),
		Directory: filepath.Dir(name),
	}

	// Extract dependencies by parsing the template source
	source, err := f.GetSource(name)
	if err == nil {
		dependencies, _ := f.extractDependencies(source)
		info.Dependencies = dependencies
	}

	return info, nil
}

// extractDependencies extracts template dependencies from source content
func (f *FileSystemLoader) extractDependencies(source string) ([]string, error) {
	var dependencies []string

	// Use pre-compiled regex patterns for dependency extraction
	patterns := []*regexp.Regexp{
		reExtendsPattern,
		reIncludePattern,
		reImportPattern,
		reFromPattern,
	}

	for _, re := range patterns {
		matches := re.FindAllStringSubmatch(source, -1)
		for _, match := range matches {
			if len(match) > 1 {
				dependencies = append(dependencies, match[1])
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueDeps []string
	for _, dep := range dependencies {
		if !seen[dep] {
			uniqueDeps = append(uniqueDeps, dep)
			seen[dep] = true
		}
	}

	return uniqueDeps, nil
}

// ClearCache clears the template cache
func (f *FileSystemLoader) ClearCache() {
	f.cacheMutex.Lock()
	defer f.cacheMutex.Unlock()
	f.cache = make(map[string]*cachedTemplate)
}

// GetCacheStats returns cache performance statistics
func (f *FileSystemLoader) GetCacheStats() CacheStats {
	f.cacheMutex.RLock()
	defer f.cacheMutex.RUnlock()

	stats := f.stats
	stats.Size = len(f.cache)
	return stats
}

// findTemplate finds the full path to a template file
func (f *FileSystemLoader) findTemplate(name string) (string, error) {
	name = f.ResolveTemplateName(name)
	if name == "" {
		return "", fmt.Errorf("invalid template name")
	}

	// Try each search path
	for _, searchPath := range f.searchPaths {
		// Try the name as-is first
		fullPath := filepath.Join(searchPath, name)
		if f.fileExists(fullPath) {
			return fullPath, nil
		}

		// If no extension, try with each valid extension
		if filepath.Ext(name) == "" {
			for _, ext := range f.extensions {
				fullPathWithExt := fullPath + ext
				if f.fileExists(fullPathWithExt) {
					return fullPathWithExt, nil
				}
			}
		}
	}

	return "", fmt.Errorf("template not found: %s", name)
}

// fileExists checks if a file exists and is readable
func (f *FileSystemLoader) fileExists(path string) bool {
	// Use Lstat to check symlinks without following them
	stat, err := os.Lstat(path)
	if err != nil {
		return false
	}

	// If followLinks is false, check if it's a symlink
	if !f.followLinks {
		if stat.Mode()&os.ModeSymlink != 0 {
			return false
		}
	}

	// If it's a symlink and we're following links, use Stat to check the target
	if stat.Mode()&os.ModeSymlink != 0 && f.followLinks {
		stat, err = os.Stat(path)
		if err != nil {
			return false
		}
	}

	// Check if it's a file and not a directory
	if stat.IsDir() {
		return false
	}

	return true
}

// isExpired checks if a cached template has expired
func (f *FileSystemLoader) isExpired(cached *cachedTemplate) bool {
	return time.Now().After(cached.expires)
}

// EmbedLoader loads templates from embedded filesystem
type EmbedLoader struct {
	fs         embed.FS
	prefix     string
	extensions []string
	cache      map[string]*cachedTemplate
	cacheMutex sync.RWMutex
	parser     TemplateParser
	stats      CacheStats
}

// NewEmbedLoader creates a new embed filesystem loader
func NewEmbedLoader(embedFS embed.FS, prefix string, parser TemplateParser) *EmbedLoader {
	return &EmbedLoader{
		fs:         embedFS,
		prefix:     prefix,
		extensions: []string{".html", ".htm", ".jinja", ".jinja2", ".j2"},
		cache:      make(map[string]*cachedTemplate),
		parser:     parser,
	}
}

// SetExtensions sets the file extensions to search for
func (e *EmbedLoader) SetExtensions(extensions []string) {
	e.extensions = extensions
}

// GetSource implements the base Loader interface
func (e *EmbedLoader) GetSource(name string) (string, error) {
	source, err := e.GetSourceWithMetadata(name)
	if err != nil {
		return "", err
	}
	return source.Content, nil
}

// IsCached checks if a template is cached
func (e *EmbedLoader) IsCached(name string) bool {
	e.cacheMutex.RLock()
	defer e.cacheMutex.RUnlock()

	_, ok := e.cache[name]
	return ok
}

// LoadTemplate loads and parses a template by name
func (e *EmbedLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	e.cacheMutex.RLock()
	if cached, ok := e.cache[name]; ok {
		e.cacheMutex.RUnlock()
		e.stats.Hits++
		return cached.template, nil
	}
	e.cacheMutex.RUnlock()

	e.stats.Misses++

	source, err := e.GetSourceWithMetadata(name)
	if err != nil {
		return nil, err
	}

	template, err := e.parser.ParseTemplate(name, source.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %v", name, err)
	}

	// Cache the template (embedded templates don't change)
	e.cacheMutex.Lock()
	e.cache[name] = &cachedTemplate{
		template: template,
		source:   source,
		expires:  time.Now().Add(24 * time.Hour), // Long cache for embedded
	}
	e.cacheMutex.Unlock()

	return template, nil
}

// GetSourceWithMetadata retrieves the source content of a template with metadata
func (e *EmbedLoader) GetSourceWithMetadata(name string) (*TemplateSource, error) {
	resolvedPath := e.resolvePath(name)

	content, err := e.fs.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return &TemplateSource{
		Name:    name,
		Content: string(content),
		ModTime: time.Time{}, // Embedded files don't have meaningful mod times
	}, nil
}

// ResolveTemplateName resolves a template name to its canonical form
func (e *EmbedLoader) ResolveTemplateName(name string) string {
	// Clean the path and remove any directory traversal attempts
	name = filepath.Clean(name)
	name = strings.TrimPrefix(name, "/")

	// Ensure no directory traversal
	if strings.Contains(name, "..") {
		return ""
	}

	return name
}

// ListTemplates returns a list of all available templates
func (e *EmbedLoader) ListTemplates() ([]string, error) {
	var templates []string

	err := fs.WalkDir(e.fs, e.prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			return nil
		}

		// Check if file has a valid extension
		ext := filepath.Ext(path)
		validExt := false
		for _, validExtension := range e.extensions {
			if ext == validExtension {
				validExt = true
				break
			}
		}

		if !validExt {
			return nil
		}

		// Get relative path from prefix
		templateName := strings.TrimPrefix(path, e.prefix)
		templateName = strings.TrimPrefix(templateName, "/")

		// Convert to forward slashes for consistency
		templateName = filepath.ToSlash(templateName)

		templates = append(templates, templateName)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk embedded filesystem: %v", err)
	}

	return templates, nil
}

// ClearCache clears the template cache
func (e *EmbedLoader) ClearCache() {
	e.cacheMutex.Lock()
	defer e.cacheMutex.Unlock()
	e.cache = make(map[string]*cachedTemplate)
}

// GetCacheStats returns cache performance statistics
func (e *EmbedLoader) GetCacheStats() CacheStats {
	e.cacheMutex.RLock()
	defer e.cacheMutex.RUnlock()

	stats := e.stats
	stats.Size = len(e.cache)
	return stats
}

// resolvePath resolves a template name to a filesystem path
func (e *EmbedLoader) resolvePath(name string) string {
	name = e.ResolveTemplateName(name)
	if name == "" {
		return ""
	}

	path := filepath.Join(e.prefix, name)

	// Try the name as-is first
	if e.fileExists(path) {
		return path
	}

	// If no extension, try with each valid extension
	if filepath.Ext(name) == "" {
		for _, ext := range e.extensions {
			pathWithExt := path + ext
			if e.fileExists(pathWithExt) {
				return pathWithExt
			}
		}
	}

	return path // Return original path even if not found for error reporting
}

// fileExists checks if a file exists in the embedded filesystem
func (e *EmbedLoader) fileExists(path string) bool {
	_, err := e.fs.Open(path)
	return err == nil
}

// ChainLoader combines multiple loaders, trying them in order
type ChainLoader struct {
	loaders []AdvancedLoader
}

// NewChainLoader creates a new chain loader
func NewChainLoader(loaders ...AdvancedLoader) *ChainLoader {
	return &ChainLoader{
		loaders: loaders,
	}
}

// AddLoader adds a loader to the chain
func (c *ChainLoader) AddLoader(loader AdvancedLoader) {
	c.loaders = append(c.loaders, loader)
}

// GetSource implements the base Loader interface
func (c *ChainLoader) GetSource(name string) (string, error) {
	var lastErr error

	for _, loader := range c.loaders {
		source, err := loader.GetSource(name)
		if err == nil {
			return source, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", lastErr
	}

	return "", fmt.Errorf("template not found: %s", name)
}

// IsCached checks if any loader has the template cached
func (c *ChainLoader) IsCached(name string) bool {
	for _, loader := range c.loaders {
		if loader.IsCached(name) {
			return true
		}
	}
	return false
}

// LoadTemplate loads a template using the first loader that succeeds
func (c *ChainLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	var lastErr error

	for _, loader := range c.loaders {
		template, err := loader.LoadTemplate(name)
		if err == nil {
			return template, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("template not found: %s", name)
}

// GetSourceWithMetadata gets template source with metadata using the first loader that succeeds
func (c *ChainLoader) GetSourceWithMetadata(name string) (*TemplateSource, error) {
	var lastErr error

	for _, loader := range c.loaders {
		source, err := loader.GetSourceWithMetadata(name)
		if err == nil {
			return source, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return nil, lastErr
	}

	return nil, fmt.Errorf("template not found: %s", name)
}

// ResolveTemplateName resolves using the first loader
func (c *ChainLoader) ResolveTemplateName(name string) string {
	if len(c.loaders) > 0 {
		return c.loaders[0].ResolveTemplateName(name)
	}
	return name
}

// ListTemplates returns templates from all loaders
func (c *ChainLoader) ListTemplates() ([]string, error) {
	seen := make(map[string]bool)
	var allTemplates []string

	for _, loader := range c.loaders {
		templates, err := loader.ListTemplates()
		if err != nil {
			continue // Skip loaders that can't list templates
		}

		for _, template := range templates {
			if !seen[template] {
				allTemplates = append(allTemplates, template)
				seen[template] = true
			}
		}
	}

	return allTemplates, nil
}

// StringLoader loads templates from string content (useful for testing)
type StringLoader struct {
	templates map[string]string
	parser    TemplateParser
}

// NewStringLoader creates a new string loader
func NewStringLoader(parser TemplateParser) *StringLoader {
	return &StringLoader{
		templates: make(map[string]string),
		parser:    parser,
	}
}

// AddTemplate adds a template with the given name and content
func (s *StringLoader) AddTemplate(name, content string) {
	s.templates[name] = content
}

// GetSource implements the base Loader interface
func (s *StringLoader) GetSource(name string) (string, error) {
	content, exists := s.templates[name]
	if !exists {
		return "", fmt.Errorf("template not found: %s", name)
	}
	return content, nil
}

// IsCached always returns true for string loader (templates are in memory)
func (s *StringLoader) IsCached(name string) bool {
	_, exists := s.templates[name]
	return exists
}

// LoadTemplate loads and parses a template by name
func (s *StringLoader) LoadTemplate(name string) (*parser.TemplateNode, error) {
	content, exists := s.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return s.parser.ParseTemplate(name, content)
}

// GetSourceWithMetadata retrieves the source content of a template with metadata
func (s *StringLoader) GetSourceWithMetadata(name string) (*TemplateSource, error) {
	content, exists := s.templates[name]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	return &TemplateSource{
		Name:    name,
		Content: content,
		ModTime: time.Now(),
	}, nil
}

// ResolveTemplateName resolves a template name (identity function for string loader)
func (s *StringLoader) ResolveTemplateName(name string) string {
	return name
}

// ListTemplates returns a list of all available templates
func (s *StringLoader) ListTemplates() ([]string, error) {
	templates := make([]string, 0, len(s.templates))
	for name := range s.templates {
		templates = append(templates, name)
	}
	return templates, nil
}
