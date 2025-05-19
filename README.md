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
- `GET /api/devices`: Retrieve all devices.
- `GET /api/devices/{id}`: Retrieve a specific device by ID.

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