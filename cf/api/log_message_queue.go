package api

import (
	"sort"
	"sync"
	"time"

	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

const MAX_INT64 int64 = 1<<63 - 1

type item struct {
	message                  *logmessage.LogMessage
	timestampWhenOutputtable int64
}

type SortedMessageQueue struct {
	clock           func() time.Time
	printTimeBuffer time.Duration
	items           []*item

	mutex sync.Mutex
}

func NewSortedMessageQueue(printTimeBuffer time.Duration, clock func() time.Time) *SortedMessageQueue {
	return &SortedMessageQueue{
		clock:           clock,
		printTimeBuffer: printTimeBuffer,
	}
}

func (pq *SortedMessageQueue) PushMessage(message *logmessage.LogMessage) {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	item := &item{message: message, timestampWhenOutputtable: pq.clock().Add(pq.printTimeBuffer).UnixNano()}
	pq.items = append(pq.items, item)
	sort.Stable(pq)
}

func (pq *SortedMessageQueue) PopMessage() *logmessage.LogMessage {
	pq.mutex.Lock()
	defer pq.mutex.Unlock()

	if len(pq.items) == 0 {
		return nil
	}

	var item *item
	item = pq.items[0]
	pq.items = pq.items[1:len(pq.items)]

	return item.message
}

func (pq *SortedMessageQueue) NextTimestamp() int64 {
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
func (pq *SortedMessageQueue) Less(i, j int) bool {
	return *pq.items[i].message.Timestamp < *pq.items[j].message.Timestamp
}

func (pq *SortedMessageQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *SortedMessageQueue) Len() int {
	return len(pq.items)
}
