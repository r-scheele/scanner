package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"cloud.google.com/go/storage"

	"minio.io/clamd"
)

func scanPath(clam *clamd.Clamd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Updated data structure to include bucket name and message ID
		var data struct {
			FilePath   string `json:"filePath"`
			BucketName string `json:"bucketName"`
			MessageID  string `json:"messageId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("Failed to create client: %v", err)
			http.Error(w, "Failed to create storage client", http.StatusInternalServerError)
			return
		}
		defer client.Close()

		bucket := client.Bucket(data.BucketName)
		obj := bucket.Object(data.FilePath)

		reader, err := obj.NewReader(ctx)
		if err != nil {
			log.Printf("Failed to open file: %v", err)
			http.Error(w, "Failed to read file from bucket", http.StatusBadRequest)
			return
		}
		defer reader.Close()

		response, err := clam.ScanStream(reader, make(chan bool))
		if err != nil {
			http.Error(w, "Failed to scan the file", http.StatusInternalServerError)
			return
		}

		result := <-response
		result.FilePath = data.FilePath
		result.MessageID = data.MessageID

		scanResults := []*clamd.ScanResult{}
		scanResults = append(scanResults, result)

		httpResponse(scanResults, w)
	}
}

func scanPaths(clam *clamd.Clamd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// Updated data structure to include a list of files with message ID and bucket name for each
		var data struct {
			Requests []struct {
				FilePath   string `json:"filePath"`
				BucketName string `json:"bucketName"`
				MessageID  string `json:"messageId"`
			} `json:"requests"`
		}

		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Failed to decode request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Printf("Failed to create client: %v", err)
			http.Error(w, "Failed to create storage client", http.StatusInternalServerError)
			return
		}
		defer client.Close()

		var wg sync.WaitGroup
		var mu sync.Mutex
		scanResults := []*clamd.ScanResult{}

		for _, request := range data.Requests {
			wg.Add(1)
			go func(request struct {
				FilePath   string `json:"filePath"`
				BucketName string `json:"bucketName"`
				MessageID  string `json:"messageId"`
			}) {
				defer wg.Done()
				obj := client.Bucket(request.BucketName).Object(request.FilePath)

				reader, err := obj.NewReader(ctx)
				if err != nil {
					log.Printf("Failed to open file %s: %v", request.FilePath, err)
					return
				}
				defer reader.Close()

				response, err := clam.ScanStream(reader, make(chan bool))
				if err != nil {
					log.Printf("Failed to scan file %s: %v", request.FilePath, err)
					return
				}

				result := <-response
				result.FilePath = request.FilePath
				result.MessageID = request.MessageID

				mu.Lock()
				scanResults = append(scanResults, result)
				mu.Unlock()
			}(request)
		}

		wg.Wait()

		httpResponse(scanResults, w)
	}
}

func ping(clam *clamd.Clamd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := clam.Ping()
		if err != nil {
			log.Println(err)
			http.Error(w, "Could not ping clamd", http.StatusInternalServerError)
			return
		}
		fmt.Fprintln(w, "OK")
	}
}

func scanFile(clam *clamd.Clamd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Parse the multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil { // Set max memory to 10 MB for multipart form
			http.Error(w, "Failed to parse multipart form", http.StatusInternalServerError)
			return
		}

		// Retrieve the file from the form data
		files := r.MultipartForm.File["file"]
		if len(files) == 0 {
			http.Error(w, "No file provided", http.StatusBadRequest)
			return
		}

		// Open the file
		file, err := files[0].Open()
		if err != nil {
			http.Error(w, "Failed to open the file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Scan the file using clamd's ScanStream
		response, err := clam.ScanStream(file, make(chan bool))
		if err != nil {
			http.Error(w, "Failed to scan the file", http.StatusInternalServerError)
			return
		}

		// Receive the scan result
		result := <-response

		// Log the scanning action
		// Assuming there's a logger configured similarly to the previous example
		log.Printf("Scanning %s and returning reply", files[0].Filename)

		var scanResults []*clamd.ScanResult
		scanResults = append(scanResults, result)

		httpResponse(scanResults, w)

	}
}

func scanFiles(clam *clamd.Clamd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form
		if err := r.ParseMultipartForm(10 << 20); err != nil { // Set max memory to 10 MB for multipart form
			http.Error(w, "Failed to parse multipart form", http.StatusInternalServerError)
			return
		}

		// Retrieve the files from the form data
		files := r.MultipartForm.File["file"]
		if len(files) == 0 {
			http.Error(w, "No file provided", http.StatusBadRequest)
			return
		}

		var scanResults []*clamd.ScanResult

		// Loop over the files and scan each one
		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				http.Error(w, "Failed to open a file", http.StatusInternalServerError)
				return
			}

			// Ensure the file is closed after processing
			defer file.Close()

			// Scan the file using clamd's ScanStream
			response, err := clam.ScanStream(file, make(chan bool))
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to scan the file: %s", fileHeader.Filename), http.StatusInternalServerError)
				return
			}

			// Receive the scan result
			result := <-response

			// Log the scanning action
			log.Printf("Scanning %s and returning reply", fileHeader.Filename)

			// Append result to the scanResults
			scanResults = append(scanResults, result)

		}

		httpResponse(scanResults, w)
	}
}
