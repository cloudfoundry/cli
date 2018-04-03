package cmd

import (
	"strings"
)

type URLArg string

func (a URLArg) IsEmpty() bool {
	return len(a) == 0
}

func (a URLArg) IsRemote() bool {
	return strings.HasPrefix(string(a), "https://") || strings.HasPrefix(string(a), "http://")
}

func (a URLArg) FilePath() string {
	return strings.TrimPrefix(string(a), "file://")
}

func (a URLArg) IsGit() bool {
	return strings.HasPrefix(string(a), "git://") || strings.HasPrefix(string(a), "git+") || strings.HasPrefix(string(a), "git@")
}

func (a URLArg) GitRepo() string {
	return strings.TrimPrefix(string(a), "git+")
}
