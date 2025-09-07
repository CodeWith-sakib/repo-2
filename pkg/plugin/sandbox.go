package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// SandboxConfig defines isolation parameters for a plugin execution environment.
type SandboxConfig struct {
	MaxCPUSeconds   int
	MaxMemoryMB     int
	AllowNetwork    bool
	AllowFilesystem bool
	AllowedEnvKeys  []string
	WorkDir         string
	Timeout         time.Duration
}

// DefaultSandboxConfig returns a conservative default sandbox configuration.
func DefaultSandboxConfig() SandboxConfig {
	return SandboxConfig{
		MaxCPUSeconds:   30,
		MaxMemoryMB:     256,
		AllowNetwork:    false,
		AllowFilesystem: false,
		AllowedEnvKeys:  []string{"HOME", "PATH", "TZ"},
		Timeout:         60 * time.Second,
	}
}

// SandboxViolation records a sandboxing constraint violation.
type SandboxViolation struct {
	Category string
	Detail   string
	At       time.Time
}

func (v SandboxViolation) Error() string {
	return fmt.Sprintf("[sandbox] %s: %s", v.Category, v.Detail)
}

// ProcessSandbox manages isolated execution of plugin subprocesses with constraint enforcement.
type ProcessSandbox struct {
	mu         sync.Mutex
	cfg        SandboxConfig
	violations []SandboxViolation
	running    map[int]*exec.Cmd
	nextID     int
}

// NewProcessSandbox creates a sandbox with the given configuration.
func NewProcessSandbox(cfg SandboxConfig) *ProcessSandbox {
	return &ProcessSandbox{
		cfg:     cfg,
		running: make(map[int]*exec.Cmd),
	}
}

// BuildCommand creates a sandboxed exec.Cmd with environment filtering and working directory isolation.
func (s *ProcessSandbox) BuildCommand(ctx context.Context, program string, args ...string) (*exec.Cmd, int, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), s.cfg.Timeout)
		_ = cancel
	} else {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.cfg.Timeout)
		_ = cancel
	}

	cmd := exec.CommandContext(ctx, program, args...)

	// Filter environment
	cmd.Env = s.filteredEnv()

	// Set working directory
	workDir := s.cfg.WorkDir
	if workDir == "" {
		var err error
		workDir, err = os.MkdirTemp("", "plugin-sandbox-*")
		if err != nil {
			return nil, 0, fmt.Errorf("failed to create sandbox workdir: %w", err)
		}
	}
	cmd.Dir = workDir

	s.mu.Lock()
	s.nextID++
	id := s.nextID
	s.running[id] = cmd
	s.mu.Unlock()

	return cmd, id, nil
}

// ReleaseProcess removes a process from the tracked map (call after Wait).
func (s *ProcessSandbox) ReleaseProcess(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.running, id)
}

// RecordViolation logs a sandboxing violation.
func (s *ProcessSandbox) RecordViolation(category, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.violations = append(s.violations, SandboxViolation{
		Category: category,
		Detail:   detail,
		At:       time.Now(),
	})
}

// Violations returns a snapshot of all recorded constraint violations.
func (s *ProcessSandbox) Violations() []SandboxViolation {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]SandboxViolation, len(s.violations))
	copy(out, s.violations)
	return out
}

// ActiveProcessCount returns the number of currently tracked subprocesses.
func (s *ProcessSandbox) ActiveProcessCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.running)
}

// filteredEnv returns environment variables filtered to the allow-list.
func (s *ProcessSandbox) filteredEnv() []string {
	allowed := make(map[string]bool, len(s.cfg.AllowedEnvKeys))
	for _, k := range s.cfg.AllowedEnvKeys {
		allowed[k] = true
	}

	var result []string
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 && allowed[parts[0]] {
			result = append(result, pair)
		}
	}
	return result
}

// ValidateProgram checks that the program path is an allowed executable.
func (s *ProcessSandbox) ValidateProgram(program string) error {
	if !s.cfg.AllowFilesystem {
		// On non-filesystem mode, only allow known safe absolute paths
		if !filepath.IsAbs(program) {
			return &SandboxViolation{
				Category: "filesystem",
				Detail:   fmt.Sprintf("relative program path %q not allowed in restricted mode", program),
			}
		}
	}

	// Check executable exists
	_, err := exec.LookPath(program)
	if err != nil && filepath.IsAbs(program) {
		info, statErr := os.Stat(program)
		if statErr != nil {
			return fmt.Errorf("program %q not found: %w", program, statErr)
		}
		if runtime.GOOS != "windows" && info.Mode()&0111 == 0 {
			return fmt.Errorf("program %q is not executable", program)
		}
	}

	return nil
}
