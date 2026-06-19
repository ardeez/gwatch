package watcher

import (
	"io/fs"
	"path/filepath"
	"time"

	"github.com/ardeez/gwatch/internal/config"
	"github.com/ardeez/gwatch/internal/logger"
)

type Watcher struct {
	cfg      *config.Config
	snapshot map[string]time.Time
}

func New(cfg *config.Config) *Watcher {
	return &Watcher{
		cfg:      cfg,
		snapshot: make(map[string]time.Time),
	}
}

func (w *Watcher) CreateSnapshot() error {
	newSnapshot := make(map[string]time.Time)
	err := filepath.WalkDir(w.cfg.Dir, func(path string, d fs.DirEntry, err error) error {
		if filepath.Ext(path) == w.cfg.Ext {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			newSnapshot[path] = info.ModTime()
		}
		return nil
	})

	if err != nil {
		return err
	}

	w.snapshot = newSnapshot
	return nil
}

func (w *Watcher) StartPolling(onChange func()) {
	ticker := time.NewTicker(time.Duration(w.cfg.Interval) * time.Millisecond)
	defer ticker.Stop()
	logger.Info("Watcher loop started. Polling every %dms...", w.cfg.Interval)

	for range ticker.C {
		func() {
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Watcher recovered from panic: %v. Restarting loop...", r)
				}
			}()

			w.checkChanges(onChange)
		}()
	}
}

func (w *Watcher) checkChanges(onChange func()) {
	hasChange := false
	err := filepath.WalkDir(w.cfg.Dir, func(path string, d fs.DirEntry, err error) error {

		if filepath.Ext(path) == w.cfg.Ext {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			modTime := info.ModTime()

			oldModTime, exists := w.snapshot[path]
			if !exists || modTime.After(oldModTime) {
				logger.Info("File change detected: %s", path)
				hasChange = true
			}
		}
		return nil
	})
	if err != nil {
		logger.Error("Error during filesystem polling: %v", err)
		return
	}

	if hasChange {
		_ = w.CreateSnapshot()
		onChange()
	}
}
