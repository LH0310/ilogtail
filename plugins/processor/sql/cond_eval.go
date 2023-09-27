package sql

import (
	"errors"
	"regexp"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type condEvaluator func(stringLogContents) bool

func compileCondExpr(e *sqlparser.Expr) (condEvaluator, error) {
	if e == nil {
		return nil, errors.New("expr is nil")
	}

	switch expr := (*e).(type) {
	case *sqlparser.NotExpr:
		condFunc, err := compileCondExpr(&expr.Expr)
		if err != nil {
			return nil, err
		}
		return func(logContents stringLogContents) bool {
			return !condFunc(logContents)
		}, nil

	case *sqlparser.AndExpr:
		leftCondFunc, err := compileCondExpr(&expr.Left)
		if err != nil {
			return nil, err
		}
		rightCondFunc, err := compileCondExpr(&expr.Right)
		if err != nil {
			return nil, err
		}
		return func(logContents stringLogContents) bool {
			return leftCondFunc(logContents) && rightCondFunc(logContents)
		}, nil

	case *sqlparser.OrExpr:
		leftCondFunc, err := compileCondExpr(&expr.Left)
		if err != nil {
			return nil, err
		}
		rightCondFunc, err := compileCondExpr(&expr.Right)
		if err != nil {
			return nil, err
		}
		return func(logContents stringLogContents) bool {
			return leftCondFunc(logContents) || rightCondFunc(logContents)
		}, nil

	case *sqlparser.ComparisonExpr:
		leftStrFunc, err := compileStringExpr(&expr.Left)
		if err != nil {
			return nil, err
		}
		rightStrFunc, err := compileStringExpr(&expr.Right)
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

	//TODO: 为什么这里编译器检测不出这条语句是无法抵达的？
	return nil, errors.New("")
}

func handleRegexpStr(leftStrFunc, rightStrFunc stringEvaluator) (func(stringLogContents) bool, error) {
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
	// 可能不够全面，可以看看别的项目实现，这种应该很好复用
	regexpPattern := SQLLikeToRegexp(pattern)
	re := regexp.MustCompile(regexpPattern)
	return re.MatchString(input)
}
