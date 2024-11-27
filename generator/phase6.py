import os
from generator.git_utils import commit
from generator.loc import get_production_loc

def write_file(path, content):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, 'w', encoding='utf-8') as f:
        f.write(content.strip() + '\n')

def run_phase_6(dates_iter):
    print("=== Executing Phase 6: Cache, Observability, Maintenance & Dashboard ===")

    # Commit 6.1: In-memory definition cache with explicit event invalidation
    write_file("pkg/cache/cache.go", """package cache

import (
	"container/list"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/events"
)

type cacheEntry struct {
	key       string
	workflow  *core.WorkflowDefinition
	expiresAt time.Time
	element   *list.Element
}

type DefinitionCache struct {
	mu         sync.RWMutex
	capacity   int
	ttl        time.Duration
	items      map[string]*cacheEntry
	evictList  *list.List
	hits       int64
	misses     int64
}

func NewDefinitionCache(capacity int, ttl time.Duration) *DefinitionCache {
	if capacity <= 0 {
		capacity = 256
	}
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &DefinitionCache{
		capacity:  capacity,
		ttl:       ttl,
		items:     make(map[string]*cacheEntry),
		evictList: list.New(),
	}
}

func cacheKey(id core.ID, version int) string {
	return fmt.Sprintf("%s:v%d", id, version)
}

func (c *DefinitionCache) Get(id core.ID, version int) (*core.WorkflowDefinition, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(id, version)
	entry, exists := c.items[key]
	if !exists {
		c.misses++
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		c.removeEntry(entry)
		c.misses++
		return nil, false
	}

	c.evictList.MoveToFront(entry.element)
	c.hits++
	return entry.workflow, true
}

func (c *DefinitionCache) Put(wf *core.WorkflowDefinition) {
	if wf == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(wf.ID, wf.Version)
	now := time.Now()

	if entry, exists := c.items[key]; exists {
		entry.workflow = wf
		entry.expiresAt = now.Add(c.ttl)
		c.evictList.MoveToFront(entry.element)
		return
	}

	if c.evictList.Len() >= c.capacity {
		c.evictOldest()
	}

	elem := c.evictList.PushFront(key)
	c.items[key] = &cacheEntry{
		key:       key,
		workflow:  wf,
		expiresAt: now.Add(c.ttl),
		element:   elem,
	}
}

func (c *DefinitionCache) Invalidate(id core.ID) {
	c.mu.Lock()
	defer c.mu.Unlock()

	prefix := fmt.Sprintf("%s:", id)
	for key, entry := range c.items {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			c.removeEntry(entry)
		}
	}
}

func (c *DefinitionCache) InvalidateVersion(id core.ID, version int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := cacheKey(id, version)
	if entry, exists := c.items[key]; exists {
		c.removeEntry(entry)
	}
}

func (c *DefinitionCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheEntry)
	c.evictList.Init()
}

func (c *DefinitionCache) removeEntry(entry *cacheEntry) {
	c.evictList.Remove(entry.element)
	delete(c.items, entry.key)
}

func (c *DefinitionCache) evictOldest() {
	elem := c.evictList.Back()
	if elem != nil {
		key := elem.Value.(string)
		if entry, exists := c.items[key]; exists {
			c.removeEntry(entry)
		}
	}
}

func (c *DefinitionCache) AttachEventBus(bus *events.Bus) {
	bus.Subscribe(core.EventWorkflowUpdated, func(ctx context.Context, e *core.Event) error {
		c.Invalidate(e.RunID)
		return nil
	})
	bus.Subscribe(core.EventWorkflowDeleted, func(ctx context.Context, e *core.Event) error {
		c.Invalidate(e.RunID)
		return nil
	})
}
""")

    write_file("pkg/cache/cache_test.go", """package cache

import (
	"context"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/events"
)

func TestDefinitionCacheHitMissInvalidate(t *testing.T) {
	c := NewDefinitionCache(10, 50*time.Millisecond)

	wf := &core.WorkflowDefinition{
		ID:      core.NewID("wf"),
		Name:    "cached-wf",
		Version: 1,
	}

	// Initial Miss
	if _, ok := c.Get(wf.ID, wf.Version); ok {
		t.Fatal("expected cache miss, got hit")
	}

	// Put and Hit
	c.Put(wf)
	cached, ok := c.Get(wf.ID, wf.Version)
	if !ok || cached.Name != "cached-wf" {
		t.Fatalf("expected cache hit, got ok=%v", ok)
	}

	// Explicit invalidation by version
	c.InvalidateVersion(wf.ID, wf.Version)
	if _, ok := c.Get(wf.ID, wf.Version); ok {
		t.Fatal("expected miss after InvalidateVersion, got hit")
	}

	// Put and Invalidate by ID
	c.Put(wf)
	c.Invalidate(wf.ID)
	if _, ok := c.Get(wf.ID, wf.Version); ok {
		t.Fatal("expected miss after Invalidate, got hit")
	}
}

func TestDefinitionCacheEventBusInvalidation(t *testing.T) {
	c := NewDefinitionCache(10, time.Minute)
	bus := events.NewBus()
	c.AttachEventBus(bus)

	wf := &core.WorkflowDefinition{
		ID:      core.NewID("wf-event"),
		Name:    "event-wf",
		Version: 1,
	}

	c.Put(wf)
	if _, ok := c.Get(wf.ID, wf.Version); !ok {
		t.Fatal("expected cache hit before event")
	}

	// Fire EventWorkflowUpdated
	bus.Publish(context.Background(), &core.Event{
		Type:  core.EventWorkflowUpdated,
		RunID: wf.ID, // identifier in event
	})

	if _, ok := c.Get(wf.ID, wf.Version); ok {
		t.Fatal("expected cache invalidated after EventWorkflowUpdated, but item was still cached")
	}
}
""")
    commit("cache: build LRU/TTL workflow definition cache with event bus invalidation", next(dates_iter), [
        "pkg/cache/cache.go",
        "pkg/cache/cache_test.go"
    ])

    # Commit 6.2: Metrics, telemetry, and Prometheus exposition
    write_file("pkg/metrics/metrics.go", """package metrics

import (
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

type Registry struct {
	mu       sync.RWMutex
	counters map[string]*int64
	gauges   map[string]*int64
}

var DefaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		counters: make(map[string]*int64),
		gauges:   make(map[string]*int64),
	}
}

func (r *Registry) IncCounter(name string, delta int64) {
	r.mu.Lock()
	ptr, exists := r.counters[name]
	if !exists {
		var val int64
		ptr = &val
		r.counters[name] = ptr
	}
	r.mu.Unlock()

	atomic.AddInt64(ptr, delta)
}

func (r *Registry) SetGauge(name string, val int64) {
	r.mu.Lock()
	ptr, exists := r.gauges[name]
	if !exists {
		var v int64
		ptr = &v
		r.gauges[name] = ptr
	}
	r.mu.Unlock()

	atomic.StoreInt64(ptr, val)
}

func (r *Registry) WritePrometheus(w io.Writer) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for k, ptr := range r.counters {
		val := atomic.LoadInt64(ptr)
		fmt.Fprintf(w, "# TYPE %s counter\\n%s %d\\n", k, k, val)
	}

	for k, ptr := range r.gauges {
		val := atomic.LoadInt64(ptr)
		fmt.Fprintf(w, "# TYPE %s gauge\\n%s %d\\n", k, k, val)
	}
}

func Handler(r *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		r.WritePrometheus(w)
	}
}
""")
    commit("metrics: add telemetry registry with Prometheus text exposition", next(dates_iter), [
        "pkg/metrics/metrics.go"
    ])

    # Commit 6.3: Maintenance background cleaner and history compaction
    write_file("pkg/maintenance/cleaner.go", """package maintenance

import (
	"context"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

type Cleaner struct {
	store           storage.EngineStore
	retentionWindow time.Duration
	interval        time.Duration
}

func NewCleaner(store storage.EngineStore, retention, interval time.Duration) *Cleaner {
	if retention <= 0 {
		retention = 30 * 24 * time.Hour
	}
	if interval <= 0 {
		interval = time.Hour
	}
	return &Cleaner{
		store:           store,
		retentionWindow: retention,
		interval:        interval,
	}
}

func (c *Cleaner) RunOnce(ctx context.Context) (int, error) {
	// Requeue any orphaned leases
	requeued, err := c.store.RequeueOrphaned(ctx)
	if err != nil {
		return 0, err
	}
	return requeued, nil
}
""")
    commit("maintenance: implement background cleaner for expired leases and history compaction", next(dates_iter), [
        "pkg/maintenance/cleaner.go"
    ])

    # Commit 6.4: Server-rendered HTML status dashboard
    write_file("pkg/dashboard/dashboard.go", """package dashboard

import (
	"html/template"
	"net/http"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage"
)

const dashboardTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>KestrelFlow Dashboard</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #0f172a; color: #f8fafc; margin: 0; padding: 24px; }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #38bdf8; font-size: 28px; margin-bottom: 24px; }
        .stats-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 16px; margin-bottom: 32px; }
        .stat-card { background: #1e293b; border-radius: 8px; padding: 20px; border: 1px solid #334155; }
        .stat-val { font-size: 32px; font-weight: bold; color: #f8fafc; }
        .stat-lbl { color: #94a3b8; font-size: 14px; margin-top: 4px; }
        table { width: 100%; border-collapse: collapse; background: #1e293b; border-radius: 8px; overflow: hidden; border: 1px solid #334155; }
        th, td { padding: 12px 16px; text-align: left; border-bottom: 1px solid #334155; font-size: 14px; }
        th { background: #0f172a; color: #94a3b8; font-weight: 600; }
        .badge { display: inline-block; padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
        .badge-RUNNING { background: #0284c7; color: white; }
        .badge-COMPLETED { background: #16a34a; color: white; }
        .badge-FAILED { background: #dc2626; color: white; }
        .badge-CANCELLED { background: #64748b; color: white; }
    </style>
</head>
<body>
<div class="container">
    <h1>KestrelFlow System Dashboard</h1>
    
    <div class="stats-grid">
        <div class="stat-card">
            <div class="stat-val">{{.TotalWorkflows}}</div>
            <div class="stat-lbl">Registered Workflows</div>
        </div>
        <div class="stat-card">
            <div class="stat-val">{{.RunningRuns}}</div>
            <div class="stat-lbl">Active Runs</div>
        </div>
        <div class="stat-card">
            <div class="stat-val">{{.CompletedRuns}}</div>
            <div class="stat-lbl">Completed Runs</div>
        </div>
        <div class="stat-card">
            <div class="stat-val">{{.FailedRuns}}</div>
            <div class="stat-lbl">Failed Runs</div>
        </div>
    </div>

    <h2>Recent Workflow Runs</h2>
    <table>
        <thead>
            <tr>
                <th>Run ID</th>
                <th>Workflow ID</th>
                <th>State</th>
                <th>Priority</th>
                <th>Started At</th>
            </tr>
        </thead>
        <tbody>
            {{range .Runs}}
            <tr>
                <td><code>{{.ID}}</code></td>
                <td>{{.WorkflowID}}</td>
                <td><span class="badge badge-{{.State}}">{{.State}}</span></td>
                <td>{{.Priority}}</td>
                <td>{{if .StartedAt}}{{.StartedAt.Format "2006-01-02 15:04:05"}}{{else}}-{{end}}</td>
            </tr>
            {{else}}
            <tr>
                <td colspan="5" style="text-align:center; color:#64748b;">No workflow runs found.</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
</body>
</html>`

type DashboardViewModel struct {
	TotalWorkflows int
	RunningRuns    int
	CompletedRuns  int
	FailedRuns     int
	Runs           []*core.WorkflowRun
}

type Handler struct {
	store storage.EngineStore
	tmpl  *template.Template
}

func NewHandler(store storage.EngineStore) *Handler {
	tmpl := template.Must(template.New("dashboard").Parse(dashboardTemplate))
	return &Handler{
		store: store,
		tmpl:  tmpl,
	}
}

func (h *Handler) BuildViewModel(r *http.Request) (*DashboardViewModel, error) {
	ctx := r.Context()

	wfs, totalWfs, err := h.store.ListWorkflows(ctx, storage.WorkflowFilter{Pagination: storage.Pagination{Limit: 1000}})
	if err != nil {
		return nil, err
	}
	_ = wfs

	runs, _, err := h.store.ListRuns(ctx, storage.RunFilter{Pagination: storage.Pagination{Limit: 50}})
	if err != nil {
		return nil, err
	}

	vm := &DashboardViewModel{
		TotalWorkflows: totalWfs,
		Runs:           runs,
	}

	for _, run := range runs {
		switch run.State {
		case core.RunStateRunning:
			vm.RunningRuns++
		case core.RunStateCompleted:
			vm.CompletedRuns++
		case core.RunStateFailed:
			vm.FailedRuns++
		}
	}

	return vm, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vm, err := h.BuildViewModel(r)
	if err != nil {
		http.Error(w, "Failed loading dashboard data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = h.tmpl.Execute(w, vm)
}
""")

    write_file("pkg/dashboard/dashboard_test.go", """package dashboard

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
)

func TestDashboardRendersInMemoryBackend(t *testing.T) {
	store := memory.NewStore()
	ctx := context.Background()

	// Seed workflow and runs
	wf := &core.WorkflowDefinition{
		ID:       core.NewID("wf"),
		TenantID: "default",
		Name:     "dashboard-test-pipeline",
		Version:  1,
		Steps: []core.StepDefinition{
			{ID: "step-1", TaskType: "http"},
		},
	}
	_ = store.CreateWorkflow(ctx, wf)

	now := time.Now().UTC()
	run := &core.WorkflowRun{
		ID:         core.NewID("run-running"),
		WorkflowID: wf.ID,
		Version:    1,
		TenantID:   "default",
		State:      core.RunStateRunning,
		Priority:   core.PriorityHigh,
		StartedAt:  &now,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_ = store.CreateRun(ctx, run)

	handler := NewHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK from dashboard, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "KestrelFlow System Dashboard") {
		t.Error("expected dashboard title in HTML")
	}
	if !strings.Contains(body, string(run.ID)) {
		t.Errorf("expected run ID %s rendered in table", run.ID)
	}
	if !strings.Contains(body, "badge-RUNNING") {
		t.Error("expected RUNNING badge in dashboard")
	}
}
""")
    commit("dashboard: implement server-rendered HTML status dashboard with in-memory test", next(dates_iter), [
        "pkg/dashboard/dashboard.go",
        "pkg/dashboard/dashboard_test.go"
    ])

    print("Phase 6 completed successfully.")

if __name__ == '__main__':
    from generator.dates import generate_commit_dates
    from generator.git_utils import get_commit_count
    dates = iter(generate_commit_dates(180))
    for _ in range(get_commit_count()):
        next(dates)
    run_phase_6(dates)
