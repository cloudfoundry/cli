package ui_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/util/ui"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("Request Logger Terminal Display", func() {
	var (
		out     *Buffer
		testUI  *UI
		display *RequestLoggerTerminalDisplay
	)

	BeforeEach(func() {
		out = NewBuffer()
		testUI = NewTestUI(nil, out, NewBuffer())
		display = testUI.RequestLoggerTerminalDisplay()
		Expect(display.Start()).ToNot(HaveOccurred())
	})

	Describe("DisplayBody", func() {
		It("displays the redacted value", func() {
			err := display.DisplayBody([]byte("some-string body"))
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("\\[PRIVATE DATA HIDDEN\\]"))
		})
	})

	Describe("DisplayDump", func() {
		It("displays the passed in string", func() {
			err := display.DisplayDump("some-dump-of-string")
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("some-dump-of-string"))
		})

		It("redacts auth tokens", func() {
			dump := `GET /apps/ce03a2e2-95c0-4f3b-abb9-32718d408c8b/stream HTTP/1.1
Host: wss://doppler.bosh-lite.com:443
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Version: 13
Sec-WebSocket-Key: [HIDDEN]
Authorization: bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImtleS0xIiwidHlwIjoiSldUIn0.eyJqdGkiOiI3YzRmYWEyZjI5MmQ0MTQ5ODM5NGE3OTU0Y2E3ZWNlMCIsInN1YiI6IjIyMjNiM2IzLTE3ZTktNDJkNi1iNzQzLThjZjcyZWIwOWRlNSIsInNjb3BlIjpbInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicm91dGluZy5yb3V0ZXJfZ3JvdXBzLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiLCJzY2ltLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLmFkbWluIiwidWFhLnVzZXIiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjIyMjNiM2IzLTE3ZTktNDJkNi1iNzQzLThjZjcyZWIwOWRlNSIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsInJldl9zaWciOiI4NDBiMDBhMyIsImlhdCI6MTQ5NjQyNTU5NiwiZXhwIjoxNDk2NDI2MTk2LCJpc3MiOiJodHRwczovL3VhYS5ib3NoLWxpdGUuY29tL29hdXRoL3Rva2VuIiwiemlkIjoidWFhIiwiYXVkIjpbInNjaW0iLCJjbG91ZF9jb250cm9sbGVyIiwicGFzc3dvcmQiLCJjZiIsInVhYSIsIm9wZW5pZCIsImRvcHBsZXIiLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMiXX0.TFDmHviKcs-eeNoz79dVwOl-k_dHTdqHkyztont2qnBDchNSpWvR5Yba54MMG8uTUHM72YbCopxdyaLY-g8s5wJFGLaBocrDgqswUh3mQRvynQG6_zne1h_0oHXnm0U-ZPnTyV8qjtHUoLvks4GOuktXc6ZE3NriWODpKIU5WdMgEbvyhuTnUEn88rQnmGJbKvHOIilulb6avSkZfTEq1o8w4VLCeRDlVLNh5JzCUtGzLfImNb31ks_Wv6HuI8kFjQZ5PQiTYjlhkuDQOcNSaAyWxQ_7425hiA7x8omBgEr-uST7GsxLvgoHqQaDH0JSTgMmO_GaN_QD52JVuru9og
Origin: wss://doppler.bosh-lite.com:443`
			err := display.DisplayDump(dump)
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Connection: Upgrade"))
			Expect(testUI.Out).To(Say("Authorization: \\[PRIVATE DATA HIDDEN\\]"))
			Expect(testUI.Out).To(Say("Origin: wss://doppler.bosh-lite.com:443"))
		})
	})

	Describe("DisplayHeader", func() {
		It("displays the header key and value", func() {
			err := display.DisplayHeader("Header", "Value")
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Header: Value"))
		})
	})

	Describe("DisplayHost", func() {
		It("displays the host", func() {
			err := display.DisplayHost("banana")
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("Host: banana"))
		})
	})

	Describe("DisplayJSONBody", func() {
		Context("when provided well formed JSON", func() {
			It("displayed a formated output", func() {
				raw := `{"a":"b", "c":"d", "don't html escape":"<&>"}`
				formatted := `{
  "a": "b",
  "c": "d",
  "don't html escape": "<&>"
}`
				err := display.DisplayJSONBody([]byte(raw))
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				Expect(testUI.Out).To(Say(formatted))
			})
		})

		Context("when the body is empty", func() {
			It("does not display the body", func() {
				err := display.DisplayJSONBody(nil)
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				Expect(string(out.Contents())).To(Equal("\n"))
			})
		})

		Context("when provided malformed JSON", func() {
			It("displays the raw body", func() {
				raw := `[{"data":1, "banana": 2}]`
				err := display.DisplayJSONBody([]byte(raw))
				Expect(err).ToNot(HaveOccurred())

				err = display.Stop()
				Expect(err).ToNot(HaveOccurred())

				buff, ok := testUI.Out.(*Buffer)
				Expect(ok).To(BeTrue())
				Expect(string(buff.Contents())).To(Equal(raw + "\n\n"))
			})
		})
	})

	Describe("DisplayMessage", func() {
		It("writes the message", func() {
			msg := "i am a message!!!!"
			err := display.DisplayMessage(msg)
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say(msg))
		})
	})

	Describe("DisplayRequestHeader", func() {
		It("displays the method, uri and http protocol", func() {
			err := display.DisplayRequestHeader("GET", "/v2/spaces/guid/summary", "HTTP/1.1")
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("GET /v2/spaces/guid/summary HTTP/1.1"))
		})
	})

	Describe("DisplayResponseHeader", func() {
		It("displays the method, uri and http protocol", func() {
			err := display.DisplayResponseHeader("HTTP/1.1", "200 OK")
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("HTTP/1.1 200 OK"))
		})
	})

	Describe("DisplayType", func() {
		It("displays the passed type and time in localized ISO 8601", func() {
			passedTime := time.Now()
			err := display.DisplayType("banana", passedTime)
			Expect(err).ToNot(HaveOccurred())

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Out).To(Say("banana: \\[%s\\]", passedTime.Format(time.RFC3339)))
		})
	})

	Describe("HandleInternalError", func() {
		It("sends error to standard error", func() {
			err := errors.New("foobar")
			display.HandleInternalError(err)

			err = display.Stop()
			Expect(err).ToNot(HaveOccurred())

			Expect(testUI.Err).To(Say("foobar"))
		})
	})

	Describe("Start and Stop", func() {
		It("locks and then unlocks the mutex properly", func() {
			c := make(chan bool)
			go func() {
				Expect(display.Start()).NotTo(HaveOccurred())
				c <- true
			}()
			Consistently(c).ShouldNot(Receive())
			Expect(display.Stop()).NotTo(HaveOccurred())
			Eventually(c).Should(Receive())
			Expect(display.Stop()).NotTo(HaveOccurred())
		})
	})
})
