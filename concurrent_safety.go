package miya

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ThreadSafeTemplate provides thread-safe template operations
type ThreadSafeTemplate struct {
	*Template
	mutex sync.RWMutex
}

// NewThreadSafeTemplate creates a new thread-safe template wrapper
func NewThreadSafeTemplate(tmpl *Template) *ThreadSafeTemplate {
	return &ThreadSafeTemplate{
		Template: tmpl,
	}
}

// RenderConcurrent renders the template in a thread-safe manner
func (tst *ThreadSafeTemplate) RenderConcurrent(ctx Context) (string, error) {
	tst.mutex.RLock()
	defer tst.mutex.RUnlock()
	return tst.Template.Render(ctx)
}

// ThreadSafeEnvironment provides thread-safe environment operations
type ThreadSafeEnvironment struct {
	*Environment
	templateMutex sync.RWMutex
	filterMutex   sync.RWMutex
	globalMutex   sync.RWMutex
}

// NewThreadSafeEnvironment creates a new thread-safe environment
func NewThreadSafeEnvironment(opts ...EnvironmentOption) *ThreadSafeEnvironment {
	return &ThreadSafeEnvironment{
		Environment: NewEnvironment(opts...),
	}
}

// GetTemplateConcurrent retrieves a template in a thread-safe manner
func (tse *ThreadSafeEnvironment) GetTemplateConcurrent(name string) (*ThreadSafeTemplate, error) {
	tse.templateMutex.RLock()
	defer tse.templateMutex.RUnlock()

	tmpl, err := tse.Environment.GetTemplate(name)
	if err != nil {
		return nil, err
	}

	return NewThreadSafeTemplate(tmpl), nil
}

// FromStringConcurrent compiles a template from string in a thread-safe manner
func (tse *ThreadSafeEnvironment) FromStringConcurrent(source string) (*ThreadSafeTemplate, error) {
	tse.templateMutex.RLock()
	defer tse.templateMutex.RUnlock()

	tmpl, err := tse.Environment.FromString(source)
	if err != nil {
		return nil, err
	}

	return NewThreadSafeTemplate(tmpl), nil
}

// AddFilterConcurrent adds a filter in a thread-safe manner
func (tse *ThreadSafeEnvironment) AddFilterConcurrent(name string, filter FilterFunc) error {
	tse.filterMutex.Lock()
	defer tse.filterMutex.Unlock()
	return tse.Environment.AddFilter(name, filter)
}

// AddGlobalConcurrent adds a global variable in a thread-safe manner
func (tse *ThreadSafeEnvironment) AddGlobalConcurrent(name string, value interface{}) {
	tse.globalMutex.Lock()
	defer tse.globalMutex.Unlock()
	tse.Environment.AddGlobal(name, value)
}

// ConcurrentTemplateRenderer provides high-performance concurrent template rendering
type ConcurrentTemplateRenderer struct {
	template *Template
	workers  int
	workChan chan renderJob
	wg       sync.WaitGroup
	started  int32
	stopChan chan struct{}
}

type renderJob struct {
	ctx      Context
	resultCh chan renderResult
}

type renderResult struct {
	output string
	err    error
}

// NewConcurrentTemplateRenderer creates a new concurrent template renderer
func NewConcurrentTemplateRenderer(template *Template, workers int) *ConcurrentTemplateRenderer {
	return &ConcurrentTemplateRenderer{
		template: template,
		workers:  workers,
		workChan: make(chan renderJob, workers*2), // Buffer for 2x workers
		stopChan: make(chan struct{}),
	}
}

// Start starts the concurrent renderer workers
func (ctr *ConcurrentTemplateRenderer) Start() {
	if !atomic.CompareAndSwapInt32(&ctr.started, 0, 1) {
		return // Already started
	}

	for i := 0; i < ctr.workers; i++ {
		ctr.wg.Add(1)
		go ctr.worker()
	}
}

// Stop stops the concurrent renderer workers
func (ctr *ConcurrentTemplateRenderer) Stop() {
	if !atomic.CompareAndSwapInt32(&ctr.started, 1, 0) {
		return // Not started or already stopped
	}

	close(ctr.stopChan)
	ctr.wg.Wait()
}

