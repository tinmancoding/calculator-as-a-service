package com.calculator.division;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.client.RestTemplate;

import java.net.InetAddress;
import java.net.UnknownHostException;
import java.time.Instant;
import java.time.format.DateTimeFormatter;
import java.util.HashMap;
import java.util.Map;

@RestController
public class DivisionController {

    private static final double ZERO_TOLERANCE = 1e-10;
    
    private final ObjectMapper objectMapper = new ObjectMapper();
    private final RestTemplate restTemplate;

    public DivisionController(RestTemplate restTemplate) {
        this.restTemplate = restTemplate;
    }

    @Value("${SERVICE_NAME:division-service}")
    private String serviceName;

    @Value("${ADDITION_SERVICE_URL:http://addition-service:8082}")
    private String additionServiceUrl;

    @Value("${SUBTRACTION_SERVICE_URL:http://subtraction-service:8083}")
    private String subtractionServiceUrl;

    @Value("${MULTIPLICATION_SERVICE_URL:http://multiplication-service:8084}")
    private String multiplicationServiceUrl;

    @Value("${DIVISION_SERVICE_URL:http://division-service:8086}")
    private String divisionServiceUrl;

    private String getHostname() {
        String hostname = System.getenv("HOSTNAME");
        if (hostname != null && !hostname.isEmpty()) {
            return hostname;
        }
        try {
            return InetAddress.getLocalHost().getHostName();
        } catch (UnknownHostException e) {
            return "unknown-host";
        }
    }

    private String getTimestamp() {
        return DateTimeFormatter.ISO_INSTANT.format(Instant.now());
    }

    private String getServiceUrlForOperator(String operator) {
        return switch (operator) {
            case "+" -> additionServiceUrl;
            case "-" -> subtractionServiceUrl;
            case "*" -> multiplicationServiceUrl;
            case "/" -> divisionServiceUrl;
            default -> throw new IllegalArgumentException("Unknown operator: " + operator);
        };
    }

    /**
     * Result of evaluating an operand.
     */
    private static class EvaluationResult {
        double value;
        ObjectNode delegation;
        ArrayNode eventLogs;

        EvaluationResult(double value, ObjectNode delegation, ArrayNode eventLogs) {
            this.value = value;
            this.delegation = delegation;
            this.eventLogs = eventLogs;
        }
    }

    /**
     * Evaluate an operand which can be either a number or a nested operation.
     */
    private EvaluationResult evaluateOperand(JsonNode operand) {
        ArrayNode emptyLogs = objectMapper.createArrayNode();

        // If operand is a simple number (numeric JSON value)
        if (operand.isNumber()) {
            return new EvaluationResult(operand.asDouble(), null, emptyLogs);
        }

        // If operand is a dict with type "number"
        if (operand.isObject() && "number".equals(operand.path("type").asText())) {
            return new EvaluationResult(operand.path("value").asDouble(), null, emptyLogs);
        }

        // If operand is an operation node, delegate to appropriate service
        if (operand.isObject() && "operation".equals(operand.path("type").asText())) {
            String operator = operand.path("operator").asText();
            String serviceUrl = getServiceUrlForOperator(operator);

            // Build request body
            ObjectNode requestBody = objectMapper.createObjectNode();
            requestBody.set("operation", operand);

            // Call the appropriate service
            ResponseEntity<JsonNode> response = restTemplate.postForEntity(
                    serviceUrl + "/execute",
                    requestBody,
                    JsonNode.class
            );

            JsonNode responseBody = response.getBody();
            if (responseBody == null) {
                throw new RuntimeException("Empty response from service: " + serviceUrl);
            }

            double result = responseBody.path("result").asDouble();
            ArrayNode eventLogs = (ArrayNode) responseBody.path("eventLog");

            // Extract the last event log entry to get hostname/service info
            ObjectNode delegation = null;
            if (eventLogs != null && !eventLogs.isEmpty()) {
                JsonNode lastEvent = eventLogs.get(eventLogs.size() - 1);
                delegation = objectMapper.createObjectNode();
                delegation.put("service", lastEvent.path("service").asText("unknown-service"));
                delegation.put("hostname", lastEvent.path("hostname").asText("unknown-hostname"));
                delegation.put("operation", operator);
                delegation.put("result", result);
            }

            return new EvaluationResult(result, delegation, eventLogs != null ? eventLogs : emptyLogs);
        }

        throw new IllegalArgumentException("Invalid operand format: " + operand);
    }

