package push

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("pushing a path with the -p flag", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when the -p and -o flags are used together", func() {
		var path string

		BeforeEach(func() {
			tempFile, err := ioutil.TempFile("", "")
			Expect(err).ToNot(HaveOccurred())
			path = tempFile.Name()
		})

		AfterEach(func() {
			err := os.Remove(path)
			Expect(err).ToNot(HaveOccurred())
		})

		It("tells the user that they cannot be used together, displays usage and fails", func() {
			session := helpers.CF(PushCommandName, appName, "-o", DockerImage, "-p", path)

			Eventually(session.Err).Should(Say("Incorrect Usage: '--docker-image, -o' and '-p' cannot be used together\\."))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("USAGE:"))

			Eventually(session).Should(Exit(1))
		})
	})

	Context("pushing a directory", func() {
		It("pushes the app from the directory", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				session := helpers.CF(PushCommandName, appName, "-p", appDir)

				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("path:\\s+%s", appDir))
				Eventually(session).Should(Say("routes:"))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
				Eventually(session).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				Eventually(session).Should(Say("Staging app and tracing logs\\.\\.\\."))
				Eventually(session).Should(Say("name:\\s+%s", appName))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("pushing a zip file", func() {
		FIt("pushes the app from the zip file", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				tmpfile, err := ioutil.TempFile("", "example")
				Expect(err).ToNot(HaveOccurred())
				tmpfileName := tmpfile.Name()

				defer os.Remove(tmpfileName)
				err = zipit(appDir, tmpfileName, "")
				Expect(err).ToNot(HaveOccurred())

				session := helpers.CF(PushCommandName, appName, "-p", tmpfileName)

				Eventually(session).Should(Say("Getting app info\\.\\.\\."))
				Eventually(session).Should(Say("Creating app with these attributes\\.\\.\\."))
				Eventually(session).Should(Say("path:\\s+%s", tmpfileName))
				Eventually(session).Should(Say("routes:"))
				Eventually(session).Should(Say("Mapping routes\\.\\.\\."))
				Eventually(session).Should(Say("Packaging files to upload\\.\\.\\."))
				Eventually(session).Should(Say("Uploading files\\.\\.\\."))
				Eventually(session).Should(Say("Waiting for API to complete processing files\\.\\.\\."))
				Eventually(session).Should(Say("Staging app and tracing logs\\.\\.\\."))
				Eventually(session).Should(Say("name:\\s+%s", appName))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})

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
