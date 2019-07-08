package manifestparser_test

import (
	. "code.cloudfoundry.org/cli/util/manifestparser"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Application", func() {
	Describe("Unmarshal", func() {
		var (
			rawYAML     []byte
			application Application
			executeErr  error
		)

		JustBeforeEach(func() {
			executeErr = yaml.Unmarshal(rawYAML, &application)
		})

		Context("when a name is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
name: spark
`)
			})

			It("unmarshals the name", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Name).To(Equal("spark"))
			})
		})

		Context("when a path is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
path: /my/path
`)
			})

			It("unmarshals the path", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Path).To(Equal("/my/path"))
			})
		})

		Context("when a docker map is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
docker:
  image: some-image
  username: some-username
`)
			})

			It("unmarshals the docker properties", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.Docker.Image).To(Equal("some-image"))
				Expect(application.Docker.Username).To(Equal("some-username"))
			})
		})

		Context("when no-route is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
no-route: true
`)
			})

			It("unmarshals the no-route property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.NoRoute).To(BeTrue())
			})
		})

		Context("when random-route is provided", func() {
			BeforeEach(func() {
				rawYAML = []byte(`---
random-route: true
`)
			})

			It("unmarshals the random-route property", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(application.RandomRoute).To(BeTrue())
			})
		})
	})
})
