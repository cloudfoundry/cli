package emitter

type ByteEmitter interface {
	Emit([]byte) error
	Close()
}
