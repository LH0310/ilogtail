package sql

import (
	"errors"
	"fmt"

	"github.com/xwb1989/sqlparser"
)

func compileStringExpr(e *sqlparser.Expr) (*stringEvaluator, error) {
	if e == nil {
		return nil, errors.New("expression is nil")
	}

	switch expr := (*e).(type) {
	case *sqlparser.SQLVal:
		// TODO: Handling for non-string types should go here
		constantVal := string(expr.Val)
		return &stringEvaluator{
			IsStatic:    true,
			StaticValue: constantVal,
		}, nil

	case *sqlparser.ColName:
		columnName := expr.Name.String()
		return &stringEvaluator{
			IsStatic: false,
			EvalFunc: func(slc stringLogContents) string {
				return slc.Get(columnName)
			},
		}, nil

	case *sqlparser.CaseExpr:
		return compileCaseExpr(expr)

	case *sqlparser.FuncExpr:
		funcName := expr.Name.Lowered()
		exprs, err := extractExprs(expr.Exprs)
		if err != nil {
			return nil, err
		}

		if handler, ok := scalarHandlerMap[funcName]; ok {
			return handler(exprs)
		} else {
			return nil, fmt.Errorf("Unsupported function: %s", funcName)
		}

	case *sqlparser.SubstrExpr:

	default:
		return nil, errors.New("Unsupported expression type: " + fmt.Sprintf("%T", expr))
	}
	return nil, errors.New("")
}

func compileCaseExpr(caseExpr *sqlparser.CaseExpr) (*stringEvaluator, error) {
	isValueCase := caseExpr.Expr != nil
	whenCont := len(caseExpr.Whens)

	valueEvaluators := make([]*stringEvaluator, whenCont)
	condEvaluators := make([]condEvaluator, whenCont)

	var caseValueEvaluator *stringEvaluator
	var defaultEvaluator *stringEvaluator

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
				return caseValueEvaluator.evaluate(slc) == compareValueEvaluator.evaluate(slc)
			}
		} else {
			cond, err := compileCondExpr(&when.Cond)
			if err != nil {
				return nil, err
			}
			condEvaluators[i] = cond
		}
	}
	return &stringEvaluator{
		IsStatic: false,
		EvalFunc: func(slc stringLogContents) string {
			for i, cond := range condEvaluators {
				if cond(slc) {
					return valueEvaluators[i].evaluate(slc)
				}
			}
			if defaultEvaluator != nil {
				return defaultEvaluator.evaluate(slc)
			}
			return ""
		},
	}, nil
}

func extractExprs(selExprs sqlparser.SelectExprs) (exprs sqlparser.Exprs, err error) {
	exprs = make(sqlparser.Exprs, len(selExprs))
	for i, se := range selExprs {
		ae, ok := se.(*sqlparser.AliasedExpr)
		if !ok {
			return nil, errors.New("not an AliasedExpr in func args")
		}
		exprs[i] = ae.Expr
	}
	return exprs, nil
}
