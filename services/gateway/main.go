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

// Configuration from environment variables
var (
	parserServiceURL         string
	additionServiceURL       string
	subtractionServiceURL    string
	multiplicationServiceURL string
	divisionServiceURL       string
	serviceName              string
	hostname                 string
	port                     string
)

func init() {
	parserServiceURL = getEnv("PARSER_SERVICE_URL", "http://parser-service:8081")
	additionServiceURL = getEnv("ADDITION_SERVICE_URL", "http://addition-service:8082")
	subtractionServiceURL = getEnv("SUBTRACTION_SERVICE_URL", "http://subtraction-service:8083")
	multiplicationServiceURL = getEnv("MULTIPLICATION_SERVICE_URL", "http://multiplication-service:8084")
	divisionServiceURL = getEnv("DIVISION_SERVICE_URL", "http://division-service:8086")
	serviceName = getEnv("SERVICE_NAME", "gateway-service")
	hostname = getEnv("HOSTNAME", getHostname())
	port = getEnv("PORT", "8080")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return h
}

// CalculateRequest represents the incoming calculation request
type CalculateRequest struct {
	Expression string `json:"expression"`
}

// CalculateResponse represents the final calculation response
type CalculateResponse struct {
	Result     float64                  `json:"result"`
	Expression string                   `json:"expression"`
	EventLog   []map[string]interface{} `json:"eventLog"`
	Metadata   Metadata                 `json:"metadata"`
}

// Metadata contains summary information about the calculation
type Metadata struct {
	TotalServices int `json:"totalServices"`
	TotalDuration int `json:"totalDuration"`
}

// ParseRequest is sent to the parser service
type ParseRequest struct {
	Expression string `json:"expression"`
}

// ParseResponse is received from the parser service
type ParseResponse struct {
	AST      map[string]interface{}   `json:"ast"`
	EventLog []map[string]interface{} `json:"eventLog"`
	Error    string                   `json:"error,omitempty"`
}

// ExecuteRequest is sent to operation services
type ExecuteRequest struct {
	Operation map[string]interface{} `json:"operation"`
}

