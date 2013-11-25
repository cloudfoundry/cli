package api

import (
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"sort"
	"time"
)

const MAX_INT64 int64 = 1<<63 - 1

type Item struct {
	message                  *logmessage.Message
	timestampWhenOutputtable int64
	index                    int
}

type PriorityQueue struct {
	items           []*Item
	printTimeBuffer time.Duration
}

func (pq PriorityQueue) Len() int { return len(pq.items) }

func (pq PriorityQueue) Less(i, j int) bool {
	return *pq.items[i].message.GetLogMessage().Timestamp < *pq.items[j].message.GetLogMessage().Timestamp
}

func (pq PriorityQueue) Swap(i, j int) {
	pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
	pq.items[i].index = i
	pq.items[j].index = j
}

func (pq *PriorityQueue) PushMessage(message *logmessage.Message) {
	item := &Item{message: message, timestampWhenOutputtable: time.Now().Add(pq.printTimeBuffer).UnixNano()}
	pq.items = append(pq.items, item)
	sort.Sort(pq)
}

func (pq *PriorityQueue) PopMessage() *logmessage.Message {
	if len(pq.items) == 0 {
		return nil
	}

	var item *Item
	item = pq.items[0]
	pq.items = pq.items[1:len(pq.items)]

	return item.message
}

func (pq *PriorityQueue) NextTimestamp() int64 {
	currentQueue := pq.items
	n := len(currentQueue)
	if n == 0 {
		return MAX_INT64
	}
	item := currentQueue[0]
	return item.timestampWhenOutputtable
}

func NewPriorityMessageQueue(printTimeBuffer time.Duration) *PriorityQueue {
	return &PriorityQueue{printTimeBuffer: printTimeBuffer}
}
