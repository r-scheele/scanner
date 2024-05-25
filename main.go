package main

import (
	"log"
	"net/http"

	"minio.io/clamd"
	"minio.io/config"
)

func main() {
	config := config.LoadConfig()

	// Initialize clamd connection
	clamConnection := clamd.NewClamd(config.ClamAddress)
	if err := clamConnection.Ping(); err != nil {
		log.Fatalf("Failed to connect to clamd at %s: %v", config.ClamAddress, err)
	}

	log.Printf("Connected to clamd on %s", config.ClamAddress)
	log.Printf("Server starting on port %s", config.ListenPort)

	http.HandleFunc("/ping", ping(clamConnection))
	http.HandleFunc("/scan/path", scanPath(clamConnection, config))
	http.HandleFunc("/scan/paths", scanPaths(clamConnection, config))
	http.HandleFunc("/scan/file", scanFile(clamConnection, config))
	http.HandleFunc("/scan/files", scanFiles(clamConnection, config))

	// Start HTTP server
	if err := http.ListenAndServe(":"+config.ListenPort, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
