package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

var ConfigPath string
var GlobalConfig *Config
var GlobalServer *Server

func main() {
	// Deal with the arguments.
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<port> <config.json>")
		os.Exit(1)
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Invalid port number:", os.Args[1])
		os.Exit(1)
	}
	ConfigPath = os.Args[2]

	// Load the configuration.
	GlobalConfig, err = LoadConfig(ConfigPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to load configuration:", err)
		os.Exit(1)
	}

	// Run the tasks before we start the servers so the configuration page isn't
	// accessible until the tasks are started.
	GlobalConfig.Lock()
	for _, t := range GlobalConfig.Tasks {
		t.StartLoop()
		if t.AutoRun {
			t.Start()
		}
	}
	GlobalConfig.Unlock()

	// Start the servers.
	GlobalServer, err = NewServer(GlobalConfig, port)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to start server:", err)
		shutdown()
	}

	// Wait for SIGTERM.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Goule shutting down...")
	shutdown()
}

func shutdown() {
	GlobalConfig.Lock()
	for _, t := range GlobalConfig.Tasks {
		t.StopLoop()
	}
	GlobalServer.Control.Stop()
	GlobalServer.HTTP.Stop()
	GlobalServer.HTTPS.Stop()
	os.Exit(1)
}
