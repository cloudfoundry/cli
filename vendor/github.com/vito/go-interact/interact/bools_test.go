package interact_test

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/vito/go-interact/interact"
)

var _ = Describe("Resolving into bools", func() {
	Context("when the destination is false", func() {
		BeforeEach(func() {
			destination = boolDst(false)
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when 'y' is entered", Example{
				Prompt: "some prompt",

				Input: "y\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [yN]: y\n",
			}),

			Entry("when 'yes' is entered", Example{
				Prompt: "some prompt",

				Input: "yes\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [yN]: yes\n",
			}),

			Entry("when 'n' is entered", Example{
				Prompt: "some prompt",

				Input: "n\n",

				ExpectedAnswer: false,
				ExpectedOutput: "some prompt [yN]: n\n",
			}),

			Entry("when 'N' is entered", Example{
				Prompt: "some prompt",

				Input: "N\n",

				ExpectedAnswer: false,
				ExpectedOutput: "some prompt [yN]: N\n",
			}),

			Entry("when 'no' is entered", Example{
				Prompt: "some prompt",

				Input: "no\n",

				ExpectedAnswer: false,
				ExpectedOutput: "some prompt [yN]: no\n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: false,
				ExpectedOutput: "some prompt [yN]: \n",
			}),

			Entry("when a non-boolean is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: false,
				ExpectedErr:    io.EOF,
				ExpectedOutput: "some prompt [yN]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [yN]: ",
			}),

			Entry("when a non-integer is entered, followed by 'y'", Example{
				Prompt: "some prompt",

				Input: "foo\ny\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [yN]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [yN]: y\n",
			}),

			Entry("when a non-integer is entered, followed by a blank line", Example{
				Prompt: "some prompt",

				Input: "foo\n\n",

				ExpectedAnswer: false,
				ExpectedOutput: "some prompt [yN]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [yN]: \n",
			}),
		)

		Context("when required", func() {
			BeforeEach(func() {
				destination = interact.Required(destination)
			})

			DescribeTable("Resolve", (Example).Run,
				Entry("when 'y' is entered", Example{
					Prompt: "some prompt",

					Input: "y\n",

					ExpectedAnswer: true,
					ExpectedOutput: "some prompt [yn]: y\n",
				}),

				Entry("when 'yes' is entered", Example{
					Prompt: "some prompt",

					Input: "yes\n",

					ExpectedAnswer: true,
					ExpectedOutput: "some prompt [yn]: yes\n",
				}),

				Entry("when 'n' is entered", Example{
					Prompt: "some prompt",

					Input: "n\n",

					ExpectedAnswer: false,
					ExpectedOutput: "some prompt [yn]: n\n",
				}),

				Entry("when 'no' is entered", Example{
					Prompt: "some prompt",

					Input: "no\n",

					ExpectedAnswer: false,
					ExpectedOutput: "some prompt [yn]: no\n",
				}),

				Entry("when a blank line is entered, followed by EOF", Example{
					Prompt: "some prompt",

					Input: "\n",

					ExpectedAnswer: false,
					ExpectedErr:    io.EOF,
					ExpectedOutput: "some prompt [yn]: \nsome prompt [yn]: ",
				}),

				Entry("when a non-boolean is entered, followed by EOF", Example{
					Prompt: "some prompt",

					Input: "foo\n",

					ExpectedAnswer: false,
					ExpectedErr:    io.EOF,
					ExpectedOutput: "some prompt [yn]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [yn]: ",
				}),

				Entry("when a non-integer is entered, followed by 'y'", Example{
					Prompt: "some prompt",

					Input: "foo\ny\n",

					ExpectedAnswer: true,
					ExpectedOutput: "some prompt [yn]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [yn]: y\n",
				}),

				Entry("when a non-integer is entered, followed by a blank line, followed by EOF", Example{
					Prompt: "some prompt",

					Input: "foo\n\n",

					ExpectedAnswer: false,
					ExpectedErr:    io.EOF,
					ExpectedOutput: "some prompt [yn]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [yn]: \nsome prompt [yn]: ",
				}),

				Entry("when a non-integer is entered, followed by a blank line, followed by 'y'", Example{
					Prompt: "some prompt",

					Input: "foo\n\ny\n",

					ExpectedAnswer: true,
					ExpectedOutput: "some prompt [yn]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [yn]: \nsome prompt [yn]: y\n",
				}),
			)
		})
	})

	Context("when the destination is true", func() {
		BeforeEach(func() {
			destination = boolDst(true)
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when 'y' is entered", Example{
				Prompt: "some prompt",

				Input: "y\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [Yn]: y\n",
			}),

			Entry("when 'Y' is entered", Example{
				Prompt: "some prompt",

				Input: "Y\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [Yn]: Y\n",
			}),

			Entry("when 'yes' is entered", Example{
				Prompt: "some prompt",

				Input: "yes\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [Yn]: yes\n",
			}),

			Entry("when 'n' is entered", Example{
				Prompt: "some prompt",

				Input: "n\n",

				ExpectedAnswer: false,
				ExpectedOutput: "some prompt [Yn]: n\n",
			}),

			Entry("when 'no' is entered", Example{
				Prompt: "some prompt",

				Input: "no\n",

				ExpectedAnswer: false,
				ExpectedOutput: "some prompt [Yn]: no\n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [Yn]: \n",
			}),

			Entry("when a non-boolean is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: true,
				ExpectedErr:    io.EOF,
				ExpectedOutput: "some prompt [Yn]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [Yn]: ",
			}),

			Entry("when a non-integer is entered, followed by 'y'", Example{
				Prompt: "some prompt",

				Input: "foo\ny\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [Yn]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [Yn]: y\n",
			}),

			Entry("when a non-integer is entered, followed by a blank line", Example{
				Prompt: "some prompt",

				Input: "foo\n\n",

				ExpectedAnswer: true,
				ExpectedOutput: "some prompt [Yn]: foo\ninvalid input (not y, n, yes, or no)\nsome prompt [Yn]: \n",
			}),
		)
	})
})

func boolDst(dst bool) *bool {
	return &dst
}
