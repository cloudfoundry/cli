package net_test

import (
	"os"
	"time"

	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/net"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ProgressReader", func() {

	var (
		testFile       *os.File
		err            error
		progressReader *ProgressReader
		ui             *testterm.FakeUI
		b              []byte
		fileStat       os.FileInfo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		testFile, err = os.Open("../../fixtures/test.file")
		Expect(err).ToNot(HaveOccurred())
		fileStat, err = testFile.Stat()
		Expect(err).ToNot(HaveOccurred())

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

		Expect(ui.UncapturedOutput).To(ContainSubstrings([]string{"\r", "uploaded..."}))
		Expect(ui.UncapturedOutput).To(ContainSubstrings([]string{"\r    "}))
		Expect(ui.Outputs).To(ContainSubstrings([]string{"\rDone "}))
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
