package config

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"time"
)

type ReloadCallback func(newCfg *Config)

type ConfigWatcher struct {
	filePath  string
	interval  time.Duration
	lastHash  string
	callback  ReloadCallback
	mu        sync.RWMutex
	stopCh    chan struct{}
	running   bool
}

func NewConfigWatcher(filePath string, interval time.Duration, cb ReloadCallback) *ConfigWatcher {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	return &ConfigWatcher{
		filePath: filePath,
		interval: interval,
		callback: cb,
		stopCh:   make(chan struct{}),
	}
}

func (w *ConfigWatcher) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("watcher already running")
	}
	w.running = true
	w.mu.Unlock()

	hash, err := w.computeHash()
	if err == nil {
		w.lastHash = hash
	}

	go w.watchLoop(ctx)
	return nil
}

func (w *ConfigWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.running {
		w.running = false
		close(w.stopCh)
	}
}

func (w *ConfigWatcher) computeHash() (string, error) {
	data, err := os.ReadFile(w.filePath)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}

func (w *ConfigWatcher) CheckOnce() (bool, error) {
	hash, err := w.computeHash()
	if err != nil {
		return false, err
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if hash != w.lastHash {
		w.lastHash = hash
		cfg, err := LoadConfigFile(w.filePath)
		if err != nil {
			return false, fmt.Errorf("failed reloading config: %w", err)
		}
		if w.callback != nil {
			w.callback(cfg)
		}
		return true, nil
	}
	return false, nil
}

func (w *ConfigWatcher) watchLoop(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case <-ticker.C:
			_, _ = w.CheckOnce()
		}
	}
}
