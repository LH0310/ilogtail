package sql

import (
	"errors"
	"regexp"

	"github.com/xwb1989/sqlparser"
)

type condEvaluator func(stringLogContents) bool

func (p *ProcessorSQL) compileCondExpr(e sqlparser.Expr) (condEvaluator, error) {
	if e == nil {
		return nil, errors.New("expr is nil")
	}

	switch expr := e.(type) {
	case *sqlparser.ParenExpr:
		return p.compileCondExpr(expr.Expr)

	case sqlparser.BoolVal:
		return func(logContents stringLogContents) bool {
			return bool(expr)
		}, nil

	case *sqlparser.NotExpr:
		condFunc, err := p.compileCondExpr(expr.Expr)
		if err != nil {
			return nil, err
		}
		return func(logContents stringLogContents) bool {
			return !condFunc(logContents)
		}, nil

	case *sqlparser.AndExpr:
		leftCondFunc, err := p.compileCondExpr(expr.Left)
		if err != nil {
			return nil, err
		}
		rightCondFunc, err := p.compileCondExpr(expr.Right)
		if err != nil {
			return nil, err
		}
		return func(logContents stringLogContents) bool {
			return leftCondFunc(logContents) && rightCondFunc(logContents)
		}, nil

	case *sqlparser.OrExpr:
		leftCondFunc, err := p.compileCondExpr(expr.Left)
		if err != nil {
			return nil, err
		}
		rightCondFunc, err := p.compileCondExpr(expr.Right)
		if err != nil {
			return nil, err
		}
		return func(logContents stringLogContents) bool {
			return leftCondFunc(logContents) || rightCondFunc(logContents)
		}, nil

	case *sqlparser.ComparisonExpr:
		leftStrFunc, err := p.compileStringExpr(expr.Left)
		if err != nil {
			return nil, err
		}
		rightStrFunc, err := p.compileStringExpr(expr.Right)
		if err != nil {
			return nil, err
		}
		switch expr.Operator {
		case sqlparser.EqualStr:
			return func(logContents stringLogContents) bool {
				return leftStrFunc.evaluate(logContents) == rightStrFunc.evaluate(logContents)
			}, nil
		case sqlparser.NotEqualStr:
			return func(logContents stringLogContents) bool {
				return leftStrFunc.evaluate(logContents) != rightStrFunc.evaluate(logContents)
			}, nil
		case sqlparser.LessThanStr:
			return func(logContents stringLogContents) bool {
				return leftStrFunc.evaluate(logContents) < rightStrFunc.evaluate(logContents)
			}, nil
		case sqlparser.LessEqualStr:
			return func(logContents stringLogContents) bool {
				return leftStrFunc.evaluate(logContents) <= rightStrFunc.evaluate(logContents)
			}, nil
		case sqlparser.GreaterThanStr:
			return func(logContents stringLogContents) bool {
				return leftStrFunc.evaluate(logContents) > rightStrFunc.evaluate(logContents)
			}, nil
		case sqlparser.GreaterEqualStr:
			return func(logContents stringLogContents) bool {
				return leftStrFunc.evaluate(logContents) >= rightStrFunc.evaluate(logContents)
			}, nil
		case sqlparser.RegexpStr:
			return handleRegexpStr(leftStrFunc, rightStrFunc)
		case sqlparser.LikeStr:
			return handleLikeStr(leftStrFunc, rightStrFunc)
		default:
			return nil, errors.New("unknown operator: " + expr.Operator)
		}
	default:
		return nil, errors.New("unsupport expression type: " + sqlparser.String(expr))
	}
}

func handleRegexpStr(leftStrFunc, rightStrFunc stringEvaluator) (condEvaluator, error) {
	switch right := rightStrFunc.(type) {
	case *staticStringEvaluator:
		re := regexp.MustCompile(right.Value)
		return func(slc stringLogContents) bool {
			return re.MatchString(leftStrFunc.evaluate(slc))
		}, nil
	case *dynamicStringEvaluator:
		return func(logContents stringLogContents) bool {
			re, err := regexp.Compile(right.evaluate(logContents))
			if err != nil {
				// TODO: handle runtime compile error
				return false
			}
			return re.MatchString(leftStrFunc.evaluate(logContents))
		}, nil
	default:
		return nil, errors.New("unknown evaluator type for rightStrFunc")
	}
}

func handleLikeStr(leftStrFunc, rightStrFunc stringEvaluator) (func(stringLogContents) bool, error) {
	switch right := rightStrFunc.(type) {
	case *staticStringEvaluator:
		pat := right.Value
		regPat := SQLLikeToRegexp(pat)
		re := regexp.MustCompile(regPat)
		return func(slc stringLogContents) bool {
			return re.MatchString(leftStrFunc.evaluate(slc))
		}, nil
	case *dynamicStringEvaluator:
		return func(logContents stringLogContents) bool {
			pat := right.evaluate(logContents)
			regPat := SQLLikeToRegexp(pat)
			re, err := regexp.Compile(regPat)
			if err != nil {
				return false
			}
			return re.MatchString(leftStrFunc.evaluate(logContents))
		}, nil
	default:
		return nil, errors.New("unknown evaluator type for rightStrFunc")
	}
}
