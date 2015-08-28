package flags_test

import (
	"github.com/simonleung8/flags"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Showing Flags Usage", func() {

	var (
		fc flags.FlagContext
	)

	BeforeEach(func() {
		fc = flags.New()
		fc.NewIntFlag("intFlag", "i", "Usage for intFlag")
		fc.NewIntFlag("m", "", "Usage for intFlag")
		fc.NewBoolFlag("boolFlag", "b", "Usage for boolFlag")
		fc.NewBoolFlag("f", "", "Usage for f")
	})

	It("prints both the full and short flag name", func() {
		outputs := fc.ShowUsage(0)
		Ω(outputs).To(ContainSubstring("-intFlag, -i"))
		Ω(outputs).To(ContainSubstring("-f"))
		Ω(outputs).To(ContainSubstring("--boolFlag, -b"))
	})

	It("prints full flag name with double dashes (--) if shortName exists", func() {
		outputs := fc.ShowUsage(1)
		Ω(outputs).To(ContainSubstring(" --intFlag"))
		Ω(outputs).To(ContainSubstring(" -m"))
		Ω(outputs).To(ContainSubstring(" -f"))
		Ω(outputs).To(ContainSubstring(" --boolFlag, -b"))
	})

	It("prefixes the flag name with spaces", func() {
		outputs := fc.ShowUsage(5)
		Ω(outputs).To(ContainSubstring("     --intFlag"))
		Ω(outputs).To(ContainSubstring("     -f"))
		Ω(outputs).To(ContainSubstring("     --boolFlag"))
	})

	It("prints the usages with non-bool flags first", func() {
		outputs := fc.ShowUsage(0)
		buffer := gbytes.BufferWithBytes([]byte(outputs))
		Eventually(buffer).Should(gbytes.Say("intFlag"))
		Eventually(buffer).Should(gbytes.Say("Usage for intFlag"))
		Eventually(buffer).Should(gbytes.Say("boolFlag"))
		Eventually(buffer).Should(gbytes.Say("Usage for boolFlag"))
		Ω(outputs).To(ContainSubstring("f"))
		Ω(outputs).To(ContainSubstring("Usage for f"))
	})

	It("prefixes the non-bool flag with '-'", func() {
		outputs := fc.ShowUsage(0)
		Ω(outputs).To(ContainSubstring("-intFlag"))
	})

	It("prefixes single character bool flags with '-'", func() {
		outputs := fc.ShowUsage(0)
		Ω(outputs).To(ContainSubstring("-f"))
	})

	It("prefixes multi-character bool flags with '--'", func() {
		outputs := fc.ShowUsage(0)
		Ω(outputs).To(ContainSubstring("--boolFlag"))
	})

	It("aligns the text by padding string with spaces", func() {
		outputs := fc.ShowUsage(0)
		Ω(outputs).To(ContainSubstring("--intFlag, -i       Usage for intFlag"))
		Ω(outputs).To(ContainSubstring("-f                  Usage for f"))
		Ω(outputs).To(ContainSubstring("--boolFlag, -b      Usage for boolFlag"))
	})
})
