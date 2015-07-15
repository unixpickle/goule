package main

import (
	"log"
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
		log.Fatal("Usage: " + os.Args[0] + " <port> <config.json>")
	}
	port, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid port number: " + os.Args[1])
	}
	ConfigPath = os.Args[2]

	// Load the configuration.
	GlobalConfig, err = LoadConfig(ConfigPath)
	if err != nil {
		log.Fatal("Failed to load configuration: " + err.Error())
	}

	// Run the tasks before we start the servers so the configuration page isn't accessible until
	// the tasks are started.
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
		log.Fatal("Failed to start server: " + err.Error())
		shutdown()
	}

	// Wait for SIGTERM.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Print("Goule shutting down...")
	shutdown()
}

func shutdown() {
	GlobalConfig.Lock()
	for _, t := range GlobalConfig.Tasks {
		t.StopLoop()
	}
	if GlobalServer != nil {
		GlobalServer.Control.Stop()
		GlobalServer.HTTP.Stop()
		GlobalServer.HTTPS.Stop()
	}
	os.Exit(1)
}
