package lager_test

import (
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("ReconfigurableSink", func() {
	var (
		testSink *lagertest.TestSink

		sink *lager.ReconfigurableSink
	)

	BeforeEach(func() {
		testSink = lagertest.NewTestSink()

		sink = lager.NewReconfigurableSink(testSink, lager.INFO)
	})

	It("returns the current level", func() {
		Expect(sink.GetMinLevel()).To(Equal(lager.INFO))
	})

	Context("when logging above the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.INFO, []byte("hello world"))
		})

		It("writes to the given sink", func() {
			Expect(testSink.Buffer()).To(gbytes.Say("hello world\n"))
		})
	})

	Context("when logging below the minimum log level", func() {
		BeforeEach(func() {
			sink.Log(lager.DEBUG, []byte("hello world"))
		})

		It("does not write to the given writer", func() {
			Expect(testSink.Buffer().Contents()).To(BeEmpty())
		})
	})

	Context("when reconfigured to a new log level", func() {
		BeforeEach(func() {
			sink.SetMinLevel(lager.DEBUG)
		})

		It("writes logs above the new log level", func() {
			sink.Log(lager.DEBUG, []byte("hello world"))
			Expect(testSink.Buffer()).To(gbytes.Say("hello world\n"))
		})

		It("returns the newly updated level", func() {
			Expect(sink.GetMinLevel()).To(Equal(lager.DEBUG))
		})
	})
})
