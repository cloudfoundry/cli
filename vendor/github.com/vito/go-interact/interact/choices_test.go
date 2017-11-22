package interact_test

import (
	"io"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"

	"github.com/vito/go-interact/interact"
)

var _ = Describe("Resolving from a set of choices", func() {
	BeforeEach(func() {
		choices = []interact.Choice{
			{Display: "Uno", Value: arbitrary{"uno"}},
			{Display: "Dos", Value: arbitrary{"dos"}},
			{Display: "Tres", Value: arbitrary{"tres"}},
		}
	})

	Context("when the destination is zero-valued", func() {
		BeforeEach(func() {
			destination = arbDst(arbitrary{})
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when '0' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "0\n",

				ExpectedAnswer: arbitrary{},
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: 0\ninvalid selection (must be 1-3)\nsome prompt: ",
			}),

			Entry("when '0' is entered, followed by '1'", Example{
				Prompt: "some prompt",

				Input: "0\n1\n",

				ExpectedAnswer: arbitrary{"uno"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: 0\ninvalid selection (must be 1-3)\nsome prompt: 1\n",
			}),

			Entry("when '1' is entered", Example{
				Prompt: "some prompt",

				Input: "1\n",

				ExpectedAnswer: arbitrary{"uno"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: 1\n",
			}),

			Entry("when '2' is entered", Example{
				Prompt: "some prompt",

				Input: "2\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: 2\n",
			}),

			Entry("when '3' is entered", Example{
				Prompt: "some prompt",

				Input: "3\n",

				ExpectedAnswer: arbitrary{"tres"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: 3\n",
			}),

			Entry("when '4' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "4\n",

				ExpectedAnswer: arbitrary{},
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: 4\ninvalid selection (must be 1-3)\nsome prompt: ",
			}),

			Entry("when '4' is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "4\n2\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: 4\ninvalid selection (must be 1-3)\nsome prompt: 2\n",
			}),

			Entry("when a blank line is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: arbitrary{},
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: \nsome prompt: ",
			}),

			Entry("when a blank line is entered, followed by '3'", Example{
				Prompt: "some prompt",

				Input: "\n3\n",

				ExpectedAnswer: arbitrary{"tres"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: \nsome prompt: 3\n",
			}),

			Entry("when a non-selection is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: arbitrary{},
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: foo\ninvalid selection (not a number)\nsome prompt: ",
			}),

			Entry("when a non-selection is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "foo\n2\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: foo\ninvalid selection (not a number)\nsome prompt: 2\n",
			}),

			Entry("when a non-integer is entered, followed by a blank line, followed by '3'", Example{
				Prompt: "some prompt",

				Input: "foo\n\n3\n",

				ExpectedAnswer: arbitrary{"tres"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt: foo\ninvalid selection (not a number)\nsome prompt: \nsome prompt: 3\n",
			}),
		)

		Context("when an unassignable choice is configured", func() {
			BeforeEach(func() {
				choices = append(choices, interact.Choice{
					Display: "Bogus",
					Value:   "bogus",
				})
			})

			DescribeTable("Resolve", (Example).Run,
				Entry("when the unassignable choice is chosen", Example{
					Prompt: "some prompt",

					Input: "4\n",

					ExpectedAnswer: arbitrary{},
					ExpectedErr: interact.NotAssignableError{
						Destination: reflect.TypeOf(arbitrary{}),
						Value:       reflect.TypeOf("bogus"),
					},
					ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: Bogus\nsome prompt: 4\n",
				}),
			)
		})
	})

	Context("when the destination is one of the choices", func() {
		BeforeEach(func() {
			destination = arbDst(arbitrary{"dos"})
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when '0' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "0\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): 0\ninvalid selection (must be 1-3)\nsome prompt (2): ",
			}),

			Entry("when '0' is entered, followed by '1'", Example{
				Prompt: "some prompt",

				Input: "0\n1\n",

				ExpectedAnswer: arbitrary{"uno"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): 0\ninvalid selection (must be 1-3)\nsome prompt (2): 1\n",
			}),

			Entry("when '1' is entered", Example{
				Prompt: "some prompt",

				Input: "1\n",

				ExpectedAnswer: arbitrary{"uno"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): 1\n",
			}),

			Entry("when '2' is entered", Example{
				Prompt: "some prompt",

				Input: "2\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): 2\n",
			}),

			Entry("when '3' is entered", Example{
				Prompt: "some prompt",

				Input: "3\n",

				ExpectedAnswer: arbitrary{"tres"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): 3\n",
			}),

			Entry("when '4' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "4\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): 4\ninvalid selection (must be 1-3)\nsome prompt (2): ",
			}),

			Entry("when '4' is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "4\n2\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): 4\ninvalid selection (must be 1-3)\nsome prompt (2): 2\n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): \n",
			}),

			Entry("when a non-selection is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): foo\ninvalid selection (not a number)\nsome prompt (2): ",
			}),

			Entry("when a non-selection is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "foo\n2\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): foo\ninvalid selection (not a number)\nsome prompt (2): 2\n",
			}),

			Entry("when a non-integer is entered, followed by a blank line", Example{
				Prompt: "some prompt",

				Input: "foo\n\n",

				ExpectedAnswer: arbitrary{"dos"},
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\nsome prompt (2): foo\ninvalid selection (not a number)\nsome prompt (2): \n",
			}),
		)
	})

	Context("when the destination is nil and one of the choices is nil", func() {
		BeforeEach(func() {
			var emptyDst *arbitrary
			destination = &emptyDst

			choices = []interact.Choice{
				{Display: "Uno", Value: &arbitrary{"uno"}},
				{Display: "Dos", Value: &arbitrary{"dos"}},
				{Display: "Tres", Value: &arbitrary{"tres"}},
				{Display: "none", Value: nil},
			}
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when '0' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "0\n",

				ExpectedAnswer: noArbAns(),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 0\ninvalid selection (must be 1-4)\nsome prompt (4): ",
			}),

			Entry("when '0' is entered, followed by '1'", Example{
				Prompt: "some prompt",

				Input: "0\n1\n",

				ExpectedAnswer: arbAns(arbitrary{"uno"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 0\ninvalid selection (must be 1-4)\nsome prompt (4): 1\n",
			}),

			Entry("when '1' is entered", Example{
				Prompt: "some prompt",

				Input: "1\n",

				ExpectedAnswer: arbAns(arbitrary{"uno"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 1\n",
			}),

			Entry("when '2' is entered", Example{
				Prompt: "some prompt",

				Input: "2\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 2\n",
			}),

			Entry("when '3' is entered", Example{
				Prompt: "some prompt",

				Input: "3\n",

				ExpectedAnswer: arbAns(arbitrary{"tres"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 3\n",
			}),

			Entry("when '4' is entered", Example{
				Prompt: "some prompt",

				Input: "4\n",

				ExpectedAnswer: noArbAns(),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 4\n",
			}),

			Entry("when '5' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "5\n",

				ExpectedAnswer: noArbAns(),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 5\ninvalid selection (must be 1-4)\nsome prompt (4): ",
			}),

			Entry("when '5' is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "5\n2\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 5\ninvalid selection (must be 1-4)\nsome prompt (4): 2\n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: noArbAns(),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): \n",
			}),

			Entry("when a non-selection is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: noArbAns(),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): foo\ninvalid selection (not a number)\nsome prompt (4): ",
			}),

			Entry("when a non-selection is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "foo\n2\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): foo\ninvalid selection (not a number)\nsome prompt (4): 2\n",
			}),

			Entry("when a non-selection is entered, followed by a blank line", Example{
				Prompt: "some prompt",

				Input: "foo\n\n",

				ExpectedAnswer: noArbAns(),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): foo\ninvalid selection (not a number)\nsome prompt (4): \n",
			}),
		)
	})

	Context("when the destination is nil and one of the choices is a typed nil", func() {
		BeforeEach(func() {
			var emptyDst *arbitrary
			destination = &emptyDst

			var nilDst *arbitrary

			choices = []interact.Choice{
				{Display: "Uno", Value: &arbitrary{"uno"}},
				{Display: "Dos", Value: &arbitrary{"dos"}},
				{Display: "Tres", Value: &arbitrary{"tres"}},
				{Display: "none", Value: nilDst},
			}
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when '0' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "0\n",

				ExpectedAnswer: noArbAns(),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 0\ninvalid selection (must be 1-4)\nsome prompt (4): ",
			}),

			Entry("when '0' is entered, followed by '1'", Example{
				Prompt: "some prompt",

				Input: "0\n1\n",

				ExpectedAnswer: arbAns(arbitrary{"uno"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 0\ninvalid selection (must be 1-4)\nsome prompt (4): 1\n",
			}),

			Entry("when '1' is entered", Example{
				Prompt: "some prompt",

				Input: "1\n",

				ExpectedAnswer: arbAns(arbitrary{"uno"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 1\n",
			}),

			Entry("when '2' is entered", Example{
				Prompt: "some prompt",

				Input: "2\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 2\n",
			}),

			Entry("when '3' is entered", Example{
				Prompt: "some prompt",

				Input: "3\n",

				ExpectedAnswer: arbAns(arbitrary{"tres"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 3\n",
			}),

			Entry("when '4' is entered", Example{
				Prompt: "some prompt",

				Input: "4\n",

				ExpectedAnswer: noArbAns(),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 4\n",
			}),

			Entry("when '5' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "5\n",

				ExpectedAnswer: noArbAns(),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 5\ninvalid selection (must be 1-4)\nsome prompt (4): ",
			}),

			Entry("when '5' is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "5\n2\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): 5\ninvalid selection (must be 1-4)\nsome prompt (4): 2\n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: noArbAns(),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): \n",
			}),

			Entry("when a non-selection is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: noArbAns(),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): foo\ninvalid selection (not a number)\nsome prompt (4): ",
			}),

			Entry("when a non-selection is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "foo\n2\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): foo\ninvalid selection (not a number)\nsome prompt (4): 2\n",
			}),

			Entry("when a non-selection is entered, followed by a blank line", Example{
				Prompt: "some prompt",

				Input: "foo\n\n",

				ExpectedAnswer: noArbAns(),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (4): foo\ninvalid selection (not a number)\nsome prompt (4): \n",
			}),
		)
	})

	Context("when the destination is by reference and one of the choices is nil", func() {
		BeforeEach(func() {
			dosDst := &arbitrary{"dos"}
			destination = &dosDst

			choices = []interact.Choice{
				{Display: "Uno", Value: &arbitrary{"uno"}},
				{Display: "Dos", Value: &arbitrary{"dos"}},
				{Display: "Tres", Value: &arbitrary{"tres"}},
				{Display: "none", Value: nil},
			}
		})

		DescribeTable("Resolve", (Example).Run,
			Entry("when '0' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "0\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): 0\ninvalid selection (must be 1-4)\nsome prompt (2): ",
			}),

			Entry("when '0' is entered, followed by '1'", Example{
				Prompt: "some prompt",

				Input: "0\n1\n",

				ExpectedAnswer: arbAns(arbitrary{"uno"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): 0\ninvalid selection (must be 1-4)\nsome prompt (2): 1\n",
			}),

			Entry("when '1' is entered", Example{
				Prompt: "some prompt",

				Input: "1\n",

				ExpectedAnswer: arbAns(arbitrary{"uno"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): 1\n",
			}),

			Entry("when '2' is entered", Example{
				Prompt: "some prompt",

				Input: "2\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): 2\n",
			}),

			Entry("when '3' is entered", Example{
				Prompt: "some prompt",

				Input: "3\n",

				ExpectedAnswer: arbAns(arbitrary{"tres"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): 3\n",
			}),

			Entry("when '4' is entered", Example{
				Prompt: "some prompt",

				Input: "4\n",

				ExpectedAnswer: noArbAns(),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): 4\n",
			}),

			Entry("when '5' is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "5\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): 5\ninvalid selection (must be 1-4)\nsome prompt (2): ",
			}),

			Entry("when '5' is entered, followed by '3'", Example{
				Prompt: "some prompt",

				Input: "5\n3\n",

				ExpectedAnswer: arbAns(arbitrary{"tres"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): 5\ninvalid selection (must be 1-4)\nsome prompt (2): 3\n",
			}),

			Entry("when a blank line is entered", Example{
				Prompt: "some prompt",

				Input: "\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): \n",
			}),

			Entry("when a non-selection is entered, followed by EOF", Example{
				Prompt: "some prompt",

				Input: "foo\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedErr:    io.EOF,
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): foo\ninvalid selection (not a number)\nsome prompt (2): ",
			}),

			Entry("when a non-selection is entered, followed by '2'", Example{
				Prompt: "some prompt",

				Input: "foo\n2\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): foo\ninvalid selection (not a number)\nsome prompt (2): 2\n",
			}),

			Entry("when a non-selection is entered, followed by a blank line", Example{
				Prompt: "some prompt",

				Input: "foo\n\n",

				ExpectedAnswer: arbAns(arbitrary{"dos"}),
				ExpectedOutput: "1: Uno\n2: Dos\n3: Tres\n4: none\nsome prompt (2): foo\ninvalid selection (not a number)\nsome prompt (2): \n",
			}),
		)
	})
})

type arbitrary struct {
	value string
}

func arbDst(dst arbitrary) *arbitrary {
	return &dst
}

func arbAns(dst arbitrary) *arbitrary {
	return &dst
}

func noArbAns() *arbitrary {
	return nil
}
