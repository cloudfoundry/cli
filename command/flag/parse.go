package flag

import "strconv"

func ParseStringToInt(str string) (int, error) {
	integer64Bit, err := strconv.ParseInt(str, 0, 0)
	if err != nil {
		return 0, err
	}
	return int(integer64Bit), nil
}
