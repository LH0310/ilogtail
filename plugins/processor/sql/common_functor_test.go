package sql

import (
	"testing"
)

func TestSubstringIndex(t *testing.T) {
	tests := []struct {
		str    string
		delim  string
		count  int
		result string
	}{
		{"www.mysql.com", ".", 2, "www.mysql"},
		{"www.mysql.com", ".", -2, "mysql.com"},
		{"www.mysql.com", ".", 0, ""},
		{"www.mysql.com", ".", 4, "www.mysql.com"},
		{"www.mysql.com", ".", -4, "www.mysql.com"},
		{"", ".", 2, ""},
		{"www.mysql.com", "", 2, ""},
	}

	for _, test := range tests {
		t.Run(test.str, func(t *testing.T) {
			output := substringIndex(test.str, test.delim, test.count)
			if output != test.result {
				t.Errorf("Expected %s, got %s", test.result, output)
			}
		})
	}
}
