package sql

import (
	"errors"
	"fmt"

	"github.com/alibaba/ilogtail/pkg/logger"
	"github.com/xwb1989/sqlparser"
)

func (p *ProcessorSQL) compileStringExpr(e sqlparser.Expr) (stringEvaluator, error) {
	if e == nil {
		return nil, errors.New("expression is nil")
	}

	switch expr := e.(type) {
	case *sqlparser.ParenExpr:
		return p.compileStringExpr(expr.Expr)

	case *sqlparser.SQLVal:
		constantVal := string(expr.Val)
		return &staticStringEvaluator{
			Value: constantVal,
		}, nil

	case *sqlparser.ColName:
		columnName := expr.Name.String()
		return &dynamicStringEvaluator{
			EvalFunc: func(slc stringLogContents) string {
				if slc.Contains(columnName) {
					return slc.Get(columnName)
				} else if p.NoKeyError {
					logger.Warning(p.context.GetRuntimeContext(), "SQL_FIND_ALARM", "cannot find key", columnName)
				}
				return ""
			},
		}, nil

	case *sqlparser.CaseExpr:
		return p.compileCaseExpr(expr)

	case *sqlparser.FuncExpr:
		funcName := expr.Name.Lowered()
		exprs, err := extractExprs(expr.Exprs)
		if err != nil {
			return nil, err
		}

		if handler, ok := scalarHandlerMap[funcName]; ok {
			return handler(exprs)
		} else {
			return nil, fmt.Errorf("unsupported function: %s", funcName)
		}

	case *sqlparser.SubstrExpr:
		from, err := evaluateIntExpr(expr.From)
		if err != nil {
			return nil, err
		}

		eval, err := p.compileStringExpr(expr.Name)
		if err != nil {
			return nil, err
		}

		if expr.To == nil {
			return &dynamicStringEvaluator{
				EvalFunc: func(slc stringLogContents) string {
					return mysqlSubstrNoLen(eval.evaluate(slc), from)
				},
			}, nil
		}

		length, err := evaluateIntExpr(expr.To)
		if err != nil {
			return nil, err
		}
		return &dynamicStringEvaluator{
			EvalFunc: func(slc stringLogContents) string {
				return mysqlSubstrWithLen(eval.evaluate(slc), from, length)
			},
		}, nil

	default:
		return nil, errors.New("Unsupported expression type: " + fmt.Sprintf("%T", expr))
	}
}

func (p *ProcessorSQL) compileCaseExpr(caseExpr *sqlparser.CaseExpr) (stringEvaluator, error) {
	isValueCase := caseExpr.Expr != nil
	whenCont := len(caseExpr.Whens)

	valueEvaluators := make([]stringEvaluator, whenCont)
	condEvaluators := make([]condEvaluator, whenCont)

	var caseValueEvaluator stringEvaluator
	var defaultEvaluator stringEvaluator

	if caseExpr.Else != nil {
		var err error
		defaultEvaluator, err = p.compileStringExpr(caseExpr.Else)
		if err != nil {
			return nil, err
		}
	}

	if isValueCase {
		var err error
		caseValueEvaluator, err = p.compileStringExpr(caseExpr.Expr)
		if err != nil {
			return nil, err
		}
	}

	for i, when := range caseExpr.Whens {
		whenValueEvaluator, err := p.compileStringExpr(when.Val)
		if err != nil {
			return nil, err
		}
		valueEvaluators[i] = whenValueEvaluator

		if isValueCase {
			compareValueEvaluator, err := p.compileStringExpr(when.Cond)
			if err != nil {
				return nil, err
			}

			condEvaluators[i] = func(slc stringLogContents) bool {
				return caseValueEvaluator.evaluate(slc) == compareValueEvaluator.evaluate(slc)
			}
		} else {
			cond, err := p.compileCondExpr(when.Cond)
			if err != nil {
				return nil, err
			}
			condEvaluators[i] = cond
		}
	}
	return &dynamicStringEvaluator{
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
