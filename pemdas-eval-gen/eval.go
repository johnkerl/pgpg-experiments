package main

import (
	"fmt"
	"math"
	"strconv"

	"github.com/johnkerl/pgpg/go/lib/pkg/asts"
)

// evaluateAST walks the AST and returns the result.
func evaluateAST(ast *asts.AST, verbose bool) (int, error) {
	var zero int
	if verbose {
		ast.Print()
	}
	if ast.RootNode == nil {
		fmt.Println("(nil AST)")
		return zero, nil
	}
	return evaluateNode(ast.RootNode)
}

func evaluateNode(node *asts.ASTNode) (int, error) {
	var zero int
	switch node.Type {
	case "int_literal", "hex_literal", "float_literal", "bin_literal":
		return evaluateLiteralNode(node)
	case "operator":
		return evaluateBinaryOperatorNode(node)
	case "unary":
		return evaluateUnaryOperatorNode(node)
	default:
		return zero, fmt.Errorf("unhandled node type %q", node.Type)
	}
}

func isLiteralNode(node *asts.ASTNode) bool {
	return node != nil && node.Token != nil &&
		(node.Type == "int_literal" || node.Type == "hex_literal" || node.Type == "float_literal" || node.Type == "bin_literal")
}

func evaluateLiteralNode(node *asts.ASTNode) (int, error) {
	var zero int
	if node.Token == nil {
		return zero, fmt.Errorf("literal node has no token")
	}
	return strconv.Atoi(string(node.Token.Lexeme))
}

func evaluateBinaryOperatorNode(node *asts.ASTNode) (int, error) {
	var zero int
	op := string(node.Token.Lexeme)
	if len(node.Children) != 2 {
		return zero, fmt.Errorf("expected two operands for operator %q; got %d", op, len(node.Children))
	}
	a, err := evaluateNode(node.Children[0])
	if err != nil {
		return zero, err
	}
	b, err := evaluateNode(node.Children[1])
	if err != nil {
		return zero, err
	}
	switch op {
	case "+":
		return a + b, nil
	case "-":
		return a - b, nil
	case "*":
		return a * b, nil
	case "/":
		return a / b, nil
	case "%":
		return a % b, nil
	case "**":
		return int(math.Pow(float64(a), float64(b))), nil
	default:
		return zero, fmt.Errorf("unhandled operator %q", op)
	}
}

func evaluateUnaryOperatorNode(node *asts.ASTNode) (int, error) {
	var zero int
	op := string(node.Token.Lexeme)
	if len(node.Children) != 1 {
		return zero, fmt.Errorf("expected one operand for unary %q; got %d", op, len(node.Children))
	}
	v, err := evaluateNode(node.Children[0])
	if err != nil {
		return zero, err
	}
	switch op {
	case "+":
		return v, nil
	case "-":
		return -v, nil
	default:
		return zero, fmt.Errorf("unhandled unary operator %q", op)
	}
}
