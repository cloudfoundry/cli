package servicebrokerstub

import "sync"

func concurrently(callback func(stub *ServiceBrokerStub), stubs []*ServiceBrokerStub) {
	var wg sync.WaitGroup
	wg.Add(len(stubs))
	for _, s := range stubs {
		go func(stub *ServiceBrokerStub) {
			callback(stub)
			wg.Done()
		}(s)
	}
	wg.Wait()
}
