package sql

type stringEvaluator interface {
	evaluate(slc stringLogContents) string
}

type staticStringEvaluator struct {
	Value string
}

func (sse *staticStringEvaluator) evaluate(_ stringLogContents) string {
	return sse.Value
}

type dynamicStringEvaluator struct {
	EvalFunc func(stringLogContents) string
}

func (dse *dynamicStringEvaluator) evaluate(slc stringLogContents) string {
	return dse.EvalFunc(slc)
}
