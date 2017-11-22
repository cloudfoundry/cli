package interact_test

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/vito/go-interact/interact"
)

var _ = Describe("Resolving into ints", func() {
	Context("when the destination is 0", func() {
		BeforeEach(func() {
			destination = intDst(0)
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when an integer is entered", Example{
				Prompt: "some prompt",

				Input: "42\n",

				ExpectedAnswer: 42,
				ExpectedOutput: "some prompt (0): 42\n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: 0,
				ExpectedOutput: "some prompt (0): \n",
			}),

			Entry("when a non-integer is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: 0,
				ExpectedErr:    io.EOF,
				ExpectedOutput: "some prompt (0): foo\ninvalid input (not a number)\nsome prompt (0): ",
			}),

			Entry("when a non-integer is entered, followed by an integer", Example{
				Prompt: "some prompt",

				Input: "foo\n42\n",

				ExpectedAnswer: 42,
				ExpectedOutput: "some prompt (0): foo\ninvalid input (not a number)\nsome prompt (0): 42\n",
			}),

			Entry("when a non-integer is entered, followed by a blank line", Example{
				Prompt: "some prompt",

				Input: "foo\n\n",

				ExpectedAnswer: 0,
				ExpectedOutput: "some prompt (0): foo\ninvalid input (not a number)\nsome prompt (0): \n",
			}),
		)

		Context("when required", func() {
			BeforeEach(func() {
				destination = interact.Required(destination)
			})

			DescribeTable("Resolving into a required int", (Example).Run,
				Entry("when an integer is entered", Example{
					Prompt: "some prompt",

					Input: "42\n",

					ExpectedAnswer: 42,
					ExpectedOutput: "some prompt: 42\n",
				}),

				Entry("when a blank line is entered, followed by EOF", Example{
					Prompt: "some prompt",

					Input: "\n",

					ExpectedAnswer: 0,
					ExpectedErr:    io.EOF,
					ExpectedOutput: "some prompt: \nsome prompt: ",
				}),

				Entry("when a non-integer is entered, followed by EOF", Example{
					Prompt: "some prompt",

					Input: "foo\n",

					ExpectedAnswer: 0,
					ExpectedErr:    io.EOF,
					ExpectedOutput: "some prompt: foo\ninvalid input (not a number)\nsome prompt: ",
				}),

				Entry("when a non-integer is entered, followed by an integer", Example{
					Prompt: "some prompt",

					Input: "foo\n42\n",

					ExpectedAnswer: 42,
					ExpectedOutput: "some prompt: foo\ninvalid input (not a number)\nsome prompt: 42\n",
				}),

				Entry("when a non-integer is entered, followed by a blank line, followed by EOF", Example{
					Prompt: "some prompt",

					Input: "foo\n\n",

					ExpectedAnswer: 0,
					ExpectedErr:    io.EOF,
					ExpectedOutput: "some prompt: foo\ninvalid input (not a number)\nsome prompt: \nsome prompt: ",
				}),

				Entry("when a non-integer is entered, followed by a blank line, followed by an integer", Example{
					Prompt: "some prompt",

					Input: "foo\n\n42\n",

					ExpectedAnswer: 42,
					ExpectedOutput: "some prompt: foo\ninvalid input (not a number)\nsome prompt: \nsome prompt: 42\n",
				}),
			)
		})
	})

	Context("when the destination is nonzero", func() {
		BeforeEach(func() {
			destination = intDst(21)
		})

		DescribeTable("Resolving into a nonzero int", (Example).Run,
			Entry("when an integer is entered", Example{
				Prompt: "some prompt",

				Input: "42\n",

				ExpectedAnswer: 42,
				ExpectedOutput: "some prompt (21): 42\n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: 21,
				ExpectedOutput: "some prompt (21): \n",
			}),

			Entry("when a non-integer is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: 21,
				ExpectedErr:    io.EOF,
				ExpectedOutput: "some prompt (21): foo\ninvalid input (not a number)\nsome prompt (21): ",
			}),

			Entry("when a non-integer is entered, followed by an integer", Example{
				Prompt: "some prompt",

				Input: "foo\n42\n",

				ExpectedAnswer: 42,
				ExpectedOutput: "some prompt (21): foo\ninvalid input (not a number)\nsome prompt (21): 42\n",
			}),

			Entry("when a non-integer is entered, followed by a blank line", Example{
				Prompt: "some prompt",

				Input: "foo\n\n",

				ExpectedAnswer: 21,
				ExpectedOutput: "some prompt (21): foo\ninvalid input (not a number)\nsome prompt (21): \n",
			}),
		)
	})
})

func intDst(dst int) *int {
	return &dst
}
