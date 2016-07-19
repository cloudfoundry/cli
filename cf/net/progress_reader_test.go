package net_test

import (
	"os"
	"time"

	. "code.cloudfoundry.org/cli/cf/net"
	"code.cloudfoundry.org/cli/cf/terminal/terminalfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProgressReader", func() {
	var (
		testFile       *os.File
		err            error
		progressReader *ProgressReader
		ui             *terminalfakes.FakeUI
		b              []byte
		fileStat       os.FileInfo
	)

	BeforeEach(func() {
		ui = new(terminalfakes.FakeUI)

		testFile, err = os.Open("../../fixtures/test.file")
		Expect(err).NotTo(HaveOccurred())

		fileStat, err = testFile.Stat()
		Expect(err).NotTo(HaveOccurred())

		b = make([]byte, 1024)
		progressReader = NewProgressReader(testFile, ui, 1*time.Millisecond)
		progressReader.SetTotalSize(fileStat.Size())
	})

	It("prints progress while content is being read", func() {
		for {
			time.Sleep(50 * time.Microsecond)
			_, err := progressReader.Read(b)
			if err != nil {
				break
			}
		}

		Expect(ui.SayCallCount()).To(Equal(1))
		Expect(ui.SayArgsForCall(0)).To(ContainSubstring("\rDone "))

		Expect(ui.PrintCapturingNoOutputCallCount()).To(BeNumerically(">", 0))
		status, _ := ui.PrintCapturingNoOutputArgsForCall(0)
		Expect(status).To(ContainSubstring("uploaded..."))
		status, _ = ui.PrintCapturingNoOutputArgsForCall(ui.PrintCapturingNoOutputCallCount() - 1)
		Expect(status).To(Equal("\r                             "))
	})

	It("reads the correct number of bytes", func() {
		bytesRead := 0

		for {
			n, err := progressReader.Read(b)
			if err != nil {
				break
			}

			bytesRead += n
		}

		Expect(int64(bytesRead)).To(Equal(fileStat.Size()))
	})
})
