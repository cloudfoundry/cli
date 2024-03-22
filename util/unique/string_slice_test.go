package unique_test

import (
	"code.cloudfoundry.org/cli/util/unique"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StringSlice", func() {
	It("returns a slice without duplicates", func() {
		input := []string{
			"flopsy",
			"mopsy",
			"cottontail",
			"mopsy",
			"peter",
			"peter",
		}

		Expect(unique.StringSlice(input)).To(Equal([]string{
			"flopsy",
			"mopsy",
			"cottontail",
			"peter",
		}))
	})
})
