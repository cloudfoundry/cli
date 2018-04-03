package fakes

type FakeTemplateEvaluationContext struct {
}

func (f FakeTemplateEvaluationContext) MarshalJSON() ([]byte, error) {
	return []byte("{}"), nil
}
