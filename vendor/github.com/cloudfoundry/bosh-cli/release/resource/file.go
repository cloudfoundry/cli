package resource

import (
	"path/filepath"
	"strings"
)

type File struct {
	Path    string
	DirPath string

	RelativePath string

	// Fingerprinting options
	UseBasename bool
	ExcludeMode bool
}

func NewFile(path, dirPath string) File {
	sep := string(filepath.Separator)

	// returns "inner" for args ("/tmp/inner", "/tmp")
	relativePath := strings.TrimPrefix(strings.TrimPrefix(path, dirPath), sep)

	return File{
		Path:         path,
		DirPath:      strings.TrimSuffix(dirPath, sep),
		RelativePath: relativePath,
	}
}

func NewFileUsesBasename(path, dirPath string) File {
	file := NewFile(path, dirPath)
	file.UseBasename = true
	return file
}

func (f File) WithNewDir(dirPath string) File {
	f.Path = filepath.Join(dirPath, f.RelativePath)
	f.DirPath = dirPath
	return f
}

type FileRelativePathSorting []File

func (s FileRelativePathSorting) Len() int { return len(s) }
func (s FileRelativePathSorting) Less(i, j int) bool {
	return s[i].RelativePath < s[j].RelativePath
}
func (s FileRelativePathSorting) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
