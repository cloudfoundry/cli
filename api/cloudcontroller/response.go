package cloudcontroller

type Response struct {
	Result      interface{}
	RawResponse []byte
	Warnings    []string
}
