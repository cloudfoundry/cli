package spellcheck

import (
	"github.com/sajari/fuzzy"
)

type CommandSuggester struct {
	model *fuzzy.Model
}

func (s CommandSuggester) Recommend(cmd string) []string {
	return s.model.Suggestions(cmd, true)
}

func NewCommandSuggester(existingCmds []string) CommandSuggester {
	model := fuzzy.NewModel()
	model.SetThreshold(1)
	model.SetDepth(1)

	model.Train(existingCmds)

	return CommandSuggester{model: model}
}
