package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"sync"
)

// DAEMON CORE 
type Daemon struct {
	Name            string
	Program         func() error
	State           uint64 // 0 - not started ; 1 - running; 2 - done
	StartupGrace    time.Duration
	ShutdownGrace   time.Duration
	mu              sync.Mutex
}


var DaemonRegister []*Daemon


//prints messages with timestamps.
func Output_Logg(Type, Source, Output string) {
	Type = strings.ToLower(Type)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logMsg := fmt.Sprintf("[%s//%s] %s\n", timestamp, Source, Output)
	
	if Type == "out" {
		os.Stdout.Write([]byte(logMsg))
	} else {
		os.Stderr.Write([]byte(logMsg))
	}
}

// First log message 
func RunDaemonCore() {
	Output_Logg("OUT", "Main", "PROJECT: Starting up")
	
	if len(DaemonRegister) == 0 {
		Output_Logg("OUT", "Main", "PROJECT: No Daemon(s) to run. Shutting down now")
		return
	}
	
	// Start all daemons
	for _, daemon := range DaemonRegister {
		Output_Logg("OUT", "Main", fmt.Sprintf("PROJECT: Daemon %s: Starting up...", daemon.Name))
		
		daemon.State = 1 // Running
		
		go func(d *Daemon) {
			if err := d.Program(); err != nil {
				Output_Logg("ERR", "Main", fmt.Sprintf("PROJECT: Daemon %s: Error: %v", d.Name, err))
				d.mu.Lock()
				d.State = 2
				d.mu.Unlock()
			} else {
				Output_Logg("OUT", "Main", fmt.Sprintf("PROJECT: Daemon %s: Finished", d.Name))
				d.mu.Lock()
				d.State = 2
				d.mu.Unlock()
			}
		}(daemon) //immediately calls again
		
		Output_Logg("OUT", "Main", fmt.Sprintf("PROJECT: Daemon %s: Up and running", daemon.Name))
	}
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	Output_Logg("OUT", "Main", "PROJECT: Running. Press Ctrl+C to shutdown.")
	<-sigChan
	
	Output_Logg("OUT", "Main", "PROJECT: Shutdown signal received")
	
	//Time for graceful shutdown
	time.Sleep(2 * time.Second)
	Output_Logg("OUT", "Main", "PROJECT: Shutdown complete")
}

//  HTTP SERVICE 
// Request payload structure
type RequestPayload struct {
	Srvc string `json:"Srvc"`
}

// Response  
type ResponsePayload struct {
	ExctnOutcomeCode int    `json:"ExctnOutcomeCode"`
	ExctnOutcomeNote string `json:"ExctnOutcomeNote"`
	Yield            string `json:"Yield"`
}

func StartSR05Service() error {
	Output_Logg("OUT", "SR05", "Starting HTTP server on :8080")
	
	// Setup HTTP handler
	http.HandleFunc("/service", handler)
	
	// Start server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}
	return nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parse JSON
	var req RequestPayload
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	// Prepare response
	resp := ResponsePayload{
		ExctnOutcomeCode: 400,
		ExctnOutcomeNote: "unknown service code",
		Yield:            "",
	}

	// Check for sr05 service code
	if req.Srvc == "sr05" {
		resp.ExctnOutcomeCode = 200
		resp.ExctnOutcomeNote = ""
		resp.Yield = "Hello world"
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		Output_Logg("ERR", "SR05", fmt.Sprintf("failed to write response: %v", err))
	}
}


func main() {
	// Register the SR05 daemon
	DaemonRegister = []*Daemon{
		{
			Name:          "sr05-service",
			Program:       StartSR05Service,
			StartupGrace:  5 * time.Second,
			ShutdownGrace: 5 * time.Second,
		},
	}
	
	RunDaemonCore()
}
