package sql

import (
	"crypto/aes"
	"crypto/cipher"
	cryptoMd5 "crypto/md5"
	"crypto/rand"
	cryptoSha1 "crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func md5(s string) string {
	hashed := cryptoMd5.Sum([]byte(s))
	return hex.EncodeToString(hashed[:])
}

func sha1(s string) string {
	hashed := cryptoSha1.Sum([]byte(s))
	return hex.EncodeToString(hashed[:])
}

// SHA2 hashes the given string using the SHA-2 family of hash functions.
// The second argument indicates the desired bit length of the result.
func sha2(str string, hashLength int) (string, error) {
	switch hashLength {
	case 224:
		hash := sha256.Sum224([]byte(str))
		return hex.EncodeToString(hash[:]), nil
	case 256, 0:
		hash := sha256.Sum256([]byte(str))
		return hex.EncodeToString(hash[:]), nil
	case 384:
		hash := sha512.Sum384([]byte(str))
		return hex.EncodeToString(hash[:]), nil
	case 512:
		hash := sha512.Sum512([]byte(str))
		return hex.EncodeToString(hash[:]), nil
	default:
		return "", fmt.Errorf("invalid hash length: %d", hashLength)
	}
}

// SHA2Generator returns a function that generates SHA-2 hashes of the specified length.
func sha2Generator(hashLength int) (func(string) string, error) {
	_, err := sha2("", hashLength)
	if err != nil {
		return nil, err
	}

	return func(data string) string {
		hashed, _ := sha2(data, hashLength)
		return hashed
	}, nil
}

func aesEncrypt(plainText string, keyStr string) string {
	plainBytes := []byte(plainText)
	key := sha256.Sum256([]byte(keyStr))
	block, _ := aes.NewCipher(key[:])
	cipherText := make([]byte, aes.BlockSize+len(plainBytes))
	iv := cipherText[:aes.BlockSize]
	io.ReadFull(rand.Reader, iv)
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainBytes)
	return hex.EncodeToString(cipherText)
}

func toBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func ltrim(s string) string {
	return strings.TrimLeftFunc(s, unicode.IsSpace)
}

func rtrim(s string) string {
	return strings.TrimRightFunc(s, unicode.IsSpace)
}

func trim(s string) string {
	return strings.TrimFunc(s, unicode.IsSpace)
}

func strLen(s string) string {
	return strconv.Itoa(len(s))
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

func locate(substr, str string, pos int) string {
	if pos < 1 || pos > len(str) {
		return "0"
	}

	position := strings.Index(str[pos-1:], substr)
	if position == -1 {
		return "0"
	}
	return strconv.Itoa(position + pos)
}

func regexpInstr(str, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringIndex(str)

	if matches == nil {
		return "0"
	}

	return strconv.Itoa(matches[0] + 1)
}

// regexpLike returns 1 if str matches the pattern, 0 otherwise.
func regexpLike(str, pattern string) string {
	re := regexp.MustCompile(pattern)
	if re.MatchString(str) {
		return "1"
	} else {
		return "0"
	}
}

func regexpReplace(str, pattern, replace string) string {
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(str, replace)
}

func regexpSubstr(str, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(str)

	if matches == nil {
		return ""
	}

	return matches[1]
}
