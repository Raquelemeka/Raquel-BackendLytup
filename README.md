# Raquel-BackendLytup

## Overview
This is my backend assignment solution for a server system.

 The main components are:
   - Daemon Core - Process manager that starts/stops services
   - HTTP Service - REST API handler on port 8080
   - Logging System - Structured logging with timestamps
   - Signal Handler - Graceful shutdown on Ctrl+C

## Program Structure
main.go
├── Daemon Core
│   ├── Daemon struct definition
│   ├── RunDaemonCore() - Main manager
│   └── Output_Logg() - Logging system
├── HTTP Service
│   ├── StartSR05Service() - Server setup
│   ├── handler() - Request processor
│   └── Request/Response structs
└── Main Function
    └── Service registration and startup
 
## To test,
### 1. Build the project
go build -o sr05-service

### 2. Run it
./sr05-service

### 3. In another terminal, test it
Correct Input (Success Case)
curl -X POST http://localhost:8080/service \
  -H "Content-Type: application/json" \
  -d '{"Srvc":"sr05"}'

The correct output should be:
{
    "ExctnOutcomeCode": 200,
    "ExctnOutcomeNote": "",
    "Yield": "Hello world"
}

### 4. (Optional) Example wrong input
curl -X POST http://localhost:8080/service \
  -H "Content-Type: application/json" \
  -d '{"Srvc":"sr05"'

  The correct output should be:
  invalid json

  Other error messages you could encouter would be the Wrong Service Code Error:
  
  {
    "ExctnOutcomeCode": 400,
    "ExctnOutcomeNote": "unknown service code",
    "Yield": ""
}



