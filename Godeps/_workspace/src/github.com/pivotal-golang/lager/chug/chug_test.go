package chug_test

import (
	"errors"
	"io"
	"time"

	"github.com/pivotal-golang/lager"
	. "github.com/pivotal-golang/lager/chug"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Chug", func() {
	var (
		logger     lager.Logger
		stream     chan Entry
		pipeReader *io.PipeReader
		pipeWriter *io.PipeWriter
	)

	BeforeEach(func() {
		pipeReader, pipeWriter = io.Pipe()
		logger = lager.NewLogger("chug-test")
		logger.RegisterSink(lager.NewWriterSink(pipeWriter, lager.DEBUG))
		stream = make(chan Entry, 100)
		go Chug(pipeReader, stream)
	})

	AfterEach(func() {
		pipeWriter.Close()
		Eventually(stream).Should(BeClosed())
	})

	Context("when fed a stream of well-formed lager messages", func() {
		It("should return parsed lager messages", func() {
			data := lager.Data{"some-float": 3.0, "some-string": "foo"}
			logger.Debug("chug", data)
			logger.Info("again", data)

			entry := <-stream
			Ω(entry.IsLager).Should(BeTrue())
			Ω(entry.Log).Should(MatchLogEntry(LogEntry{
				LogLevel: lager.DEBUG,
				Source:   "chug-test",
				Message:  "chug",
				Data:     data,
			}))

			entry = <-stream
			Ω(entry.IsLager).Should(BeTrue())
			Ω(entry.Log).Should(MatchLogEntry(LogEntry{
				LogLevel: lager.INFO,
				Source:   "chug-test",
				Message:  "again",
				Data:     data,
			}))
		})

		It("should parse the timestamp", func() {
			logger.Debug("chug")
			entry := <-stream
			Ω(entry.Log.Timestamp).Should(BeTemporally("~", time.Now(), 10*time.Millisecond))
		})

		Context("when parsing an error message", func() {
			It("should include the error", func() {
				data := lager.Data{"some-float": 3.0, "some-string": "foo"}
				logger.Error("chug", errors.New("some-error"), data)
				Ω((<-stream).Log).Should(MatchLogEntry(LogEntry{
					LogLevel: lager.ERROR,
					Source:   "chug-test",
					Message:  "chug",
					Error:    errors.New("some-error"),
					Data:     lager.Data{"some-float": 3.0, "some-string": "foo"},
				}))
			})
		})

		Context("when multiple sessions have been established", func() {
			It("should build up the task array appropriately", func() {
				firstSession := logger.Session("first-session")
				firstSession.Info("encabulate")
				nestedSession := firstSession.Session("nested-session-1")
				nestedSession.Info("baconize")
				firstSession.Info("remodulate")
				nestedSession.Info("ergonomize")
				nestedSession = firstSession.Session("nested-session-2")
				nestedSession.Info("modernify")

				Ω((<-stream).Log).Should(MatchLogEntry(LogEntry{
					LogLevel: lager.INFO,
					Source:   "chug-test",
					Message:  "first-session.encabulate",
					Session:  "1",
					Data:     lager.Data{},
				}))

				Ω((<-stream).Log).Should(MatchLogEntry(LogEntry{
					LogLevel: lager.INFO,
					Source:   "chug-test",
					Message:  "first-session.nested-session-1.baconize",
					Session:  "1.1",
					Data:     lager.Data{},
				}))

				Ω((<-stream).Log).Should(MatchLogEntry(LogEntry{
					LogLevel: lager.INFO,
					Source:   "chug-test",
					Message:  "first-session.remodulate",
					Session:  "1",
					Data:     lager.Data{},
				}))

				Ω((<-stream).Log).Should(MatchLogEntry(LogEntry{
					LogLevel: lager.INFO,
					Source:   "chug-test",
					Message:  "first-session.nested-session-1.ergonomize",
					Session:  "1.1",
					Data:     lager.Data{},
				}))

				Ω((<-stream).Log).Should(MatchLogEntry(LogEntry{
					LogLevel: lager.INFO,
					Source:   "chug-test",
					Message:  "first-session.nested-session-2.modernify",
					Session:  "1.2",
					Data:     lager.Data{},
				}))
			})
		})
	})

	Context("handling lager JSON that is surrounded by non-JSON", func() {
		var input []byte
		var entry Entry

		BeforeEach(func() {
			input = []byte(`[some-component][e]{"timestamp":"1407102779.028711081","source":"chug-test","message":"chug-test.chug","log_level":0,"data":{"some-float":3,"some-string":"foo"}}...some trailing stuff`)
			pipeWriter.Write(input)
			pipeWriter.Write([]byte("\n"))

			Eventually(stream).Should(Receive(&entry))
		})

		It("should be a lager message", func() {
			Ω(entry.IsLager).Should(BeTrue())
		})

		It("should contain all the data in Raw", func() {
			Ω(entry.Raw).Should(Equal(input))
		})

		It("should succesfully parse the lager message", func() {
			Ω(entry.Log.Source).Should(Equal("chug-test"))
		})
	})

	Context("handling malformed/non-lager data", func() {
		var input []byte
		var entry Entry

		JustBeforeEach(func() {
			pipeWriter.Write(input)
			pipeWriter.Write([]byte("\n"))

			Eventually(stream).Should(Receive(&entry))
		})

		itReturnsRawData := func() {
			It("returns raw data", func() {
				Ω(entry.IsLager).Should(BeFalse())
				Ω(entry.Log).Should(BeZero())
				Ω(entry.Raw).Should(Equal(input))
			})
		}

		Context("when fed a stream of malformed lager messages", func() {
			Context("when the timestamp is invalid", func() {
				BeforeEach(func() {
					input = []byte(`{"timestamp":"tomorrow","source":"chug-test","message":"chug-test.chug","log_level":3,"data":{"some-float":3,"some-string":"foo","error":7}}`)
				})

				itReturnsRawData()
			})

			Context("when the error does not parse", func() {
				BeforeEach(func() {
					input = []byte(`{"timestamp":"1407102779.028711081","source":"chug-test","message":"chug-test.chug","log_level":3,"data":{"some-float":3,"some-string":"foo","error":7}}`)
				})

				itReturnsRawData()
			})

			Context("when the trace does not parse", func() {
				BeforeEach(func() {
					input = []byte(`{"timestamp":"1407102779.028711081","source":"chug-test","message":"chug-test.chug","log_level":3,"data":{"some-float":3,"some-string":"foo","trace":7}}`)
				})

				itReturnsRawData()
			})

			Context("when the session does not parse", func() {
				BeforeEach(func() {
					input = []byte(`{"timestamp":"1407102779.028711081","source":"chug-test","message":"chug-test.chug","log_level":3,"data":{"some-float":3,"some-string":"foo","session":7}}`)
				})

				itReturnsRawData()
			})

			Context("when there aren't enough message components", func() {
				BeforeEach(func() {
					input = []byte(`{"timestamp":"1407102779.028711081","source":"chug-test","message":"chug-test","log_level":1,"data":{}}`)
				})

				itReturnsRawData()
			})
		})

		Context("When fed JSON that is not a lager message at all", func() {
			BeforeEach(func() {
				input = []byte(`{"source":"chattanooga"}`)
			})

			itReturnsRawData()
		})

		Context("When fed none-JSON that is not a lager message at all", func() {
			BeforeEach(func() {
				input = []byte(`ß`)
			})

			itReturnsRawData()
		})
	})
})
