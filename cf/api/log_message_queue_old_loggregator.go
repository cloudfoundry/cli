package api

import (
	"sort"
	"sync"
	"time"

	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

// const MAX_INT64 int64 = 1<<63 - 1

type loggregator_item struct {
	message                  *logmessage.LogMessage
	timestampWhenOutputtable int64
}

type Loggregator_SortedMessageQueue struct {
	clock           func() time.Time
	printTimeBuffer time.Duration
	items           []*loggregator_item

	mutex sync.Mutex
}

func NewLoggregator_SortedMessageQueue(printTimeBuffer time.Duration, clock func() time.Time) *Loggregator_SortedMessageQueue {
	return &Loggregator_SortedMessageQueue{
		clock:           clock,
		printTimeBuffer: printTimeBuffer,
	}
}

func (pq *Loggregator_SortedMessageQueue) PushMessage(message *logmessage.LogMessage) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	item := &loggregator_item{message: message, timestampWhenOutputtable: pq.clock().Add(pq.printTimeBuffer).UnixNano()}
	pq.items = append(pq.items, item)
	sort.Stable(pq)
}

func (pq *Loggregator_SortedMessageQueue) PopMessage() *logmessage.LogMessage {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	var item *loggregator_item
	item = pq.items[0]
	pq.items = pq.items[1:len(pq.items)]

	return item.message
}

func (pq *Loggregator_SortedMessageQueue) NextTimestamp() int64 {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	currentQueue := pq.items
	n := len(currentQueue)
	if n == 0 {
		return MAX_INT64
	}
	item := currentQueue[0]
	return item.timestampWhenOutputtable
}

// implement sort interface so we can sort messages as we receive them in PushMessage
func (pq *Loggregator_SortedMessageQueue) Less(i, j int) bool {
	return *pq.items[i].message.Timestamp < *pq.items[j].message.Timestamp
}

func (pq *Loggregator_SortedMessageQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *Loggregator_SortedMessageQueue) Len() int {
	return len(pq.items)
}
