package sql

import (
	"errors"
	"strconv"

	"github.com/xwb1989/sqlparser"
)

func evaluateIntExpr(expr sqlparser.Expr) (int, error) {
	if expr == nil {
		return 0, errors.New("expression is nil")
	}
	valueExpr, ok := expr.(*sqlparser.SQLVal)
	if !ok {
		return 0, errors.New("not a SQLVal expr")
	}

	if valueExpr.Type != sqlparser.IntVal {
		return 0, errors.New("not a int value")
	}

	return strconv.Atoi(string(valueExpr.Val))
}
