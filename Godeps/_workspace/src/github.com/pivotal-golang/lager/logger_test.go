package lager_test

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Logger", func() {
	var logger lager.Logger
	var testSink *lagertest.TestSink

	var component = "my-component"
	var action = "my-action"
	var logData = lager.Data{
		"foo":      "bar",
		"a-number": 7,
	}

	BeforeEach(func() {
		logger = lager.NewLogger(component)
		testSink = lagertest.NewTestSink()
		logger.RegisterSink(testSink)
	})

	var TestCommonLogFeatures = func(level lager.LogLevel) {
		var log lager.LogFormat

		BeforeEach(func() {
			log = testSink.Logs()[0]
		})

		It("writes a log to the sink", func() {
			Ω(testSink.Logs()).Should(HaveLen(1))
		})

		It("records the source component", func() {
			Ω(log.Source).Should(Equal(component))
		})

		It("outputs a properly-formatted message", func() {
			Ω(log.Message).Should(Equal(fmt.Sprintf("%s.%s", component, action)))
		})

		It("has a timestamp", func() {
			expectedTime := float64(time.Now().UnixNano()) / 1e9
			parsedTimestamp, err := strconv.ParseFloat(log.Timestamp, 64)
			Ω(err).ShouldNot(HaveOccurred())
			Ω(parsedTimestamp).Should(BeNumerically("~", expectedTime, 1.0))
		})

		It("sets the proper output level", func() {
			Ω(log.LogLevel).Should(Equal(level))
		})
	}

	var TestLogData = func() {
		var log lager.LogFormat

		BeforeEach(func() {
			log = testSink.Logs()[0]
		})

		It("data contains custom user data", func() {
			Ω(log.Data["foo"]).Should(Equal("bar"))
			Ω(log.Data["a-number"]).Should(BeNumerically("==", 7))
		})
	}

	Describe("Session", func() {
		var session lager.Logger

		BeforeEach(func() {
			session = logger.Session("sub-action")
		})

		Describe("the returned logger", func() {
			JustBeforeEach(func() {
				session.Debug("some-debug-action", lager.Data{"level": "debug"})
				session.Info("some-info-action", lager.Data{"level": "info"})
				session.Error("some-error-action", errors.New("oh no!"), lager.Data{"level": "error"})

				defer func() {
					recover()
				}()

				session.Fatal("some-fatal-action", errors.New("oh no!"), lager.Data{"level": "fatal"})
			})

			It("logs with a shared session id in the data", func() {
				Ω(testSink.Logs()[0].Data["session"]).Should(Equal("1"))
				Ω(testSink.Logs()[1].Data["session"]).Should(Equal("1"))
				Ω(testSink.Logs()[2].Data["session"]).Should(Equal("1"))
				Ω(testSink.Logs()[3].Data["session"]).Should(Equal("1"))
			})

			It("logs with the task added to the message", func() {
				Ω(testSink.Logs()[0].Message).Should(Equal("my-component.sub-action.some-debug-action"))
				Ω(testSink.Logs()[1].Message).Should(Equal("my-component.sub-action.some-info-action"))
				Ω(testSink.Logs()[2].Message).Should(Equal("my-component.sub-action.some-error-action"))
				Ω(testSink.Logs()[3].Message).Should(Equal("my-component.sub-action.some-fatal-action"))
			})

			It("logs with the original data", func() {
				Ω(testSink.Logs()[0].Data["level"]).Should(Equal("debug"))
				Ω(testSink.Logs()[1].Data["level"]).Should(Equal("info"))
				Ω(testSink.Logs()[2].Data["level"]).Should(Equal("error"))
				Ω(testSink.Logs()[3].Data["level"]).Should(Equal("fatal"))
			})

			Context("with data", func() {
				BeforeEach(func() {
					session = logger.Session("sub-action", lager.Data{"foo": "bar"})
				})

				It("logs with the data added to the message", func() {
					Ω(testSink.Logs()[0].Data["foo"]).Should(Equal("bar"))
					Ω(testSink.Logs()[1].Data["foo"]).Should(Equal("bar"))
					Ω(testSink.Logs()[2].Data["foo"]).Should(Equal("bar"))
					Ω(testSink.Logs()[3].Data["foo"]).Should(Equal("bar"))
				})

				It("keeps the original data", func() {
					Ω(testSink.Logs()[0].Data["level"]).Should(Equal("debug"))
					Ω(testSink.Logs()[1].Data["level"]).Should(Equal("info"))
					Ω(testSink.Logs()[2].Data["level"]).Should(Equal("error"))
					Ω(testSink.Logs()[3].Data["level"]).Should(Equal("fatal"))
				})
			})

			Context("with another session", func() {
				BeforeEach(func() {
					session = logger.Session("next-sub-action")
				})

				It("logs with a shared session id in the data", func() {
					Ω(testSink.Logs()[0].Data["session"]).Should(Equal("2"))
					Ω(testSink.Logs()[1].Data["session"]).Should(Equal("2"))
					Ω(testSink.Logs()[2].Data["session"]).Should(Equal("2"))
					Ω(testSink.Logs()[3].Data["session"]).Should(Equal("2"))
				})

				It("logs with the task added to the message", func() {
					Ω(testSink.Logs()[0].Message).Should(Equal("my-component.next-sub-action.some-debug-action"))
					Ω(testSink.Logs()[1].Message).Should(Equal("my-component.next-sub-action.some-info-action"))
					Ω(testSink.Logs()[2].Message).Should(Equal("my-component.next-sub-action.some-error-action"))
					Ω(testSink.Logs()[3].Message).Should(Equal("my-component.next-sub-action.some-fatal-action"))
				})
			})

			Context("with a nested session", func() {
				BeforeEach(func() {
					session = session.Session("sub-sub-action")
				})

				It("logs with a shared session id in the data", func() {
					Ω(testSink.Logs()[0].Data["session"]).Should(Equal("1.1"))
					Ω(testSink.Logs()[1].Data["session"]).Should(Equal("1.1"))
					Ω(testSink.Logs()[2].Data["session"]).Should(Equal("1.1"))
					Ω(testSink.Logs()[3].Data["session"]).Should(Equal("1.1"))
				})

				It("logs with the task added to the message", func() {
					Ω(testSink.Logs()[0].Message).Should(Equal("my-component.sub-action.sub-sub-action.some-debug-action"))
					Ω(testSink.Logs()[1].Message).Should(Equal("my-component.sub-action.sub-sub-action.some-info-action"))
					Ω(testSink.Logs()[2].Message).Should(Equal("my-component.sub-action.sub-sub-action.some-error-action"))
					Ω(testSink.Logs()[3].Message).Should(Equal("my-component.sub-action.sub-sub-action.some-fatal-action"))
				})
			})
		})
	})

	Describe("Debug", func() {
		Context("with log data", func() {
			BeforeEach(func() {
				logger.Debug(action, logData)
			})

			TestCommonLogFeatures(lager.DEBUG)
			TestLogData()
		})

		Context("with no log data", func() {
			BeforeEach(func() {
				logger.Debug(action)
			})

			TestCommonLogFeatures(lager.DEBUG)
		})
	})

	Describe("Info", func() {
		Context("with log data", func() {
			BeforeEach(func() {
				logger.Info(action, logData)
			})

			TestCommonLogFeatures(lager.INFO)
			TestLogData()
		})

		Context("with no log data", func() {
			BeforeEach(func() {
				logger.Info(action)
			})

			TestCommonLogFeatures(lager.INFO)
		})
	})

	Describe("Error", func() {
		var err = errors.New("oh noes!")
		Context("with log data", func() {
			BeforeEach(func() {
				logger.Error(action, err, logData)
			})

			TestCommonLogFeatures(lager.ERROR)
			TestLogData()

			It("data contains error message", func() {
				Ω(testSink.Logs()[0].Data["error"]).Should(Equal(err.Error()))
			})
		})

		Context("with no log data", func() {
			BeforeEach(func() {
				logger.Error(action, err)
			})

			TestCommonLogFeatures(lager.ERROR)

			It("data contains error message", func() {
				Ω(testSink.Logs()[0].Data["error"]).Should(Equal(err.Error()))
			})
		})

		Context("with no error", func() {
			BeforeEach(func() {
				logger.Error(action, nil)
			})

			TestCommonLogFeatures(lager.ERROR)

			It("does not contain the error message", func() {
				Ω(testSink.Logs()[0].Data).ShouldNot(HaveKey("error"))
			})
		})
	})

	Describe("Fatal", func() {
		var err = errors.New("oh noes!")
		var fatalErr interface{}

		Context("with log data", func() {
			BeforeEach(func() {
				defer func() {
					fatalErr = recover()
				}()

				logger.Fatal(action, err, logData)
			})

			TestCommonLogFeatures(lager.FATAL)
			TestLogData()

			It("data contains error message", func() {
				Ω(testSink.Logs()[0].Data["error"]).Should(Equal(err.Error()))
			})

			It("data contains stack trace", func() {
				Ω(testSink.Logs()[0].Data["trace"]).ShouldNot(BeEmpty())
			})

			It("panics with the provided error", func() {
				Ω(fatalErr).Should(Equal(err))
			})
		})

		Context("with no log data", func() {
			BeforeEach(func() {
				defer func() {
					fatalErr = recover()
				}()

				logger.Fatal(action, err)
			})

			TestCommonLogFeatures(lager.FATAL)

			It("data contains error message", func() {
				Ω(testSink.Logs()[0].Data["error"]).Should(Equal(err.Error()))
			})

			It("data contains stack trace", func() {
				Ω(testSink.Logs()[0].Data["trace"]).ShouldNot(BeEmpty())
			})

			It("panics with the provided error", func() {
				Ω(fatalErr).Should(Equal(err))
			})
		})

		Context("with no error", func() {
			BeforeEach(func() {
				defer func() {
					fatalErr = recover()
				}()

				logger.Fatal(action, nil)
			})

			TestCommonLogFeatures(lager.FATAL)

			It("does not contain the error message", func() {
				Ω(testSink.Logs()[0].Data).ShouldNot(HaveKey("error"))
			})

			It("data contains stack trace", func() {
				Ω(testSink.Logs()[0].Data["trace"]).ShouldNot(BeEmpty())
			})

			It("panics with the provided error (i.e. nil)", func() {
				Ω(fatalErr).Should(BeNil())
			})
		})
	})
})
