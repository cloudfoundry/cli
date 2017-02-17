package flag

import (
	"fmt"
	"path/filepath"

	flags "github.com/jessevdk/go-flags"
)

type FilenameWithAt string

func (f FilenameWithAt) Complete(match string) []flags.Completion {
	if len(match) > 0 && match[0] == '@' {
		fileMatches, _ := filepath.Glob(fmt.Sprintf("%s*", match[1:len(match)]))
		matches := make([]flags.Completion, len(fileMatches))
		for i, fileMatch := range fileMatches {
			matches[i].Item = fmt.Sprintf("@%s", fileMatch)
		}

		return matches
	}

	return []flags.Completion{}
}
