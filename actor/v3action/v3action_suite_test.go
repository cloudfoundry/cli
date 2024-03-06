package v3action_test

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"

	. "code.cloudfoundry.org/cli/actor/v3action"
	"code.cloudfoundry.org/cli/actor/v3action/v3actionfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"testing"
)

func TestV3Action(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V3 Actions Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

func NewTestActor() (*Actor, *v3actionfakes.FakeCloudControllerClient, *v3actionfakes.FakeConfig, *v3actionfakes.FakeSharedActor, *v3actionfakes.FakeUAAClient) {
	fakeCloudControllerClient := new(v3actionfakes.FakeCloudControllerClient)
	fakeConfig := new(v3actionfakes.FakeConfig)
	fakeSharedActor := new(v3actionfakes.FakeSharedActor)
	fakeUAAClient := new(v3actionfakes.FakeUAAClient)
	actor := NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, fakeUAAClient)

	return actor, fakeCloudControllerClient, fakeConfig, fakeSharedActor, fakeUAAClient
}

// Thanks to Svett Ralchev
// http://blog.ralch.com/tutorial/golang-working-with-zip/
func zipit(source, target, prefix string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	if prefix != "" {
		_, err = io.WriteString(zipfile, prefix)
		if err != nil {
			return err
		}
	}

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(path, source)

		if info.IsDir() {
			header.Name += string(os.PathSeparator)
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return err
}
