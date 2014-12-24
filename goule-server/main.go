package main

import (
	"fmt"
	"github.com/unixpickle/goule"
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
	config, err := goule.ReadConfig(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read configuration:", err)
		os.Exit(1)
	}
	instance := goule.NewGoule(config)
	if err := instance.Start(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to start:", err)
		os.Exit(1)
	}
	fmt.Println("Goule running.")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Goule shutting down...")
	instance.Stop()
}