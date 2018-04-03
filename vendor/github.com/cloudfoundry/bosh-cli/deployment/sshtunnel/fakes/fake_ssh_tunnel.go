package fakes

type FakeTunnel struct {
	startOutput *startOutput
	Started     bool
}

type startOutput struct {
	ReadyErrChOutput error
	ErrChOutput      error
}

func NewFakeTunnel() *FakeTunnel {
	return &FakeTunnel{}
}

func (s *FakeTunnel) Start(readyErrCh chan<- error, errCh chan<- error) {
	s.Started = true

	if s.startOutput != nil {
		readyErrCh <- s.startOutput.ReadyErrChOutput
		errCh <- s.startOutput.ErrChOutput
	}
}

func (s *FakeTunnel) SetStartBehavior(readyErrChOutput error, errChOutput error) {
	s.startOutput = &startOutput{
		ReadyErrChOutput: readyErrChOutput,
		ErrChOutput:      errChOutput,
	}
}
