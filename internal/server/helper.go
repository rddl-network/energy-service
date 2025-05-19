package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/rddl-network/logger-service/internal/model"
)

func (s *Server) writeJSON2File(data model.EnergyData) {
	// Store data in a JSON file (append as JSON Lines)
	s.energyDataFileMutex.Lock()
	f, err := os.OpenFile(cfg.Server.DataFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open JSON file: %v", err)
	} else {
		enc := json.NewEncoder(f)
		if err := enc.Encode(data); err != nil {
			log.Printf("Failed to write energy data to JSON file: %v", err)
		}
		f.Close()
	}
	s.energyDataFileMutex.Unlock()
}

func (s *Server) write2InfluxDB(data model.EnergyData) error {
	writeAPI := s.influxWriteAPI
	if writeAPI == nil {
		log.Printf("No InfluxDB write API set")
		return nil
	}

	for i := 0; i < 96; i++ {
		hour, minutes := s.utils.Index2Time(i)
		ts, err := s.utils.CreateTimestamp(data.Date, hour, minutes)
		if err != nil {
			log.Printf("Error creating timestamp: %v", err)
			continue
		}
		err = writeAPI.WritePoint(
			context.Background(),
			"energy_data",
			map[string]string{"zigbee_id": data.ZigbeeID},
			map[string]interface{}{"overall_kwh": data.Data[i]},
			ts,
		)
		if err != nil {
			log.Printf("Failed to write to InfluxDB: %v", err)
			return err
		}
	}
	return nil
}

// sendJSONResponse sends a JSON response with the given status code
func sendJSONResponse(w http.ResponseWriter, resp Response, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Printf("Failed to encode devices %v", err.Error())
	}
}

