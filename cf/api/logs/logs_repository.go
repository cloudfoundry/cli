package logs

import "time"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Loggable
type Loggable interface {
	ToLog(loc *time.Location) string
	ToSimpleLog() string
	GetSourceName() string
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Repository

type Repository interface {
	RecentLogsFor(appGUID string) ([]Loggable, error)
	TailLogsFor(appGUID string, onConnect func(), logChan chan<- Loggable, errChan chan<- error)
	Close()
}

const defaultBufferTime time.Duration = 25 * time.Millisecond

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
