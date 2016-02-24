package api

import (
	"sort"
	"sync"

	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

type Loggregator_SortedMessageQueue struct {
	messages []*logmessage.LogMessage
	mutex    sync.Mutex
}

func NewLoggregator_SortedMessageQueue() *Loggregator_SortedMessageQueue {
	return &Loggregator_SortedMessageQueue{}
}

func (pq *Loggregator_SortedMessageQueue) PushMessage(message *logmessage.LogMessage) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	pq.messages = append(pq.messages, message)
}

// implement sort interface so we can sort messages as we receive them in PushMessage
func (pq *Loggregator_SortedMessageQueue) Less(i, j int) bool {
	return *pq.messages[i].Timestamp < *pq.messages[j].Timestamp
}

func (pq *Loggregator_SortedMessageQueue) Swap(i, j int) {
	pq.messages[i], pq.messages[j] = pq.messages[j], pq.messages[i]
}

func (pq *Loggregator_SortedMessageQueue) Len() int {
	return len(pq.messages)
}

func (pq *Loggregator_SortedMessageQueue) EnumerateAndClear(onMessage func(*logmessage.LogMessage)) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	sort.Stable(pq)

	for _, x := range pq.messages {
		onMessage(x)
	}

	pq.messages = []*logmessage.LogMessage{}
}