// CreateTemplates creates templates and static directories
func CreateTemplates() error {
	// Create templates directory
	if err := os.MkdirAll("templates", 0755); err != nil {
		return err
	}

	// Create static directory
	if err := os.MkdirAll("static", 0755); err != nil {
		return err
	}

	// Write HTML template file
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Zigbee ID Registration</title>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
</head>
<body class="bg-gray-100">
    <div class="container mx-auto p-6 max-w-4xl">
        <h1 class="text-2xl font-bold mb-6">Zigbee ID Registration</h1>
        
        <div class="bg-white shadow-md rounded-lg p-6 mb-6">
            <form id="registration-form" class="space-y-4">
                <div>
                    <label class="block text-gray-700 font-medium mb-2" for="zigbee_id">
                        Zigbee ID*
                    </label>
                    <input
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        id="zigbee_id"
                        name="zigbee_id"
                        type="text"
                        placeholder="Enter Zigbee ID"
                        required
                    >
                    <p class="text-sm text-gray-500 mt-1">Enter a valid Zigbee ID.</p>
                </div>
                
                <div>
                    <label class="block text-gray-700 font-medium mb-2" for="liquid_address">
                        Liquid Address*
                    </label>
                    <input
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        id="liquid_address"
                        name="liquid_address"
                        type="text"
                        placeholder="Enter liquid address"
                        required
                    >
                </div>
                
                <div>
                    <label class="block text-gray-700 font-medium mb-2" for="device_name">
                        Device Name*
                    </label>
                    <input
                        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                        id="device_name"
                        name="device_name"
                        type="text"
                        placeholder="Enter device name"
                        required
                    >
                </div>
                
                <div class="flex justify-between">
                    <button
                        type="submit"
                        class="bg-blue-600 text-white py-2 px-4 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
                    >
                        Register Device
                    </button>
                    
                    <button
                        type="button"
                        id="toggle-db-btn"
                        class="bg-gray-200 text-gray-800 py-2 px-4 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-400"
                    >
                        View Database
                    </button>
                </div>
            </form>
            
            <div id="alert" class="mt-4 p-3 border rounded hidden"></div>
        </div>
        
        <div id="database-container" class="hidden">
            <h2 class="text-xl font-semibold mb-4">Database Entries</h2>
            <div class="overflow-x-auto">
                <table class="w-full border-collapse table-auto">
                    <thead>
                        <tr class="bg-gray-100">
                            <th class="border px-4 py-2">Zigbee ID</th>
                            <th class="border px-4 py-2">Liquid Address</th>
                            <th class="border px-4 py-2">Device Name</th>
                            <th class="border px-4 py-2">Timestamp</th>
                        </tr>
                    </thead>
                    <tbody id="devices-table">
                        <!-- Devices will be added here dynamically -->
                    </tbody>
                </table>
            </div>
        </div>
    </div>

    <script>
        document.addEventListener('DOMContentLoaded', function() {
            const form = document.getElementById('registration-form');
            const alertBox = document.getElementById('alert');
            const toggleDbBtn = document.getElementById('toggle-db-btn');
            const dbContainer = document.getElementById('database-container');
            const devicesTable = document.getElementById('devices-table');
            
            // Toggle database visibility
            toggleDbBtn.addEventListener('click', function() {
                const isHidden = dbContainer.classList.contains('hidden');
                
                if (isHidden) {
                    fetchDevices();
                    dbContainer.classList.remove('hidden');
                    toggleDbBtn.textContent = 'Hide Database';
                } else {
                    dbContainer.classList.add('hidden');
                    toggleDbBtn.textContent = 'View Database';
                }
            });
            
            // Submit form
            form.addEventListener('submit', function(e) {
                e.preventDefault();
                
                const formData = new FormData(form);
                
                fetch('/register', {
                    method: 'POST',
                    body: formData
                })
                .then(response => response.json())
                .then(data => {
                    if (data.error) {
                        showAlert(data.error, 'error');
                    } else {
                        showAlert(data.message, 'success');
                        form.reset();
                        
                        // Update database table if visible
                        if (!dbContainer.classList.contains('hidden')) {
                            fetchDevices();
                        }
                    }
                })
                .catch(error => {
                    showAlert('An error occurred. Please try again.', 'error');
                });
            });
            
            function showAlert(message, type) {
                alertBox.textContent = message;
                alertBox.classList.remove('hidden', 'bg-red-100', 'border-red-400', 'text-red-700', 'bg-green-100', 'border-green-400', 'text-green-700');
                
                if (type === 'error') {
                    alertBox.classList.add('bg-red-100', 'border-red-400', 'text-red-700');
                } else {
                    alertBox.classList.add('bg-green-100', 'border-green-400', 'text-green-700');
                }
                
                // Hide alert after 5 seconds
                setTimeout(() => {
                    alertBox.classList.add('hidden');
                }, 5000);
            }
            
            function fetchDevices() {
                fetch('/api/devices')
                .then(response => response.json())
                .then(devices => {
                    devicesTable.innerHTML = '';
                    
                    if (Object.keys(devices).length === 0) {
                        const emptyRow = document.createElement('tr');
                        emptyRow.innerHTML = '<td colspan="4" class="border px-4 py-2 text-center text-gray-500">No devices registered yet.</td>';
                        devicesTable.appendChild(emptyRow);
                        return;
                    }
                    
                    for (const [zigbee_id, data] of Object.entries(devices)) {
                        const row = document.createElement('tr');
                        
                        row.innerHTML = '<td class="border px-4 py-2 font-mono">' + zigbee_id + '</td>' +
                            '<td class="border px-4 py-2">' + data.liquid_address + '</td>' +
                            '<td class="border px-4 py-2">' + data.device_name + '</td>' +
                            '<td class="border px-4 py-2 text-sm">' + new Date(data.timestamp).toLocaleString() + '</td>';
                        
                        devicesTable.appendChild(row);
                    }
                })
                .catch(error => {
                    console.error('Error fetching devices:', error);
                });
            }
        });
    </script>
</body>
</html>`

	return os.WriteFile("templates/index.html", []byte(indexHTML), 0644)
}