// ExecuteResponse is received from operation services
type ExecuteResponse struct {
	Result   float64                  `json:"result"`
	EventLog []map[string]interface{} `json:"eventLog"`
	Error    string                   `json:"error,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// HTTP client with timeout
var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

// getOperatorServiceURL returns the service URL for a given operator
func getOperatorServiceURL(operator string) (string, error) {
	switch operator {
	case "+":
		return additionServiceURL, nil
	case "-":
		return subtractionServiceURL, nil
	case "*":
		return multiplicationServiceURL, nil
	case "/":
		return divisionServiceURL, nil
	default:
		return "", fmt.Errorf("unknown operator: %s", operator)
	}
}

// callParserService calls the parser service to convert expression to AST
func callParserService(expression string) (*ParseResponse, error) {
	reqBody := ParseRequest{Expression: expression}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal parse request: %w", err)
	}

	resp, err := httpClient.Post(parserServiceURL+"/parse", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call parser service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read parser response: %w", err)
	}

	var parseResp ParseResponse
	if err := json.Unmarshal(body, &parseResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal parser response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if parseResp.Error != "" {
			return nil, fmt.Errorf("parser error: %s", parseResp.Error)
		}
		return nil, fmt.Errorf("parser service returned status %d", resp.StatusCode)
	}

	return &parseResp, nil
}

// callOperationService calls the appropriate operation service to execute an operation
func callOperationService(operation map[string]interface{}) (*ExecuteResponse, error) {
	operator, ok := operation["operator"].(string)
	if !ok {
		return nil, fmt.Errorf("operation missing operator field")
	}

	serviceURL, err := getOperatorServiceURL(operator)
	if err != nil {
		return nil, err
	}

	reqBody := ExecuteRequest{Operation: operation}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal execute request: %w", err)
	}

	resp, err := httpClient.Post(serviceURL+"/execute", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call operation service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read operation response: %w", err)
	}

	var execResp ExecuteResponse
	if err := json.Unmarshal(body, &execResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal operation response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if execResp.Error != "" {
			return nil, fmt.Errorf("operation error: %s", execResp.Error)
		}
		return nil, fmt.Errorf("operation service returned status %d", resp.StatusCode)
	}

	return &execResp, nil
}

// evaluateAST evaluates the AST by calling appropriate operation services
func evaluateAST(ast map[string]interface{}) (*ExecuteResponse, error) {
	nodeType, ok := ast["type"].(string)
	if !ok {
		return nil, fmt.Errorf("AST node missing type field")
	}

	// If it's just a number, return it directly
	if nodeType == "number" {
		value, ok := ast["value"].(float64)
		if !ok {
			// Try int
			if intVal, ok := ast["value"].(int); ok {
				value = float64(intVal)
			} else if jsonNum, ok := ast["value"].(json.Number); ok {
				var err error
				value, err = jsonNum.Float64()
				if err != nil {
					return nil, fmt.Errorf("invalid number value in AST")
				}
			} else {
				return nil, fmt.Errorf("invalid number value in AST")
			}
		}
		return &ExecuteResponse{
			Result:   value,
			EventLog: []map[string]interface{}{},
		}, nil
	}

	// If it's an operation, call the appropriate service
	if nodeType == "operation" {
		return callOperationService(ast)
	}

	return nil, fmt.Errorf("unknown AST node type: %s", nodeType)
}

// countUniqueServices counts unique services in the event log
func countUniqueServices(eventLog []map[string]interface{}) int {
	services := make(map[string]bool)
	for _, event := range eventLog {
		if service, ok := event["service"].(string); ok {
			services[service] = true
		}
	}
	return len(services)
}

// calculateTotalDuration sums up all durations in the event log
func calculateTotalDuration(eventLog []map[string]interface{}) int {
	total := 0
	for _, event := range eventLog {
		if duration, ok := event["duration"].(float64); ok {
			total += int(duration)
		} else if duration, ok := event["duration"].(int); ok {
			total += duration
		}
	}
	return total
}

// calculateHandler handles POST /calculate requests
func calculateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON request body"})
		return
	}

	if req.Expression == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Missing 'expression' field in request"})
		return
	}

	// Step 1: Call parser service to get AST
	parseResp, err := callParserService(req.Expression)
	if err != nil {
		log.Printf("Parser error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	// Step 2: Evaluate the AST by calling operation services
	evalResp, err := evaluateAST(parseResp.AST)
	if err != nil {
		log.Printf("Evaluation error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: err.Error()})
		return
	}

	// Combine event logs from parser and evaluation
	allEventLogs := append(parseResp.EventLog, evalResp.EventLog...)

	// Build response
	response := CalculateResponse{
		Result:     evalResp.Result,
		Expression: req.Expression,
		EventLog:   allEventLogs,
		Metadata: Metadata{
			TotalServices: countUniqueServices(allEventLogs),
			TotalDuration: calculateTotalDuration(allEventLogs),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// healthHandler handles GET /health requests (liveness probe)
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "healthy",
		"service":  serviceName,
		"hostname": hostname,
	})
}

// readyHandler handles GET /ready requests (readiness probe)
func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "ready",
		"service":  serviceName,
		"hostname": hostname,
	})
}

// rootHandler handles GET / requests
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":  serviceName,
		"version":  "1.0.0",
		"hostname": hostname,
		"endpoints": map[string]string{
			"calculate": "POST /calculate",
			"health":    "GET /health",
			"ready":     "GET /ready",
		},
	})
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/calculate", calculateHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/ready", readyHandler)

	log.Printf("Starting %s on port %s", serviceName, port)
	log.Printf("Hostname: %s", hostname)
	log.Printf("Parser Service URL: %s", parserServiceURL)
	log.Printf("Addition Service URL: %s", additionServiceURL)
	log.Printf("Subtraction Service URL: %s", subtractionServiceURL)
	log.Printf("Multiplication Service URL: %s", multiplicationServiceURL)
	log.Printf("Division Service URL: %s", divisionServiceURL)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
