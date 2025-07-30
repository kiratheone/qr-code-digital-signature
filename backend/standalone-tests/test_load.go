package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// LoadTestConfig holds configuration for load testing
type LoadTestConfig struct {
	BaseURL     string
	Concurrency int
	Requests    int
	Timeout     time.Duration
}

// TestResult holds the result of a single test
type TestResult struct {
	Duration time.Duration
	Success  bool
	Error    error
}

// LoadTestResults holds aggregated test results
type LoadTestResults struct {
	TotalRequests   int
	SuccessRequests int
	FailedRequests  int
	AverageTime     time.Duration
	MinTime         time.Duration
	MaxTime         time.Duration
	RequestsPerSec  float64
}

func main() {
	config := LoadTestConfig{
		BaseURL:     getEnv("API_URL", "http://localhost:8000"),
		Concurrency: 10,
		Requests:    100,
		Timeout:     30 * time.Second,
	}

	fmt.Printf("Starting load test with %d concurrent users, %d total requests\n", config.Concurrency, config.Requests)
	fmt.Printf("Target URL: %s\n", config.BaseURL)
	fmt.Println(strings.Repeat("=", 60))

	// Test different endpoints
	tests := []struct {
		name string
		fn   func(LoadTestConfig) LoadTestResults
	}{
		{"Health Check", testHealthEndpoint},
		{"Document List", testDocumentListEndpoint},
		{"Verification Info", testVerificationEndpoint},
	}

	for _, test := range tests {
		fmt.Printf("\nTesting: %s\n", test.name)
		fmt.Println(strings.Repeat("-", 40))
		
		start := time.Now()
		results := test.fn(config)
		totalTime := time.Since(start)
		
		printResults(results, totalTime)
	}
}

func testHealthEndpoint(config LoadTestConfig) LoadTestResults {
	return runLoadTest(config, func() TestResult {
		start := time.Now()
		
		client := &http.Client{Timeout: config.Timeout}
		resp, err := client.Get(config.BaseURL + "/health")
		
		duration := time.Since(start)
		
		if err != nil {
			return TestResult{Duration: duration, Success: false, Error: err}
		}
		defer resp.Body.Close()
		
		success := resp.StatusCode == http.StatusOK
		return TestResult{Duration: duration, Success: success, Error: nil}
	})
}

func testDocumentListEndpoint(config LoadTestConfig) LoadTestResults {
	return runLoadTest(config, func() TestResult {
		start := time.Now()
		
		client := &http.Client{Timeout: config.Timeout}
		req, err := http.NewRequest("GET", config.BaseURL+"/api/documents?page=1&limit=10", nil)
		if err != nil {
			return TestResult{Duration: time.Since(start), Success: false, Error: err}
		}
		
		// Add mock authorization header for testing
		req.Header.Set("Authorization", "Bearer test-token")
		
		resp, err := client.Do(req)
		duration := time.Since(start)
		
		if err != nil {
			return TestResult{Duration: duration, Success: false, Error: err}
		}
		defer resp.Body.Close()
		
		// Accept both 200 (success) and 401 (unauthorized) as valid responses for load testing
		success := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized
		return TestResult{Duration: duration, Success: success, Error: nil}
	})
}

func testVerificationEndpoint(config LoadTestConfig) LoadTestResults {
	return runLoadTest(config, func() TestResult {
		start := time.Now()
		
		client := &http.Client{Timeout: config.Timeout}
		// Use a test document ID
		resp, err := client.Get(config.BaseURL + "/api/verify/test-doc-id")
		
		duration := time.Since(start)
		
		if err != nil {
			return TestResult{Duration: duration, Success: false, Error: err}
		}
		defer resp.Body.Close()
		
		// Accept 200, 404 (not found) as valid responses for load testing
		success := resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound
		return TestResult{Duration: duration, Success: success, Error: nil}
	})
}

func runLoadTest(config LoadTestConfig, testFunc func() TestResult) LoadTestResults {
	var wg sync.WaitGroup
	results := make(chan TestResult, config.Requests)
	
	// Control concurrency
	semaphore := make(chan struct{}, config.Concurrency)
	
	start := time.Now()
	
	for i := 0; i < config.Requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			
			result := testFunc()
			results <- result
		}()
	}
	
	wg.Wait()
	close(results)
	
	totalTime := time.Since(start)
	
	// Aggregate results
	var totalDuration time.Duration
	var minTime, maxTime time.Duration
	successCount := 0
	failedCount := 0
	
	first := true
	for result := range results {
		if result.Success {
			successCount++
		} else {
			failedCount++
		}
		
		totalDuration += result.Duration
		
		if first {
			minTime = result.Duration
			maxTime = result.Duration
			first = false
		} else {
			if result.Duration < minTime {
				minTime = result.Duration
			}
			if result.Duration > maxTime {
				maxTime = result.Duration
			}
		}
	}
	
	avgTime := totalDuration / time.Duration(config.Requests)
	requestsPerSec := float64(config.Requests) / totalTime.Seconds()
	
	return LoadTestResults{
		TotalRequests:   config.Requests,
		SuccessRequests: successCount,
		FailedRequests:  failedCount,
		AverageTime:     avgTime,
		MinTime:         minTime,
		MaxTime:         maxTime,
		RequestsPerSec:  requestsPerSec,
	}
}

func printResults(results LoadTestResults, totalTime time.Duration) {
	fmt.Printf("Total Requests: %d\n", results.TotalRequests)
	fmt.Printf("Successful: %d (%.2f%%)\n", results.SuccessRequests, 
		float64(results.SuccessRequests)/float64(results.TotalRequests)*100)
	fmt.Printf("Failed: %d (%.2f%%)\n", results.FailedRequests,
		float64(results.FailedRequests)/float64(results.TotalRequests)*100)
	fmt.Printf("Average Response Time: %v\n", results.AverageTime)
	fmt.Printf("Min Response Time: %v\n", results.MinTime)
	fmt.Printf("Max Response Time: %v\n", results.MaxTime)
	fmt.Printf("Requests/Second: %.2f\n", results.RequestsPerSec)
	fmt.Printf("Total Test Time: %v\n", totalTime)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}