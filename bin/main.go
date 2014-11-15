package main

import (
	"../src"
	"os"
	"fmt"
	"net/http"
	"encoding/json"
	"io/ioutil"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: goule <config.json>")
		os.Exit(1)
	}
	configData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Failed to read", os.Args[1])
		os.Exit(1)
	}
	config := goule.MakeConfiguration()
	if err := json.Unmarshal(configData, config); err != nil {
		fmt.Println("Failed to read JSON:", err)
		os.Exit(1)
	}
	http.HandleFunc("/", goule.CreateHandler(config, false))
	err = http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("failed to listen :(")
	}
}
