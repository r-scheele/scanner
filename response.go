package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"minio.io/clamd"
	"minio.io/config"
)

func httpResponse(scanResults []*clamd.ScanResult, w http.ResponseWriter) {
	// Marshal the scan results into JSON
	encodedResults, err := json.Marshal(scanResults)
	if err != nil {
		http.Error(w, "Failed to encode scan results", http.StatusInternalServerError)
		return
	}

	// Create a reader for the JSON payload
	postBody := bytes.NewReader(encodedResults)

	// Create a new HTTP POST request with the body and the subnet endpoint
	postRequest, err := http.NewRequest("POST", config.GetConfigString("SUBNET_ENDPOINT", ""), postBody)
	if err != nil {
		log.Printf("Failed to create post request: %v", err)
		http.Error(w, "Failed to post scan results", http.StatusInternalServerError)
		return
	}

	// Set the Content-Type header to indicate the type of data being sent
	postRequest.Header.Set("Content-Type", "application/json")

	// Retrieve the API token dynamically and add to the Authorization header
	apiToken := getAPIToken(w)
	postRequest.Header.Set("Authorization", "Bearer "+apiToken)

	// Create an HTTP client and send the POST request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(postRequest)
	if err != nil {
		log.Printf("Failed to post results: %v", err)
		http.Error(w, "Failed to post scan results", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Write the JSON encoded scan results back to the client
	w.Header().Set("Content-Type", "application/json")
	w.Write(encodedResults)
}

// Function to retrieve API token dynamically, can be adjusted as per actual implementation needs
func getAPIToken(w http.ResponseWriter) string {
	// Here you could access a secure store, or fetch the token from a configuration file

	token := config.GetConfigString("API_TOKEN", "")
	if token == "" {
		log.Println("API Token not set, using default")
		http.Error(w, "API Token not set", http.StatusInternalServerError)
		return ""

	}

	return token
}
