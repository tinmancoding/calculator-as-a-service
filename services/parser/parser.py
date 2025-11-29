"""
Expression Parser for Calculator-as-a-Service
Implements a recursive descent parser with operator precedence
"""

import re
from typing import Dict, Any


class ParserError(Exception):
    """Custom exception for parsing errors"""
    pass


class ExpressionParser:
    """
    Recursive descent parser for arithmetic expressions

    Grammar:
    expression  : term (('+' | '-') term)*
    term        : factor (('*' | '/') factor)*
    factor      : NUMBER | '(' expression ')'
    NUMBER      : [0-9]+('.'[0-9]+)?
    """

    def __init__(self, expression: str):
        self.expression = expression.strip()
        self.tokens = self._tokenize(expression)
        self.pos = 0

    def _tokenize(self, expression: str) -> list:
        """
        Tokenize the expression into numbers, operators, and parentheses
        """
        # Remove all whitespace
        expression = re.sub(r'\s+', '', expression)

        # Pattern for numbers (including decimals) and operators
        pattern = r'(\d+\.?\d*|[+\-*/()])'
        tokens = re.findall(pattern, expression)

        # Validate that we captured the entire expression
        reconstructed = ''.join(tokens)
        if reconstructed != expression:
            raise ParserError(f"Invalid characters in expression: {expression}")

        if not tokens:
            raise ParserError("Empty expression")

        return tokens

    def _current_token(self) -> str:
        """Get the current token without consuming it"""
        if self.pos < len(self.tokens):
            return self.tokens[self.pos]
        return None

    def _consume_token(self) -> str:
        """Consume and return the current token"""
        token = self._current_token()
        self.pos += 1
        return token

    def _is_number(self, token: str) -> bool:
        """Check if token is a number"""
        if token is None:
            return False
        try:
            float(token)
            return True
        except (ValueError, TypeError):
            return False

    def parse(self) -> Dict[str, Any]:
        """
        Parse the expression and return an AST
        """
        if not self.tokens:
            raise ParserError("Empty expression")

        ast = self._parse_expression()

        # Ensure we consumed all tokens
        if self.pos < len(self.tokens):
            raise ParserError(f"Unexpected token: {self._current_token()}")

        return ast

    def _parse_expression(self) -> Dict[str, Any]:
        """
        Parse an expression (handles + and - with lower precedence)
        expression : term (('+' | '-') term)*
        """
        left = self._parse_term()

        while self._current_token() in ['+', '-']:
            operator = self._consume_token()
            right = self._parse_term()
            left = {
                "type": "operation",
                "operator": operator,
                "left": left,
                "right": right
            }

        return left

    def _parse_term(self) -> Dict[str, Any]:
        """
        Parse a term (handles * and / with higher precedence)
        term : factor (('*' | '/') factor)*
        """
        left = self._parse_factor()

        while self._current_token() in ['*', '/']:
            operator = self._consume_token()
            right = self._parse_factor()
            left = {
                "type": "operation",
                "operator": operator,
                "left": left,
                "right": right
            }

        return left

    def _parse_factor(self) -> Dict[str, Any]:
        """
        Parse a factor (handles numbers and parenthesized expressions)
        factor : NUMBER | '(' expression ')'
        """
        token = self._current_token()

        if token is None:
            raise ParserError("Unexpected end of expression")

        # Handle parentheses
        if token == '(':
            self._consume_token()  # consume '('
            expr = self._parse_expression()
            if self._current_token() != ')':
                raise ParserError("Missing closing parenthesis")
            self._consume_token()  # consume ')'
            return expr

        # Handle numbers
        if self._is_number(token):
            self._consume_token()
            # Convert to float, but if it's a whole number, convert to int
            value = float(token)
            if value.is_integer():
                value = int(value)
            return {
                "type": "number",
                "value": value
            }

        # Handle invalid tokens
        raise ParserError(f"Unexpected token: {token}")


def parse_expression(expression: str) -> Dict[str, Any]:
    """
    Parse an arithmetic expression and return its AST

    Args:
        expression: String containing arithmetic expression

    Returns:
        Dictionary representing the AST

    Raises:
        ParserError: If the expression is invalid
    """
    parser = ExpressionParser(expression)
    return parser.parse()
