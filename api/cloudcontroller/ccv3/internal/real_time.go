package internal

import "time"

type RealTime struct{}

func (RealTime) Now() time.Time {
	return time.Now()
}

func (RealTime) Sleep(d time.Duration) {
	time.Sleep(d)
}
