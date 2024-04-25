package main

import (
	"bytes"
	"io"
	"sync"

	"github.com/dutchcoders/go-clamd"
	"github.com/labstack/echo/v4"
)

func pingHandler(clam *clamd.Clamd) echo.HandlerFunc {
	return func(c echo.Context) error {
		err := clam.Ping()
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(500, "Could not ping clamd")
		}
		return c.JSON(200, "OK")
	}
}

// Modify your scanResponseHandler function
func scanResponseHandler(clam *clamd.Clamd) echo.HandlerFunc {

	return func(c echo.Context) error {
		name := c.FormValue("name")
		file, err := c.FormFile("file")
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(500, "Could not get file")
		}
		src, err := file.Open()
		if err != nil {
			c.Logger().Error(err)
			return echo.NewHTTPError(500, "Could not open file")
		}
		defer src.Close()
		filesScanned.Inc()

		// Define chunk size
		const chunkSize = 40 * 1024 * 1024 // 40MB

		// Create a buffered channel to collect scan results
		results := make(chan *clamd.ScanResult, 1)

		// Start a pool of worker goroutines
		var wg sync.WaitGroup
		for {
			buf := make([]byte, chunkSize)
			n, err := src.Read(buf)
			if err != nil {
				if err == io.EOF {
					break // Reached end of file, exit the loop
				} else {
					c.Logger().Error(err)
					results <- &clamd.ScanResult{Status: "ERROR", Raw: "Could not read file", Description: err.Error()}
					return c.JSON(500, "Could not read file")
				}
			}
			wg.Add(1)
			go func(chunk []byte, size int) {
				defer wg.Done()
				// Send the chunk for scanning
				response, err := clam.ScanStream(bytes.NewReader(chunk[:size]), make(chan bool))
				if err != nil {
					c.Logger().Error(err)
					results <- &clamd.ScanResult{Status: "ERROR", Raw: "Could not scan chunk", Description: err.Error()}
					return
				}
				result := <-response
				results <- result // Send the scan result to the channel
			}(buf, n)
		}

		// Close the results channel after all chunks have been processed
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect scan results
		var scanResults []*clamd.ScanResult
		for result := range results {
			scanResults = append(scanResults, result)
		}

		// Check the scan results
		for _, result := range scanResults {
			if result.Status == "FOUND" {
				c.Logger().Errorf("Malware detected in file %v -- %v", name, result.Raw)
				filesPositive.Inc()
				return c.JSON(451, result) // Return error response with scan result
			} else if result.Status == "ERROR" {
				c.Logger().Errorf("Error scanning file %v -- %v", name, result.Description)
				return c.JSON(500, result) // Return error response with scan result
			}
		}

		c.Logger().Infof("No malware detected in file %v", name)
		filesNegative.Inc()

		// Return the response similar to the previous code
		scanResult := &clamd.ScanResult{Status: "OK", Raw: "stream: OK", Description: "", Path: "stream", Hash: "", Size: 0}
		return c.JSON(200, scanResult)
	}
}
