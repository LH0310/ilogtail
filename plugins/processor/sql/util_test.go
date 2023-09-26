package sql

import "testing"

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

func TestLikeOperator(t *testing.T) {
	tests := []struct {
		input    string
		pattern  string
		expected bool
	}{
		{"hello", "%", true},
		{"h", "_", true},
		{"hello", "%lo", true},
		{"hello", "he%", true},
		{"world", "or%", false},
		{"world", "%o_", false},
		{"Chrome on iOS. Mozilla/5.0 (iPhone; CPU iPhone OS 16_5_1 like Mac OS X)",
			"%iPhone OS%",
			true},
	}

	for _, test := range tests {
		actual := LikeOperator(test.input, test.pattern)
		if actual != test.expected {
			t.Errorf("Expected %v but got %v for input '%s' and pattern '%s'", test.expected, actual, test.input, test.pattern)
		}
	}
}
