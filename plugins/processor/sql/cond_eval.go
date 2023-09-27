package sql

import (
	"errors"
	"regexp"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type condEvaluator func(stringLogContents) bool

func (p *ProcessorSQL) compileCondExpr(e sqlparser.Expr) (condEvaluator, error) {
	if e == nil {
		return nil, errors.New("expr is nil")
	}

	switch expr := e.(type) {
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

		case sqlparser.RegexpStr:
			return handleRegexpStr(leftStrFunc, rightStrFunc)

		case sqlparser.LikeStr:
			return handleLikeStr(leftStrFunc, rightStrFunc)
		}
	default:
		return nil, errors.New("not a cond expr")
	}
	return nil, errors.New("")
}

func handleRegexpStr(leftStrFunc, rightStrFunc stringEvaluator) (condEvaluator, error) {
	// 可以通过类型断言只允许静态的 pattern
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
			return LikeOperator(leftStrFunc.evaluate(logContents), right.evaluate(logContents))
		}, nil
	default:
		return nil, errors.New("unknown evaluator type for rightStrFunc")
	}
}

func SQLLikeToRegexp(sqlLike string) string {
	regexpLike := strings.ReplaceAll(sqlLike, "%", ".*")
	regexpLike = strings.ReplaceAll(regexpLike, "_", ".")
	return "^" + regexpLike + "$"
}

func LikeOperator(input, pattern string) bool {
	regexpPattern := SQLLikeToRegexp(pattern)
	re := regexp.MustCompile(regexpPattern)
	return re.MatchString(input)
}
