package interact_test

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/vito/go-interact/interact"
)

var _ = Describe("Resolving into passwords", func() {
	Context("when the destination is empty", func() {
		BeforeEach(func() {
			destination = passDst("")
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when a string is entered", Example{
				Prompt: "some prompt",

				Input: "forty two\n",

				ExpectedAnswer: interact.Password("forty two"),
				ExpectedOutput: "some prompt (): \n",
			}),

			Entry("when a blank line is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: interact.Password(""),
				ExpectedOutput: "some prompt (): \n",
			}),
		)

		Context("when required", func() {
			BeforeEach(func() {
				destination = interact.Required(destination)
			})

			DescribeTable("Resolve", (Example).Run,
				Entry("when a string is entered", Example{
					Prompt: "some prompt",

					Input: "forty two\n",

					ExpectedAnswer: interact.Password("forty two"),
					ExpectedOutput: "some prompt: \n",
				}),

				Entry("when a blank line is entered, followed by EOF", Example{
					Prompt: "some prompt",

					Input: "\n",

					ExpectedAnswer: interact.Password(""),
					ExpectedErr:    io.EOF,
					ExpectedOutput: "some prompt: \nsome prompt: ",
				}),

				Entry("when a blank line is entered, followed by a string", Example{
					Prompt: "some prompt",

					Input: "\nforty two\n",

					ExpectedAnswer: interact.Password("forty two"),
					ExpectedOutput: "some prompt: \nsome prompt: \n",
				}),
			)
		})
	})

	Context("when the destination is not empty", func() {
		BeforeEach(func() {
			destination = passDst("some default")
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when a string is entered", Example{
				Prompt: "some prompt",

				Input: "forty two\n",

				ExpectedAnswer: interact.Password("forty two"),
				ExpectedOutput: "some prompt (has default): \n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: interact.Password("some default"),
				ExpectedOutput: "some prompt (has default): \n",
			}),
		)
	})
})

func passDst(dst interact.Password) *interact.Password {
	return &dst
}
