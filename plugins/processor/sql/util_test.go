package sql

import (
	"errors"
	"testing"

	"github.com/xwb1989/sqlparser"
)

func TestSQLLikeToRegexp(t *testing.T) {
	tests := []struct {
		sqlLike  string
		expected string
	}{
		{"%", "^.*$"},
		{"_", "^.$"},
		{"%like%", "^.*like.*$"},
		{"_like_", "^.like.$"},
	}

	for _, test := range tests {
		actual := SQLLikeToRegexp(test.sqlLike)
		if actual != test.expected {
			t.Errorf("Expected %s but got %s", test.expected, actual)
		}
	}
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
