package types_test

import (
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/jsonry"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("optional boolean", func() {
	It("has an unset zero value", func() {
		var s types.OptionalBoolean
		Expect(s.IsSet).To(BeFalse())
		Expect(s.Value).To(BeFalse())
	})

	DescribeTable(
		"marshaling and unmarshalling",
		func(o types.OptionalBoolean, j string) {
			By("marshalling", func() {
				container := struct {
					A types.OptionalBoolean `jsonry:"a"`
				}{
					A: o,
				}

				data, err := jsonry.Marshal(container)
				Expect(err).NotTo(HaveOccurred())
				Expect(data).To(MatchJSON(j))
			})

			By("unmarshalling", func() {
				var receiver struct {
					A types.OptionalBoolean `jsonry:"a"`
				}
				err := jsonry.Unmarshal([]byte(j), &receiver)
				Expect(err).NotTo(HaveOccurred())
				Expect(receiver.A).To(Equal(o))
			})
		},
		Entry("true", types.NewOptionalBoolean(true), `{"a": true}`),
		Entry("false", types.NewOptionalBoolean(false), `{"a": false}`),
		Entry("unset", types.OptionalBoolean{}, `{}`),
	)
})
