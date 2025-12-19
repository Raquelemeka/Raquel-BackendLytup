package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Daemon structure
type Daemon struct{
	Name            string
	Program         func() error
	State           uint64 // 0 - Initial; 1 - Running; 2 - Done
	StartupGrace    time.Duration
	ShutdownGrace   time.Duration
	mu              sync.Mutex
}

var DaemonRegister []*Daemon

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

func RunDaemonCore(){
	Output_Logg("OUT", "Main", "PROJECT: Starting up")

	if len(DaemonRegister) == 0 {
		Output_Logg("OUT", "Main", "PROJECT: No Daemon(s) to run. Shutting down now")
		return
	}

	for _, daemon := range DaemonRegister {
		Output_Logg("OUT", "Main", fmt.Sprintf("PROJECT: Daemon %s: Starting up...", daemon.Name))

		daemon.mu.Lock()
		daemon.State = 1
		daemon.mu.Unlock()

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
		}(daemon)

		Output_Logg("OUT", "Main", fmt.Sprintf("PROJECT: Daemon %s: Up and running", daemon.Name))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	Output_Logg("OUT", "Main", "PROJECT: Running. Press Ctrl+C to shutdown.")
	<-sigChan

	Output_Logg("OUT", "Main", "PROJECT: Shutdown signal received")
	time.Sleep(2 * time.Second)
	Output_Logg("OUT", "Main", "PROJECT: Shutdown complete")
}

// HTTPS
type RequestPayload struct{
	Srvc string       `json:"Srvc"`
	Seed *WeatherSeed `json:"Seed,omitempty"`
}

// Weather seed data for ex02 
type WeatherSeed struct{
	Lat   float64 `json:"lat"`
	Lon   float64 `json:"lon"`
	Start string  `json:"start"`
	End   string  `json:"end"`
}

type ResponsePayload struct{
	ExctnOutcomeCode int    `json:"ExctnOutcomeCode"`
	ExctnOutcomeNote string `json:"ExctnOutcomeNote"`
	Yield            string `json:"Yield"`
}

// Weather response structure for ex02
type WeatherResponse struct{
	ExctnOutcomeCode int          `json:"ExctnOutcomeCode"`
	ExctnOutcomeNote string       `json:"ExctnOutcomeNote"`
	Yield            *WeatherData `json:"Yield,omitempty"`
}

type WeatherData struct{
	Hourly HourlyWeather `json:"hourly"`
	Daily  DailySunInfo  `json:"daily"`
}

type HourlyWeather struct{
	Time          []string  `json:"time"`
	Temperature   []float64 `json:"temperature"`
	WindSpeed     []float64 `json:"wind_speed"`
	WindDirection []int     `json:"wind_direction"`
	Humidity      []int     `json:"humidity"`
	WeatherCode   []int     `json:"weather_code"`
}

type DailySunInfo struct{
	Date    []string `json:"date"`
	Sunrise []string `json:"sunrise"`
	Sunset  []string `json:"sunset"`
}

/*
	Open-Meteo API
*/
type OpenMeteoResponse struct {
	Hourly OpenMeteoHourly `json:"hourly"`
	Daily  OpenMeteoDaily  `json:"daily"`
}

type OpenMeteoHourly struct {
	Time          []string  `json:"time"`
	Temperature   []float64 `json:"temperature_2m"`
	WindSpeed     []float64 `json:"wind_speed_10m"`
	WindDirection []int     `json:"wind_direction_10m"`
	Humidity      []int     `json:"relative_humidity_2m"`
	WeatherCode   []int     `json:"weather_code"`
}

type OpenMeteoDaily struct {
	Time    []string `json:"time"`
	Sunrise []string `json:"sunrise"`
	Sunset  []string `json:"sunset"`
}

func StartSR05Service() error{
	Output_Logg("OUT", "SR05", "Starting HTTP server on :8080")

	http.HandleFunc("/service", handler)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func handler(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodPost {
		http.Error(w, "only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req RequestPayload
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	switch req.Srvc{
	case "sr05":
		resp := ResponsePayload{
			ExctnOutcomeCode: 200,
			ExctnOutcomeNote: "",
			Yield:            "Hello world",
		}
		json.NewEncoder(w).Encode(resp)

	case "ex02":
		handleEx02Service(w, req)

	default:
		resp := ResponsePayload{
			ExctnOutcomeCode: 400,
			ExctnOutcomeNote: "unknown service code",
			Yield:            "",
		}
		json.NewEncoder(w).Encode(resp)
	}
}

func handleEx02Service(w http.ResponseWriter, req RequestPayload) {
	if req.Seed == nil {
		resp := WeatherResponse{
			ExctnOutcomeCode: 400,
			ExctnOutcomeNote: "missing Seed data for ex02 service",
			Yield:            nil,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	if req.Seed.Lat < -90 || req.Seed.Lat > 90 || req.Seed.Lon < -180 || req.Seed.Lon > 180 {
		resp := WeatherResponse{
			ExctnOutcomeCode: 400,
			ExctnOutcomeNote: "invalid latitude or longitude values",
			Yield:            nil,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	weatherData, err := fetchWeatherFromOpenMeteo(*req.Seed)
	if err != nil {
		resp := WeatherResponse{
			ExctnOutcomeCode: 500,
			ExctnOutcomeNote: "failed to fetch weather data",
			Yield:            nil,
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := WeatherResponse{
		ExctnOutcomeCode: 200,
		ExctnOutcomeNote: "",
		Yield:            &weatherData,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		Output_Logg("ERR", "SR05", fmt.Sprintf("failed to write response: %v", err))
	}
}

/*
   Fetch real weather data from Open-Meteo instead of generating mock data
*/
func fetchWeatherFromOpenMeteo(seed WeatherSeed) (WeatherData, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=temperature_2m,relative_humidity_2m,wind_speed_10m,wind_direction_10m,weather_code&daily=sunrise,sunset&timezone=UTC",
		seed.Lat,
		seed.Lon,
	)

	resp, err := http.Get(url)
	if err != nil {
		return WeatherData{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WeatherData{}, err
	}

	var apiResp OpenMeteoResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return WeatherData{}, err
	}

	return WeatherData{
		Hourly: HourlyWeather{
			Time:          apiResp.Hourly.Time,
			Temperature:   apiResp.Hourly.Temperature,
			WindSpeed:     apiResp.Hourly.WindSpeed,
			WindDirection: apiResp.Hourly.WindDirection,
			Humidity:      apiResp.Hourly.Humidity,
			WeatherCode:   apiResp.Hourly.WeatherCode,
		},
		Daily: DailySunInfo{
			Date:    apiResp.Daily.Time,
			Sunrise: apiResp.Daily.Sunrise,
			Sunset:  apiResp.Daily.Sunset,
		},
	}, nil
}

func main() {
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
