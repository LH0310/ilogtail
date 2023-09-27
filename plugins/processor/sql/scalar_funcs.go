package sql

import (
	"errors"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type (
	scalarHandler func(sqlparser.Exprs) (*stringEvaluator, error)
	strToStrFunc  func(string) string
)

var (
	scalarHandlerMap map[string]scalarHandler
	ErrArg           = errors.New("wrong type/number for arguments")
)

func initScalarFuncs() {
	scalarHandlerMap = map[string]scalarHandler{
		"coalesce":       handleCoalesce,
		"concat":         handleConcat,
		"concat_ws":      handleConcatWs,
		"substringindex": handleSubstringIndex,
		"md5":            handleMd5,
		"lower":          handleLower,
		"ltrim":          handleLtrim,
	}
}

func handleOneArgFunc(exprs sqlparser.Exprs, transform strToStrFunc) (*stringEvaluator, error) {
	if len(exprs) != 1 {
		return nil, ErrArg
	}
	argEvaluator, err := compileStringExpr(&exprs[0])
	if err != nil {
		return nil, err
	}

	eval := &stringEvaluator{
		IsStatic: argEvaluator.IsStatic,
	}

	if eval.IsStatic {
		eval.StaticValue = transform(argEvaluator.StaticValue)
	} else {
		eval.EvalFunc = func(slc stringLogContents) string {
			return transform(argEvaluator.evaluate(slc))
		}
	}
	return eval, nil
}

// Only support string type arguments currently.
// Returns the first non-empty string among its arguments, for that in go, string can't be nil,
// differing from MySQL's behavior, which returns the first non-NULL value.
func handleCoalesce(exprs sqlparser.Exprs) (*stringEvaluator, error) {
	argEvaluators := make([]*stringEvaluator, len(exprs))
	for i, expr := range exprs {
		var err error
		argEvaluators[i], err = compileStringExpr(&expr)
		if err != nil {
			return nil, err
		}
	}
	// 这里其实可以做编译期优化，如果前面有常量的话这就变成一个常量函数了，但应该没人会这么写吧
	return &stringEvaluator{
		IsStatic: false,
		EvalFunc: func(slc stringLogContents) string {
			for _, argEvaluator := range argEvaluators {
				str := argEvaluator.evaluate(slc)
				if str != "" {
					return str
				}
			}
			return ""
		},
	}, nil
}

func handleConcat(exprs sqlparser.Exprs) (*stringEvaluator, error) {
	argEvaluators := make([]*stringEvaluator, len(exprs))
	for i, expr := range exprs {
		var err error
		argEvaluators[i], err = compileStringExpr(&expr)
		if err != nil {
			return nil, err
		}
	}

	return &stringEvaluator{
		IsStatic: false,
		EvalFunc: func(slc stringLogContents) string {
			result := make([]string, len(argEvaluators))
			for i, eval := range argEvaluators {
				result[i] = eval.evaluate(slc)
			}
			return strings.Join(result, "")
		},
	}, nil
}

func handleConcatWs(exprs sqlparser.Exprs) (*stringEvaluator, error) {
	argEvaluators := make([]*stringEvaluator, len(exprs))
	for i, expr := range exprs {
		var err error
		argEvaluators[i], err = compileStringExpr(&expr)
		if err != nil {
			return nil, err
		}
	}

	return &stringEvaluator{
		IsStatic: false,
		EvalFunc: func(slc stringLogContents) string {
			result := make([]string, len(argEvaluators))
			for i, eval := range argEvaluators {
				result[i] = eval.evaluate(slc)
			}
			return strings.Join(result[1:], result[0])
		},
	}, nil
}

func handleSubstringIndex(exprs sqlparser.Exprs) (*stringEvaluator, error) {
	if len(exprs) != 3 {
		return nil, ErrArg
	}

	strEvaluator, err := compileStringExpr(&exprs[0])
	if err != nil {
		return nil, err
	}

	delimEvaluator, err := compileStringExpr(&exprs[1])
	if err != nil {
		return nil, err
	}

	count, err := evaluateIntExpr(&exprs[2])
	if err != nil {
		return nil, err
	}

	return &stringEvaluator{
		IsStatic: false,
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			delim := delimEvaluator.evaluate(slc)
			return substringIndex(str, delim, count)
		},
	}, nil
}

func handleMd5(exprs sqlparser.Exprs) (*stringEvaluator, error) {
	return handleOneArgFunc(exprs, md5)
}

func handleLower(exprs sqlparser.Exprs) (*stringEvaluator, error) {
	return handleOneArgFunc(exprs, strings.ToLower)
}

func handleLtrim(exprs sqlparser.Exprs) (*stringEvaluator, error) {
	return handleOneArgFunc(exprs, ltrim)
}
