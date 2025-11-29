import express, { Request, Response } from 'express';
import axios from 'axios';
import * as os from 'os';

const app = express();
app.use(express.json());

// Service configuration
const SERVICE_NAME = process.env.SERVICE_NAME || 'subtraction-service';
const HOSTNAME = process.env.HOSTNAME || os.hostname();
const PORT = parseInt(process.env.PORT || '8083', 10);

// Other service URLs for delegation
const ADDITION_SERVICE_URL = process.env.ADDITION_SERVICE_URL || 'http://addition-service:8082';
const SUBTRACTION_SERVICE_URL = process.env.SUBTRACTION_SERVICE_URL || 'http://subtraction-service:8083';
const MULTIPLICATION_SERVICE_URL = process.env.MULTIPLICATION_SERVICE_URL || 'http://multiplication-service:8084';
const DIVISION_SERVICE_URL = process.env.DIVISION_SERVICE_URL || 'http://division-service:8086';

// Map operators to service URLs
const OPERATOR_SERVICE_MAP: Record<string, string> = {
  '+': ADDITION_SERVICE_URL,
  '-': SUBTRACTION_SERVICE_URL,
  '*': MULTIPLICATION_SERVICE_URL,
  '/': DIVISION_SERVICE_URL,
};

// Types
interface NumberNode {
  type: 'number';
  value: number;
}

interface OperationNode {
  type: 'operation';
  operator: string;
  left: ASTNode | number;
  right: ASTNode | number;
}

type ASTNode = NumberNode | OperationNode;

interface Delegation {
  service: string;
  hostname: string;
  operation: string;
  result: number;
}

interface EventLogEntry {
  timestamp: string;
  hostname: string;
  service: string;
  operation: string;
  operands: {
    left: number;
    right: number;
  };
  result: number;
  delegations: {
    left: Delegation | null;
    right: Delegation | null;
  };
  duration: number;
}

interface EvaluationResult {
  value: number;
  delegation: Delegation | null;
  eventLogs: EventLogEntry[];
}

interface ExecuteResponse {
  result: number;
  eventLog: EventLogEntry[];
}

/**
 * Get current timestamp in ISO 8601 format.
 */
function getTimestamp(): string {
  return new Date().toISOString();
}

/**
 * Evaluate an operand which can be either a number or a nested operation.
 */
async function evaluateOperand(operand: ASTNode | number): Promise<EvaluationResult> {
  // If operand is a simple number
  if (typeof operand === 'number') {
    return {
      value: operand,
      delegation: null,
      eventLogs: [],
    };
  }

  // If operand is a dict with type "number"
  if (typeof operand === 'object' && operand.type === 'number') {
    return {
      value: operand.value,
      delegation: null,
      eventLogs: [],
    };
  }

  // If operand is an operation node, delegate to appropriate service
  if (typeof operand === 'object' && operand.type === 'operation') {
    const operator = operand.operator;
    const serviceUrl = OPERATOR_SERVICE_MAP[operator];

    if (!serviceUrl) {
      throw new Error(`Unknown operator: ${operator}`);
    }

    // Call the appropriate service
    const response = await axios.post<ExecuteResponse>(
      `${serviceUrl}/execute`,
      { operation: operand },
      { timeout: 30000 }
    );

    const result = response.data;
    const eventLogs = result.eventLog || [];
    const lastEvent = eventLogs.length > 0 ? eventLogs[eventLogs.length - 1] : null;

    return {
      value: result.result,
      delegation: {
        service: lastEvent?.service || 'unknown-service',
        hostname: lastEvent?.hostname || 'unknown-hostname',
        operation: operator,
        result: result.result,
      },
      eventLogs,
    };
  }

  // Handle unexpected operand format
  throw new Error(`Invalid operand format: ${JSON.stringify(operand)}`);
}

/**
 * Execute a subtraction operation.
 */
app.post('/execute', async (req: Request, res: Response) => {
  const startTime = Date.now();

  try {
    const data = req.body;

    if (!data || !data.operation) {
      res.status(400).json({ error: "Missing 'operation' field" });
      return;
    }

    const operation = data.operation;

    // Validate operation
    if (typeof operation !== 'object') {
      res.status(400).json({ error: 'Invalid operation format' });
      return;
    }

    const operator = operation.operator;
    if (operator !== '-') {
      res.status(400).json({ error: `This service only handles subtraction (-), got: ${operator}` });
      return;
    }

    const leftOperand = operation.left;
    const rightOperand = operation.right;

    if (leftOperand === undefined || leftOperand === null || rightOperand === undefined || rightOperand === null) {
      res.status(400).json({ error: 'Missing left or right operand' });
      return;
    }

    // Evaluate both operands (may involve delegation)
    const leftEval = await evaluateOperand(leftOperand);
    const rightEval = await evaluateOperand(rightOperand);

    // Perform the subtraction
    const result = leftEval.value - rightEval.value;

    // Calculate duration
    const duration = Date.now() - startTime;

    // Build our event log entry
    const myEvent: EventLogEntry = {
      timestamp: getTimestamp(),
      hostname: HOSTNAME,
      service: SERVICE_NAME,
      operation: '-',
      operands: {
        left: leftEval.value,
        right: rightEval.value,
      },
      result,
      delegations: {
        left: leftEval.delegation,
        right: rightEval.delegation,
      },
      duration,
    };

    // Combine all event logs in chronological order
    const allEvents = [...leftEval.eventLogs, ...rightEval.eventLogs, myEvent];

    res.json({
      result,
      eventLog: allEvents,
    });
  } catch (error) {
    if (axios.isAxiosError(error)) {
      res.status(502).json({ error: `Service delegation failed: ${error.message}` });
      return;
    }
    if (error instanceof Error) {
      res.status(400).json({ error: error.message });
      return;
    }
    res.status(500).json({ error: 'Internal error' });
  }
});

/**
 * Liveness probe endpoint.
 */
app.get('/health', (_req: Request, res: Response) => {
  res.json({ status: 'healthy', service: SERVICE_NAME });
});

/**
 * Readiness probe endpoint.
 */
app.get('/ready', (_req: Request, res: Response) => {
  res.json({ status: 'ready', service: SERVICE_NAME });
});

// Start the server
app.listen(PORT, '0.0.0.0', () => {
  console.log(`${SERVICE_NAME} listening on port ${PORT}`);
});
