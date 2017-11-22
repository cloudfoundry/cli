package scp_test

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/diego-ssh/scp"
	"code.cloudfoundry.org/diego-ssh/scp/atime"
	"code.cloudfoundry.org/diego-ssh/test_helpers/fake_io"
	"code.cloudfoundry.org/lager/lagertest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TimeMessage", func() {
	var (
		tempDir  string
		tempFile string

		logger *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")

		var err error
		tempDir, err = ioutil.TempDir("", "scp")
		Expect(err).NotTo(HaveOccurred())

		fileContents := make([]byte, 1024)
		tempFile = filepath.Join(tempDir, "binary.dat")

		_, err = rand.Read(fileContents)
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(tempFile, fileContents, 0400)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("when creating a time message from file information", func() {
		var (
			timeMessage *scp.TimeMessage

			expectedModificationTime time.Time
			expectedAccessTime       time.Time
		)

		BeforeEach(func() {
			fileInfo, ferr := os.Stat(tempFile)
			Expect(ferr).NotTo(HaveOccurred())

			expectedAccessTime, ferr = atime.AccessTime(fileInfo)
			Expect(ferr).NotTo(HaveOccurred())

			expectedModificationTime = fileInfo.ModTime()

			timeMessage = scp.NewTimeMessage(fileInfo)
		})

		It("acquires the correct modification time", func() {
			Expect(timeMessage.ModificationTime()).To(Equal(expectedModificationTime))
		})

		It("acquires the correct access time", func() {
			Expect(timeMessage.AccessTime()).To(Equal(expectedAccessTime))
		})
	})

	Context("when sending the time information to an scp sink", func() {
		var timeMessage *scp.TimeMessage

		BeforeEach(func() {
			modificationTime := time.Unix(123456789, 12345678)
			accessTime := time.Unix(987654321, 987654321)
			os.Chtimes(tempFile, accessTime, modificationTime)

			fileInfo, ferr := os.Stat(tempFile)
			Expect(ferr).NotTo(HaveOccurred())

			timeMessage = scp.NewTimeMessage(fileInfo)
		})

		It("sends the message with the appropriate times", func() {
			stdin := bytes.NewReader([]byte{0})
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			session := scp.NewSession(stdin, stdout, stderr, true, logger)

			err := timeMessage.Send(session)
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.String()).To(Equal("T123456789 0 987654321 0\n"))
		})

		It("writes the message before waiting for an acknowledgement", func() {
			stdin := &fake_io.FakeReader{}
			stdout := &fake_io.FakeWriter{}
			stdoutBuffer := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			session := scp.NewSession(stdin, stdout, stderr, true, logger)

			stdout.WriteStub = stdoutBuffer.Write
			stdin.ReadStub = func(buffer []byte) (int, error) {
				Expect(stdout.WriteCallCount()).To(BeNumerically(">", 0))

				buffer[0] = 0
				return 1, nil
			}

			err := timeMessage.Send(session)
			Expect(err).NotTo(HaveOccurred())

			Expect(stdin.ReadCallCount()).To(BeNumerically(">", 0))
			Expect(stdoutBuffer.String()).To(Equal("T123456789 0 987654321 0\n"))
		})

		It("does not return before the message is confirmed", func() {
			stdin, pw := io.Pipe()
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			session := scp.NewSession(stdin, stdout, stderr, true, logger)

			errCh := make(chan error, 1)
			go func() {
				errCh <- timeMessage.Send(session)
			}()

			Consistently(errCh).ShouldNot(Receive(HaveOccurred()))

			n, err := pw.Write([]byte{0})
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(1))

			Expect(stdout.String()).To(Equal("T123456789 0 987654321 0\n"))
		})

		Context("when the sink responds with a warning", func() {
			var stdin, stdout, stderr *bytes.Buffer
			var session *scp.Session

			BeforeEach(func() {
				stdin = &bytes.Buffer{}
				stdout = &bytes.Buffer{}
				stderr = &bytes.Buffer{}

				session = scp.NewSession(stdin, stdout, stderr, true, logger)

				stdin.WriteByte(1)
				stdin.WriteString("Danger!\n")
			})

			It("returns without an error", func() {
				err := timeMessage.Send(session)
				Expect(err).NotTo(HaveOccurred())

				Expect(stdout.String()).To(Equal("T123456789 0 987654321 0\n"))
			})

			It("writes the message to stderr", func() {
				timeMessage.Send(session)
				Expect(stderr.String()).To(Equal("Danger!"))
			})
		})

		Context("when the sink responds with an error ", func() {
			var stdin, stdout, stderr *bytes.Buffer
			var session *scp.Session

			BeforeEach(func() {
				stdin = &bytes.Buffer{}
				stdout = &bytes.Buffer{}
				stderr = &bytes.Buffer{}

				session = scp.NewSession(stdin, stdout, stderr, true, logger)

				stdin.WriteByte(2)
				stdin.WriteString("oops...\n")
			})

			It("returns with an error", func() {
				err := timeMessage.Send(session)
				Expect(err).To(MatchError("oops..."))
			})
		})
	})

	Context("when receiving a time message from an scp source", func() {
		var timeMessage *scp.TimeMessage

		BeforeEach(func() {
			timeMessage = &scp.TimeMessage{}
		})

		It("creates a time message with the appropriate information", func() {
			stdin := strings.NewReader("T123456789 0 987654321 0\n")
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			session := scp.NewSession(stdin, stdout, stderr, true, logger)

			err := timeMessage.Receive(session)
			Expect(err).NotTo(HaveOccurred())

			Expect(timeMessage.ModificationTime()).To(Equal(time.Unix(123456789, 0)))
			Expect(timeMessage.AccessTime()).To(Equal(time.Unix(987654321, 0)))
		})

		It("sends a confirmation after the message is received", func() {
			reader := strings.NewReader("T123456789 0 987654321 0\n")
			stdin := &fake_io.FakeReader{}
			stdout := &fake_io.FakeWriter{}
			stderr := &bytes.Buffer{}
			session := scp.NewSession(stdin, stdout, stderr, true, logger)

			stdin.ReadStub = reader.Read
			stdout.WriteStub = func(message []byte) (int, error) {
				Expect(stdin.ReadCallCount()).To(BeNumerically(">", 0))
				Expect(reader.Len()).To(Equal(0))

				Expect(message).To(HaveLen(1))
				Expect(message[0]).To(BeEquivalentTo(0))

				return 1, nil
			}

			err := timeMessage.Receive(session)
			Expect(err).NotTo(HaveOccurred())

			Expect(stdout.WriteCallCount()).To(BeNumerically(">", 0))
			Expect(stdout.WriteArgsForCall(0)).To(Equal([]byte{0}))
		})

		Context("when the message is not a time message", func() {
			It("fails with an error", func() {
				stdin := strings.NewReader("$123456789 0 987654321 0\n")
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}
				session := scp.NewSession(stdin, stdout, stderr, true, logger)

				err := timeMessage.Receive(session)
				Expect(err).To(MatchError("unexpected message type: $"))
			})
		})

		Context("when the modification time field is not a number", func() {
			It("fails with an error", func() {
				stdin := strings.NewReader("Tmodification 0 987654321 0\n")
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}
				session := scp.NewSession(stdin, stdout, stderr, true, logger)

				err := timeMessage.Receive(session)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the access time field is not a number", func() {
			It("fails with an error", func() {
				stdin := strings.NewReader("T123456789 0 access 0\n")
				stdout := &bytes.Buffer{}
				stderr := &bytes.Buffer{}
				session := scp.NewSession(stdin, stdout, stderr, true, logger)

				err := timeMessage.Receive(session)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
