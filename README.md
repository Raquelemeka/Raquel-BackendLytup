# Raquel-BackendLytup

## Overview
This is my backend solution for a server system that supports both a basic service response and a weather service.

The main components are:
- **Daemon Core** – Process manager that starts and stops services
- **HTTP Service** – REST API handler on port 8080
- **Weather Service** – Handles weather-related requests using seeded input data
- **Logging System** – Structured logging with timestamps
- **Signal Handler** – Graceful shutdown on Ctrl+C

## Program Structure

```text
main.go
├── Daemon Core
│   ├── Daemon struct definition
│   ├── RunDaemonCore() - Main manager
│   └── Output_Logg() - Logging system
├── HTTP Service
│   ├── StartSR05Service() - Server setup
│   ├── handler() - Request processor
│   └── Request/Response structs
├── Weather Service
│   ├── Weather request handling
│   └── Seeded input processing (lat, lon, date range)
└── Main Function
    └── Service registration and startup
```

## Supported Services

- **sr05**
- **weather** – Accepts location and date seed data and returns a weather-related response

## To Test

### 1. Build the project
```bash
go build -o sr05-service
```

### 2. Run it
```bash
./sr05-service
```

### 3. In another terminal, test it

#### Base Service (Success Case)
```bash
curl -X POST http://localhost:8080/service \
  -H "Content-Type: application/json" \
  -d '{"Srvc":"sr05"}'
```

Expected output:
```json
{
  "ExctnOutcomeCode": 200,
  "ExctnOutcomeNote": "",
  "Yield": "Hello world"
}
```

#### Weather Service (Example Request)
```bash
curl -X POST http://localhost:8080/service \
  -H "Content-Type: application/json" \
  -d '{
    "Srvc": "weather",
    "Seed": {
      "lat": 40,
      "lon": -74,
      "start": "2024-01-01",
      "end": "2024-01-07"
    }
  }'
```

Expected output:
- A successful execution response containing weather-related data based on the provided seed input.

### 4. Example Wrong Input
```bash
curl -X POST http://localhost:8080/service \
  -H "Content-Type: application/json" \
  -d '{"Srvc":"sr05"'
```

Expected output:
```
invalid json
```

### 5. Wrong Service Code Error
```json
{
  "ExctnOutcomeCode": 400,
  "ExctnOutcomeNote": "unknown service code",
  "Yield": ""
}
```
