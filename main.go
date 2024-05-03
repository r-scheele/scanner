package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/minio/pkg/v2/env"
	"minio.io/clamd"
)

func getEnv(key, fallback string) string {
	return env.Get(key, fallback)
}

func main() {
	clamHost := getEnv("CLAMD_HOST", "localhost")
	clamPort := getEnv("CLAMD_PORT", "3310")
	listenPort := getEnv("LISTEN_PORT", "8080")

	clamConnection := clamd.NewClamd(fmt.Sprintf("tcp://%v:%v", clamHost, clamPort))

	err := clamConnection.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to clamd on %v:%v", clamHost, clamPort)
	log.Printf("Listening on port %v", listenPort)

	http.HandleFunc("/scan/stream", scanStream(clamConnection))
	http.HandleFunc("/ping", ping(clamConnection))
	http.HandleFunc("/scan/file", scanFile(clamConnection))
	http.HandleFunc("/scan/files", scanFiles(clamConnection))

	err = http.ListenAndServe(":"+listenPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}
