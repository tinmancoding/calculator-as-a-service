import os
import socket
import time
from datetime import datetime, timezone

import requests
from flask import Flask, jsonify, request

app = Flask(__name__)

# Service configuration
SERVICE_NAME = os.environ.get("SERVICE_NAME", "addition-service")
HOSTNAME = os.environ.get("HOSTNAME", socket.gethostname())

# Other service URLs for delegation
ADDITION_SERVICE_URL = os.environ.get("ADDITION_SERVICE_URL", "http://addition-service:8082")
SUBTRACTION_SERVICE_URL = os.environ.get("SUBTRACTION_SERVICE_URL", "http://subtraction-service:8083")
MULTIPLICATION_SERVICE_URL = os.environ.get("MULTIPLICATION_SERVICE_URL", "http://multiplication-service:8084")
DIVISION_SERVICE_URL = os.environ.get("DIVISION_SERVICE_URL", "http://division-service:8086")

# Map operators to service URLs
OPERATOR_SERVICE_MAP = {
    "+": ADDITION_SERVICE_URL,
    "-": SUBTRACTION_SERVICE_URL,
    "*": MULTIPLICATION_SERVICE_URL,
    "/": DIVISION_SERVICE_URL,
}


def get_timestamp():
    """Get current timestamp in ISO 8601 format."""
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%S.%f")[:-3] + "Z"


def evaluate_operand(operand):
    """
    Evaluate an operand which can be either a number or a nested operation.
    
    Returns a dict with:
        - value: the numeric result
        - delegation: delegation info if delegated, None otherwise
        - event_logs: list of event logs from delegated operations
    """
    # If operand is a simple number
    if isinstance(operand, (int, float)):
        return {
            "value": operand,
            "delegation": None,
            "event_logs": []
        }
    
    # If operand is a dict with type "number"
    if isinstance(operand, dict) and operand.get("type") == "number":
        return {
            "value": operand.get("value"),
            "delegation": None,
            "event_logs": []
        }
    
    # If operand is an operation node, delegate to appropriate service
    if isinstance(operand, dict) and operand.get("type") == "operation":
        operator = operand.get("operator")
        service_url = OPERATOR_SERVICE_MAP.get(operator)
        
        if not service_url:
            raise ValueError(f"Unknown operator: {operator}")
        
        # Call the appropriate service
        response = requests.post(
            f"{service_url}/execute",
            json={"operation": operand},
            timeout=30
        )
        response.raise_for_status()
        result = response.json()
        
        # Extract the last event log entry to get hostname/service info
        event_logs = result.get("eventLog", [])
        last_event = event_logs[-1] if event_logs else {}
        
        return {
            "value": result.get("result"),
            "delegation": {
                "service": last_event.get("service", "unknown-service"),
                "hostname": last_event.get("hostname", "unknown-hostname"),
                "operation": operator,
                "result": result.get("result")
            },
            "event_logs": event_logs
        }
    
    # Handle unexpected operand format
    raise ValueError(f"Invalid operand format: {operand}")


@app.route("/execute", methods=["POST"])
def execute():
    """Execute an addition operation."""
    start_time = time.time()
    
    try:
        data = request.get_json()
        
        if not data or "operation" not in data:
            return jsonify({"error": "Missing 'operation' field"}), 400
        
        operation = data["operation"]
        
        # Validate operation
        if not isinstance(operation, dict):
            return jsonify({"error": "Invalid operation format"}), 400
        
        operator = operation.get("operator")
        if operator != "+":
            return jsonify({"error": f"This service only handles addition (+), got: {operator}"}), 400
        
        left_operand = operation.get("left")
        right_operand = operation.get("right")
        
        if left_operand is None or right_operand is None:
            return jsonify({"error": "Missing left or right operand"}), 400
        
        # Evaluate both operands (may involve delegation)
        left_eval = evaluate_operand(left_operand)
        right_eval = evaluate_operand(right_operand)
        
        # Perform the addition
        result = left_eval["value"] + right_eval["value"] + 1
        
        # Calculate duration
        duration = int((time.time() - start_time) * 1000)
        
        # Build our event log entry
        my_event = {
            "timestamp": get_timestamp(),
            "hostname": HOSTNAME,
            "service": SERVICE_NAME,
            "operation": "+",
            "operands": {
                "left": left_eval["value"],
                "right": right_eval["value"]
            },
            "result": result,
            "delegations": {
                "left": left_eval["delegation"],
                "right": right_eval["delegation"]
            },
            "duration": duration
        }
        
        # Combine all event logs in chronological order
        all_events = left_eval["event_logs"] + right_eval["event_logs"] + [my_event]
        
        return jsonify({
            "result": result,
            "eventLog": all_events
        })
        
    except requests.RequestException as e:
        return jsonify({"error": f"Service delegation failed: {str(e)}"}), 502
    except ValueError as e:
        return jsonify({"error": str(e)}), 400
    except Exception as e:
        return jsonify({"error": f"Internal error: {str(e)}"}), 500


@app.route("/health", methods=["GET"])
def health():
    """Liveness probe endpoint."""
    return jsonify({"status": "healthy", "service": SERVICE_NAME})


@app.route("/ready", methods=["GET"])
def ready():
    """Readiness probe endpoint."""
    return jsonify({"status": "ready", "service": SERVICE_NAME})


if __name__ == "__main__":
    port = int(os.environ.get("PORT", 8082))
    app.run(host="0.0.0.0", port=port, debug=False)
