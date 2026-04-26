package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"base":"USD","target":"EUR","rate":0.85}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	rate, err := svc.GetRate("USD", "EUR")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if rate != 0.85 {
		t.Errorf("expected rate 0.85, got %f", rate)
	}
}

func TestGetRate_APIBusinessError_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"invalid currency pair"}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "XYZ")
	if err == nil {
		t.Fatal("expected error for 404 with error message, got nil")
	}
}

func TestGetRate_APIBusinessError_400(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid currency pair"}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("INVALID", "EUR")
	if err == nil {
		t.Fatal("expected error for 400 with error message, got nil")
	}
}

func TestGetRate_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected decode error for malformed JSON, got nil")
	}
}

func TestGetRate_SlowResponseTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	svc.Client.Timeout = 100 * time.Millisecond

	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected timeout/network error, got nil")
	}
}

func TestGetRate_ServerPanic500(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

func TestGetRate_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// write nothing — empty body
	}))
	defer server.Close()

	svc := NewExchangeService(server.URL)
	_, err := svc.GetRate("USD", "EUR")
	if err == nil {
		t.Fatal("expected decode error for empty body, got nil")
	}
}
