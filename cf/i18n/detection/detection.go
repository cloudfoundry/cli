package detection

import "github.com/pivotal-cf-experimental/jibber_jabber"

type Detector interface {
	DetectIETF() (string, error)
	DetectLanguage() (string, error)
}

type JibberJabberDetector struct{}

func (detector *JibberJabberDetector) DetectIETF() (string, error) {
	return jibber_jabber.DetectIETF()
}

func (detector *JibberJabberDetector) DetectLanguage() (string, error) {
	return jibber_jabber.DetectLanguage()
}
