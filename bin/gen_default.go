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
	config.HTTPSettings.Enabled = true
	config.HTTPSettings.Port = 1337
	x, _ := json.Marshal(*config)
	var out bytes.Buffer
	json.Indent(&out, x, "", "  ")
	out.WriteTo(os.Stdout)
}
