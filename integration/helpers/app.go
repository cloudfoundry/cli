package helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
)

func WithSimpleApp(f func(dir string)) {
	dir, err := ioutil.TempDir("", "simple-app")
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "index.html")
	err = ioutil.WriteFile(tempfile, []byte("hello world"), 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}
