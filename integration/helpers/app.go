package helpers

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"
)

// WithHelloWorldApp creates a simple application to use with your CLI command
// (typically CF Push). When pushing, be aware of specifying '-b
// staticfile_buildpack" so that your app will correctly start up with the
// proper buildpack.
func WithHelloWorldApp(f func(dir string)) {
	dir, err := ioutil.TempDir("", "simple-app")
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "index.html")
	err = ioutil.WriteFile(tempfile, []byte("hello world"), 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}

// WithBananaPantsApp creates a simple application to use with your CLI command
// (typically CF Push). When pushing, be aware of specifying '-b
// staticfile_buildpack" so that your app will correctly start up with the
// proper buildpack.
func WithBananaPantsApp(f func(dir string)) {
	dir, err := ioutil.TempDir("", "simple-app")
	Expect(err).ToNot(HaveOccurred())
	defer os.RemoveAll(dir)

	tempfile := filepath.Join(dir, "index.html")
	err = ioutil.WriteFile(tempfile, []byte("Banana Pants"), 0666)
	Expect(err).ToNot(HaveOccurred())

	f(dir)
}
