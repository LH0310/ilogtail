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

func Test_locate(t *testing.T) {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := locate(tt.args.substr, tt.args.str, tt.args.pos); got != tt.want {
				t.Errorf("locate() = %v, want %v", got, tt.want)
			}
		})
	}
}
