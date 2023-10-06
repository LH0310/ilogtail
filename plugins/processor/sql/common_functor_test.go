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

func TestLocate(t *testing.T) {
	type args struct {
		substr string
		str    string
		pos    int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1", args{"bar", "foobarbar", 1}, "4"},
		{"2", args{"xbar", "foobarbar", 1}, "0"},
		{"3", args{"bar", "foobarbar", 5}, "7"},
		{"4", args{"xbar", "foobarbar", 0}, "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := locate(tt.args.substr, tt.args.str, tt.args.pos); got != tt.want {
				t.Errorf("locate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMysqlSubstrNoLen(t *testing.T) {
	tests := []struct {
		str      string
		pos      int
		expected string
	}{
		{"Hello, world!", 8, "world!"},
		{"Hello, world!", 0, ""},
		{"Hello, world!", -5, "orld!"},
		{"Hello, world!", -15, ""},
		{"Hello, world!", 15, ""},
		{"", 1, ""},
	}

	for _, test := range tests {
		result := mysqlSubstrNoLen(test.str, test.pos)
		if result != test.expected {
			t.Errorf("mysqlSubstrNoLen(%q, %d) = %q; want %q", test.str, test.pos, result, test.expected)
		}
	}
}

func TestMysqlSubstrWithLen(t *testing.T) {
	tests := []struct {
		str      string
		pos      int
		subLen   int
		expected string
	}{
		{"Hello, world!", 8, 5, "world"},
		{"Hello, world!", 0, 5, ""},
		{"Hello, world!", -5, 3, "orl"},
		{"Hello, world!", -15, 5, ""},
		{"Hello, world!", 15, 5, ""},
		{"Hello, world!", 8, 50, "world!"},
		{"", 1, 5, ""},
	}

	for _, test := range tests {
		result := mysqlSubstrWithLen(test.str, test.pos, test.subLen)
		if result != test.expected {
			t.Errorf("mysqlSubstrWithLen(%q, %d, %d) = %q; want %q", test.str, test.pos, test.subLen, result, test.expected)
		}
	}
}

func TestRegexpLike(t *testing.T) {
	testCases := []struct {
		str     string
		pattern string
		want    string
	}{
		{"hello", "he.*o", "1"},
		{"world", "he.*o", "0"},
		{"hello", "he.*", "1"},
		{"hello", "he.*$", "1"},
	}

	for _, tc := range testCases {
		got := regexpLike(tc.str, tc.pattern)
		if got != tc.want {
			t.Errorf("regexpLike(%q, %q) = %q; want %q", tc.str, tc.pattern, got, tc.want)
		}
	}
}

func TestRegexpReplace(t *testing.T) {
	testCases := []struct {
		str     string
		pattern string
		replace string
		want    string
	}{
		{"hello world", "o", "0", "hell0 w0rld"},
		{"hello world", "l+", "L", "heLo worLd"},
		{"hello world", "l+", "", "heo word"},
	}

	for _, tc := range testCases {
		got := regexpReplace(tc.str, tc.pattern, tc.replace)
		if got != tc.want {
			t.Errorf("regexpReplace(%q, %q, %q) = %q; want %q", tc.str, tc.pattern, tc.replace, got, tc.want)
		}
	}
}

func TestRegexpSubstr(t *testing.T) {
	testCases := []struct {
		str     string
		pattern string
		want    string
	}{
		{"hello world", "l+", "ll"},
		{"hello world", "o", "o"},
		{"hello world", "z", ""},
	}

	for _, tc := range testCases {
		got := regexpSubstr(tc.str, tc.pattern)
		if got != tc.want {
			t.Errorf("regexpSubstr(%q, %q) = %q; want %q", tc.str, tc.pattern, got, tc.want)
		}
	}
}

func TestRegexpInstr(t *testing.T) {
	testCases := []struct {
		str     string
		pattern string
		want    string
	}{
		{"hello world", "l+", "3"},
		{"hello world", "o", "5"},
		{"hello world", "z", "0"},
	}

	for _, tc := range testCases {
		got := regexpInstr(tc.str, tc.pattern)
		if got != tc.want {
			t.Errorf("regexpInstr(%q, %q) = %q; want %q", tc.str, tc.pattern, got, tc.want)
		}
	}
}
