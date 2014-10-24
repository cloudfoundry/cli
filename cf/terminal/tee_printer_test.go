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

	})

	Describe("GetOutputAndReset", func() {
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
	})
})
