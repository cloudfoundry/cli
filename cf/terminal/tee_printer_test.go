package terminal_test

import (
	"os"

	. "code.cloudfoundry.org/cli/cf/terminal"

	io_helpers "code.cloudfoundry.org/cli/util/testhelpers/io"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("TeePrinter", func() {
	var (
		output  []string
		printer *TeePrinter
	)

	Describe(".Print", func() {
		var bucket *gbytes.Buffer

		BeforeEach(func() {
			bucket = gbytes.NewBuffer()

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter(os.Stdout)
				printer.SetOutputBucket(bucket)
				printer.Print("Hello ")
				printer.Print("Mom!")
			})
		})

		It("should delegate to fmt.Print", func() {
			Expect(output[0]).To(Equal("Hello Mom!"))
		})

		It("should save the output to the slice", func() {
			Expect(bucket).To(gbytes.Say("Hello "))
			Expect(bucket).To(gbytes.Say("Mom!"))
		})

		It("should decolorize text", func() {
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter(os.Stdout)
				printer.SetOutputBucket(bucket)
				printer.Print("hi " + EntityNameColor("foo"))
			})

			Expect(bucket).To(gbytes.Say("hi foo"))
		})
	})

	Describe(".Printf", func() {
		var bucket *gbytes.Buffer

		BeforeEach(func() {
			bucket = gbytes.NewBuffer()

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter(os.Stdout)
				printer.SetOutputBucket(bucket)
				printer.Printf("Hello %s", "everybody")
			})
		})

		It("should delegate to fmt.Printf", func() {
			Expect(output[0]).To(Equal("Hello everybody"))
		})

		It("should save the output to the slice", func() {
			Expect(bucket).To(gbytes.Say("Hello everybody"))
		})

		It("should decolorize text", func() {
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter(os.Stdout)
				printer.SetOutputBucket(bucket)
				printer.Printf("hi %s", EntityNameColor("foo"))
			})

			Expect(bucket).To(gbytes.Say("hi foo"))
		})
	})

	Describe(".Println", func() {
		var bucket *gbytes.Buffer
		BeforeEach(func() {
			bucket = gbytes.NewBuffer()

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter(os.Stdout)
				printer.SetOutputBucket(bucket)
				printer.Println("Hello ", "everybody")
			})
		})

		It("should delegate to fmt.Printf", func() {
			Expect(output[0]).To(Equal("Hello everybody"))
		})

		It("should save the output to the slice", func() {
			Expect(bucket).To(gbytes.Say("Hello everybody"))
		})

		It("should decolorize text", func() {
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter(os.Stdout)
				printer.SetOutputBucket(bucket)
				printer.Println("hi " + EntityNameColor("foo"))
			})

			Expect(bucket).To(gbytes.Say("hi foo"))
		})
	})

	Describe(".SetOutputBucket", func() {
		It("sets the []string used to save the output", func() {
			bucket := gbytes.NewBuffer()

			output := io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter(os.Stdout)
				printer.SetOutputBucket(bucket)
				printer.Printf("Hello %s", "everybody")
			})

			Expect(bucket).To(gbytes.Say("Hello everybody"))
			Expect(output).To(ContainElement("Hello everybody"))
		})

		It("disables the output saving when set to nil", func() {
			output := io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter(os.Stdout)
				printer.SetOutputBucket(nil)
				printer.Printf("Hello %s", "everybody")
			})
			Expect(output).To(ContainElement("Hello everybody"))
		})
	})
})
