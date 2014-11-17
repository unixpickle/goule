package main

import (
	"../src"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: goule <config.json>")
		os.Exit(1)
	}
	config, err := goule.ReadConfiguration(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read configuration:", err)
		os.Exit(1)
	}
	router := goule.NewRouter(config)
	if err := router.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to run router:", err)
	}
	fmt.Println("Goule running.")
	select {}
}
