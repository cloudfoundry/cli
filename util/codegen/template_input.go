package codegen

import (
	"strings"
	"unicode"
)

type TemplateInput struct {
	EntityName       string
	EntityNameSnake  string
	EntityNameDashes string
	EntityNameVar    string
}

func NewTemplateInput(entityName string) TemplateInput {
	return TemplateInput{
		EntityName:       entityName,
		EntityNameSnake:  toSnakeCase(entityName),
		EntityNameDashes: toDashes(entityName),
		EntityNameVar:    toVarName(entityName),
	}
}

func toSnakeCase(s string) string {
	var result []rune

	for i, roon := range s {
		if i > 0 && unicode.IsUpper(roon) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(roon))
	}

	return string(result)
}

func toDashes(s string) string {
	return strings.Replace(toSnakeCase(s), "_", "-", -1)
}

func toVarName(s string) string {
	var runes []rune

	runes = append(runes, unicode.ToLower(rune(s[0])))
	runes = append(runes, []rune(s[1:len(s)])...)

	return string(runes)
}
