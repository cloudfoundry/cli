package actionerror

// SSHHostkeyFingerprintNotSetError is returned when staging an application fails.
type SSHHostKeyFingerprintNotSetError struct {
}

func (e SSHHostKeyFingerprintNotSetError) Error() string {
	return "SSH hostkey fingerprint not set"
}
