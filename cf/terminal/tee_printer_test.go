package terminal_test

import (
	. "github.com/cloudfoundry/cli/cf/terminal"

	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TeePrinter", func() {
	var (
		output  []string
		printer *TeePrinter
	)

	Describe(".Print", func() {
		BeforeEach(func() {
			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.Print("Hello ")
				printer.Print("Mom!")
			})
		})

		It("should delegate to fmt.Print", func() {
			Expect(output[0]).To(Equal("Hello Mom!"))
		})

		It("should save the output to the slice", func() {
			outputs := printer.GetOutputAndReset()
			Expect(outputs[0]).To(Equal("Hello "))
			Expect(outputs[1]).To(Equal("Mom!"))
		})

		It("should decolorize text", func() {
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.Print("hi " + EntityNameColor("foo"))
			})

			output = printer.GetOutputAndReset()
			Expect(output[0]).To(Equal("hi foo"))
		})
	})

	Describe(".Printf", func() {
		BeforeEach(func() {
			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.Printf("Hello %s", "everybody")
			})
		})

		It("should delegate to fmt.Printf", func() {
			Expect(output[0]).To(Equal("Hello everybody"))
		})

		It("should save the output to the slice", func() {
			Expect(printer.GetOutputAndReset()[0]).To(Equal("Hello everybody"))
		})

		It("should decolorize text", func() {
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.Printf("hi %s", EntityNameColor("foo"))
			})

			output = printer.GetOutputAndReset()
			Expect(output[0]).To(Equal("hi foo"))
		})
	})

	Describe(".Println", func() {
		BeforeEach(func() {
			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.Println("Hello ", "everybody")
			})
		})

		It("should delegate to fmt.Printf", func() {
			Expect(output[0]).To(Equal("Hello everybody"))
		})

		It("should save the output to the slice", func() {
			Expect(printer.GetOutputAndReset()[0]).To(Equal("Hello everybody"))
		})

		It("should decolorize text", func() {
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.Println("hi " + EntityNameColor("foo"))
			})

			output = printer.GetOutputAndReset()
			Expect(output[0]).To(Equal("hi foo"))
		})
	})

	Describe(".GetOutputAndReset", func() {
		BeforeEach(func() {
			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.Print("Hello")
				printer.Print("Mom!")
			})
		})

		It("should clear the slice after access", func() {
			printer.GetOutputAndReset()
			Expect(printer.GetOutputAndReset()).To(BeEmpty())
		})
	})

	Describe("Pausing Output", func() {
		BeforeEach(func() {
			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.DisableTerminalOutput(true)
				printer.Print("Hello")
				printer.Println("Mom!")
				printer.Printf("Dad!")
			})
		})

		It("should print no terminal output", func() {
			Expect(output).To(Equal([]string{""}))
		})

		It("should still capture output", func() {
			Expect(printer.GetOutputAndReset()[0]).To(Equal("Hello"))
		})

		Describe(".ResumeOutput", func() {
			BeforeEach(func() {
				printer.GetOutputAndReset()
				output = io_helpers.CaptureOutput(func() {
					printer.DisableTerminalOutput(false)
					printer.Print("Hello")
					printer.Println("Mom!")
					printer.Printf("Dad!")
					printer.Println("Grandpa!")
				})
			})

			It("should print all output", func() {
				Expect(output[1]).To(Equal("Dad!Grandpa!"))
			})

			It("should print terminal output", func() {
				Expect(printer.GetOutputAndReset()[3]).To(Equal("Grandpa!"))
			})
		})
	})
})
