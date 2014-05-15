package i18n_test

import (
	"github.com/cloudfoundry/cli/cf/i18n"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Initialize i18n handler", func() {
	Context("#Init", func() {
		// Confirm T works at all
		FIt("Returns a usable function", func() {
			t, err := i18n.Init("main", "test_fixtures")
			Ω(err).ShouldNot(HaveOccurred())
			Ω(t).ShouldNot(BeNil())
		})

	})
	//confirm T converts if the key is in the json
	//confirm t returns the value if the key doesnt exist in the json
	// Confirm that it reads LC_ALL
	// Confirm that it reads LANG if LC_ALL not set
	// Confirm that it defaults to English if LC_ALL and LANG not set
})
