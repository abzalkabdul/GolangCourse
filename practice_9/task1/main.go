package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"
)

type PaymentClient struct {
	HTTPClient *http.Client
	MaxRetries int
	BaseDelay  time.Duration
}

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	switch resp.StatusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	case http.StatusUnauthorized, // 401
		http.StatusNotFound: // 404
		return false
	}
	return false
}

func CalculateBackoff(attempt int, baseDelay time.Duration) time.Duration {
	exp := attempt
	if exp > 10 {
		exp = 10
	}
	maxDelay := float64(baseDelay) * math.Pow(2, float64(exp))

	jittered := rand.Float64() * maxDelay
	return time.Duration(jittered)
}

func (c *PaymentClient) ExecutePayment(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	var lastResp *http.Response

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		fmt.Printf("Attempt %d: sending request...\n", attempt+1)
		resp, err := c.HTTPClient.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				return nil, fmt.Errorf("failed to read response: %w", readErr)
			}
			fmt.Printf("Attempt %d: Success! Status: %d\n", attempt+1, resp.StatusCode)
			return body, nil
		}

		if !IsRetryable(resp, err) {
			if resp != nil {
				resp.Body.Close()
				return nil, fmt.Errorf("non-retryable error: status %d", resp.StatusCode)
			}
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}

		lastErr = err
		if resp != nil {
			lastResp = resp
			resp.Body.Close()
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		if attempt == c.MaxRetries {
			break
		}

		backoff := CalculateBackoff(attempt, c.BaseDelay)
		fmt.Printf("Attempt %d failed: waiting %v...\n", attempt+1, backoff.Round(time.Millisecond))

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
		case <-time.After(backoff):
		}
	}

	_ = lastResp
	return nil, fmt.Errorf("all %d attempts failed, last error: %v", c.MaxRetries+1, lastErr)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var requestCount int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&requestCount, 1)
		fmt.Printf("[Server] Received request #%d\n", count)

		if count <= 3 {
			// First 3 requests fail with 503
			fmt.Printf("[Server] Returning 503 for request #%d\n", count)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "service unavailable"}`))
			return
		}

		// 4th request succeeds
		fmt.Printf("[Server] Returning 200 for request #%d\n", count)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	fmt.Println("=== Payment Gateway Retry Demo ===")
	fmt.Printf("Server URL: %s\n\n", server.URL)

	client := &PaymentClient{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	body, err := client.ExecutePayment(ctx, server.URL+"/pay")
	if err != nil {
		fmt.Printf("\nFailed to process payment: %v\n", err)
		return
	}

	fmt.Printf("\nPayment response: %s\n", body)
	fmt.Println("=== Payment processed successfully! ===")
}