// RenderAsync renders a template asynchronously
func (ctr *ConcurrentTemplateRenderer) RenderAsync(ctx Context) <-chan renderResult {
	resultCh := make(chan renderResult, 1)

	// Check if stopped first to avoid race condition where both channels are ready
	select {
	case <-ctr.stopChan:
		resultCh <- renderResult{err: fmt.Errorf("renderer is stopped")}
		return resultCh
	default:
	}

	// Try to send work, but also check for stop
	select {
	case ctr.workChan <- renderJob{ctx: ctx, resultCh: resultCh}:
		return resultCh
	case <-ctr.stopChan:
		resultCh <- renderResult{err: fmt.Errorf("renderer is stopped")}
		return resultCh
	}
}

// worker processes render jobs
func (ctr *ConcurrentTemplateRenderer) worker() {
	defer ctr.wg.Done()

	for {
		select {
		case job := <-ctr.workChan:
			// Process job with panic recovery and guaranteed channel close
			func() {
				var result renderResult
				defer func() {
					// Always close the channel after sending result
					defer close(job.resultCh)
					if r := recover(); r != nil {
						result = renderResult{
							output: "",
							err:    fmt.Errorf("panic during template render: %v", r),
						}
					}
					job.resultCh <- result
				}()
				output, err := ctr.template.Render(job.ctx)
				result = renderResult{output: output, err: err}
			}()
		case <-ctr.stopChan:
			return
		}
	}
}

// RenderBatch renders multiple contexts in parallel
func (ctr *ConcurrentTemplateRenderer) RenderBatch(contexts []Context) ([]string, []error) {
	results := make([]string, len(contexts))
	errors := make([]error, len(contexts))

	var wg sync.WaitGroup
	for i, ctx := range contexts {
		wg.Add(1)
		go func(idx int, context Context) {
			defer wg.Done()
			// Recover from panics to prevent one bad render from affecting others
			defer func() {
				if r := recover(); r != nil {
					errors[idx] = fmt.Errorf("panic during template render: %v", r)
				}
			}()
			result, err := ctr.template.Render(context)
			results[idx] = result
			errors[idx] = err
		}(i, ctx)
	}

	wg.Wait()
	return results, errors
}

// TemplatePool provides a pool of template instances for concurrent rendering
type TemplatePool struct {
	pool     sync.Pool
	template *Template
}

// NewTemplatePool creates a new template pool
func NewTemplatePool(template *Template) *TemplatePool {
	return &TemplatePool{
		template: template,
		pool: sync.Pool{
			New: func() interface{} {
				// Return a copy of the template for concurrent use
				return &Template{
					name:   template.name,
					source: template.source,
					env:    template.env,
					ast:    template.ast,
				}
			},
		},
	}
}

// Get gets a template from the pool
func (tp *TemplatePool) Get() *Template {
	if tmpl, ok := tp.pool.Get().(*Template); ok {
		return tmpl
	}
	// Fallback: create new template if pool returns unexpected type
	return &Template{
		name:   tp.template.name,
		source: tp.template.source,
		env:    tp.template.env,
		ast:    tp.template.ast,
	}
}

// Put returns a template to the pool
func (tp *TemplatePool) Put(tmpl *Template) {
	tp.pool.Put(tmpl)
}

// RenderConcurrent renders using a pooled template
func (tp *TemplatePool) RenderConcurrent(ctx Context) (string, error) {
	tmpl := tp.Get()
	defer tp.Put(tmpl)
	return tmpl.Render(ctx)
}

// ConcurrentContextPool provides thread-safe context pooling
type ConcurrentContextPool struct {
	pool  sync.Pool
	mutex sync.Mutex
	stats struct {
		gets int64
		puts int64
	}
}

// NewConcurrentContextPool creates a new concurrent context pool
func NewConcurrentContextPool() *ConcurrentContextPool {
	return &ConcurrentContextPool{
		pool: sync.Pool{
			New: func() interface{} {
				return NewContext()
			},
		},
	}
}

// Get gets a context from the pool
func (ccp *ConcurrentContextPool) Get() Context {
	atomic.AddInt64(&ccp.stats.gets, 1)
	if ctx, ok := ccp.pool.Get().(Context); ok {
		return ctx
	}
	// Fallback: create new context if pool returns unexpected type
	return NewContext()
}

