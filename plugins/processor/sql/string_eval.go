package sql

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type stringEvaluator func(stringLogContents) string

func compileStringExpr(e *sqlparser.Expr) (stringEvaluator, error) {
	if e == nil {
		return nil, errors.New("expression is nil")
	}

	switch expr := (*e).(type) {
	case *sqlparser.SQLVal:
		// Handling for non-string types should go here
		constantVal := string(expr.Val)
		return func(logContents stringLogContents) string {
			return constantVal
		}, nil

	case *sqlparser.ColName:
		columnName := expr.Name.String()
		return func(logContents stringLogContents) string {
			return logContents.Get(columnName)
		}, nil

	case *sqlparser.CaseExpr:
		return compileCaseExpr(expr)

	case *sqlparser.FuncExpr:
		funcName := strings.ToLower(expr.Name.String())
		funcList, err := extractFuncs(expr.Exprs)
		if err != nil {
			return nil, err
		}

		switch funcName {
		case "coalesce":
			return handleCoalesce(funcList), nil

		case "concat":
			return handleConcat(funcList), nil

		case "concat_ws":
			return handleConcatWs(funcList), nil

		case "md5":
			if len(expr.Exprs) != 1 {
				return nil, errors.New("wrong number of args for md5")
			}
			return handleMd5(funcList[0]), nil

		case "lower":
			if len(expr.Exprs) != 1 {
				return nil, errors.New("wrong number of args for lower")
			}
			return handleLower(funcList[0]), nil

		case "ltrim":
			if len(expr.Exprs) != 1 {
				return nil, errors.New("wrong number of args for ltrim")
			}
			return handleLtrim(funcList[0]), nil

		default:
			return nil, errors.New("Unsupported SQL function: " + funcName)
		}
	default:
		return nil, errors.New("Unsupported expression type: " + fmt.Sprintf("%T", expr))
	}
}

func handleCoalesce(funcList []stringEvaluator) stringEvaluator {
	return func(logContents stringLogContents) string {
		for _, evalFunc := range funcList {
			val := evalFunc(logContents)
			if val != "" {
				return val
			}
		}
		return ""
	}
}

func handleConcat(funcList []stringEvaluator) stringEvaluator {
	return func(logContents stringLogContents) string {
		result := make([]string, len(funcList))
		for i, evalFunc := range funcList {
			result[i] = evalFunc(logContents)
		}
		return strings.Join(result, "")
	}
}

func handleConcatWs(funcList []stringEvaluator) stringEvaluator {
	return func(logContents stringLogContents) string {
		result := make([]string, len(funcList))
		for i, evalFunc := range funcList {
			result[i] = evalFunc(logContents)
		}
		return strings.Join(result[1:], result[0])
	}
}

func handleMd5(eval stringEvaluator) stringEvaluator {
	return func(logContents stringLogContents) string {
		str := eval(logContents)
		return fmt.Sprintf("%x", md5.Sum([]byte(str)))
	}
}

func handleLower(eval stringEvaluator) stringEvaluator {
	return func(slc stringLogContents) string {
		str := eval(slc)
		return strings.ToLower(str)
	}
}

func handleLtrim(eval stringEvaluator) stringEvaluator {
	return func(logContents stringLogContents) string {
		str := eval(logContents)
		return strings.TrimLeft(str, " ")
	}
}

func compileCaseExpr(caseExpr *sqlparser.CaseExpr) (stringEvaluator, error) {
	isValueCase := caseExpr.Expr != nil
	whenCont := len(caseExpr.Whens)

	valueEvaluators := make([]stringEvaluator, whenCont)
	condEvaluators := make([]condEvaluator, whenCont)

	var caseValueEvaluator stringEvaluator
	var defaultEvaluator stringEvaluator

	if caseExpr.Else != nil {
		var err error
		defaultEvaluator, err = compileStringExpr(&caseExpr.Else)
		if err != nil {
			return nil, err
		}
	}

	if isValueCase {
		var err error
		caseValueEvaluator, err = compileStringExpr(&caseExpr.Expr)
		if err != nil {
			return nil, err
		}
	}

	for i, when := range caseExpr.Whens {
		whenValueEvaluator, err := compileStringExpr(&when.Val)
		if err != nil {
			return nil, err
		}
		valueEvaluators[i] = whenValueEvaluator

		if isValueCase {
			compareValueEvaluator, err := compileStringExpr(&when.Cond)
			if err != nil {
				return nil, err
			}

			condEvaluators[i] = func(slc stringLogContents) bool {
				return caseValueEvaluator(slc) == compareValueEvaluator(slc)
			}
		} else {
			cond, err := compileCondExpr(&when.Cond)
			if err != nil {
				return nil, err
			}
			condEvaluators[i] = cond
		}
	}
	return func(slc stringLogContents) string {
		for i, cond := range condEvaluators {
			if cond(slc) {
				return valueEvaluators[i](slc)
			}
		}
		if defaultEvaluator != nil {
			return defaultEvaluator(slc)
		}
		return ""
	}, nil
}

func extractFuncs(exprs sqlparser.SelectExprs) ([]stringEvaluator, error) {
	funcs := make([]stringEvaluator, len(exprs))
	for i, se := range exprs {
		ae, ok := se.(*sqlparser.AliasedExpr)
		if !ok {
			return nil, errors.New("not an AliasedExpr")
		}
		var err error
		funcs[i], err = compileStringExpr(&ae.Expr)
		if err != nil {
			return nil, err
		}
	}
	return funcs, nil
}
