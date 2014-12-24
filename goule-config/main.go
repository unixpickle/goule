package main

import (
	"fmt"
	"github.com/unixpickle/goule"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: goule-config <output.json>")
		os.Exit(1)
	}
	config := goule.NewConfig()
	config.Path = os.Args[1]
	config.Admin.Hash = readPasswordHash()
	config.Admin.Port = readPort()
	config.Save()
}

func readPort() int {
	for {
		fmt.Print("    Enter admin port: ")
		var port int
		if _, err := fmt.Scanln(&port); err != nil {
			fmt.Println("Invalid port number. Try again.")
			continue
		} else {
			return port
		}
	}
}

func readPasswordHash() string {
	fmt.Print("Enter admin password: ")
	setTTYEcho(false)
	var password string
	fmt.Scanln(&password)
	setTTYEcho(true)
	fmt.Println("")
	return goule.Hash(password)
}

func setTTYEcho(enabled bool) {
	stty, err := exec.LookPath("stty")
	if err != nil {
		return
	}
	arg := "echo"
	if !enabled {
		arg = "-echo"
	}
	cmd := exec.Command(stty, arg)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Run()
}
