package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"minio.io/clamd"
)

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

func scanStream(clam *clamd.Clamd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read request body stream - Adjust maxFileSize as needed
		const maxFileSize = 5000 * 1024 * 1024 // 5000 MB
		body := r.Body
		defer body.Close()

		var buf bytes.Buffer
		if _, err := io.CopyN(&buf, body, maxFileSize+1); err != nil && err != io.EOF {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		fileSize := int64(buf.Len())
		if fileSize > maxFileSize {
			http.Error(w, "File size exceeds the limit", http.StatusBadRequest)
			return
		}

		// Determine chunk size
		const chunkSize = 450 * 1024 * 1024 // 450 MB
		numChunks := (fileSize + chunkSize - 1) / chunkSize

		// Start streaming scan
		resChan := make(chan *clamd.ScanResult)
		var wg sync.WaitGroup
		for i := int64(0); i < numChunks; i++ {
			start := i * chunkSize
			end := (i + 1) * chunkSize
			if end > fileSize {
				end = fileSize
			}

			chunkReader := bytes.NewReader(buf.Bytes()[start:end])

			wg.Add(1)
			go func() {
				defer wg.Done()
				ch, err := clam.ScanStream(chunkReader, make(chan bool))
				if err != nil {
					fmt.Printf("Error scanning stream: %s\n", err.Error())
					return
				}
				for res := range ch {
					resChan <- res
				}
			}()
		}

		// Wait for all scans to complete
		go func() {
			wg.Wait()
			close(resChan)
		}()

		// Collect scan results
		var scanResults []*clamd.ScanResult
		for res := range resChan {
			scanResults = append(scanResults, res)
		}

		// Encode scan results as JSON and send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(scanResults); err != nil {
			http.Error(w, "Failed to encode scan results", http.StatusInternalServerError)
			return
		}
	}
}

func scanFile(clam *clamd.Clamd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body to extract file path
		var path string
		if err := json.NewDecoder(r.Body).Decode(&path); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		// Scan the file path
		ch, err := clam.ScanFile(path)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error scanning file %s: %v", path, err), http.StatusInternalServerError)
			return
		}

		// Collect scan results
		var scanResults []*clamd.ScanResult
		for res := range ch {
			scanResults = append(scanResults, res)
		}

		// Encode scan results as JSON and send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(scanResults); err != nil {
			http.Error(w, "Failed to encode scan results", http.StatusInternalServerError)
			return
		}
	}
}

func scanFiles(clam *clamd.Clamd) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body to extract file paths
		var paths []string
		if err := json.NewDecoder(r.Body).Decode(&paths); err != nil {
			http.Error(w, "Failed to decode request body", http.StatusBadRequest)
			return
		}

		// Scan each file path using MultiScanFile
		results := make(chan *clamd.ScanResult)
		var wg sync.WaitGroup
		for _, path := range paths {
			wg.Add(1)
			go func(p string) {
				defer wg.Done()
				ch, err := clam.MultiScanFile(p)
				if err != nil {
					log.Printf("Error scanning file %s: %v", p, err)
					return
				}
				for res := range ch {
					results <- res
				}
			}(path)
		}

		// Wait for all scans to complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect scan results and send them to the client
		var scanResults []*clamd.ScanResult
		for res := range results {
			scanResults = append(scanResults, res)
		}

		// Encode scan results as JSON and send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(scanResults); err != nil {
			http.Error(w, "Failed to encode scan results", http.StatusInternalServerError)
			return
		}
	}
}
