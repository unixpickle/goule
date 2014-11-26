package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/unixpickle/goule/pkg/config"
	"io/ioutil"
	"os"
	"os/exec"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: goule-config <output_config.json>")
		os.Exit(1)
	}

	hostname := readHostname()
	port := readPort()
	hash := readPasswordHash()

	cfg := config.NewConfig()
	adminRule := config.SourceURL{Scheme: "http", Hostname: hostname,
		Path: "/admin"}
	cfg.Admin.Rules = append(cfg.Admin.Rules, adminRule)
	// Default password: "password"
	cfg.Admin.PasswordHash = hash
	cfg.HTTPSettings.Enabled = true
	cfg.HTTPSettings.Port = port

	saveConfiguration(cfg)
	fmt.Printf("Saved configuration. Admin accessible via http://%s:%d/admin\n",
		hostname, port)
}

func readPasswordHash() string {
	fmt.Print("New Password: ")
	setTTYEcho(false)
	var password string
	fmt.Scanln(&password)
	setTTYEcho(true)
	fmt.Println("")
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func readPort() int {
	for {
		fmt.Print("HTTP Server port: ")
		var port int
		if _, err := fmt.Scan(&port); err != nil {
			fmt.Println("Invalid port number. Try again.")
			continue
		} else {
			return port
		}
	}
}

func readHostname() string {
	fmt.Print("Enter admin hostname: ")
	var hostname string
	fmt.Scanln(&hostname)
	return hostname
}

func saveConfiguration(cfg *config.Config) {
	x, _ := json.Marshal(*cfg)
	var out bytes.Buffer
	json.Indent(&out, x, "", "  ")
	ioutil.WriteFile(os.Args[1], out.Bytes(), 0700)
}

func setTTYEcho(enabled bool) {
	stty, err := exec.LookPath("stty")
	if err != nil {
		fmt.Println("popy")
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
