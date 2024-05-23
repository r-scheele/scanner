package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"minio.io/clamd"
	"minio.io/config"
)

func httpResponse(scanResults []*clamd.ScanResult, w http.ResponseWriter, config *config.AppConfig) {

	encodedResults, err := json.Marshal(scanResults)
	if err != nil {
		http.Error(w, "Failed to encode scan results", http.StatusInternalServerError)
		return
	}

	postBody := bytes.NewReader(encodedResults)
	postRequest, err := http.NewRequest("POST", config.SubnetEndpoint, postBody)
	if err != nil {
		log.Printf("Failed to create post request: %v", err)
		http.Error(w, "Failed to post scan results", http.StatusInternalServerError)
		return
	}

	postRequest.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(postRequest)
	if err != nil {
		log.Printf("Failed to post results: %v", err)
		http.Error(w, "Failed to post scan results", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.Write(encodedResults)

}
