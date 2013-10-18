package api

import (
	"container/heap"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"time"
)

const MAX_INT64 = 1<<63 - 1

type Item struct {
	message                  *logmessage.Message
	timestampWhenOutputtable int64
	index                    int
}

func NewPriorityMessageQueue(printTimeBuffer time.Duration) *PriorityQueue {
	pq := &PriorityQueue{printTimeBuffer: printTimeBuffer}
	heap.Init(pq)

	return pq
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
	heap.Push(pq, item)
}

func (pq *PriorityQueue) PopMessage() *logmessage.Message {
	if len(pq.items) == 0 {
		return nil
	}
	item := heap.Pop(pq).(*Item)
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

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*Item)
	item.index = n
	pq.items = append(pq.items, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := pq.items
	n := len(old)
	item := old[n-1]
	item.index = -1
	pq.items = old[0 : n-1]
	return item
}