// Put returns a context to the pool
func (ccp *ConcurrentContextPool) Put(ctx Context) {
	atomic.AddInt64(&ccp.stats.puts, 1)
	// In a real implementation, we'd want to reset the context
	ccp.pool.Put(NewContext()) // Create fresh context for safety
}

// GetStats returns pool statistics
func (ccp *ConcurrentContextPool) GetStats() (gets, puts int64) {
	return atomic.LoadInt64(&ccp.stats.gets), atomic.LoadInt64(&ccp.stats.puts)
}

// GlobalConcurrentContextPool is a global context pool for concurrent rendering
var GlobalConcurrentContextPool = NewConcurrentContextPool()

// RateLimitedRenderer provides rate-limited template rendering
type RateLimitedRenderer struct {
	template  *Template
	semaphore chan struct{}
}

// NewRateLimitedRenderer creates a new rate-limited renderer
func NewRateLimitedRenderer(template *Template, maxConcurrent int) *RateLimitedRenderer {
	return &RateLimitedRenderer{
		template:  template,
		semaphore: make(chan struct{}, maxConcurrent),
	}
}

// Render renders the template with rate limiting
func (rlr *RateLimitedRenderer) Render(ctx Context) (string, error) {
	// Acquire semaphore
	rlr.semaphore <- struct{}{}
	defer func() { <-rlr.semaphore }() // Release semaphore

	return rlr.template.Render(ctx)
}

// ConcurrentCacheManager provides thread-safe cache management
type ConcurrentCacheManager struct {
	cache sync.Map
	stats struct {
		hits   int64
		misses int64
	}
}

// NewConcurrentCacheManager creates a new concurrent cache manager
func NewConcurrentCacheManager() *ConcurrentCacheManager {
	return &ConcurrentCacheManager{}
}

// Get gets a value from the cache
func (ccm *ConcurrentCacheManager) Get(key string) (interface{}, bool) {
	value, ok := ccm.cache.Load(key)
	if ok {
		atomic.AddInt64(&ccm.stats.hits, 1)
	} else {
		atomic.AddInt64(&ccm.stats.misses, 1)
	}
	return value, ok
}

// Set sets a value in the cache
func (ccm *ConcurrentCacheManager) Set(key string, value interface{}) {
	ccm.cache.Store(key, value)
}

// Delete deletes a value from the cache
func (ccm *ConcurrentCacheManager) Delete(key string) {
	ccm.cache.Delete(key)
}

// GetStats returns cache statistics
func (ccm *ConcurrentCacheManager) GetStats() (hits, misses int64) {
	return atomic.LoadInt64(&ccm.stats.hits), atomic.LoadInt64(&ccm.stats.misses)
}

// GlobalConcurrentCache is a global concurrent cache for template rendering
var GlobalConcurrentCache = NewConcurrentCacheManager()

// ConcurrentEnvironmentRegistry manages multiple environments concurrently
type ConcurrentEnvironmentRegistry struct {
	environments sync.Map
	defaultEnv   *ThreadSafeEnvironment
}

// NewConcurrentEnvironmentRegistry creates a new concurrent environment registry
func NewConcurrentEnvironmentRegistry() *ConcurrentEnvironmentRegistry {
	return &ConcurrentEnvironmentRegistry{
		defaultEnv: NewThreadSafeEnvironment(),
	}
}

// RegisterEnvironment registers an environment with a name
func (cer *ConcurrentEnvironmentRegistry) RegisterEnvironment(name string, env *ThreadSafeEnvironment) {
	cer.environments.Store(name, env)
}

// GetEnvironment gets an environment by name
func (cer *ConcurrentEnvironmentRegistry) GetEnvironment(name string) (*ThreadSafeEnvironment, bool) {
	if env, ok := cer.environments.Load(name); ok {
		if tsEnv, ok := env.(*ThreadSafeEnvironment); ok {
			return tsEnv, true
		}
	}
	return nil, false
}

// GetDefaultEnvironment returns the default environment
func (cer *ConcurrentEnvironmentRegistry) GetDefaultEnvironment() *ThreadSafeEnvironment {
	return cer.defaultEnv
}

// GlobalEnvironmentRegistry is a global registry for concurrent environments
var GlobalEnvironmentRegistry = NewConcurrentEnvironmentRegistry()
