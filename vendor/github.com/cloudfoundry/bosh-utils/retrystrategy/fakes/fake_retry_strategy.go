package fakes

type FakeRetryStrategy struct {
	TryCalled bool
	TryErr    error
}

func NewFakeRetryStrategy() *FakeRetryStrategy {
	return &FakeRetryStrategy{}
}

func (s *FakeRetryStrategy) Try() error {
	s.TryCalled = true
	return s.TryErr
}
