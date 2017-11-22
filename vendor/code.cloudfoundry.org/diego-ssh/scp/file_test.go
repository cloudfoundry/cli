package scp_test

import (
	"bytes"
	"crypto/rand"
	"errors"
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

var _ = Describe("File Message", func() {
	var (
		tempDir  string
		tempFile string

		logger     *lagertest.TestLogger
		testCopier TestCopier
	)

	newTestCopier := func(stdin io.Reader, stdout io.Writer, stderr io.Writer, preserveTimeAndMode bool) TestCopier {
		options := &scp.Options{
			PreserveTimesAndMode: preserveTimeAndMode,
		}
		secureCopier, ok := scp.New(options, stdin, stdout, stderr, logger).(TestCopier)
		Expect(ok).To(BeTrue())
		return secureCopier
	}

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		var err error
		tempDir, err = ioutil.TempDir("", "scp")
		Expect(err).NotTo(HaveOccurred())

		fileContents := make([]byte, 1024)
		tempFile = filepath.Join(tempDir, "binary.dat")

		_, err = rand.Read(fileContents)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(tempFile, fileContents, 0640)
		Expect(err).NotTo(HaveOccurred())

		modificationTime := time.Unix(123456789, 12345678)
		accessTime := time.Unix(987654321, 987654321)
		err = os.Chtimes(tempFile, accessTime, modificationTime)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("when sending the file to an scp sink", func() {
		var (
			file     *os.File
			fileInfo os.FileInfo
			err      error
		)

		BeforeEach(func() {
			file, err = os.Open(tempFile)
			Expect(err).NotTo(HaveOccurred())

			fileInfo, err = file.Stat()
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			file.Close()
		})

		It("sends the file message and contents to the sink", func() {
			stdin := bytes.NewReader([]byte{0, 0})
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			testCopier = newTestCopier(stdin, stdout, stderr, false)
			err := testCopier.SendFile(file, fileInfo)
			Expect(err).NotTo(HaveOccurred())

			cMessage, err := stdout.ReadString('\n')
			Expect(cMessage).To(Equal("C0640 1024 binary.dat\n"))

			contents := make([]byte, 1024)
			n, err := stdout.Read(contents)
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(1024))

			expectedContents, err := ioutil.ReadFile(tempFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal(expectedContents))

			confirmation := make([]byte, 1)
			bytesRead, err := stdout.Read(confirmation)
			Expect(err).NotTo(HaveOccurred())
			Expect(bytesRead).To(Equal(1))
			Expect(confirmation).To(Equal([]byte{0}))
		})

		It("waits for confirmation before sending the file contents", func() {
			stdin := &fake_io.FakeReader{}
			stdout := &fake_io.FakeWriter{}
			stdoutBuffer := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			stdout.WriteStub = stdoutBuffer.Write
			stdin.ReadStub = func(buffer []byte) (int, error) {
				if stdin.ReadCallCount() == 1 {
					cMessage, err := stdoutBuffer.ReadString('\n')
					Expect(err).NotTo(HaveOccurred())
					Expect(cMessage).To(Equal("C0640 1024 binary.dat\n"))
					Expect(stdoutBuffer.Len()).To(Equal(0))
				} else {
					contents := make([]byte, 1024)
					n, err := stdoutBuffer.Read(contents)
					Expect(err).NotTo(HaveOccurred())
					Expect(n).To(Equal(1024))

					expectedContents, err := ioutil.ReadFile(tempFile)
					Expect(err).NotTo(HaveOccurred())
					Expect(contents).To(Equal(expectedContents))
				}

				buffer[0] = 0
				return 1, nil
			}

			testCopier = newTestCopier(stdin, stdout, stderr, false)
			err := testCopier.SendFile(file, fileInfo)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not return before the contents are confirmed", func() {
			stdin, pw := io.Pipe()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			errCh := make(chan error, 1)
			go func() {
				testCopier = newTestCopier(stdin, stdout, stderr, false)
				errCh <- testCopier.SendFile(file, fileInfo)
			}()

			Consistently(errCh).ShouldNot(Receive())

			pw.Write([]byte{0})
			Consistently(errCh).ShouldNot(Receive())

			pw.Write([]byte{0})
			Eventually(errCh).Should(Receive(BeNil()))
		})

		Context("when preserving time stamps", func() {
			It("sends the time information before the file message", func() {
				stdin := bytes.NewReader([]byte{0, 0, 0})
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				testCopier = newTestCopier(stdin, stdout, stderr, true)
				err := testCopier.SendFile(file, fileInfo)
				Expect(err).NotTo(HaveOccurred())

				tMessage, err := stdout.ReadString('\n')
				Expect(tMessage).To(Equal("T123456789 0 987654321 0\n"))

				cMessage, err := stdout.ReadString('\n')
				Expect(cMessage).To(Equal("C0640 1024 binary.dat\n"))

				contents := make([]byte, 1024)
				n, err := stdout.Read(contents)
				Expect(err).NotTo(HaveOccurred())
				Expect(n).To(Equal(1024))

				expectedContents, err := ioutil.ReadFile(tempFile)
				Expect(err).NotTo(HaveOccurred())
				Expect(contents).To(Equal(expectedContents))
			})
		})

		Context("when copy encounters a short read", func() {
			It("returns with an error", func() {
				stdin := bytes.NewReader([]byte{0, 0})
				stdout := &fake_io.FakeWriter{}
				stderr := &bytes.Buffer{}

				stdout.WriteStub = func(buffer []byte) (int, error) {
					f, err := os.OpenFile(tempFile, os.O_RDWR, 0640)
					Expect(err).NotTo(HaveOccurred())

					err = f.Truncate(512)
					Expect(err).NotTo(HaveOccurred())

					err = f.Close()
					Expect(err).NotTo(HaveOccurred())

					return len(buffer), nil
				}

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.SendFile(file, fileInfo)
				Expect(err).To(Equal(io.EOF))
			})
		})

		Context("when the sink responds with a warning", func() {
			var stdin, stdout, stderr *bytes.Buffer

			BeforeEach(func() {
				stdin = &bytes.Buffer{}
				stdout = &bytes.Buffer{}
				stderr = &bytes.Buffer{}

				testCopier = newTestCopier(stdin, stdout, stderr, false)

				stdin.WriteByte(1)
				stdin.WriteString("Danger!\n")

				stdin.WriteByte(0)
			})

			It("returns without an error", func() {
				err := testCopier.SendFile(file, fileInfo)
				Expect(err).NotTo(HaveOccurred())
			})

			It("writes the message to stderr", func() {
				testCopier.SendFile(file, fileInfo)
				Expect(stderr.String()).To(Equal("Danger!"))
			})
		})

		Context("when the sink responds with a warning", func() {
			var stdin, stdout, stderr *bytes.Buffer

			BeforeEach(func() {
				stdin = &bytes.Buffer{}
				stdout = &bytes.Buffer{}
				stderr = &bytes.Buffer{}

				testCopier = newTestCopier(stdin, stdout, stderr, false)

				stdin.WriteByte(2)
				stdin.WriteString("oops...\n")

				stdin.WriteByte(0)
			})

			It("returns with an error", func() {
				err := testCopier.SendFile(file, fileInfo)
				Expect(err).To(MatchError("oops..."))
			})
		})

		Context("when the sink responds with an invalid acknowledgement", func() {
			var stdin, stdout, stderr *bytes.Buffer

			BeforeEach(func() {
				stdin = &bytes.Buffer{}
				stdout = &bytes.Buffer{}
				stderr = &bytes.Buffer{}

				testCopier = newTestCopier(stdin, stdout, stderr, false)

				stdin.WriteByte('a')
			})

			It("returns with an error", func() {
				err := testCopier.SendFile(file, fileInfo)
				Expect(err).To(MatchError("invalid acknowledgement identifier: 61"))
			})
		})

		Context("when the file is a directory", func() {
			It("fails and returns an error", func() {
				dir, err := os.Open(tempDir)
				Expect(err).NotTo(HaveOccurred())

				dirInfo, err := dir.Stat()
				Expect(err).NotTo(HaveOccurred())

				testCopier = newTestCopier(nil, nil, nil, false)
				Expect(testCopier.SendFile(dir, dirInfo)).To(HaveOccurred())
			})
		})

		Context("when sending the confirmation fails", func() {
			It("returns an error", func() {
				stdin := bytes.NewReader([]byte{0, 0})
				stdout := &fake_io.FakeWriter{}
				stderr := &bytes.Buffer{}

				stdout.WriteStub = func(buffer []byte) (int, error) {
					if stdout.WriteCallCount() == 3 {
						return 0, errors.New("BOOM")
					} else {
						return len(buffer), nil
					}
				}

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.SendFile(file, fileInfo)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("when receiving a file message from an scp source", func() {
		It("creates the file with the received contents", func() {
			stdin := &bytes.Buffer{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			stdin.WriteString("C0640 5 hello.txt\n")
			stdin.WriteString("hello")
			stdin.WriteByte(0)

			testCopier = newTestCopier(stdin, stdout, stderr, false)
			err := testCopier.ReceiveFile(tempFile, false, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(ioutil.ReadFile(tempFile)).To(BeEquivalentTo("hello"))
		})

		It("sends a confirmation after each message is received", func() {
			stdinBuffer := &bytes.Buffer{}
			stdin := &fake_io.FakeReader{}
			stdout := &fake_io.FakeWriter{}
			stderr := &bytes.Buffer{}

			stdinBuffer.WriteString("C0640 5 hello.txt\n")
			stdinBuffer.WriteString("hello")

			stdin.ReadStub = func(buffer []byte) (int, error) {
				b, err := stdinBuffer.ReadByte()
				if err != nil {
					return 0, err
				}
				buffer[0] = b
				return 1, nil
			}

			stdout.WriteStub = func(message []byte) (int, error) {
				if stdout.WriteCallCount() == 1 {
					Expect(stdin.ReadCallCount()).To(BeNumerically(">", 0))
					Expect(stdinBuffer.Len()).To(Equal(len("hello")))
					stdinBuffer.WriteByte(0)

					Expect(message).To(HaveLen(1))
					Expect(message[0]).To(BeEquivalentTo(0))
				} else {
					Expect(stdinBuffer.Len()).To(Equal(0))

					Expect(message).To(HaveLen(1))
					Expect(message[0]).To(BeEquivalentTo(0))
				}

				return 1, nil
			}

			testCopier = newTestCopier(stdin, stdout, stderr, false)
			err := testCopier.ReceiveFile(tempFile, false, nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.WriteCallCount()).To(Equal(2))

			Expect(ioutil.ReadFile(tempFile)).To(BeEquivalentTo("hello"))
		})

		It("sets the permissions of the file", func() {
			stdin := &bytes.Buffer{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			stdin.WriteString("C0444 5 hello.txt\n")
			stdin.WriteString("hello")
			stdin.WriteByte(0)

			testCopier = newTestCopier(stdin, stdout, stderr, true)
			err := testCopier.ReceiveFile(tempDir, true, nil)
			Expect(err).NotTo(HaveOccurred())

			fileInfo, err := os.Stat(filepath.Join(tempDir, "hello.txt"))
			Expect(err).NotTo(HaveOccurred())

			Expect(fileInfo.Mode()).To(Equal(os.FileMode(0444)))
		})

		It("sets the timestamp of the file if present", func() {
			stdin := &bytes.Buffer{}
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			stdin.WriteString("C0444 5 hello.txt\n")
			stdin.WriteString("hello")
			stdin.WriteByte(0)

			tempFileInfo, err := os.Stat(tempFile)
			Expect(err).NotTo(HaveOccurred())
			timestamp := scp.NewTimeMessage(tempFileInfo)

			testCopier = newTestCopier(stdin, stdout, stderr, true)
			err = testCopier.ReceiveFile(tempDir, true, timestamp)
			Expect(err).NotTo(HaveOccurred())

			fileInfo, err := os.Stat(filepath.Join(tempDir, "hello.txt"))
			Expect(err).NotTo(HaveOccurred())

			Expect(fileInfo.ModTime()).To(Equal(tempFileInfo.ModTime()))

			fileAtime, err := atime.AccessTime(fileInfo)
			Expect(err).NotTo(HaveOccurred())

			tempAtime, err := atime.AccessTime(tempFileInfo)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileInfo.ModTime()).To(Equal(tempFileInfo.ModTime()))
			Expect(fileAtime).To(Equal(tempAtime))
		})

		It("waits for a confirmation that the file has been sent", func() {
			stdin, pw := io.Pipe()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			errCh := make(chan error, 1)
			go func() {
				testCopier = newTestCopier(stdin, stdout, stderr, false)
				errCh <- testCopier.ReceiveFile(tempFile, false, nil)
			}()

			pw.Write([]byte("C0640 5 hello.txt\n"))
			pw.Write([]byte("hello"))

			Consistently(errCh).ShouldNot(Receive())

			pw.Write([]byte{0})
			Eventually(errCh).Should(Receive(BeNil()))
		})

		Context("when preserving time stamps and mode", func() {
			It("restores the access time and modification time", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("C0444 5 hello.txt\n")
				stdin.WriteString("hello")
				stdin.WriteByte(0)

				tempFileInfo, err := os.Stat(tempFile)
				Expect(err).NotTo(HaveOccurred())

				timeMessage := scp.NewTimeMessage(tempFileInfo)

				testCopier = newTestCopier(stdin, stdout, stderr, true)
				err = testCopier.ReceiveFile(tempFile, false, timeMessage)
				Expect(err).NotTo(HaveOccurred())

				fileInfo, err := os.Stat(tempFile)
				Expect(err).NotTo(HaveOccurred())

				fileAccessTime, err := atime.AccessTime(fileInfo)
				Expect(err).NotTo(HaveOccurred())

				expectedAccessTime, err := atime.AccessTime(tempFileInfo)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileInfo.ModTime()).To(Equal(tempFileInfo.ModTime()))
				Expect(fileAccessTime).To(Equal(expectedAccessTime))
			})
		})

		Context("when the file already exists", func() {
			var (
				preserveTimestampsAndMode bool
				target                    string
			)

			BeforeEach(func() {
				preserveTimestampsAndMode = false
			})

			JustBeforeEach(func() {
				target = filepath.Join(tempDir, "hello.txt")
				err := ioutil.WriteFile(target, []byte("goodbye"), 0600)
				Expect(err).NotTo(HaveOccurred())

				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("C0640 5 hello.txt\n")
				stdin.WriteString("hello")
				stdin.WriteByte(0)

				testCopier = newTestCopier(stdin, stdout, stderr, preserveTimestampsAndMode)
				err = testCopier.ReceiveFile(target, false, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("replaces the file with the received contents", func() {
				Expect(ioutil.ReadFile(target)).To(BeEquivalentTo("hello"))
			})

			It("does not change the permissions of the file", func() {
				file, err := os.Open(target)
				Expect(err).NotTo(HaveOccurred())
				defer file.Close()

				fileInfo, err := file.Stat()
				Expect(err).NotTo(HaveOccurred())

				Expect(fileInfo.Mode()).To(Equal(os.FileMode(0600 & 0777)))
			})

			Context("and preserving mode is set", func() {
				BeforeEach(func() {
					preserveTimestampsAndMode = true
				})

				It("changes permissions of the file", func() {
					file, err := os.Open(target)
					Expect(err).NotTo(HaveOccurred())
					defer file.Close()

					fileInfo, err := file.Stat()
					Expect(err).NotTo(HaveOccurred())

					Expect(fileInfo.Mode()).To(Equal(os.FileMode(0640)))
				})
			})
		})

		Context("when opening the target file fails", func() {
			BeforeEach(func() {
				err := os.Chmod(tempFile, 0400)
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("C0640 5 hello.txt\n")
				stdin.WriteString("hello")

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.ReceiveFile(tempFile, false, nil)
				Expect(err).To(MatchError(MatchRegexp("permission denied")))
			})
		})

		Context("when the message is not a file message", func() {
			It("fails with an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("c0640 5 hello.txt\n")
				stdin.WriteString("hello")

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.ReceiveFile(tempFile, false, nil)
				Expect(err).To(MatchError(`unexpected message type: c`))
			})
		})

		Context("when the file length field is not a number", func() {
			It("fails with an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("C0640 five hello.txt\n")
				stdin.WriteString("hello")

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.ReceiveFile(tempFile, false, nil)
				Expect(err).To(MatchError(`strconv.ParseInt: parsing "five": invalid syntax`))
			})
		})

		Context("when the file mode field is not an octal number", func() {
			It("fails with an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("C0999 5 hello.txt\n")
				stdin.WriteString("hello")

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.ReceiveFile(tempFile, false, nil)
				Expect(err).To(MatchError(`strconv.ParseUint: parsing "0999": invalid syntax`))
			})
		})

		Context("when the source does not send enough data for the file", func() {
			BeforeEach(func() {
				target := filepath.Join(tempDir, "hello.txt")
				err := ioutil.WriteFile(target, []byte("h"), 0660)
				Expect(err).NotTo(HaveOccurred())
			})

			It("fails with an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("C0640 512 hello.txt\n")
				stdin.WriteString("hello")

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.ReceiveFile(tempFile, false, nil)
				Expect(err).To(Equal(io.EOF))
			})
		})

		Context("when the target is a directory", func() {
			It("copies the file into the directory", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.WriteString("C0640 5 hello.txt\n")
				stdin.WriteString("hello")
				stdin.WriteByte(0)

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.ReceiveFile(tempDir, true, nil)
				Expect(err).NotTo(HaveOccurred())

				Expect(ioutil.ReadFile(filepath.Join(tempDir, "hello.txt"))).To(BeEquivalentTo("hello"))
			})
		})

		Context("when the confirmation of the file fails", func() {
			It("returns an error", func() {
				stdin := &bytes.Buffer{}
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}

				stdin.Write([]byte("C0640 5 hello.txt\n"))
				stdin.Write([]byte("hello"))
				stdin.Write([]byte{2})
				stdin.Write([]byte("BOOM\n"))

				testCopier = newTestCopier(stdin, stdout, stderr, false)
				err := testCopier.ReceiveFile(tempFile, false, nil)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
