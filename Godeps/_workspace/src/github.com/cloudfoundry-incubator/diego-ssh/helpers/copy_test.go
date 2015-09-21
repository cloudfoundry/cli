package helpers_test

import (
	"io"
	"strings"
	"sync"

	"github.com/cloudfoundry-incubator/diego-ssh/helpers"
	"github.com/cloudfoundry-incubator/diego-ssh/test_helpers/fake_io"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Copy", func() {
	var logger lager.Logger

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
	})

	Describe("Copy", func() {
		var reader io.Reader
		var fakeWriter *fake_io.FakeWriter
		var wg *sync.WaitGroup

		BeforeEach(func() {
			reader = strings.NewReader("message")
			fakeWriter = &fake_io.FakeWriter{}
			wg = nil
		})

		JustBeforeEach(func() {
			helpers.Copy(logger, wg, fakeWriter, reader)
		})

		It("copies from source to target", func() {
			Expect(fakeWriter.WriteCallCount()).To(Equal(1))
			Expect(string(fakeWriter.WriteArgsForCall(0))).To(Equal("message"))
		})

		Context("when a wait group is provided", func() {
			BeforeEach(func() {
				wg = &sync.WaitGroup{}
				wg.Add(1)
			})

			It("calls done before returning", func() {
				wg.Wait()
			})
		})
	})

	Describe("CopyAndClose", func() {
		var reader io.Reader
		var fakeWriteCloser *fake_io.FakeWriteCloser
		var wg *sync.WaitGroup

		BeforeEach(func() {
			reader = strings.NewReader("message")
			fakeWriteCloser = &fake_io.FakeWriteCloser{}
			wg = nil
		})

		JustBeforeEach(func() {
			helpers.CopyAndClose(logger, wg, fakeWriteCloser, reader)
		})

		It("copies from source to target", func() {
			Expect(fakeWriteCloser.WriteCallCount()).To(Equal(1))
			Expect(string(fakeWriteCloser.WriteArgsForCall(0))).To(Equal("message"))
		})

		It("closes the target when the copy is complete", func() {
			Expect(fakeWriteCloser.CloseCallCount()).To(Equal(1))
		})

		Context("when a wait group is provided", func() {
			BeforeEach(func() {
				wg = &sync.WaitGroup{}
				wg.Add(1)
			})

			It("calls done before returning", func() {
				wg.Wait()
			})
		})
	})
})