    @PostMapping("/execute")
    public ResponseEntity<ObjectNode> execute(@RequestBody JsonNode request) {
        long startTime = System.currentTimeMillis();

        try {
            JsonNode operation = request.path("operation");

            if (operation.isMissingNode()) {
                ObjectNode error = objectMapper.createObjectNode();
                error.put("error", "Missing 'operation' field");
                return ResponseEntity.badRequest().body(error);
            }

            if (!operation.isObject()) {
                ObjectNode error = objectMapper.createObjectNode();
                error.put("error", "Invalid operation format");
                return ResponseEntity.badRequest().body(error);
            }

            String operator = operation.path("operator").asText();
            if (!"/".equals(operator)) {
                ObjectNode error = objectMapper.createObjectNode();
                error.put("error", "This service only handles division (/), got: " + operator);
                return ResponseEntity.badRequest().body(error);
            }

            JsonNode leftOperand = operation.path("left");
            JsonNode rightOperand = operation.path("right");

            if (leftOperand.isMissingNode() || rightOperand.isMissingNode()) {
                ObjectNode error = objectMapper.createObjectNode();
                error.put("error", "Missing left or right operand");
                return ResponseEntity.badRequest().body(error);
            }

            // Evaluate both operands (may involve delegation)
            EvaluationResult leftEval = evaluateOperand(leftOperand);
            EvaluationResult rightEval = evaluateOperand(rightOperand);

            // Check for division by zero (using tolerance for floating-point comparison)
            if (Math.abs(rightEval.value) < ZERO_TOLERANCE) {
                ObjectNode error = objectMapper.createObjectNode();
                error.put("error", "Division by zero");
                return ResponseEntity.badRequest().body(error);
            }

            // Perform the division
            double result = leftEval.value / rightEval.value;

            // Calculate duration
            long duration = System.currentTimeMillis() - startTime;

            // Build our event log entry
            ObjectNode myEvent = objectMapper.createObjectNode();
            myEvent.put("timestamp", getTimestamp());
            myEvent.put("hostname", getHostname());
            myEvent.put("service", serviceName);
            myEvent.put("operation", "/");

            ObjectNode operands = objectMapper.createObjectNode();
            operands.put("left", leftEval.value);
            operands.put("right", rightEval.value);
            myEvent.set("operands", operands);

            myEvent.put("result", result);

            ObjectNode delegations = objectMapper.createObjectNode();
            if (leftEval.delegation != null) {
                delegations.set("left", leftEval.delegation);
            } else {
                delegations.putNull("left");
            }
            if (rightEval.delegation != null) {
                delegations.set("right", rightEval.delegation);
            } else {
                delegations.putNull("right");
            }
            myEvent.set("delegations", delegations);

            myEvent.put("duration", duration);

            // Combine all event logs in chronological order
            ArrayNode allEvents = objectMapper.createArrayNode();
            allEvents.addAll(leftEval.eventLogs);
            allEvents.addAll(rightEval.eventLogs);
            allEvents.add(myEvent);

            // Build response
            ObjectNode response = objectMapper.createObjectNode();
            response.put("result", result);
            response.set("eventLog", allEvents);

            return ResponseEntity.ok(response);

        } catch (IllegalArgumentException e) {
            ObjectNode error = objectMapper.createObjectNode();
            error.put("error", e.getMessage());
            return ResponseEntity.badRequest().body(error);
        } catch (Exception e) {
            ObjectNode error = objectMapper.createObjectNode();
            error.put("error", "Service delegation failed: " + e.getMessage());
            return ResponseEntity.status(HttpStatus.BAD_GATEWAY).body(error);
        }
    }

    @GetMapping("/health")
    public ResponseEntity<Map<String, String>> health() {
        Map<String, String> response = new HashMap<>();
        response.put("status", "healthy");
        response.put("service", serviceName);
        return ResponseEntity.ok(response);
    }

    @GetMapping("/ready")
    public ResponseEntity<Map<String, String>> ready() {
        Map<String, String> response = new HashMap<>();
        response.put("status", "ready");
        response.put("service", serviceName);
        return ResponseEntity.ok(response);
    }
}
