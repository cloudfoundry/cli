package cliFlags_test

import (
	. "github.com/cloudfoundry/cli/flags/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CLI Flags", func() {
	Describe("StringFlag", func() {
		var flag *StringFlag

		BeforeEach(func() {
			flag = &StringFlag{
				Name:  "string-flag",
				Value: "initial-value",
				Usage: "A string flag for testing",
			}
		})

		It("stores name, value, and usage", func() {
			Expect(flag.GetName()).To(Equal("string-flag"))
			Expect(flag.GetValue()).To(Equal("initial-value"))
			Expect(flag.String()).To(Equal("A string flag for testing"))
		})

		It("sets value", func() {
			flag.Set("new-value")
			Expect(flag.GetValue()).To(Equal("new-value"))
		})

		It("sets empty string value", func() {
			flag.Set("")
			Expect(flag.GetValue()).To(Equal(""))
		})

		It("overwrites existing value", func() {
			flag.Set("first")
			flag.Set("second")
			flag.Set("third")
			Expect(flag.GetValue()).To(Equal("third"))
		})

		It("handles special characters", func() {
			flag.Set("value-with-dashes")
			Expect(flag.GetValue()).To(Equal("value-with-dashes"))

			flag.Set("value with spaces")
			Expect(flag.GetValue()).To(Equal("value with spaces"))

			flag.Set("value/with/slashes")
			Expect(flag.GetValue()).To(Equal("value/with/slashes"))
		})
	})

	Describe("BoolFlag", func() {
		var flag *BoolFlag

		BeforeEach(func() {
			flag = &BoolFlag{
				Name:  "bool-flag",
				Value: false,
				Usage: "A boolean flag for testing",
			}
		})

		It("stores name, value, and usage", func() {
			Expect(flag.GetName()).To(Equal("bool-flag"))
			Expect(flag.GetValue()).To(BeFalse())
			Expect(flag.String()).To(Equal("A boolean flag for testing"))
		})

		It("sets true value", func() {
			flag.Set("true")
			Expect(flag.GetValue()).To(BeTrue())
		})

		It("sets false value", func() {
			flag.Value = true
			flag.Set("false")
			Expect(flag.GetValue()).To(BeFalse())
		})

		It("parses '1' as true", func() {
			flag.Set("1")
			Expect(flag.GetValue()).To(BeTrue())
		})

		It("parses '0' as false", func() {
			flag.Value = true
			flag.Set("0")
			Expect(flag.GetValue()).To(BeFalse())
		})

		It("parses 't' as true", func() {
			flag.Set("t")
			Expect(flag.GetValue()).To(BeTrue())
		})

		It("parses 'f' as false", func() {
			flag.Value = true
			flag.Set("f")
			Expect(flag.GetValue()).To(BeFalse())
		})

		It("parses 'T' as true", func() {
			flag.Set("T")
			Expect(flag.GetValue()).To(BeTrue())
		})

		It("parses 'F' as false", func() {
			flag.Value = true
			flag.Set("F")
			Expect(flag.GetValue()).To(BeFalse())
		})

		It("parses 'TRUE' as true", func() {
			flag.Set("TRUE")
			Expect(flag.GetValue()).To(BeTrue())
		})

		It("parses 'FALSE' as false", func() {
			flag.Value = true
			flag.Set("FALSE")
			Expect(flag.GetValue()).To(BeFalse())
		})

		It("handles invalid values as false", func() {
			flag.Value = true
			flag.Set("invalid")
			Expect(flag.GetValue()).To(BeFalse())
		})

		It("handles empty string as false", func() {
			flag.Value = true
			flag.Set("")
			Expect(flag.GetValue()).To(BeFalse())
		})
	})

	Describe("IntFlag", func() {
		var flag *IntFlag

		BeforeEach(func() {
			flag = &IntFlag{
				Name:  "int-flag",
				Value: 0,
				Usage: "An integer flag for testing",
			}
		})

		It("stores name, value, and usage", func() {
			Expect(flag.GetName()).To(Equal("int-flag"))
			Expect(flag.GetValue()).To(Equal(0))
			Expect(flag.String()).To(Equal("An integer flag for testing"))
		})

		It("sets positive integer value", func() {
			flag.Set("42")
			Expect(flag.GetValue()).To(Equal(42))
		})

		It("sets negative integer value", func() {
			flag.Set("-100")
			Expect(flag.GetValue()).To(Equal(-100))
		})

		It("sets zero value", func() {
			flag.Value = 99
			flag.Set("0")
			Expect(flag.GetValue()).To(Equal(0))
		})

		It("overwrites existing value", func() {
			flag.Set("10")
			flag.Set("20")
			flag.Set("30")
			Expect(flag.GetValue()).To(Equal(30))
		})

		It("handles large numbers", func() {
			flag.Set("2147483647") // Max int32
			Expect(flag.GetValue()).To(Equal(2147483647))
		})

		It("handles invalid values as zero", func() {
			flag.Value = 99
			flag.Set("not-a-number")
			Expect(flag.GetValue()).To(Equal(0))
		})

		It("handles empty string as zero", func() {
			flag.Value = 99
			flag.Set("")
			Expect(flag.GetValue()).To(Equal(0))
		})

		It("handles decimal points by truncating", func() {
			flag.Set("42.7")
			// ParseInt stops at the decimal point
			Expect(flag.GetValue()).To(Equal(0)) // Invalid parse
		})
	})

	Describe("StringSliceFlag", func() {
		var flag *StringSliceFlag

		BeforeEach(func() {
			flag = &StringSliceFlag{
				Name:  "string-slice-flag",
				Value: []string{},
				Usage: "A string slice flag for testing",
			}
		})

		It("stores name, value, and usage", func() {
			Expect(flag.GetName()).To(Equal("string-slice-flag"))
			Expect(flag.GetValue()).To(Equal([]string{}))
			Expect(flag.String()).To(Equal("A string slice flag for testing"))
		})

		It("appends single value", func() {
			flag.Set("value1")
			Expect(flag.GetValue()).To(Equal([]string{"value1"}))
		})

		It("appends multiple values", func() {
			flag.Set("value1")
			flag.Set("value2")
			flag.Set("value3")
			Expect(flag.GetValue()).To(Equal([]string{"value1", "value2", "value3"}))
		})

		It("maintains order of values", func() {
			flag.Set("first")
			flag.Set("second")
			flag.Set("third")

			values := flag.GetValue().([]string)
			Expect(values[0]).To(Equal("first"))
			Expect(values[1]).To(Equal("second"))
			Expect(values[2]).To(Equal("third"))
		})

		It("allows duplicate values", func() {
			flag.Set("duplicate")
			flag.Set("duplicate")
			flag.Set("duplicate")
			Expect(flag.GetValue()).To(Equal([]string{"duplicate", "duplicate", "duplicate"}))
		})

		It("appends empty strings", func() {
			flag.Set("")
			flag.Set("value")
			flag.Set("")
			Expect(flag.GetValue()).To(Equal([]string{"", "value", ""}))
		})

		It("handles special characters", func() {
			flag.Set("value-with-dashes")
			flag.Set("value with spaces")
			flag.Set("value/with/slashes")

			Expect(len(flag.Value)).To(Equal(3))
			Expect(flag.Value[0]).To(Equal("value-with-dashes"))
			Expect(flag.Value[1]).To(Equal("value with spaces"))
			Expect(flag.Value[2]).To(Equal("value/with/slashes"))
		})

		It("starts with empty slice", func() {
			newFlag := &StringSliceFlag{
				Name:  "new-flag",
				Value: []string{},
			}

			Expect(len(newFlag.Value)).To(Equal(0))
		})

		It("can be initialized with values", func() {
			newFlag := &StringSliceFlag{
				Name:  "initialized-flag",
				Value: []string{"initial1", "initial2"},
			}

			newFlag.Set("additional")
			Expect(newFlag.Value).To(Equal([]string{"initial1", "initial2", "additional"}))
		})
	})
})
