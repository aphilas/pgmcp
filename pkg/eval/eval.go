// eval implements a simple expression evaluator in Go.
// See: https://thorstenball.com/blog/2016/11/16/putting-eval-in-go/
package eval

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
)

func Eval(input string) (*int, error) {
	exp, err := parser.ParseExpr(input)
	if err != nil {
		return nil, fmt.Errorf("parsing expression: %w", err)
	}

	res, err := evalExpr(exp)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func evalExpr(exp ast.Expr) (int, error) {
	switch exp := exp.(type) {
	case *ast.BinaryExpr:
		return evalBinaryExpr(exp)
	case *ast.ParenExpr:
		return evalExpr(exp.X)
	case *ast.BasicLit:
		if exp.Kind == token.INT {
			i, err := strconv.Atoi(exp.Value)
			if err != nil {
				return 0, fmt.Errorf("invalid integer %q: %w", exp.Value, err)
			}
			return i, nil
		}
		return 0, fmt.Errorf("unsupported literal type: %s", exp.Kind)
	}

	return 0, fmt.Errorf("unsupported expression type: %T", exp)
}

func evalBinaryExpr(exp *ast.BinaryExpr) (int, error) {
	left, err := evalExpr(exp.X)
	if err != nil {
		return 0, err
	}
	right, err := evalExpr(exp.Y)
	if err != nil {
		return 0, err
	}

	switch exp.Op {
	case token.ADD:
		return left + right, nil
	case token.SUB:
		return left - right, nil
	case token.MUL:
		return left * right, nil
	case token.QUO:
		if right == 0 {
			return 0, fmt.Errorf("division by zero")
		}
		return left / right, nil
	}

	return 0, fmt.Errorf("unsupported operator: %s", exp.Op)
}
