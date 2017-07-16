package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Docker", func() {
	var docker DockerImage

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			docker = DockerImage{}
		})

		Context("when the docker image URL is a valid user/repo:tag URL", func() {
			It("set the path and does not return an error", func() {
				err := docker.UnmarshalFlag("user/repo:tag")
				Expect(err).ToNot(HaveOccurred())
				Expect(docker.Path).To(Equal("user/repo:tag"))
			})
		})

		Context("when the docker image URL is an HTTP URL ", func() {
			It("set the path and does not return an error", func() {
				err := docker.UnmarshalFlag("registry.example.com:5000/user/repository/tag")
				Expect(err).ToNot(HaveOccurred())
				Expect(docker.Path).To(Equal("registry.example.com:5000/user/repository/tag"))
			})
		})

		Context("when the docker image URL is invalid", func() {
			It("returns an error", func() {
				err := docker.UnmarshalFlag("AAAAAA")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: "invalid docker reference: repository name must be lowercase",
				}))
				Expect(docker.Path).To(BeEmpty())
			})
		})
	})
})
