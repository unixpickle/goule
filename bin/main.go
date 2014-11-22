package main

import (
	"../src"
	"fmt"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: goule <config.json>")
		os.Exit(1)
	}
	config := goule.NewConfiguration()
	if err := config.Read(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read configuration:", err)
		os.Exit(1)
	}
	overseer := goule.NewOverseer()
	overseer.Update(config)
	if !overseer.IsRunning() {
		fmt.Fprintln(os.Stderr, "No webservers are running!")
		os.Exit(1)
	}
	fmt.Println("Goule running.")
	select {}
}
