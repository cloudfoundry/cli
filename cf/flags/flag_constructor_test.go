package flags_test

import (
	"code.cloudfoundry.org/cli/cf/flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Flag Constructors", func() {

	var (
		fc flags.FlagContext
	)

	BeforeEach(func() {
		fc = flags.New()
	})

	Describe("NewStringFlag()", func() {
		It("init the flag context with a new string flagset", func() {
			fc.Parse("-s", "test")
			Expect(fc.IsSet("s")).To(BeFalse())
			Expect(fc.String("s")).To(Equal(""))

			fc.NewStringFlag("s", "s2", "setting new string flag")
			fc.Parse("-s", "test2")
			Expect(fc.IsSet("s")).To(BeTrue())
			Expect(fc.IsSet("s2")).To(BeTrue())
			Expect(fc.String("s")).To(Equal("test2"))
			Expect(fc.String("s2")).To(Equal("test2"))
		})
	})

	Describe("NewStringFlagWithDefault()", func() {
		It("init the flag context with a new string flagset with default value", func() {
			fc.Parse("-s", "test")
			Expect(fc.IsSet("s")).To(BeFalse())
			Expect(fc.String("s")).To(Equal(""))

			fc.NewStringFlagWithDefault("s", "s2", "setting new string flag", "barz")
			fc.Parse()
			Expect(fc.IsSet("s")).To(BeTrue())
			Expect(fc.IsSet("s2")).To(BeTrue())
			Expect(fc.String("s")).To(Equal("barz"))
			Expect(fc.String("s2")).To(Equal("barz"))
		})
	})

	Describe("NewBoolFlag()", func() {
		It("init the flag context with a new bool flagset", func() {
			fc.Parse("--force")
			Expect(fc.IsSet("force")).To(BeFalse())

			fc.NewBoolFlag("force", "f", "force process")
			fc.Parse("--force")
			Expect(fc.IsSet("force")).To(BeTrue())
			Expect(fc.IsSet("f")).To(BeTrue())
			Expect(fc.Bool("force")).To(BeTrue())
			Expect(fc.Bool("f")).To(BeTrue())
		})
	})

	Describe("NewIntFlag()", func() {
		It("init the flag context with a new int flagset", func() {
			fc.Parse("-i", "5")
			Expect(fc.IsSet("i")).To(BeFalse())
			Expect(fc.Int("i")).To(Equal(0))

			fc.NewIntFlag("i", "i2", "setting new int flag")
			fc.Parse("-i", "5")
			Expect(fc.IsSet("i")).To(BeTrue())
			Expect(fc.IsSet("i2")).To(BeTrue())
			Expect(fc.Int("i")).To(Equal(5))
			Expect(fc.Int("i2")).To(Equal(5))
		})
	})

	Describe("NewIntFlagWithDefault()", func() {
		It("init the flag context with a new int flagset with default value", func() {
			fc.Parse("-i", "5")
			Expect(fc.IsSet("i")).To(BeFalse())
			Expect(fc.Int("i")).To(Equal(0))

			fc.NewIntFlagWithDefault("i", "i2", "setting new int flag", 10)
			fc.Parse()
			Expect(fc.IsSet("i")).To(BeTrue())
			Expect(fc.IsSet("i2")).To(BeTrue())
			Expect(fc.Int("i")).To(Equal(10))
			Expect(fc.Int("i2")).To(Equal(10))
		})
	})

	Describe("NewFloat64Flag()", func() {
		It("init the flag context with a new float64 flagset", func() {
			fc.Parse("-f", "5.5")
			Expect(fc.IsSet("f")).To(BeFalse())
			Expect(fc.Float64("f")).To(Equal(float64(0)))

			fc.NewFloat64Flag("f", "f2", "setting new flag")
			fc.Parse("-f", "5.5")
			Expect(fc.IsSet("f")).To(BeTrue())
			Expect(fc.IsSet("f2")).To(BeTrue())
			Expect(fc.Float64("f")).To(Equal(5.5))
			Expect(fc.Float64("f2")).To(Equal(5.5))
		})
	})

	Describe("NewFloat64FlagWithDefault()", func() {
		It("init the flag context with a new Float64 flagset with default value", func() {
			fc.Parse()
			Expect(fc.IsSet("i")).To(BeFalse())
			Expect(fc.Float64("i")).To(Equal(float64(0)))

			fc.NewFloat64FlagWithDefault("i", "i2", "setting new flag", 5.5)
			fc.Parse()
			Expect(fc.IsSet("i")).To(BeTrue())
			Expect(fc.IsSet("i2")).To(BeTrue())
			Expect(fc.Float64("i")).To(Equal(5.5))
			Expect(fc.Float64("i2")).To(Equal(5.5))
		})
	})

	Describe("NewStringSliceFlag()", func() {
		It("init the flag context with a new StringSlice flagset", func() {
			fc.Parse("-s", "5", "-s", "6")
			Expect(fc.IsSet("s")).To(BeFalse())
			Expect(fc.StringSlice("s")).To(Equal([]string{}))

			fc.NewStringSliceFlag("s", "s2", "setting new StringSlice flag")
			fc.Parse("-s", "5", "-s", "6")
			Expect(fc.IsSet("s")).To(BeTrue())
			Expect(fc.IsSet("s2")).To(BeTrue())
			Expect(fc.StringSlice("s")).To(Equal([]string{"5", "6"}))
			Expect(fc.StringSlice("s2")).To(Equal([]string{"5", "6"}))
		})
	})

	Describe("NewStringSliceFlagWithDefault()", func() {
		It("init the flag context with a new StringSlice flagset with default value", func() {
			fc.Parse()
			Expect(fc.IsSet("s")).To(BeFalse())
			Expect(fc.StringSlice("s")).To(Equal([]string{}))

			fc.NewStringSliceFlagWithDefault("s", "s2", "setting new StringSlice flag", []string{"5", "6", "7"})
			fc.Parse()
			Expect(fc.IsSet("s")).To(BeTrue())
			Expect(fc.IsSet("s2")).To(BeTrue())
			Expect(fc.StringSlice("s")).To(Equal([]string{"5", "6", "7"}))
			Expect(fc.StringSlice("s2")).To(Equal([]string{"5", "6", "7"}))
		})
	})

})
