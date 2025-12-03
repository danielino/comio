package main

import (
"bytes"
"crypto/rand"
"fmt"
"io"
"net/http"
"os"
"sync"
"sync/atomic"
"time"
)

const (
	baseURL          = "http://localhost:8080"
	bucket           = "test"
	numSmallFiles    = 500  // 1-100KB
	numMediumFiles   = 200  // 100KB-1MB
	numLargeFiles    = 100  // 1-5MB
	concurrency      = 20   // parallel uploads
)

type TestResult struct {
	Success bool
	Key     string
	Size    int
	Error   string
}

func main() {
	fmt.Println("======================================")
	fmt.Println("ComIO Parallel Stress Test")
	fmt.Println("======================================")
	fmt.Printf("Small files (1-100KB): %d\n", numSmallFiles)
	fmt.Printf("Medium files (100KB-1MB): %d\n", numMediumFiles)
	fmt.Printf("Large files (1-5MB): %d\n", numLargeFiles)
	fmt.Printf("Concurrency: %d\n", concurrency)
	fmt.Println("======================================")
	fmt.Println()

	startTime := time.Now()
	
	var (
successCount atomic.Int64
errorCount   atomic.Int64
wg           sync.WaitGroup
results      = make(chan TestResult, numSmallFiles+numMediumFiles+numLargeFiles)
semaphore    = make(chan struct{}, concurrency)
)

	// Test small files
	fmt.Println("Testing small files (1-100KB)...")
	for i := 1; i <= numSmallFiles; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}        // acquire
			defer func() { <-semaphore }() // release

			size := (idx%100 + 1) * 1024 // 1-100KB
			key := fmt.Sprintf("small_%d_%dk", idx, size/1024)
			
			result := uploadFile(key, size)
			results <- result
			
			if result.Success {
				successCount.Add(1)
				fmt.Print("\033[0;32m.\033[0m")
			} else {
				errorCount.Add(1)
				fmt.Print("\033[0;31mF\033[0m")
			}
			
			if idx%10 == 0 {
				fmt.Printf(" [%d/%d]\n", idx, numSmallFiles)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println()

	// Test medium files
	fmt.Println("Testing medium files (100KB-1MB)...")
	for i := 1; i <= numMediumFiles; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			size := (100 + (idx%900)) * 1024 // 100-1000KB
			key := fmt.Sprintf("medium_%d_%dk", idx, size/1024)
			
			result := uploadFile(key, size)
			results <- result
			
			if result.Success {
				successCount.Add(1)
				fmt.Print("\033[0;32m.\033[0m")
			} else {
				errorCount.Add(1)
				fmt.Print("\033[0;31mF\033[0m")
			}
			
			if idx%5 == 0 {
				fmt.Printf(" [%d/%d]\n", idx, numMediumFiles)
			}
		}(i)
	}
	wg.Wait()
	fmt.Println()

	// Test large files
	fmt.Println("Testing large files (1-5MB)...")
	for i := 1; i <= numLargeFiles; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			size := (1 + (idx % 5)) * 1024 * 1024 // 1-5MB
			key := fmt.Sprintf("large_%d_%dmb", idx, size/(1024*1024))
			
			result := uploadFile(key, size)
			results <- result
			
			if result.Success {
				successCount.Add(1)
				fmt.Print("\033[0;32m.\033[0m")
			} else {
				errorCount.Add(1)
				fmt.Print("\033[0;31mF\033[0m")
			}
			
			if idx%2 == 0 {
				fmt.Printf(" [%d/%d]\n", idx, numLargeFiles)
			}
		}(i)
	}
	wg.Wait()
	close(results)
	fmt.Println()

	// Collect errors
	var errors []TestResult
	for result := range results {
		if !result.Success {
			errors = append(errors, result)
		}
	}

	duration := time.Since(startTime)

	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("Stress Test Results")
	fmt.Println("======================================")
	fmt.Printf("Successful uploads: \033[0;32m%d\033[0m\n", successCount.Load())
	fmt.Printf("Failed uploads: \033[0;31m%d\033[0m\n", errorCount.Load())
	fmt.Printf("Total duration: %.2fs\n", duration.Seconds())
	fmt.Printf("Throughput: %.2f req/s\n", float64(successCount.Load()+errorCount.Load())/duration.Seconds())
	fmt.Println("======================================")

	// Write errors to file
	if len(errors) > 0 {
		f, err := os.Create("stress_errors.log")
		if err == nil {
			for _, e := range errors {
				fmt.Fprintf(f, "Failed: %s (size: %d bytes, error: %s)\n", e.Key, e.Size, e.Error)
			}
			f.Close()
			fmt.Println()
			fmt.Println("\033[1;33mErrors logged to stress_errors.log\033[0m")
		}
	}

	// Show metrics
	fmt.Println()
	fmt.Println("======================================")
	fmt.Println("Server Metrics")
	fmt.Println("======================================")
	showMetrics()

	if errorCount.Load() > 0 {
		os.Exit(1)
	}
}

func uploadFile(key string, size int) TestResult {
	// Generate random data
	data := make([]byte, size)
	rand.Read(data)

	url := fmt.Sprintf("%s/%s/%s", baseURL, bucket, key)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		return TestResult{Success: false, Key: key, Size: size, Error: err.Error()}
	}

	req.Header.Set("Content-Type", "application/octet-stream")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TestResult{Success: false, Key: key, Size: size, Error: err.Error()}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return TestResult{Success: false, Key: key, Size: size, Error: string(body)}
	}

	return TestResult{Success: true, Key: key, Size: size}
}

func showMetrics() {
	resp, err := http.Get(baseURL + "/admin/metrics")
	if err != nil {
		fmt.Printf("Error fetching metrics: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println(string(body))
}
