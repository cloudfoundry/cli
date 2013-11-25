package api

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

const MAX_INT64 int64 = 1<<63 - 1

type Item struct {
	message                  *logmessage.Message
	timestampWhenOutputtable int64
	index                    int
}

type SortedMessageQueue struct {
	items           []*Item
	printTimeBuffer time.Duration
}

func (pq *SortedMessageQueue) PushMessage(message *logmessage.Message) {
	item := &Item{message: message, timestampWhenOutputtable: time.Now().Add(pq.printTimeBuffer).UnixNano()}
	pq.items = append(pq.items, item)
	pq.insertionSort()
}

func (pq *SortedMessageQueue) PopMessage() *logmessage.Message {
	if len(pq.items) == 0 {
		return nil
	}

	var item *Item
	item = pq.items[0]
	pq.items = pq.items[1:len(pq.items)]

	return item.message
}

func (pq *SortedMessageQueue) NextTimestamp() int64 {
	currentQueue := pq.items
	n := len(currentQueue)
	if n == 0 {
		return MAX_INT64
	}
	item := currentQueue[0]
	return item.timestampWhenOutputtable
}

func (pq SortedMessageQueue) less(i, j int) bool {
	return *pq.items[i].message.GetLogMessage().Timestamp < *pq.items[j].message.GetLogMessage().Timestamp
}

func (pq SortedMessageQueue) swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq SortedMessageQueue) insertionSort() {
	for i := 0 + 1; i < len(pq.items); i++ {
		for j := i; j > 0 && pq.less(j, j-1); j-- {
			pq.swap(j, j-1)
		}
	}
}
