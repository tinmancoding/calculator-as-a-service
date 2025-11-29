package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", response["status"])
	}
}

func TestReadyHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/ready", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(readyHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "ready" {
		t.Errorf("Expected status 'ready', got '%s'", response["status"])
	}
}

func TestRootHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(rootHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["service"] != serviceName {
		t.Errorf("Expected service '%s', got '%s'", serviceName, response["service"])
	}

	if response["version"] != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response["version"])
	}

	endpoints, ok := response["endpoints"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'endpoints' to be a map")
	}

	if endpoints["calculate"] != "POST /calculate" {
		t.Errorf("Expected calculate endpoint 'POST /calculate', got '%s'", endpoints["calculate"])
	}
}

func TestCalculateHandler_MethodNotAllowed(t *testing.T) {
	req, err := http.NewRequest("GET", "/calculate", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(calculateHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestCalculateHandler_InvalidJSON(t *testing.T) {
	req, err := http.NewRequest("POST", "/calculate", bytes.NewBufferString("invalid json"))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(calculateHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestCalculateHandler_MissingExpression(t *testing.T) {
	reqBody := map[string]string{}
	body, _ := json.Marshal(reqBody)
	
	req, err := http.NewRequest("POST", "/calculate", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(calculateHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	var response ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Error != "Missing 'expression' field in request" {
		t.Errorf("Expected error about missing expression, got: '%s'", response.Error)
	}
}

func TestGetOperatorServiceURL(t *testing.T) {
	tests := []struct {
		operator string
		expected string
		hasError bool
	}{
		{"+", additionServiceURL, false},
		{"-", subtractionServiceURL, false},
		{"*", multiplicationServiceURL, false},
		{"/", divisionServiceURL, false},
		{"%", "", true},
	}

	for _, test := range tests {
		url, err := getOperatorServiceURL(test.operator)
		if test.hasError && err == nil {
			t.Errorf("Expected error for operator '%s', got none", test.operator)
		}
		if !test.hasError && err != nil {
			t.Errorf("Unexpected error for operator '%s': %v", test.operator, err)
		}
		if !test.hasError && url != test.expected {
			t.Errorf("Expected URL '%s' for operator '%s', got '%s'", test.expected, test.operator, url)
		}
	}
}

func TestCountUniqueServices(t *testing.T) {
	eventLog := []map[string]interface{}{
		{"service": "parser-service"},
		{"service": "addition-service"},
		{"service": "multiplication-service"},
		{"service": "addition-service"}, // duplicate
	}

	count := countUniqueServices(eventLog)
	if count != 3 {
		t.Errorf("Expected 3 unique services, got %d", count)
	}
}

func TestCalculateTotalDuration(t *testing.T) {
	eventLog := []map[string]interface{}{
		{"service": "parser-service", "duration": float64(5)},
		{"service": "addition-service", "duration": float64(3)},
		{"service": "multiplication-service", "duration": float64(2)},
	}

	total := calculateTotalDuration(eventLog)
	if total != 10 {
		t.Errorf("Expected total duration 10, got %d", total)
	}
}

func TestEvaluateAST_NumberNode(t *testing.T) {
	ast := map[string]interface{}{
		"type":  "number",
		"value": float64(42),
	}

	resp, err := evaluateAST(ast)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp.Result != 42 {
		t.Errorf("Expected result 42, got %f", resp.Result)
	}

	if len(resp.EventLog) != 0 {
		t.Errorf("Expected empty event log for number node, got %d events", len(resp.EventLog))
	}
}
