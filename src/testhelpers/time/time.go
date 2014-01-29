package time

import (
	"time"
)

func MustParse(format, timeString string) (result time.Time) {
	result, err := time.Parse(format, timeString)
	if err != nil {
		panic(err)
	}
	return
}
