package ccerror

// V3Error represents a cloud controller error.
type V3Error struct {
	Code   int64  `json:"code"`
	Detail string `json:"detail"`
	Title  string `json:"title"`
}
