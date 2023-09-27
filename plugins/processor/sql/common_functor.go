package sql

import (
	cryptoMd5 "crypto/md5"
	"fmt"
	"strings"
)

func md5(s string) string {
	return fmt.Sprintf("%x", cryptoMd5.Sum([]byte(s)))
}

func ltrim(s string) string {
	return strings.TrimLeft(s, " ")
}

func substringIndex(str, delim string, count int) string {
	if delim == "" {
		return ""
	}

	parts := strings.Split(str, delim)

	if count > 0 {
		if count > len(parts) {
			count = len(parts)
		}
		return strings.Join(parts[:count], delim)
	} else if count < 0 {
		if -count > len(parts) {
			count = -len(parts)
		}
		return strings.Join(parts[len(parts)+count:], delim)
	}

	return ""
}

func mysqlSubstrNoLen(str string, pos int) string {
	strLen := len(str)

	if pos == 0 || pos > strLen || pos < -strLen {
		return ""
	}

	if pos < 0 {
		pos = strLen + pos + 1
	}

	return str[pos-1:]
}

func mysqlSubstrWithLen(str string, pos int, subLen int) string {
	strLen := len(str)

	if pos == 0 || subLen < 1 || pos > strLen || pos < -strLen {
		return ""
	}

	if pos < 0 {
		pos = strLen + pos + 1
	}

	endPos := pos + subLen - 1

	if endPos > strLen {
		endPos = strLen
	}

	return str[pos-1 : endPos]
}
