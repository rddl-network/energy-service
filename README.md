# Energy Service and Client

## Energy Client

### Overview
The `energy-client` is a command-line tool designed to send JSON payloads containing energy data to a server. It supports configurable options such as protocol, host, port, Zigbee ID, and date. The client validates input data and sends HTTP POST requests to the server.

### Parameters
The `energy-client` accepts the following parameters:

- `--protocol` (default: `http`): Protocol to use (http or https).
- `--host` (default: `localhost`): Hostname or IP address of the server.
- `--port` (default: `8080`): Port of the server.
- `--zigbee_id` (required): Zigbee ID to include in the JSON payload.
- `--production` (default: `false`): Use for production purposes. Ensures exactly 96 data values are provided.
- `--date` (default: current date in `YYYY-MM-DD` format): Date to include in the JSON payload.
- `--data` (default: `[]`): 96 float values to be sent in the JSON payload. If not provided, random data will be generated in non-production mode.

### Example Usage
```bash
./energy-client --protocol http --host localhost --port 8080 --zigbee_id 12345 --date 2025-05-15 --data "1.0 2.0 3.0 ..."
```

### Development
To build the client, run:
```bash
go build -o energy-client ./cmd/energy-client
```

---

## Energy Service

### Overview
The `energy-service` is a server application that handles device registration and data storage. It provides RESTful APIs for registering devices, retrieving device information, and managing data. The service includes a web interface for interacting with the database and supports template rendering.

### Features
- RESTful API for device registration and data retrieval.
- Web interface for database management.
- Template rendering for dynamic content.
- SQLite database for persistent storage.

### API Endpoints
- `POST /register`: Register a new device.
- `GET /api/devices`: Retrieve all registered devices as a JSON array.
- `POST /api/energy`: Upload energy data (JSON payload, see client for format).
- `GET /api/energy/download?pwd=YOUR_PASSWORD`: Download all stored energy data as a JSON array. **Password-protected.**

#### /register
- **Method:** POST
- **Request Body:** JSON object with device registration details. Example fields:
  - `zigbee_id` (string, required): Unique Zigbee ID for the device
  - `device_name` (string, required): Human-readable name
  - `device_type` (string, required): Type/category of the device
  - `liquid_address` (string, required): Liquid address for the device
  - `planetmint_address` (string, required): Planetmint address for the device
- **Response:**
  - On success: `{ "message": "Device registered successfully" }`
  - On error: `{ "error": "..." }` with appropriate HTTP status code (e.g., 400 for validation errors, 409 for duplicate Zigbee ID)

**Example:**
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "zigbee_id": "12345",
    "device_name": "Living Room Plug",
    "device_type": "Plug",
    "liquid_address": "liq1...",
    "planetmint_address": "pm1..."
  }'
```

#### /api/devices
- **Method:** GET
- **Response:**
  - On success: Returns a JSON array of all registered devices, each with their properties (e.g., `zigbee_id`, `device_name`, `device_type`, etc.).
  - If no devices are registered: Returns `[]` (empty array).

**Example:**
```bash
curl http://localhost:8080/api/devices
```

#### /api/energy
- **Method:** POST
- **Request Body:** JSON object with the following fields:
  - `version` (int, required): Version of the payload format
  - `zigbee_id` (string, required): Unique Zigbee ID for the device
  - `date` (string, required): Date for the energy data (YYYY-MM-DD)
  - `timezone_name` (string, required): Name of the timezone (e.g., "Europe/Vienna")
  - `data` (array of 96 objects, required): Each object is:
    - `value` (float): The energy value
    - `timestamp` (string): UTC timestamp in the format `YYYY-MM-DD HH:MM:SS`
- **Response:**
  - On success: `{ "message": "Energy data received and written to database successfully" }`
  - On error: `{ "error": "..." }` with appropriate HTTP status code (e.g., 400 for validation errors, 409 for duplicate, 500 for server error)

**Example:**
```json
{
  "version": 1,
  "zigbee_id": "12345",
  "date": "2025-06-04",
  "timezone_name": "Europe/Vienna",
  "data": [
    { "value": 1.23, "timestamp": "2025-06-04 00:00:00" },
    { "value": 1.24, "timestamp": "2025-06-04 00:15:00" },
    ... (total 96 entries) ...
  ]
}
```

**Note:** The `data` array must contain exactly 96 entries, each with a value and a UTC timestamp string in the specified format.

#### /api/energy/download
- **Method:** GET
- **Query Parameter:** `pwd` (required, must match the configured server password)
- **Response:**
  - On success: Returns a JSON array of all uploaded energy data entries (may be empty if no data).
  - If the file is empty: Returns `[]` (empty array).
  - If the file is corrupted or contains invalid JSON: Returns HTTP 500 with an error message.
  - If the password is missing or incorrect: Returns HTTP 401 Unauthorized.

**Example:**
```bash
curl "http://localhost:8080/api/energy/download?pwd=YOUR_PASSWORD"
```

**Note:** The download endpoint streams all valid JSON entries from the server's data file. Each entry matches the format uploaded via `/api/energy`.

### Usage
Run the `energy-service` with the following command:
```bash
./energy-service
```
The server will start on `http://localhost:8080` by default.

### Development
To build the service, run:
```bash
go build -o energy-service ./cmd/energy-service
```