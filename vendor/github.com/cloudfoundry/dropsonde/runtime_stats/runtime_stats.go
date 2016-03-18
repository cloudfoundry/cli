package runtime_stats

import (
	"log"
	"runtime"
	"time"

	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

type RuntimeStats struct {
	eventEmitter emitter.EventEmitter
	interval     time.Duration
}

func NewRuntimeStats(eventEmitter emitter.EventEmitter, interval time.Duration) *RuntimeStats {
	return &RuntimeStats{
		eventEmitter: eventEmitter,
		interval:     interval,
	}
}

func (rs *RuntimeStats) Run(stopChan <-chan struct{}) {
	ticker := time.NewTicker(rs.interval)
	defer ticker.Stop()
	for {
		rs.emit("numCPUS", float64(runtime.NumCPU()))
		rs.emit("numGoRoutines", float64(runtime.NumGoroutine()))
		rs.emitMemMetrics()

		select {
		case <-ticker.C:
		case <-stopChan:
			return
		}
	}
}

func (rs *RuntimeStats) emitMemMetrics() {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	rs.emit("memoryStats.numBytesAllocatedHeap", float64(stats.HeapAlloc))
	rs.emit("memoryStats.numBytesAllocatedStack", float64(stats.StackInuse))
	rs.emit("memoryStats.numBytesAllocated", float64(stats.Alloc))
	rs.emit("memoryStats.numMallocs", float64(stats.Mallocs))
	rs.emit("memoryStats.numFrees", float64(stats.Frees))
	rs.emit("memoryStats.lastGCPauseTimeNS", float64(stats.PauseNs[(stats.NumGC+255)%256]))
}

func (rs *RuntimeStats) emit(name string, value float64) {
	err := rs.eventEmitter.Emit(&events.ValueMetric{
		Name:  &name,
		Value: &value,
		Unit:  proto.String("count"),
	})
	if err != nil {
		log.Printf("RuntimeStats: failed to emit: %v", err)
	}
}
