package v7action_test

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/clock/fakeclock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"testing"
)

func TestV3Action(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V7 Actions Suite")
}

var _ = BeforeEach(func() {
	log.SetLevel(log.PanicLevel)
})

func NewTestActor() (*Actor, *v7actionfakes.FakeCloudControllerClient, *v7actionfakes.FakeConfig, *v7actionfakes.FakeSharedActor, *v7actionfakes.FakeUAAClient, *v7actionfakes.FakeRoutingClient, *fakeclock.FakeClock) {
	fakeCloudControllerClient := new(v7actionfakes.FakeCloudControllerClient)
	fakeConfig := new(v7actionfakes.FakeConfig)
	fakeSharedActor := new(v7actionfakes.FakeSharedActor)
	fakeUAAClient := new(v7actionfakes.FakeUAAClient)
	fakeRoutingClient := new(v7actionfakes.FakeRoutingClient)
	fakeClock := fakeclock.NewFakeClock(time.Now())
	actor := NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, fakeUAAClient, fakeRoutingClient, fakeClock)

	return actor, fakeCloudControllerClient, fakeConfig, fakeSharedActor, fakeUAAClient, fakeRoutingClient, fakeClock
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
