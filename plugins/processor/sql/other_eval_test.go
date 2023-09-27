package sql

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xwb1989/sqlparser"
)

func TestEvaluateIntExpr(t *testing.T) {
	tests := []struct {
		in  sqlparser.Expr
		out int
		err error
	}{
		{&sqlparser.SQLVal{Type: sqlparser.IntVal, Val: []byte("42")}, 42, nil},
		{&sqlparser.SQLVal{Type: sqlparser.IntVal, Val: []byte("0")}, 0, nil},
		{&sqlparser.SQLVal{Type: sqlparser.StrVal, Val: []byte("strval")}, 0, errors.New("not a int value")},
	}

	for _, test := range tests {
		result, err := evaluateIntExpr(&test.in)

		if test.err == nil {
			assert.NoError(t, err)
			assert.Equal(t, test.out, result)
		} else {
			assert.Error(t, err)
			assert.Equal(t, test.err.Error(), err.Error())
		}
	}
}
