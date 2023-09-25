package sql

import (
	"errors"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type stringEvaluator func(stringLogContents) string

func compileStringExpr(e *sqlparser.Expr) (stringEvaluator, error) {
	if e == nil {
		return nil, errors.New("expr is nil")
	}
	switch expr := (*e).(type) {
	case *sqlparser.SQLVal:
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
		funcName := expr.Name.Lowered()
		switch funcName {
		case "coalesce":
			funcList, err := extractFuncs(expr.Exprs)
			if err != nil {
				return nil, err
			}
			return func(logContents stringLogContents) string {
				for _, evalFunc := range funcList {
					val := evalFunc(logContents)
					if val != "" {
						return val
					}
				}
				return ""
			}, nil

		case "concat":
			funcList, err := extractFuncs(expr.Exprs)
			if err != nil {
				return nil, err
			}
			return func(logContents stringLogContents) string {
				var result []string
				for _, evalFunc := range funcList {
					result = append(result, evalFunc(logContents))
				}
				return strings.Join(result, "")
			}, nil

		case "ltrim":
			if len(expr.Exprs) != 1 {
				return nil, errors.New("wrong number of args for ltrim")
			}
			innerExpr := expr.Exprs[0].(*sqlparser.AliasedExpr).Expr
			innerEval, err := compileStringExpr(&innerExpr)
			if err != nil {
				return nil, err
			}
			return func(logContents stringLogContents) string {
				str := innerEval(logContents)
				return strings.TrimLeft(str, " ")
			}, nil

		default:
			return nil, errors.New("unsupported function")
		}
	default:
		return nil, errors.New("unsupported expression type")
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
