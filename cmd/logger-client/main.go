package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	var dataStr string

	utcTime := time.Now().UTC()
	// Format as YYYY-MM-DD (same format you're using in your JSON)
	currentDate := utcTime.Format("2006-01-02")
	defaultData := "[]"
	// Define CLI flags
	protocol := flag.String("protocol", "http", "Protocol to use (http or https)")
	host := flag.String("host", "localhost", "Hostname or IP address of the server")
	port := flag.String("port", "8080", "Port of the server")
	zigbeeID := flag.String("zigbee_id", "", "Zigbee ID to include in the JSON payload")
	production := flag.Bool("production", false, "Use for production purposes")
	date := flag.String("date", currentDate, "Date in YYYY-MM-DD format")
	flag.StringVar(&dataStr, "data", defaultData, "96 float value to be sent in the JSON payload")

	flag.Parse()

	// Validate required flags
	if *zigbeeID == "" {
		fmt.Println("Error: zigbee_id is required")
		flag.Usage()
		os.Exit(1)
	}

	strValues := strings.Fields(dataStr)
	// Create a slice to hold the float values
	dataSlice := make([]float64, 0, len(strValues))

	// Convert each string to float64
	for _, strVal := range strValues {
		floatVal, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", strVal, err)
			continue
		}
		dataSlice = append(dataSlice, floatVal)
	}
	// var dataSlice []float64
	// err := json.Unmarshal([]byte(dataStr), &dataSlice)
	// if err != nil {
	// 	log.Fatalf("Failed to parse data array: %v", err)
	// }

	generateRandomData := false
	if !*production {
		if dataStr == defaultData {
			generateRandomData = true
		} else if len(dataSlice) != 96 {
			log.Fatalf("Expected 96 values, got %d", len(dataSlice))
		}
	} else {
		// Ensure we have exactly 96 values
		if len(dataSlice) != 96 {
			log.Fatalf("Expected 96 values, got %d", len(dataSlice))
		}
	}

	// Convert slice to array
	var dataArray [96]float64
	for i := 0; i < 96; i++ {
		if generateRandomData {
			dataArray[i] = rand.Float64()
		} else {
			dataArray[i] = dataSlice[i]
		}
	}

	// Construct the URL
	url := fmt.Sprintf("%s://%s:%s/api/energy", *protocol, *host, *port)

	// Create the JSON payload
	payload := map[string]interface{}{
		"version":   1,
		"zigbee_id": *zigbeeID,
		"date":      *date, // Define the date
		"data":      [96]float64{},
	}

	// Reassign the modified array back to the payload
	payload["data"] = dataArray

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
