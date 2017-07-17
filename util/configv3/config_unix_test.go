// +build !windows

package configv3_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Verbose", func() {
	var homeDir string

	BeforeEach(func() {
		homeDir = setup()
	})

	AfterEach(func() {
		teardown(homeDir)
	})

	DescribeTable("absolute paths",
		func(env string, configTrace string, flag bool, expected bool, location []string) {
			rawConfig := fmt.Sprintf(`{ "Trace":"%s" }`, configTrace)
			setConfig(homeDir, rawConfig)

			defer os.Unsetenv("CF_TRACE")
			if env == "" {
				Expect(os.Unsetenv("CF_TRACE")).ToNot(HaveOccurred())
			} else {
				Expect(os.Setenv("CF_TRACE", env)).ToNot(HaveOccurred())
			}

			config, err := LoadConfig(FlagOverride{
				Verbose: flag,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())

			verbose, parsedLocation := config.Verbose()
			Expect(verbose).To(Equal(expected))
			Expect(parsedLocation).To(Equal(location))
		},

		Entry("CF_TRACE true: enables verbose", "true", "", false, true, nil),
		Entry("CF_TRACE true, config trace false: enables verbose", "true", "false", false, true, nil),
		Entry("CF_TRACE true, config trace file path: enables verbose AND logging to file", "true", "/foo/bar", false, true, []string{"/foo/bar"}),

		Entry("CF_TRACE false: disables verbose", "false", "", false, false, nil),
		Entry("CF_TRACE false, '-v': enables verbose", "false", "", true, true, nil),
		Entry("CF_TRACE false, config trace true: disables verbose", "false", "true", false, false, nil),
		Entry("CF_TRACE false, config trace file path: enables logging to file", "false", "/foo/bar", false, false, []string{"/foo/bar"}),
		Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo/bar", true, true, []string{"/foo/bar"}),

		Entry("CF_TRACE empty: disables verbose", "", "", false, false, nil),
		Entry("CF_TRACE empty:, '-v': enables verbose", "", "", true, true, nil),
		Entry("CF_TRACE empty, config trace true: enables verbose", "", "true", false, true, nil),
		Entry("CF_TRACE empty, config trace file path: enables logging to file", "", "/foo/bar", false, false, []string{"/foo/bar"}),
		Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo/bar", true, true, []string{"/foo/bar"}),

		Entry("CF_TRACE filepath: enables logging to file", "/foo/bar", "", false, false, []string{"/foo/bar"}),
		Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo/bar", "", true, true, []string{"/foo/bar"}),
		Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo/bar", "true", false, true, []string{"/foo/bar"}),
		Entry("CF_TRACE filepath, config trace filepath: enables logging to file for BOTH paths", "/foo/bar", "/baz", false, false, []string{"/foo/bar", "/baz"}),
		Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo/bar", "/baz", true, true, []string{"/foo/bar", "/baz"}),
	)

	Context("relative paths (cannot be tested in DescribeTable)", func() {
		It("resolves relative paths into absolute paths", func() {
			configTrace := "foo/bar"

			rawConfig := fmt.Sprintf(`{ "Trace":"%s" }`, configTrace)
			setConfig(homeDir, rawConfig)

			cfTrace := "foo2/bar2"
			Expect(os.Setenv("CF_TRACE", cfTrace)).ToNot(HaveOccurred())
			defer os.Unsetenv("CF_TRACE")

			config, err := LoadConfig(FlagOverride{})
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())

			cwd, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			_, parsedLocation := config.Verbose()
			Expect(parsedLocation).To(Equal([]string{filepath.Join(cwd, cfTrace), filepath.Join(cwd, configTrace)}))
		})
	})
})
