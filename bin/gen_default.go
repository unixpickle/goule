package main

import (
	"../src"
	"encoding/json"
	"bytes"
	"os"
)

func main() {
	config := goule.NewConfiguration()
	adminRule := goule.SourceURL{Scheme: "http", Hostname: "localhost",
		Path: "/admin"}
	config.Admin.Rules = append(config.Admin.Rules, adminRule)
	// Default password: "password"
	config.Admin.PasswordHash = 
		"5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"
	config.HTTPSettings.Enabled = true
	config.HTTPSettings.Port = 1337
	x, _ := json.Marshal(*config)
	var out bytes.Buffer
	json.Indent(&out, x, "", "  ")
	out.WriteTo(os.Stdout)
}
