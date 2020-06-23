package servicebrokerstub

import (
	"sync"

	. "github.com/onsi/ginkgo"
)

func concurrently(callback func(stub *ServiceBrokerStub), stubs []*ServiceBrokerStub) {
	var wg sync.WaitGroup
	wg.Add(len(stubs))
	for _, s := range stubs {
		go func(stub *ServiceBrokerStub) {
			defer wg.Done()
			defer GinkgoRecover()
			callback(stub)
		}(s)
	}
	wg.Wait()
}
