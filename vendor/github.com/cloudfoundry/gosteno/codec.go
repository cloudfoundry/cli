package gosteno

type Codec interface {
	EncodeRecord(record *Record) ([]byte, error)
}
