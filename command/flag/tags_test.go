package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tags", func() {
	var tags Tags

	BeforeEach(func() {
		tags = Tags{}
	})

	Describe("UnmarshalFlag", func() {
		When("the empty string is provided", func() {
			It("should return empty list", func() {
				Expect(tags.UnmarshalFlag("")).To(Succeed())
				Expect(tags).To(ConsistOf([]string{}))
			})
		})

		When("single tag is provided", func() {
			It("should return the string", func() {
				Expect(tags.UnmarshalFlag("tag")).To(Succeed())
				Expect(tags).To(ConsistOf([]string{"tag"}))
			})
		})

		When("multiple comma separated tags are provided", func() {
			It("should return the string", func() {
				Expect(tags.UnmarshalFlag("tag1,tag2")).To(Succeed())
				Expect(tags).To(ConsistOf([]string{"tag1", "tag2"}))
			})
		})

		When("multiple tags with spaces are provided", func() {
			It("should return the trimmed tags", func() {
				Expect(tags.UnmarshalFlag(" tag1,  tag2")).To(Succeed())
				Expect(tags).To(ConsistOf([]string{"tag1", "tag2"}))
			})
		})

		When("multiple tags with excessive commas are provided", func() {
			It("should return just the tags", func() {
				Expect(tags.UnmarshalFlag(",tag1,tag2,")).To(Succeed())
				Expect(tags).To(ConsistOf([]string{"tag1", "tag2"}))
			})
		})
	})
})
