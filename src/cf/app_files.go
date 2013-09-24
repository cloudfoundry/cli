package cf

import (
	"io"
	"os"
	"path/filepath"
)

func AppFilesInDir(dir string) (appFiles []AppFile, err error) {
	err = walkAppFiles(dir, func(fileName string, _ string) {
		appFiles = append(appFiles, AppFile{Path: fileName})
	})
	return
}

func TempDirForApp(app Application) (dir string) {
	dir = filepath.Join(os.TempDir(), "cf", app.Guid)
	return
}

func InitializeDir(dir string) (err error) {
	err = os.RemoveAll(dir)
	if err != nil {
		return
	}
	err = os.MkdirAll(dir, os.ModeDir|os.ModeTemporary|os.ModePerm)
	return
}

func CopyFiles(appFiles []AppFile, fromDir, toDir string) (err error) {
	if err != nil {
		return
	}

	for _, file := range appFiles {
		fromPath := filepath.Join(fromDir, file.Path)
		toPath := filepath.Join(toDir, file.Path)
		err = copyFile(fromPath, toPath)
		if err != nil {
			return
		}
	}
	return
}

func copyFile(fromPath, toPath string) (err error) {
	err = os.MkdirAll(filepath.Dir(toPath), os.ModeDir|os.ModeTemporary|os.ModePerm)
	if err != nil {
		return
	}

	src, err := os.Open(fromPath)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(toPath)
	if err != nil {
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return
}

type walkAppFileFunc func(fileName, fullPath string)

func walkAppFiles(dir string, onEachFile walkAppFileFunc) (err error) {
	exclusions := readCfIgnore(dir)

	walkFunc := func(fullPath string, f os.FileInfo, inErr error) (err error) {
		err = inErr
		if err != nil {
			return
		}

		if f.IsDir() {
			return
		}

		fileName, _ := filepath.Rel(dir, fullPath)
		if fileShouldBeIgnored(exclusions, fileName) {
			return
		}

		onEachFile(fileName, fullPath)

		return
	}

	err = filepath.Walk(dir, walkFunc)
	return
}
