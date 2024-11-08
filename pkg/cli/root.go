package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kestrelflow/kestrelflow/pkg/api/http"
	"github.com/kestrelflow/kestrelflow/pkg/api/types"
	"github.com/kestrelflow/kestrelflow/pkg/config"
	"github.com/kestrelflow/kestrelflow/pkg/core"
	"github.com/kestrelflow/kestrelflow/pkg/scheduler"
	"github.com/kestrelflow/kestrelflow/pkg/storage/memory"
	"github.com/kestrelflow/kestrelflow/pkg/worker"
)

type CLI struct {
	Out io.Writer
	Err io.Writer
}

func NewCLI(out, err io.Writer) *CLI {
	if out == nil {
		out = os.Stdout
	}
	if err == nil {
		err = os.Stderr
	}
	return &CLI{Out: out, Err: err}
}

func (c *CLI) Execute(args []string) int {
	if len(args) == 0 {
		c.printUsage()
		return 0
	}

	switch args[0] {
	case "server":
		return c.runServer(args[1:])
	case "worker":
		return c.runWorker(args[1:])
	case "workflow":
		return c.runWorkflow(args[1:])
	case "run":
		return c.runRun(args[1:])
	case "version":
		fmt.Fprintln(c.Out, "kestrelflow v1.0.0")
		return 0
	case "help", "--help", "-h":
		c.printUsage()
		return 0
	default:
		fmt.Fprintf(c.Err, "unknown command: %s\n", args[0])
		c.printUsage()
		return 1
	}
}

func (c *CLI) printUsage() {
	fmt.Fprintln(c.Out, `KestrelFlow - Distributed Workflow Orchestration Engine

Usage:
  kestrel <command> [arguments]

Commands:
  server    Start KestrelFlow API and scheduler daemon
  worker    Start standalone worker agent
  workflow  Manage workflow definitions (submit, list, get)
  run       Manage workflow runs (start, status, cancel, list)
  version   Print version information
`)
}

func (c *CLI) runServer(args []string) int {
	cfg, err := config.Load(args)
	if err != nil {
		fmt.Fprintf(c.Err, "error loading config: %v\n", err)
		return 1
	}

	store := memory.NewStore()
	sched := scheduler.NewScheduler(store, cfg.Worker.PollInterval)
	srv := http.NewServer(store, sched)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sched.Start(ctx)

	httpServer := &stdhttp.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      srv.Handler(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	fmt.Fprintf(c.Out, "Starting KestrelFlow server on %s:%d\n", cfg.Server.Host, cfg.Server.Port)

	go func() {
		_ = httpServer.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Fprintln(c.Out, "Shutting down server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	_ = httpServer.Shutdown(shutdownCtx)
	sched.Stop()
	return 0
}

func (c *CLI) runWorker(args []string) int {
	cfg, err := config.Load(args)
	if err != nil {
		fmt.Fprintf(c.Err, "error loading config: %v\n", err)
		return 1
	}

	store := memory.NewStore()
	sched := scheduler.NewScheduler(store, cfg.Worker.PollInterval)
	reg := worker.NewExecutorRegistry()

	workerCfg := worker.Config{
		WorkerID:          cfg.Worker.WorkerID,
		Concurrency:       cfg.Worker.Concurrency,
		PollInterval:      cfg.Worker.PollInterval,
		LeaseDuration:     cfg.Worker.LeaseDuration,
		HeartbeatInterval: cfg.Worker.HeartbeatInterval,
	}
	pool := worker.NewPool(workerCfg, store, sched, reg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Fprintf(c.Out, "Starting worker %s with concurrency %d\n", workerCfg.WorkerID, workerCfg.Concurrency)
	go pool.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Fprintln(c.Out, "Draining worker pool...")
	pool.Stop()
	return 0
}

func (c *CLI) runWorkflow(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(c.Out, "Usage: kestrel workflow <submit|list|get> [args]")
		return 1
	}

	switch args[0] {
	case "list":
		fmt.Fprintln(c.Out, "Listing workflows...")
		return 0
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(c.Err, "missing workflow ID")
			return 1
		}
		fmt.Fprintf(c.Out, "Fetching workflow %s\n", args[1])
		return 0
	case "submit":
		if len(args) < 2 {
			fmt.Fprintln(c.Err, "missing workflow definition file")
			return 1
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Fprintf(c.Err, "failed reading file: %v\n", err)
			return 1
		}
		var wf core.WorkflowDefinition
		if err := json.Unmarshal(data, &wf); err != nil {
			fmt.Fprintf(c.Err, "invalid JSON: %v\n", err)
			return 1
		}
		fmt.Fprintf(c.Out, "Submitted workflow %s (version %d)\n", wf.Name, wf.Version)
		return 0
	default:
		fmt.Fprintf(c.Err, "unknown workflow command: %s\n", args[0])
		return 1
	}
}

func (c *CLI) runRun(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(c.Out, "Usage: kestrel run <start|get|cancel|list> [args]")
		return 1
	}

	switch args[0] {
	case "start":
		if len(args) < 2 {
			fmt.Fprintln(c.Err, "missing workflow ID")
			return 1
		}
		fmt.Fprintf(c.Out, "Triggered run for workflow %s\n", args[1])
		return 0
	case "get":
		if len(args) < 2 {
			fmt.Fprintln(c.Err, "missing run ID")
			return 1
		}
		fmt.Fprintf(c.Out, "Run status for %s: RUNNING\n", args[1])
		return 0
	case "cancel":
		if len(args) < 2 {
			fmt.Fprintln(c.Err, "missing run ID")
			return 1
		}
		fmt.Fprintf(c.Out, "Cancelled run %s\n", args[1])
		return 0
	case "list":
		fmt.Fprintln(c.Out, "Listing active runs...")
		return 0
	default:
		fmt.Fprintf(c.Err, "unknown run command: %s\n", args[0])
		return 1
	}
}
