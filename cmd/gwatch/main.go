package main

import (
	"flag"
	"os"

	"github.com/ardeez/gwatch/internal/config"
	"github.com/ardeez/gwatch/internal/logger"
	"github.com/ardeez/gwatch/internal/runner"
	"github.com/ardeez/gwatch/internal/watcher"
)

func main() {
	cfg, err := config.ParseConfig()
	if err != nil {
		logger.Error("Configuration Error: %v", err)
		flag.Usage()
		os.Exit(1)
	}

	logger.Info("gwatch configuration loaded successfully")
	logger.Debug("Entry Point : %s", cfg.Entry)
	logger.Debug("Watch Dir   : %s", cfg.Dir)
	logger.Debug("Extension   : %s", cfg.Ext)
	logger.Debug("Excludes    : %v", cfg.Exclude)
	logger.Debug("Interval    : %dms", cfg.Interval)
	logger.Info("Starting gwatch...")

	runEngine := runner.New(cfg)

	logger.Info("Performing initial build & run...")
	if err := runEngine.Build(); err != nil {
		logger.Error("Initial build failed:\n%s", err.Error())
		logger.Error("gwatch cannot start because the initial build is broken. Please fix it first.")
		os.Exit(1)
	}
	runEngine.Run()

	rebuildChan := make(chan struct{}, 1)

	go runEngine.StartListening(rebuildChan)

	logger.Info("Runner pipeline is ready and listening to channels.")

	fileWatcher := watcher.New(cfg)
	if err := fileWatcher.CreateSnapshot(); err != nil {
		logger.Error("Failed to initialize filesystem snapshot: %v", err)
		os.Exit(1)
	}
	fileWatcher.StartPolling(func() {
		select {
		case rebuildChan <- struct{}{}:
		default:
		}
	})

}
