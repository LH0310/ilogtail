package sql

import (
	"errors"
	"strings"

	"github.com/xwb1989/sqlparser"
)

type (
	scalarHandler func(sqlparser.Exprs) (stringEvaluator, error)
	strToStrFunc  func(string) string
)

var (
	scalarHandlerMap map[string]scalarHandler
	ErrArg           = errors.New("wrong type/number for arguments")
)

func (p *ProcessorSQL) initScalarFuncs() {
	scalarHandlerMap = map[string]scalarHandler{
		"coalesce":       p.handleCoalesce,
		"concat":         p.handleConcat,
		"concat_ws":      p.handleConcatWs,
		"substringindex": p.handleSubstringIndex,
		"md5":            p.handleMd5,
		"lower":          p.handleLower,
		"ltrim":          p.handleLtrim,
		"rtrim":          p.handleRtrim,
		"upper":          p.handleUpper,
		"length":         p.handleLength,
		"trim":           p.handleTrim,
		"sha1":           p.handleSha1,
		"to_base64":      p.handleToBase64,
		"locate":         p.handleLocate,
		"left":           p.handleLeft,
		"right":          p.handleRight,
		"regexp_instr":   p.handleRegexpInstr,
		"regexp_like":    p.handleRegexpLike,
		"regexp_replace": p.handleRegexpReplace,
		"regexp_substr":  p.handleRegexpSubstr,
		"replace":        p.handleReplace,
		"sha2":           p.handleSha2,
		"aes_encrypt":    p.handleAesEncrypt,
	}
}

func (p *ProcessorSQL) handleOneArgFunc(exprs sqlparser.Exprs, transform strToStrFunc) (stringEvaluator, error) {
	if len(exprs) != 1 {
		return nil, ErrArg
	}
	argEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}

	switch evaluator := argEvaluator.(type) {
	case *staticStringEvaluator:
		return &staticStringEvaluator{
			Value: transform(evaluator.Value),
		}, nil
	case *dynamicStringEvaluator:
		return &dynamicStringEvaluator{
			EvalFunc: func(slc stringLogContents) string {
				return transform(evaluator.evaluate(slc))
			},
		}, nil
	default:
		return nil, errors.New("unknown evaluator type")
	}
}

// Only support string type arguments currently.
// Returns the first non-empty string among its arguments, for that in go, string can't be nil,
// differing from MySQL's behavior, which returns the first non-NULL value.
func (p *ProcessorSQL) handleCoalesce(exprs sqlparser.Exprs) (stringEvaluator, error) {
	argEvaluators := make([]stringEvaluator, len(exprs))
	for i, expr := range exprs {
		var err error
		argEvaluators[i], err = p.compileStringExpr(expr)
		if err != nil {
			return nil, err
		}
	}
	return &dynamicStringEvaluator{
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

func (p *ProcessorSQL) handleConcat(exprs sqlparser.Exprs) (stringEvaluator, error) {
	argEvaluators := make([]stringEvaluator, len(exprs))
	for i, expr := range exprs {
		var err error
		argEvaluators[i], err = p.compileStringExpr(expr)
		if err != nil {
			return nil, err
		}
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			result := make([]string, len(argEvaluators))
			for i, eval := range argEvaluators {
				result[i] = eval.evaluate(slc)
			}
			return strings.Join(result, "")
		},
	}, nil
}

func (p *ProcessorSQL) handleConcatWs(exprs sqlparser.Exprs) (stringEvaluator, error) {
	argEvaluators := make([]stringEvaluator, len(exprs))
	for i, expr := range exprs {
		var err error
		argEvaluators[i], err = p.compileStringExpr(expr)
		if err != nil {
			return nil, err
		}
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			result := make([]string, len(argEvaluators))
			for i, eval := range argEvaluators {
				result[i] = eval.evaluate(slc)
			}
			return strings.Join(result[1:], result[0])
		},
	}, nil
}

func (p *ProcessorSQL) handleSubstringIndex(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 3 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}

	delimEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	count, err := evaluateIntExpr(exprs[2])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			delim := delimEvaluator.evaluate(slc)
			return substringIndex(str, delim, count)
		},
	}, nil
}

