package main

import (
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/rs/cors"
	"minio.io/clamd"
	"minio.io/config"
)

func main() {

	err := config.LoadAppConfig()
	if err != nil {
		color.Yellow("Error loading app config: [%s]. Proceeding with default config values.", err.Error())
	}

	// Initialize clamd connection
	clamAddress := config.GetConfigString("CLAM_ADDRESS", "tcp://localhost:3310")
	clamConnection := clamd.NewClamd(clamAddress)
	if err := clamConnection.Ping(); err != nil {
		log.Fatalf("Failed to connect to clamd at %s: %v", clamAddress, err)
	}

	log.Printf("Connected to clamd on %s", clamAddress)

	listenPort := config.GetConfigString("LISTEN_PORT", "8080")

	log.Printf("Server starting on port %s", listenPort)

	http.HandleFunc("/ping", ping(clamConnection))
	http.HandleFunc("/scan/path", scanPath(clamConnection))
	http.HandleFunc("/scan/paths", scanPaths(clamConnection))
	http.HandleFunc("/scan/file", scanFile(clamConnection))
	http.HandleFunc("/scan/files", scanFiles(clamConnection))

	// CORS configuration
	allowedOrigins := config.GetConfigStringSlice("allowed_origins")
	corsHandler := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedHeaders: []string{"*"},
	}).Handler(http.DefaultServeMux)

	// Start HTTP server
	if err := http.ListenAndServe(":"+listenPort, corsHandler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
