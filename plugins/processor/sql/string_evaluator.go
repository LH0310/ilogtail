package sql

type stringEvaluator struct {
	IsStatic    bool
	StaticValue string
	EvalFunc    func(stringLogContents) string
}

func (se *stringEvaluator) evaluate(slc stringLogContents) string {
	if se.IsStatic {
		return se.StaticValue
	} else {
		return se.EvalFunc(slc)
	}
}
