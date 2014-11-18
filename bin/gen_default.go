package main

import (
	"../src"
	"encoding/json"
	"bytes"
	"os"
)

func main() {
	config := goule.NewConfiguration()
	x, _ := json.Marshal(*config)
	var out bytes.Buffer
	json.Indent(&out, x, "", "  ")
	out.WriteTo(os.Stdout)
}
