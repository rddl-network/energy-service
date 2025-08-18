package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
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
	currentDate := utcTime.Format("2006-01-02")
	defaultData := ""
	// Define CLI flags
	protocol := flag.String("protocol", "http", "Protocol to use (http or https)")
	host := flag.String("host", "localhost", "Hostname or IP address of the server")
	port := flag.String("port", "8080", "Port of the server")
	id := flag.String("id", "", "ID to include in the JSON payload")
	production := flag.Bool("production", false, "Use for production purposes")
	date := flag.String("date", currentDate, "Date in YYYY-MM-DD format")
	tzName := flag.String("timezone", "", "Timezone name (e.g., Europe/Vienna). If empty, uses system timezone or UTC.")
	flag.StringVar(&dataStr, "data", defaultData, "96 float value to be sent in the JSON payload")

	flag.Parse()

	// Validate required flags
	if *id == "" {
		fmt.Println("Error: ID is required")
		flag.Usage()
		os.Exit(1)
	}

	// Determine timezone name
	tz := *tzName
	if tz == "" {
		loc, err := time.LoadLocation("")
		if err == nil && loc.String() != "Local" {
			tz = loc.String()
		} else {
			tz = "UTC"
		}
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

	// Prepare data array of objects with value and timestamp
	var dataArray [96]map[string]interface{}
	baseTime, _ := time.ParseInLocation("2006-01-02", *date, time.UTC)
	for i := 0; i < 96; i++ {
		var val float64
		if generateRandomData {
			if i == 0 {
				val = rand.Float64()
			} else {
				val = dataArray[i-1]["value"].(float64) + rand.Float64()
			}
		} else {
			val = dataSlice[i]
		}
		// Each 15 minutes
		ts := baseTime.Add(time.Duration(i*15) * time.Minute).UTC().Format("2006-01-02 15:04:05")
		dataArray[i] = map[string]interface{}{
			"value":     val,
			"timestamp": ts,
		}
	}

	// Construct the URL
	url := fmt.Sprintf("%s://%s:%s/api/energy", *protocol, *host, *port)

	// Create the JSON payload
	payload := map[string]interface{}{
		"version":       1,
		"id":            *id,
		"date":          *date,
		"timezone_name": tz,
		"data":          dataArray,
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
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("Error closing response body: %v\n", cerr)
		}
	}()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Print the response
	fmt.Printf("Response from server: %s\n", string(body))
}
