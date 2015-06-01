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
		var bucket *[]string

		BeforeEach(func() {
			bucket = &[]string{}

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.Print("Hello ")
				printer.Print("Mom!")
			})
		})

		It("should delegate to fmt.Print", func() {
			Expect(output[0]).To(Equal("Hello Mom!"))
		})

		It("should save the output to the slice", func() {
			Expect((*bucket)[0]).To(Equal("Hello "))
			Expect((*bucket)[1]).To(Equal("Mom!"))
		})

		It("should decolorize text", func() {
			bucket = &[]string{}
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.Print("hi " + EntityNameColor("foo"))
			})

			Expect((*bucket)[0]).To(Equal("hi foo"))
		})
	})

	Describe(".Printf", func() {
		var bucket *[]string

		BeforeEach(func() {
			bucket = &[]string{}

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.Printf("Hello %s", "everybody")
			})
		})

		It("should delegate to fmt.Printf", func() {
			Expect(output[0]).To(Equal("Hello everybody"))
		})

		It("should save the output to the slice", func() {
			Expect((*bucket)[0]).To(Equal("Hello everybody"))
		})

		It("should decolorize text", func() {
			bucket = &[]string{}
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.Printf("hi %s", EntityNameColor("foo"))
			})

			Expect((*bucket)[0]).To(Equal("hi foo"))
		})
	})

	Describe(".Println", func() {
		var bucket *[]string
		BeforeEach(func() {
			bucket = &[]string{}

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.Println("Hello ", "everybody")
			})
		})

		It("should delegate to fmt.Printf", func() {
			Expect(output[0]).To(Equal("Hello everybody"))
		})

		It("should save the output to the slice", func() {
			Expect((*bucket)[0]).To(Equal("Hello everybody"))
		})

		It("should decolorize text", func() {
			bucket = &[]string{}
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.Println("hi " + EntityNameColor("foo"))
			})

			Expect((*bucket)[0]).To(Equal("hi foo"))
		})
	})

	Describe(".ForcePrintf", func() {
		var bucket *[]string

		BeforeEach(func() {
			bucket = &[]string{}

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.ForcePrintf("Hello %s", "everybody")
			})
		})

		It("should delegate to fmt.Printf", func() {
			Expect(output[0]).To(Equal("Hello everybody"))
		})

		It("should save the output to the slice", func() {
			Expect((*bucket)[0]).To(Equal("Hello everybody"))
		})

		It("should decolorize text", func() {
			bucket = &[]string{}
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.Printf("hi %s", EntityNameColor("foo"))
			})

			Expect((*bucket)[0]).To(Equal("hi foo"))
		})
	})

	Describe(".ForcePrintln", func() {
		var bucket *[]string

		BeforeEach(func() {
			bucket = &[]string{}

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.ForcePrintln("Hello ", "everybody")
			})
		})

		It("should delegate to fmt.Printf", func() {
			Expect(output[0]).To(Equal("Hello everybody"))
		})

		It("should save the output to the slice", func() {
			Expect((*bucket)[0]).To(Equal("Hello everybody"))
		})

		It("should decolorize text", func() {
			bucket = &[]string{}
			io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.Println("hi " + EntityNameColor("foo"))
			})

			Expect((*bucket)[0]).To(Equal("hi foo"))
		})
	})

	Describe(".SetOutputBucket", func() {
		var bucket *[]string

		output = io_helpers.CaptureOutput(func() {
			bucket = &[]string{}
			printer = NewTeePrinter()
			printer.SetOutputBucket(bucket)
			printer.ForcePrintf("Hello %s", "everybody")
		})

		It("sets the []string used to save the output", func() {
			Expect((*bucket)[0]).To(Equal("Hello everybody"))
		})

		It("disables the output saving when set to nil", func() {
			printer.SetOutputBucket(nil)
			Expect((*bucket)[0]).To(Equal("Hello everybody"))
		})
	})

	Describe("Pausing Output", func() {
		var bucket *[]string

		BeforeEach(func() {
			bucket = &[]string{}

			output = io_helpers.CaptureOutput(func() {
				printer = NewTeePrinter()
				printer.SetOutputBucket(bucket)
				printer.DisableTerminalOutput(true)
				printer.Print("Hello")
				printer.Println("Mom!")
				printer.Printf("Dad!")
				printer.ForcePrint("Forced Hello")
				printer.ForcePrintln("Forced Mom")
				printer.ForcePrintf("Forced Dad")
			})
		})

		It("should print only forced terminal output", func() {
			Expect(output).To(Equal([]string{"Forced HelloForced Mom", "Forced Dad"}))
		})

		It("should still capture all output", func() {
			Expect(*bucket).To(Equal([]string{"Hello", "Mom!", "Dad!", "Forced Hello", "Forced Mom", "Forced Dad"}))
		})

		Describe(".ResumeOutput", func() {
			var bucket *[]string
			BeforeEach(func() {
				bucket = &[]string{}

				output = io_helpers.CaptureOutput(func() {
					printer.SetOutputBucket(bucket)
					printer.DisableTerminalOutput(false)
					printer.Print("Hello")
					printer.Println("Mom!")
					printer.Printf("Dad!")
					printer.Println("Grandpa!")
					printer.ForcePrint("ForcePrint")
					printer.ForcePrintln("ForcePrintln")
					printer.ForcePrintf("ForcePrintf")
				})
			})

			It("should print all output", func() {
				Expect(output).To(Equal([]string{"HelloMom!", "Dad!Grandpa!", "ForcePrintForcePrintln", "ForcePrintf"}))
			})

			It("should capture all output", func() {
				Expect(*bucket).To(Equal([]string{"Hello", "Mom!", "Dad!", "Grandpa!", "ForcePrint", "ForcePrintln", "ForcePrintf"}))
			})
		})
	})
})
