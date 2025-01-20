//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
	"os/signal"
	"slate-nexus-agent/logger"
	"syscall"
)

func runAsService() {
	logger.LogInfo("Starting Slate Nexus Agent...")

	stop := make(chan struct{})

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.LogInfo("%s", fmt.Sprintf("Received signal: %v", sig))
		close(stop)
	}()

	// Run the agent
	runAgent(stop)

	logger.LogInfo("Slate Nexus Agent stopped.")
}
