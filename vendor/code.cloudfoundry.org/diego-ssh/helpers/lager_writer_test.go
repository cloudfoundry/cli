package helpers_test

import (
	"io"

	"code.cloudfoundry.org/diego-ssh/helpers"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("LagerWriter", func() {
	var logger *lagertest.TestLogger
	var lagerWriter io.Writer

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		lagerWriter = helpers.NewLagerWriter(logger)
	})

	It("writes the payload as lager.Data", func() {
		payload := []byte("Hello, world!\n")

		n, err := lagerWriter.Write(payload)
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(len(payload)))

		Expect(logger.Logs()).To(HaveLen(1))

		log := logger.Logs()[0]
		Expect(log.Source).To(Equal("test"))
		Expect(log.Message).To(Equal("test.write"))
		Expect(log.LogLevel).To(Equal(lager.INFO))
		Expect(log.Data).To(Equal(lager.Data{"payload": string(payload)}))
	})
})
