package multiopt_test

import (
	"github.com/cloudfoundry-incubator/diego-ssh/cf-plugin/options/multiopt"
	"github.com/pborman/getopt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Multiopt", func() {
	var (
		opts *getopt.Set
		mv   *multiopt.MultiValue
		args []string
	)

	BeforeEach(func() {
		opts = getopt.New()
		mv = &multiopt.MultiValue{}

		opts.Var(mv, 'L', "help string")

		args = []string{"ssh", "-L8080:example.com:80", "-L9443:example.com:443"}
	})

	JustBeforeEach(func() {
		err := opts.Getopt(args, nil)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Values", func() {
		It("aggregates values for an option", func() {
			Expect(mv.Values()).To(ConsistOf("8080:example.com:80", "9443:example.com:443"))
		})
	})

	Describe("String", func() {
		It("returns all arguments separated by a comma", func() {
			Expect(mv.String()).To(Equal("8080:example.com:80,9443:example.com:443"))
		})
	})
})
