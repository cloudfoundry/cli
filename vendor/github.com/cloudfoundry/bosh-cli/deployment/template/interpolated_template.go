package template

import ()

type InterpolatedTemplate struct {
	content []byte
	sha     string
}

func (t *InterpolatedTemplate) Content() []byte {
	return t.content
}

func (t *InterpolatedTemplate) SHA() string {
	return t.sha
}

func NewInterpolatedTemplate(content []byte, shaSumString string) InterpolatedTemplate {
	return InterpolatedTemplate{
		content: content,
		sha:     shaSumString,
	}
}
