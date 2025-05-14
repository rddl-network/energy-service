package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

func main() {
	// Define CLI flags
	protocol := flag.String("protocol", "http", "Protocol to use (http or https)")
	host := flag.String("host", "localhost", "Hostname or IP address of the server")
	port := flag.String("port", "8080", "Port of the server")
	zigbeeID := flag.String("zigbee_id", "", "Zigbee ID to include in the JSON payload")
	flag.Parse()

	// Validate required flags
	if *zigbeeID == "" {
		fmt.Println("Error: zigbee_id is required")
		flag.Usage()
		os.Exit(1)
	}

	// Construct the URL
	url := fmt.Sprintf("%s://%s:%s/api/energy", *protocol, *host, *port)

	// Create the JSON payload
	payload := map[string]interface{}{
		"version":   1,
		"zigbee_id": *zigbeeID,
		"date":      "2025-05-14", // Define the date
		"data":      [96]int{},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Send the HTTP POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response from server: %s\n", string(body))
}