func (p *ProcessorSQL) handleLocate(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) == 3 {
		return p.handleLocatePos(exprs)
	}
	if len(exprs) != 2 {
		return nil, ErrArg
	}

	substrEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}
	strEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			substr := substrEvaluator.evaluate(slc)
			str := strEvaluator.evaluate(slc)
			return locate(substr, str, 1)
		},
	}, nil
}

func (p *ProcessorSQL) handleLocatePos(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 3 {
		return nil, ErrArg
	}

	substrEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}
	strEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}
	pos, err := evaluateIntExpr(exprs[2])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			substr := substrEvaluator.evaluate(slc)
			str := strEvaluator.evaluate(slc)
			return locate(substr, str, pos)
		},
	}, nil
}

func (p *ProcessorSQL) handleLeft(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 2 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}

	count, err := evaluateIntExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			return str[:count]
		},
	}, nil
}

func (p *ProcessorSQL) handleRight(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 2 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}

	count, err := evaluateIntExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			return str[len(str)-count:]
		},
	}, nil
}

func (p *ProcessorSQL) handleRegexpInstr(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 2 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}

	patternEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			pattern := patternEvaluator.evaluate(slc)
			return regexpInstr(str, pattern)
		},
	}, nil
}

func (p *ProcessorSQL) handleRegexpLike(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 2 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}

	patternEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			pattern := patternEvaluator.evaluate(slc)
			return regexpLike(str, pattern)
		},
	}, nil
}

func (p *ProcessorSQL) handleRegexpReplace(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 3 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}

	patternEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	replaceEvaluator, err := p.compileStringExpr(exprs[2])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			pattern := patternEvaluator.evaluate(slc)
			replace := replaceEvaluator.evaluate(slc)
			return regexpReplace(str, pattern, replace)
		},
	}, nil
}

func (p *ProcessorSQL) handleRegexpSubstr(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 2 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}
	patternEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			pattern := patternEvaluator.evaluate(slc)
			return regexpSubstr(str, pattern)
		},
	}, nil
}

func (p *ProcessorSQL) handleReplace(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 3 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}
	oldEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}
	newEvaluator, err := p.compileStringExpr(exprs[2])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			old := oldEvaluator.evaluate(slc)
			new := newEvaluator.evaluate(slc)
			return strings.Replace(str, old, new, -1)
		},
	}, nil
}

func (p *ProcessorSQL) handleSha2(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 2 {
		return nil, ErrArg
	}

	hashLen, err := evaluateIntExpr(exprs[1])
	if err != nil {
		return nil, err
	}
	sha2Func, err := sha2Generator(hashLen)
	if err != nil {
		return nil, err
	}
	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			return sha2Func(str)
		},
	}, nil
}

// AES_ENCRYPT
func (p *ProcessorSQL) handleAesEncrypt(exprs sqlparser.Exprs) (stringEvaluator, error) {
	if len(exprs) != 2 {
		return nil, ErrArg
	}

	strEvaluator, err := p.compileStringExpr(exprs[0])
	if err != nil {
		return nil, err
	}
	keyEvaluator, err := p.compileStringExpr(exprs[1])
	if err != nil {
		return nil, err
	}

	return &dynamicStringEvaluator{
		EvalFunc: func(slc stringLogContents) string {
			str := strEvaluator.evaluate(slc)
			key := keyEvaluator.evaluate(slc)
			return aesEncrypt(str, key)
		},
	}, nil
}

func (p *ProcessorSQL) handleLower(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, strings.ToLower)
}

func (p *ProcessorSQL) handleUpper(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, strings.ToUpper)
}

func (p *ProcessorSQL) handleLtrim(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, ltrim)
}

func (p *ProcessorSQL) handleRtrim(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, rtrim)
}

func (p *ProcessorSQL) handleTrim(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, trim)
}

func (p *ProcessorSQL) handleLength(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, strLen)
}

func (p *ProcessorSQL) handleMd5(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, md5)
}

func (p *ProcessorSQL) handleSha1(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, sha1)
}

func (p *ProcessorSQL) handleToBase64(exprs sqlparser.Exprs) (stringEvaluator, error) {
	return p.handleOneArgFunc(exprs, toBase64)
}
