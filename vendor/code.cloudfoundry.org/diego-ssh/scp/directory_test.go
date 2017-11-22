package scp_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/diego-ssh/scp"
	"code.cloudfoundry.org/diego-ssh/scp/atime"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_io"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Directory Message", func() {
	var (
		tempDir string
		logger  *lagertest.TestLogger
		err     error

		copier TestCopier
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		tempDir, err = ioutil.TempDir("", "scp")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	newTestCopier := func(stdin io.Reader, stdout io.Writer, stderr io.Writer, preserveTimeAndMode bool) TestCopier {
		options := &scp.Options{
			PreserveTimesAndMode: preserveTimeAndMode,
		}
		secureCopier, ok := scp.New(options, stdin, stdout, stderr, logger).(TestCopier)
		Expect(ok).To(BeTrue())
		return secureCopier
	}

	Context("when sending an empty directory to an scp sink", func() {
		var (
			emptySubdir  string
			emptyDirInfo os.FileInfo
		)

		BeforeEach(func() {
			emptySubdir = filepath.Join(tempDir, "empty-dir")
			err := os.Mkdir(emptySubdir, os.FileMode(0775))
			Expect(err).NotTo(HaveOccurred())

			err = os.Chmod(emptySubdir, 0775)
			Expect(err).NotTo(HaveOccurred())

			modificationTime := time.Unix(123456789, 12345678)
			accessTime := time.Unix(987654321, 987654321)
			err = os.Chtimes(emptySubdir, accessTime, modificationTime)
			Expect(err).NotTo(HaveOccurred())

			emptyDirInfo, err = os.Stat(emptySubdir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("sends the directory start and end messages", func() {
			stdin := bytes.NewReader([]byte{0, 0})
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			copier = newTestCopier(stdin, stdout, stderr, false)
			err := copier.SendDirectory(emptySubdir, emptyDirInfo)
			Expect(err).NotTo(HaveOccurred())

			dirMessage, err := stdout.ReadString('\n')
			Expect(dirMessage).To(Equal("D0775 0 empty-dir\n"))

			endMessage, err := stdout.ReadString('\n')
			Expect(endMessage).To(Equal("E\n"))
		})

		It("waits for confirmation of each message", func() {
			stdin := &fake_io.FakeReader{}
			stdout := &fake_io.FakeWriter{}
			stdoutBuffer := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			stdout.WriteStub = stdoutBuffer.Write
			stdin.ReadStub = func(buffer []byte) (int, error) {
				if stdin.ReadCallCount() == 1 {
					dMessage, err := stdoutBuffer.ReadString('\n')
					Expect(err).NotTo(HaveOccurred())
					Expect(dMessage).To(Equal("D0775 0 empty-dir\n"))
					Expect(stdoutBuffer.Len()).To(Equal(0))
				} else {
					eMessage, err := stdoutBuffer.ReadString('\n')
					Expect(err).NotTo(HaveOccurred())
					Expect(eMessage).To(Equal("E\n"))
					Expect(stdoutBuffer.Len()).To(Equal(0))
				}

				buffer[0] = 0
				return 1, nil
			}

			copier = newTestCopier(stdin, stdout, stderr, false)
			err := copier.SendDirectory(emptySubdir, emptyDirInfo)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return before the end message is confirmed", func() {
			stdin, pw := io.Pipe()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			errCh := make(chan error, 1)
			go func() {
				copier = newTestCopier(stdin, stdout, stderr, false)
				errCh <- copier.SendDirectory(emptySubdir, emptyDirInfo)
			}()

			Consistently(errCh).ShouldNot(Receive())

			pw.Write([]byte{0})
			Consistently(errCh).ShouldNot(Receive())

			pw.Write([]byte{0})
			Eventually(errCh).Should(Receive(BeNil()))
		})

		Context("when the directory cannot be opened", func() {
			BeforeEach(func() {
				err := os.Chmod(emptySubdir, 0222)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns an error", func() {
				stdin := bytes.NewReader([]byte{0, 0})
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				copier = newTestCopier(stdin, stdout, stderr, false)

				err := copier.SendDirectory(emptySubdir, emptyDirInfo)
				Expect(err).To(MatchError(MatchRegexp("permission denied")))
			})
		})

		Context("when preserving time stamps", func() {
			It("sends the time information before the file message", func() {
				stdin := bytes.NewReader([]byte{0, 0, 0})
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				copier = newTestCopier(stdin, stdout, stderr, true)
				err := copier.SendDirectory(emptySubdir, emptyDirInfo)
				Expect(err).NotTo(HaveOccurred())

				tMessage, err := stdout.ReadString('\n')
				Expect(tMessage).To(Equal("T123456789 0 987654321 0\n"))

				dMessage, err := stdout.ReadString('\n')
				Expect(err).NotTo(HaveOccurred())
				Expect(dMessage).To(Equal("D0775 0 empty-dir\n"))

				eMessage, err := stdout.ReadString('\n')
				Expect(err).NotTo(HaveOccurred())
				Expect(eMessage).To(Equal("E\n"))
			})
		})
	})

	Context("when sending an directory that contains files and directories", func() {
		var (
			dirInfo    os.FileInfo
			subdir     string
			subdirFile string
			tempFile   string
		)

		BeforeEach(func() {
			tempFile = filepath.Join(tempDir, "tempfile.txt")
			err := ioutil.WriteFile(tempFile, []byte("temporary-file-contents\n"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = os.Chmod(tempFile, 0644)
			Expect(err).NotTo(HaveOccurred())

			subdir = filepath.Join(tempDir, "subdir")
			err = os.Mkdir(subdir, os.FileMode(0700))
			Expect(err).NotTo(HaveOccurred())

			err = os.Chmod(subdir, 0700)
			Expect(err).NotTo(HaveOccurred())

			subdirFile = filepath.Join(subdir, "subdir-file.txt")
			err = ioutil.WriteFile(subdirFile, []byte("subdir-file-contents\n"), os.FileMode(0644))
			Expect(err).NotTo(HaveOccurred())

			err = os.Chmod(subdirFile, 0644)
			Expect(err).NotTo(HaveOccurred())

			emptySubdir := filepath.Join(tempDir, "empty-dir")
			err = os.Mkdir(emptySubdir, os.FileMode(0775))
			Expect(err).NotTo(HaveOccurred())

			err = os.Chmod(emptySubdir, 0775)
			Expect(err).NotTo(HaveOccurred())

			modificationTime := time.Unix(123456789, 12345678)
			accessTime := time.Unix(987654321, 987654321)
			err = os.Chtimes(emptySubdir, accessTime, modificationTime)
			Expect(err).NotTo(HaveOccurred())

			dirInfo, err = os.Stat(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("sends the correct messages", func() {
			stdin := bytes.NewReader(bytes.Repeat([]byte{0}, 10))
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			copier = newTestCopier(stdin, stdout, stderr, false)
			err := copier.SendDirectory(tempDir, dirInfo)
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.ReadString('\n')).To(Equal("D0700 0 " + filepath.Base(tempDir) + "\n"))
			Expect(stdout.ReadString('\n')).To(Equal("D0775 0 empty-dir\n"))
			Expect(stdout.ReadString('\n')).To(Equal("E\n"))
			Expect(stdout.ReadString('\n')).To(Equal("D0700 0 subdir\n"))
			Expect(stdout.ReadString('\n')).To(Equal("C0644 21 subdir-file.txt\n"))
			Expect(stdout.ReadString('\n')).To(Equal("subdir-file-contents\n"))
			Expect(stdout.ReadByte()).To(BeEquivalentTo(0))
			Expect(stdout.ReadString('\n')).To(Equal("E\n"))
			Expect(stdout.ReadString('\n')).To(Equal("C0644 24 tempfile.txt\n"))
			Expect(stdout.ReadString('\n')).To(Equal("temporary-file-contents\n"))
			Expect(stdout.ReadByte()).To(BeEquivalentTo(0))
			Expect(stdout.ReadString('\n')).To(Equal("E\n"))
		})

		Context("when sending a file fails", func() {
			var subdirFile2 string

			BeforeEach(func() {
				subdirFile2 = filepath.Join(subdir, "tempfile.txt")
				err := ioutil.WriteFile(subdirFile2, []byte("temporary-file-contents\n"), os.FileMode(0200))
				Expect(err).NotTo(HaveOccurred())

				err = os.Chmod(subdirFile, 0200)
				Expect(err).NotTo(HaveOccurred())
			})

			It("continues to send the other files", func() {
				stdin := bytes.NewReader(bytes.Repeat([]byte{0}, 10))
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				subdirInfo, err := os.Stat(subdir)
				Expect(err).NotTo(HaveOccurred())

				copier = newTestCopier(stdin, stdout, stderr, false)
				err = copier.SendDirectory(subdir, subdirInfo)
				Expect(err).NotTo(HaveOccurred())

				Expect(stdout.ReadString('\n')).To(Equal("D0700 0 subdir\n"))

				Expect(stdout.ReadByte()).To(BeEquivalentTo(1))
				Expect(stdout.ReadString('\n')).To(ContainSubstring("permission denied"))
				Expect(stdout.ReadByte()).To(BeEquivalentTo(1))
				Expect(stdout.ReadString('\n')).To(ContainSubstring("permission denied"))

				Expect(stdout.ReadString('\n')).To(Equal("E\n"))
			})
		})
	})

	Context("when receiving a directory from an scp source", func() {
		It("populates the directory with the received contents", func() {
			stdin := &bytes.Buffer{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			stdin.WriteString("D0700 0 received-dir\n")
			stdin.WriteString("D0755 0 empty-dir\n")
			stdin.WriteString("E\n")
			stdin.WriteString("D0700 0 subdir\n")
			stdin.WriteString("C0644 21 subdir-file.txt\n")
			stdin.WriteString("subdir-file-contents\n")
			stdin.WriteByte(0)
			stdin.WriteString("E\n")
			stdin.WriteString("C0600 24 tempfile.txt\n")
			stdin.WriteString("temporary-file-contents\n")
			stdin.WriteByte(0)
			stdin.WriteString("E\n")

			copier = newTestCopier(stdin, stdout, stderr, false)
			err := copier.ReceiveDirectory(tempDir, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(tempDir, "received-dir")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "received-dir", "empty-dir")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "received-dir", "subdir")).To(BeADirectory())
			Expect(filepath.Join(tempDir, "received-dir", "subdir", "subdir-file.txt")).To(BeARegularFile())
			Expect(filepath.Join(tempDir, "received-dir", "tempfile.txt")).To(BeARegularFile())

			info, err := os.Stat(filepath.Join(tempDir, "received-dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0700)))

			info, err = os.Stat(filepath.Join(tempDir, "received-dir", "empty-dir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0755)))

			info, err = os.Stat(filepath.Join(tempDir, "received-dir", "subdir"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0700)))

			info, err = os.Stat(filepath.Join(tempDir, "received-dir", "subdir", "subdir-file.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0644)))

			info, err = os.Stat(filepath.Join(tempDir, "received-dir", "tempfile.txt"))
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode() & 0777).To(Equal(os.FileMode(0600)))

			contents, err := ioutil.ReadFile(filepath.Join(tempDir, "received-dir", "subdir", "subdir-file.txt"))
			Expect(contents).To(BeEquivalentTo("subdir-file-contents\n"))

			contents, err = ioutil.ReadFile(filepath.Join(tempDir, "received-dir", "tempfile.txt"))
			Expect(contents).To(BeEquivalentTo("temporary-file-contents\n"))
		})

		Context("when preserving time stamps", func() {
			It("restores the access time and modification time", func() {
				timeStdin := &bytes.Buffer{}
				timeStdout := &bytes.Buffer{}
				timeStderr := &bytes.Buffer{}

				timeStdin.WriteString("T123456789 0 987654321 0\n")
				timeSession := scp.NewSession(timeStdin, timeStdout, timeStderr, true, logger)

				timeMessage := &scp.TimeMessage{}
				timeMessage.Receive(timeSession)

				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("D0755 0 empty-dir\n")
				stdin.WriteString("E\n")

				copier = newTestCopier(stdin, stdout, stderr, true)
				err := copier.ReceiveDirectory(tempDir, timeMessage)
				Expect(err).NotTo(HaveOccurred())

				Expect(filepath.Join(tempDir, "empty-dir")).To(BeADirectory())

				info, err := os.Stat(filepath.Join(tempDir, "empty-dir"))
				Expect(err).NotTo(HaveOccurred())

				accessTime, err := atime.AccessTime(info)
				Expect(err).NotTo(HaveOccurred())

				Expect(info.ModTime()).To(Equal(time.Unix(123456789, 0)))
				Expect(accessTime).To(Equal(time.Unix(987654321, 0)))
			})
		})

		Context("when the message is not a directory message", func() {
			It("raises an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("C0755 0 empty-dir\n")
				stdin.WriteString("E\n")

				copier = newTestCopier(stdin, stdout, stderr, false)
				err := copier.ReceiveDirectory(tempDir, nil)
				Expect(err).To(MatchError("unexpected message type: C"))
			})
		})

		Context("when the directory mode is not octal", func() {
			It("raises an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("D0999 0 empty-dir\n")
				stdin.WriteString("E\n")

				copier = newTestCopier(stdin, stdout, stderr, false)
				err := copier.ReceiveDirectory(tempDir, nil)
				Expect(err).To(MatchError(`strconv.ParseUint: parsing "0999": invalid syntax`))
			})
		})

		Context("when the ignored length field is not sent", func() {
			It("raises an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("D0755 empty-dir\n")
				stdin.WriteString("E\n")

				copier = newTestCopier(stdin, stdout, stderr, false)
				err := copier.ReceiveDirectory(tempDir, nil)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the directory end message is not sent", func() {
			It("raises an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("D0755 empty-dir\n")

				copier = newTestCopier(stdin, stdout, stderr, false)
				err := copier.ReceiveDirectory(tempDir, nil)
				Expect(err).To(Equal(io.EOF))
			})
		})

		Context("when creating the target directory fails", func() {
			var targetDir string

			BeforeEach(func() {
				targetDir = filepath.Join(tempDir, "target")
				err := os.Mkdir(targetDir, os.FileMode(0555))
				Expect(err).NotTo(HaveOccurred())

				err = os.Chmod(targetDir, 0555)
				Expect(err).NotTo(HaveOccurred())
			})

			It("raises an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("D0755 0 empty-dir\n")
				stdin.WriteString("E\n")

				copier = newTestCopier(stdin, stdout, stderr, false)
				err := copier.ReceiveDirectory(targetDir, nil)
				Expect(err).To(MatchError(MatchRegexp("permission denied")))
			})
		})

		Context("when the target directory does not exist", func() {
			Context("but the target enclosing directory does", func() {
				var targetDir string

				BeforeEach(func() {
					targetDir = filepath.Join(tempDir, "target")
					err := os.Mkdir(targetDir, os.FileMode(0777))
					Expect(err).NotTo(HaveOccurred())
				})

				It("makes the new target directory and populates it with the sources contents", func() {
					stdin := &bytes.Buffer{}
					stdout := &bytes.Buffer{}
					stderr := &bytes.Buffer{}

					stdin.WriteString("D0700 0 subdir\n")
					stdin.WriteString("C0644 21 subdir-file.txt\n")
					stdin.WriteString("subdir-file-contents\n")
					stdin.WriteByte(0)
					stdin.WriteString("E\n")

					copier = newTestCopier(stdin, stdout, stderr, false)
					err := copier.ReceiveDirectory(filepath.Join(targetDir, "newdir"), nil)
					Expect(err).NotTo(HaveOccurred())

					info, err := os.Stat(filepath.Join(tempDir, "target", "newdir"))
					Expect(err).NotTo(HaveOccurred())
					Expect(info.Mode() & 0777).To(Equal(os.FileMode(0700)))

					info, err = os.Stat(filepath.Join(tempDir, "target", "newdir", "subdir-file.txt"))
					Expect(err).NotTo(HaveOccurred())
					Expect(info.Mode() & 0777).To(Equal(os.FileMode(0644)))

					contents, err := ioutil.ReadFile(filepath.Join(tempDir, "target", "newdir", "subdir-file.txt"))
					Expect(contents).To(BeEquivalentTo("subdir-file-contents\n"))
				})
			})

			Context("and the enclosing target directory does not exist", func() {
				var targetDir string

				BeforeEach(func() {
					targetDir = filepath.Join(tempDir, "target")
					err := os.Mkdir(targetDir, os.FileMode(0777))
					Expect(err).NotTo(HaveOccurred())
				})

				It("fails", func() {
					stdin := &bytes.Buffer{}
					stdout := &bytes.Buffer{}
					stderr := &bytes.Buffer{}

					stdin.WriteString("D0700 0 empty-dir\n")
					stdin.WriteString("E\n")

					copier = newTestCopier(stdin, stdout, stderr, false)
					err := copier.ReceiveDirectory(filepath.Join(targetDir, "newdir", "newer-dir"), nil)
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("when the target directory already exists", func() {
			BeforeEach(func() {
				dir := filepath.Join(tempDir, "empty-dir")
				err := os.Mkdir(dir, os.FileMode(0775))
				Expect(err).NotTo(HaveOccurred())

				err = os.Chmod(dir, 0775)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not raise an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("D0755 0 empty-dir\n")
				stdin.WriteString("E\n")

				copier = newTestCopier(stdin, stdout, stderr, false)
				err := copier.ReceiveDirectory(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not change the permissions", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("D0755 0 empty-dir\n")
				stdin.WriteString("E\n")

				copier = newTestCopier(stdin, stdout, stderr, false)
				err := copier.ReceiveDirectory(tempDir, nil)
				Expect(err).NotTo(HaveOccurred())

				info, err := os.Stat(filepath.Join(tempDir, "empty-dir"))
				Expect(err).NotTo(HaveOccurred())
				Expect(info.Mode() & 0777).To(Equal(os.FileMode(0775)))
			})
		})

		Context("when the target directory is really a file", func() {
			BeforeEach(func() {
				target := filepath.Join(tempDir, "empty-dir")
				err := ioutil.WriteFile(target, []byte("ego existo!"), 0660)
				Expect(err).NotTo(HaveOccurred())

				err = os.Chmod(target, 0660)
				Expect(err).NotTo(HaveOccurred())
			})

			It("raises an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("D0755 0 empty-dir\n")
				stdin.WriteString("E\n")

				copier = newTestCopier(stdin, stdout, stderr, false)
				err := copier.ReceiveDirectory(tempDir, nil)
				Expect(err).To(HaveOccurred())

				Expect(filepath.Join(tempDir, "empty-dir")).To(BeARegularFile())
			})
		})
	})
})
