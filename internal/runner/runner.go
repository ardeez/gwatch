package runner

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/ardeez/gwatch/internal/config"
	"github.com/ardeez/gwatch/internal/logger"
)

type Runner struct {
	config     *config.Config
	binaryPath string
	activeCmd  *exec.Cmd
	mu         sync.Mutex
}

func New(config *config.Config) *Runner {
	return &Runner{config: config, binaryPath: filepath.Join(".", "tmp", "app")}
}

func (r *Runner) Build() error {
	logger.Info("Compiling: go build  -o %s %s", r.binaryPath, r.config.Entry)

	err := os.MkdirAll(filepath.Dir(r.binaryPath), 0755)
	if err != nil {
		logger.Error("Failed to create tmp directory: %v", err)
		return err
	}
	cmd := exec.Command("go", "build", "-o", r.binaryPath, r.config.Entry)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s", stderr.String())
	}
	logger.Info("Build successful!")

	return nil
}

func (r *Runner) Run() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.activeCmd != nil && r.activeCmd.Process != nil {
		logger.Warn("Stopping current process (PID: %d)...", r.activeCmd.Process.Pid)
		_ = r.activeCmd.Process.Kill()
		_ = r.activeCmd.Wait()
	}
	cmd := exec.Command(r.binaryPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	logger.Info("Starting binary: %s", r.binaryPath)
	if err := cmd.Start(); err != nil {
		logger.Error("Failed to start binary: %v", err)
		return
	}

	logger.Info("Application is running (PID: %d)", cmd.Process.Pid)
	r.activeCmd = cmd

	go func(c *exec.Cmd) {
		_ = c.Wait()
	}(cmd)

}


func (r *Runner) StartListening(rebuildChan <-chan struct{}) {
	for range rebuildChan {
		logger.Warn("Rebuild signal received. Triggering build lifecycle...")
		if err := r.Build(); err != nil {
			logger.Error("Build Failed!\n%s", err.Error())
			logger.Info("Keeping the previous process running. Fix the typo and save again.")
			continue
		}
		r.Run()
	}
}