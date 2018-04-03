package task

type Reporter interface {
	TaskStarted(int)
	TaskFinished(int, string)
	TaskOutputChunk(int, []byte)
}

type Task interface {
	ID() int
	State() string
}
