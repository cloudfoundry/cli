package util_test

import (
	"fmt"
	"strings"
	"sync"

	. "code.cloudfoundry.org/cli/v8/util"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SynchronizedBuffer", func() {
	var buffer *SynchronizedBuffer

	BeforeEach(func() {
		buffer = &SynchronizedBuffer{}
	})

	It("keeps every write when written to concurrently", func() {
		var wg sync.WaitGroup
		for w := 0; w < 2; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < 500; i++ {
					fmt.Fprintln(buffer, "line")
				}
			}()
		}
		wg.Wait()

		lines := strings.Split(strings.TrimSuffix(buffer.String(), "\n"), "\n")
		Expect(lines).To(HaveLen(1000))
		Expect(lines).To(HaveEach("line"))
	})

	It("is empty after Reset", func() {
		fmt.Fprint(buffer, "some output")
		Expect(buffer.String()).To(Equal("some output"))

		buffer.Reset()
		Expect(buffer.String()).To(BeEmpty())

		fmt.Fprint(buffer, "new output")
		Expect(buffer.String()).To(Equal("new output"))
	})
})
