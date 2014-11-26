package main

import (
	"fmt"
	"github.com/unixpickle/goule/pkg/config"
	"github.com/unixpickle/goule/pkg/handler"
	"github.com/unixpickle/goule/pkg/overseer"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: goule-server <config.json>")
		os.Exit(1)
	}
	config := config.NewConfig()
	if err := config.Read(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read configuration:", err)
		os.Exit(1)
	}
	instance := overseer.NewOverseer(config, handler.Handle)
	instance.Start()
	if !instance.IsRunning() {
		fmt.Fprintln(os.Stderr, "No webservers are running!")
		os.Exit(1)
	}
	fmt.Println("Goule running.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Goule shutting down...")
	instance.Stop()
}
