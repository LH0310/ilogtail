package sql

import (
	"errors"
	"regexp"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type condEvaluator func(stringLogContents) bool

func compileCondExpr(e *sqlparser.Expr) (condEvaluator, error) {
	// if e == nil {
	// 	// TODO: when will a nil expr be passed
	// 	return nil, errors.New("expr is nil")
	// }

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
				return leftStrFunc(logContents) == rightStrFunc(logContents)
			}, nil

		case sqlparser.RegexpStr:
			return func(logContents stringLogContents) bool {
				re, err := regexp.Compile(rightStrFunc(logContents))
				if err != nil {
					// TODO: 处理错误，如果是运行时的，得想一下，也许得加一个配置
				}
				return re.MatchString(leftStrFunc(logContents))
			}, nil

		case sqlparser.LikeStr:
			return func(logContents stringLogContents) bool {
				return LikeOperator(leftStrFunc(logContents), rightStrFunc(logContents))
			}, nil
		}
	default:
		return nil, errors.New("not a cond expr")
	}

	//TODO: 为什么这里编译器检测不出这条语句是无法抵达的？
	return nil, errors.New("")
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
