/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package formatters_test

import (
	. "github.com/cloudfoundry/cli/cf/formatters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestByteSize", func() {
		Expect(ByteSize(100 * MEGABYTE)).To(Equal("100M"))
		Expect(ByteSize(uint64(100.5 * MEGABYTE))).To(Equal("100.5M"))
	})

	It("parses byte amounts with short units (e.g. M, G)", func() {
		var (
			megabytes uint64
			err       error
		)

		megabytes, err = ToMegabytes("5M")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("5m")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("2G")
		Expect(megabytes).To(Equal(uint64(2 * 1024)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("3T")
		Expect(megabytes).To(Equal(uint64(3 * 1024 * 1024)))
		Expect(err).NotTo(HaveOccurred())
	})

	It("parses byte amounts with long units (e.g MB, GB)", func() {
		var (
			megabytes uint64
			err       error
		)

		megabytes, err = ToMegabytes("5MB")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("5mb")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("2GB")
		Expect(megabytes).To(Equal(uint64(2 * 1024)))
		Expect(err).NotTo(HaveOccurred())

		megabytes, err = ToMegabytes("3TB")
		Expect(megabytes).To(Equal(uint64(3 * 1024 * 1024)))
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error when the unit is missing", func() {
		_, err := ToMegabytes("5")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unit of measurement"))
	})

	It("returns an error when the unit is unrecognized", func() {
		_, err := ToMegabytes("5MBB")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unit of measurement"))
	})

	It("allows whitespace before and after the value", func() {
		megabytes, err := ToMegabytes("\t\n\r 5MB ")
		Expect(megabytes).To(Equal(uint64(5)))
		Expect(err).NotTo(HaveOccurred())
	})

	It("returns an error for negative values", func() {
		_, err := ToMegabytes("-5MB")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unit of measurement"))
	})

	It("returns an error for zero values", func() {
		_, err := ToMegabytes("0TB")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unit of measurement"))
	})
})
