"""
Parser Service - Flask Application
Provides /parse endpoint for converting arithmetic expressions to AST
"""

import os
import socket
from datetime import datetime
from time import time
from flask import Flask, request, jsonify
from parser import parse_expression, ParserError

app = Flask(__name__)

# Configuration
SERVICE_NAME = os.getenv('SERVICE_NAME', 'parser-service')
PORT = int(os.getenv('PORT', '8081'))


def get_hostname():
    """Get the hostname (pod name in Kubernetes)"""
    return os.getenv('HOSTNAME', socket.gethostname())


def get_iso_timestamp():
    """Get current timestamp in ISO 8601 format"""
    return datetime.utcnow().isoformat() + 'Z'


def create_event_log(operation: str, input_expr: str, result: str, duration: int):
    """
    Create an event log entry

    Args:
        operation: Type of operation performed
        input_expr: Input expression
        result: Result description
        duration: Duration in milliseconds

    Returns:
        Dictionary representing the event log entry
    """
    return {
        "timestamp": get_iso_timestamp(),
        "hostname": get_hostname(),
        "service": SERVICE_NAME,
        "operation": operation,
        "input": input_expr,
        "result": result,
        "duration": duration
    }


@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint for liveness probe"""
    return jsonify({
        "status": "healthy",
        "service": SERVICE_NAME,
        "hostname": get_hostname()
    }), 200


@app.route('/ready', methods=['GET'])
def ready():
    """Readiness check endpoint"""
    return jsonify({
        "status": "ready",
        "service": SERVICE_NAME,
        "hostname": get_hostname()
    }), 200


@app.route('/parse', methods=['POST'])
def parse():
    """
    Parse endpoint - converts arithmetic expression to AST

    Request:
        {
            "expression": "2 + 3 * 4"
        }

    Response:
        {
            "ast": { ... },
            "eventLog": [ ... ]
        }

    Error Response:
        {
            "error": "Error message",
            "eventLog": [ ... ]
        }
    """
    start_time = time()

    try:
        # Validate request
        if not request.is_json:
            return jsonify({
                "error": "Content-Type must be application/json"
            }), 400

        data = request.get_json()

        if not data:
            return jsonify({
                "error": "Request body is required"
            }), 400

        expression = data.get('expression')

        if not expression:
            return jsonify({
                "error": "Missing 'expression' field in request"
            }), 400

        if not isinstance(expression, str):
            return jsonify({
                "error": "'expression' must be a string"
            }), 400

        # Parse the expression
        try:
            ast = parse_expression(expression)
        except ParserError as e:
            duration = int((time() - start_time) * 1000)
            event_log = create_event_log(
                operation="parse",
                input_expr=expression,
                result=f"Parse error: {str(e)}",
                duration=duration
            )

            return jsonify({
                "error": str(e),
                "eventLog": [event_log]
            }), 400

        # Calculate duration
        duration = int((time() - start_time) * 1000)

        # Create event log
        event_log = create_event_log(
            operation="parse",
            input_expr=expression,
            result="AST generated",
            duration=duration
        )

        # Return successful response
        return jsonify({
            "ast": ast,
            "eventLog": [event_log]
        }), 200

    except Exception as e:
        # Handle unexpected errors
        duration = int((time() - start_time) * 1000)
        event_log = create_event_log(
            operation="parse",
            input_expr=data.get('expression', 'unknown') if 'data' in locals() else 'unknown',
            result=f"Internal error: {str(e)}",
            duration=duration
        )

        return jsonify({
            "error": f"Internal server error: {str(e)}",
            "eventLog": [event_log]
        }), 500


@app.route('/', methods=['GET'])
def root():
    """Root endpoint - service information"""
    return jsonify({
        "service": SERVICE_NAME,
        "version": "1.0.0",
        "hostname": get_hostname(),
        "endpoints": {
            "parse": "POST /parse",
            "health": "GET /health",
            "ready": "GET /ready"
        }
    }), 200


if __name__ == '__main__':
    print(f"Starting {SERVICE_NAME} on port {PORT}")
    print(f"Hostname: {get_hostname()}")
    app.run(host='0.0.0.0', port=PORT, debug=False)
