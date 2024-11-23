package dashboard

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
