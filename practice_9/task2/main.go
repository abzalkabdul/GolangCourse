package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	client *redis.Client
}

func NewRedisStore() *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Cannot connect to Redis: %v", err))
	}

	fmt.Println("[Redis] Connected successfully")
	return &RedisStore{client: client}
}

func (s *RedisStore) SetNX(ctx context.Context, key string) (bool, error) {
	redisKey := "idempotency:" + key
	return s.client.SetNX(ctx, redisKey, "processing", 60*time.Second).Result()
}

func (s *RedisStore) Get(ctx context.Context, key string) (string, error) {
	redisKey := "idempotency:" + key
	return s.client.Get(ctx, redisKey).Result()
}

func (s *RedisStore) Finish(ctx context.Context, key string, statusCode int, body []byte) error {
	redisKey := "idempotency:" + key

	payload := map[string]interface{}{
		"status_code": statusCode,
		"body":        string(body),
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return s.client.Set(ctx, redisKey, string(data), 30*time.Second).Err()
}

func IdempotencyMiddleware(store *RedisStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		key := r.Header.Get("Idempotency-Key")

		if key == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"Idempotency-Key header required"}`, http.StatusBadRequest)
			return
		}

		won, err := store.SetNX(ctx, key)
		if err != nil {
			http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
			return
		}

		if !won {
			val, err := store.Get(ctx, key)
			if err != nil {
				http.Error(w, `{"error":"storage error"}`, http.StatusInternalServerError)
				return
			}

			if val == "processing" {
				fmt.Printf("[Middleware] Key=%s still processing, returning 409\n", key)
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"Duplicate request in progress"}`, http.StatusConflict)
				return
			}

			var cached map[string]interface{}
			if err := json.Unmarshal([]byte(val), &cached); err != nil {
				http.Error(w, `{"error":"failed to parse cached response"}`, http.StatusInternalServerError)
				return
			}

			statusCode := int(cached["status_code"].(float64))
			body := cached["body"].(string)

			fmt.Printf("[Middleware] Key=%s already completed, returning cached response\n", key)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Idempotent-Replayed", "true")
			w.WriteHeader(statusCode)
			w.Write([]byte(body))
			return
		}

		fmt.Printf("[Middleware] Key=%s new request, starting processing\n", key)

		recorder := httptest.NewRecorder()
		next.ServeHTTP(recorder, r)

		if err := store.Finish(ctx, key, recorder.Code, recorder.Body.Bytes()); err != nil {
			fmt.Printf("[Middleware] Warning: failed to save result to Redis: %v\n", err)
		}

		fmt.Printf("[Middleware] Key=%s processing finished, result saved to Redis\n", key)

		for k, vals := range recorder.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(recorder.Code)
		w.Write(recorder.Body.Bytes())
	})
}

var businessLogicCalls int32

func paymentHandler(w http.ResponseWriter, r *http.Request) {
	callNum := atomic.AddInt32(&businessLogicCalls, 1)
	fmt.Printf("[Handler] Processing started (business logic call #%d)\n", callNum)

	time.Sleep(2 * time.Second)

	transactionID := uuid.New().String()
	response := map[string]interface{}{
		"status":         "paid",
		"amount":         1000,
		"transaction_id": transactionID,
	}

	fmt.Printf("[Handler] Processing completed, transaction_id=%s\n", transactionID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	store := NewRedisStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/payment", paymentHandler)

	handler := IdempotencyMiddleware(store, mux)
	server := httptest.NewServer(handler)
	defer server.Close()

	fmt.Println("\n=== Idempotency Middleware Demo Redis ===")
	fmt.Printf("Server: %s\n\n", server.URL)

	const idempotencyKey = "pay-order-XYZ-2026"
	const numRequests = 8

	store.client.Del(context.Background(), "idempotency:"+idempotencyKey)

	fmt.Printf("Launching %d simultaneous requests with key: %s\n\n", numRequests, idempotencyKey)

	type result struct {
		id     int
		status int
		body   string
		replay bool
	}

	results := make(chan result, numRequests)
	var wg sync.WaitGroup

	for i := 1; i <= numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			req, _ := http.NewRequest(http.MethodPost, server.URL+"/payment", nil)
			req.Header.Set("Idempotency-Key", idempotencyKey)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				results <- result{id: id, status: -1, body: err.Error()}
				return
			}
			defer resp.Body.Close()

			var bodyMap map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&bodyMap)
			bodyJSON, _ := json.Marshal(bodyMap)

			results <- result{
				id:     id,
				status: resp.StatusCode,
				body:   string(bodyJSON),
				replay: resp.Header.Get("X-Idempotent-Replayed") == "true",
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	fmt.Println("\n=== Results ===")
	var successCount, conflictCount, replayCount int
	for r := range results {
		tag := ""
		switch r.status {
		case http.StatusOK:
			if r.replay {
				tag = "✅ 200 OK (replayed cached)"
				replayCount++
			} else {
				tag = "✅ 200 OK (first execution)"
				successCount++
			}
		case http.StatusConflict:
			tag = "⚠️  409 Conflict (duplicate in progress)"
			conflictCount++
		default:
			tag = fmt.Sprintf("❓ %d", r.status)
		}
		fmt.Printf("Request #%d -> %s | body: %s\n", r.id, tag, r.body)
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("First execution (business logic ran):  %d\n", successCount)
	fmt.Printf("Replayed (cached result returned):     %d\n", replayCount)
	fmt.Printf("409 Conflicts (duplicate in progress): %d\n", conflictCount)

	time.Sleep(35 * time.Second)

	fmt.Printf("Business logic was called %d time(s) — idempotency %s!\n",
		atomic.LoadInt32(&businessLogicCalls),
		func() string {
			if atomic.LoadInt32(&businessLogicCalls) == 1 {
				return " HOLDS"
			}
			return " VIOLATED"
		}(),
	)
}
