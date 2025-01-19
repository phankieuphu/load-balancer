package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"sync"
	"time"
)

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatalf("Error loading configuration :%s", err.Error())

	}
	fmt.Println(config.HealthCheckInterval)
	healthCheckInterval, err := time.ParseDuration(config.HealthCheckInterval)
	if err != nil {
		log.Fatalf("Invalid health check interval %s", err.Error())
	}
	var servers []*Server

	for _, serverUrl := range config.Servers {
		u, _ := url.Parse(serverUrl)
		servers = append(servers, &Server{URL: u})

	}
	// health check  server
	for _, server := range servers {
		go func(s *Server) {
			for range time.Tick(healthCheckInterval) {
				res, err := http.Get(s.URL.String())
				if err != nil || res.StatusCode >= 500 {
					s.Healthy = false
				} else {
					s.Healthy = true
				}

			}
		}(server)
	}
	lb := LoadBalancer{Current: 0}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server := lb.nextServerLeastActive(servers)
		fmt.Println("new request coming")

		if server == nil {
			http.Error(w, "No healthy server available", http.StatusServiceUnavailable)
			return
		}
		// server.Mutex.Lock()
		// server.ActiveConnections++
		// fmt.Println(server.URL, " ", server.ActiveConnections)
		// server.Mutex.Unlock()
		server.Proxy().ServeHTTP(w, r)
		server.Mutex.Lock()
		server.ActiveConnections--
		server.Mutex.Unlock()
	})

	log.Println("Starting server on port", config.ListenPort)
	err = http.ListenAndServe(config.ListenPort, nil)
	if err != nil {
		log.Fatalf("Error starting server: %s\n", err)
	}

}

type LoadBalancer struct {
	Current int
	Mutex   sync.Mutex
}
type Server struct {
	URL               *url.URL
	ActiveConnections int        // Count of active connections
	Mutex             sync.Mutex //  A mutex for safe concurrency
	Healthy           bool
}

type Config struct {
	HealthCheckInterval string   `json:"healthCheckInterval"`
	Servers             []string `json:"servers"`
	ListenPort          string   `json:"listenPort"`
}

func loadConfig(file string) (Config, error) {
	var config Config

	bytes, err := os.ReadFile(file)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, err

	}

	return config, nil

}

func (lb *LoadBalancer) nextServerLeastActive(servers []*Server) *Server {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()

	var leastActiveServer *Server
	leastActiveConnections := -1

	for _, server := range servers {
		server.Mutex.Lock()
		isHealthy := server.Healthy
		activeConnections := server.ActiveConnections
		server.Mutex.Unlock()

		if isHealthy && (leastActiveServer == nil || activeConnections < leastActiveConnections) {
			leastActiveServer = server
			leastActiveConnections = activeConnections
		}
	}

	if leastActiveServer != nil {
		// Increment ActiveConnections for the selected server
		leastActiveServer.Mutex.Lock()
		leastActiveServer.ActiveConnections++
		leastActiveServer.Mutex.Unlock()
	}

	return leastActiveServer
}

func (s *Server) Proxy() *httputil.ReverseProxy {
	return httputil.NewSingleHostReverseProxy(s.URL)
}
