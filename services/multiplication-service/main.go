package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// Service configuration
var (
	serviceName string
	hostname    string

	additionServiceURL       string
	subtractionServiceURL    string
	multiplicationServiceURL string
	divisionServiceURL       string
)

// OperatorServiceMap maps operators to their service URLs
var operatorServiceMap map[string]string

// Operation represents an AST operation node
type Operation struct {
	Type     string      `json:"type"`
	Operator string      `json:"operator"`
	Left     interface{} `json:"left"`
	Right    interface{} `json:"right"`
}

// ExecuteRequest is the request body for /execute
type ExecuteRequest struct {
	Operation Operation `json:"operation"`
}

// Delegation represents delegation info for an operand
type Delegation struct {
	Service   string `json:"service"`
	Hostname  string `json:"hostname"`
	Operation string `json:"operation"`
	Result    float64 `json:"result"`
}

// Operands represents the operand values
type Operands struct {
	Left  float64 `json:"left"`
	Right float64 `json:"right"`
}

// Delegations represents delegation info for both operands
type Delegations struct {
	Left  *Delegation `json:"left"`
	Right *Delegation `json:"right"`
}

// EventLogEntry represents a single event log entry
type EventLogEntry struct {
	Timestamp   string      `json:"timestamp"`
	Hostname    string      `json:"hostname"`
	Service     string      `json:"service"`
	Operation   string      `json:"operation"`
	Operands    Operands    `json:"operands"`
	Result      float64     `json:"result"`
	Delegations Delegations `json:"delegations"`
	Duration    int64       `json:"duration"`
}

// ExecuteResponse is the response body for /execute
type ExecuteResponse struct {
	Result   float64         `json:"result"`
	EventLog []EventLogEntry `json:"eventLog"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// OperandEvaluation represents the result of evaluating an operand
type OperandEvaluation struct {
	Value      float64
	Delegation *Delegation
	EventLogs  []EventLogEntry
}

func init() {
	serviceName = getEnv("SERVICE_NAME", "multiplication-service")
	hostname = getHostname()

	additionServiceURL = getEnv("ADDITION_SERVICE_URL", "http://addition-service:8082")
	subtractionServiceURL = getEnv("SUBTRACTION_SERVICE_URL", "http://subtraction-service:8083")
	multiplicationServiceURL = getEnv("MULTIPLICATION_SERVICE_URL", "http://multiplication-service:8084")
	divisionServiceURL = getEnv("DIVISION_SERVICE_URL", "http://division-service:8086")

	operatorServiceMap = map[string]string{
		"+": additionServiceURL,
		"-": subtractionServiceURL,
		"*": multiplicationServiceURL,
		"/": divisionServiceURL,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getHostname() string {
	if h := os.Getenv("HOSTNAME"); h != "" {
		return h
	}
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

func getTimestamp() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

// evaluateOperand evaluates an operand which can be either a number or a nested operation
func evaluateOperand(operand interface{}) (*OperandEvaluation, error) {
	// If operand is a simple number (float64 from JSON)
	if num, ok := operand.(float64); ok {
		return &OperandEvaluation{
			Value:      num,
			Delegation: nil,
			EventLogs:  []EventLogEntry{},
		}, nil
	}

	// If operand is a map (dict in JSON)
	if opMap, ok := operand.(map[string]interface{}); ok {
		opType, _ := opMap["type"].(string)

		// If it's a number type
		if opType == "number" {
			value, _ := opMap["value"].(float64)
			return &OperandEvaluation{
				Value:      value,
				Delegation: nil,
				EventLogs:  []EventLogEntry{},
			}, nil
		}

		// If it's an operation type, delegate to appropriate service
		if opType == "operation" {
			operator, _ := opMap["operator"].(string)
			serviceURL, exists := operatorServiceMap[operator]
			if !exists {
				return nil, fmt.Errorf("unknown operator: %s", operator)
			}

			// Call the appropriate service
			requestBody, err := json.Marshal(map[string]interface{}{
				"operation": opMap,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request: %w", err)
			}

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Post(serviceURL+"/execute", "application/json", bytes.NewBuffer(requestBody))
			if err != nil {
				return nil, fmt.Errorf("service delegation failed: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				var errResp ErrorResponse
				if err := json.Unmarshal(body, &errResp); err == nil {
					return nil, fmt.Errorf("service returned error: %s", errResp.Error)
				}
				return nil, fmt.Errorf("service returned status %d", resp.StatusCode)
			}

			var result ExecuteResponse
			if err := json.Unmarshal(body, &result); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}

			// Extract the last event log entry to get hostname/service info
			var lastEvent EventLogEntry
			if len(result.EventLog) > 0 {
				lastEvent = result.EventLog[len(result.EventLog)-1]
			}

			return &OperandEvaluation{
				Value: result.Result,
				Delegation: &Delegation{
					Service:   lastEvent.Service,
					Hostname:  lastEvent.Hostname,
					Operation: operator,
					Result:    result.Result,
				},
				EventLogs: result.EventLog,
			}, nil
		}
	}

	return nil, fmt.Errorf("invalid operand format")
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	if r.Method != http.MethodPost {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}

	operation := req.Operation
	if operation.Operator != "*" {
		sendError(w, fmt.Sprintf("This service only handles multiplication (*), got: %s", operation.Operator), http.StatusBadRequest)
		return
	}

	// Evaluate both operands (may involve delegation)
	leftEval, err := evaluateOperand(operation.Left)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadGateway)
		return
	}

	rightEval, err := evaluateOperand(operation.Right)
	if err != nil {
		sendError(w, err.Error(), http.StatusBadGateway)
		return
	}

	// Perform the multiplication
	result := leftEval.Value * rightEval.Value

	// Calculate duration in milliseconds
	duration := time.Since(startTime).Milliseconds()

	// Build our event log entry
	myEvent := EventLogEntry{
		Timestamp: getTimestamp(),
		Hostname:  hostname,
		Service:   serviceName,
		Operation: "*",
		Operands: Operands{
			Left:  leftEval.Value,
			Right: rightEval.Value,
		},
		Result: result,
		Delegations: Delegations{
			Left:  leftEval.Delegation,
			Right: rightEval.Delegation,
		},
		Duration: duration,
	}

	// Combine all event logs in chronological order
	allEvents := append(leftEval.EventLogs, rightEval.EventLogs...)
	allEvents = append(allEvents, myEvent)

	response := ExecuteResponse{
		Result:   result,
		EventLog: allEvents,
	}

	sendJSON(w, response, http.StatusOK)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:  "healthy",
		Service: serviceName,
	}
	sendJSON(w, response, http.StatusOK)
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HealthResponse{
		Status:  "ready",
		Service: serviceName,
	}
	sendJSON(w, response, http.StatusOK)
}

func sendJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func sendError(w http.ResponseWriter, message string, statusCode int) {
	sendJSON(w, ErrorResponse{Error: message}, statusCode)
}

func main() {
	port := getEnv("PORT", "8084")

	http.HandleFunc("/execute", executeHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)

	log.Printf("Starting %s on port %s", serviceName, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
