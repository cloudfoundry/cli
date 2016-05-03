package gosteno

type Sink interface {
	AddRecord(record *Record)
	Flush()

	SetCodec(codec Codec)
	GetCodec() Codec
}
