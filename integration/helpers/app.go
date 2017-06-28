package helpers

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
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
	err = ioutil.WriteFile(tempfile, []byte(fmt.Sprintf("hello world %d", rand.Int())), 0666)
	Expect(err).ToNot(HaveOccurred())

	err = ioutil.WriteFile(filepath.Join(dir, "Staticfile"), nil, 0666)
	Expect(err).ToNot(HaveOccurred())

	prevDir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())

	err = os.Chdir(dir)
	Expect(err).ToNot(HaveOccurred())
	defer func() {
		err = os.Chdir(prevDir)
		Expect(err).ToNot(HaveOccurred())
	}()

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

	err = ioutil.WriteFile(filepath.Join(dir, "Staticfile"), nil, 0666)
	Expect(err).ToNot(HaveOccurred())

	prevDir, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())

	err = os.Chdir(dir)
	Expect(err).ToNot(HaveOccurred())
	defer func() {
		err = os.Chdir(prevDir)
		Expect(err).ToNot(HaveOccurred())
	}()

	f(dir)
}

// AppGUID returns the GUID for an app in the currently targeted space.
func AppGUID(appName string) string {
	session := CF("app", appName, "--guid")
	Eventually(session).Should(gexec.Exit(0))
	return strings.TrimSpace(string(session.Out.Contents()))
}
