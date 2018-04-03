package erbrenderer

type TemplateEvaluationContext interface {
	MarshalJSON() ([]byte, error)
}
