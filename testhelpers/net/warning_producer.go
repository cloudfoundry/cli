package net

type warningProducer struct {
	warnings []string
}

func NewWarningProducer(warnings []string) (warning_producer warningProducer) {
	warning_producer.warnings = warnings
	return
}

func (warning_producer warningProducer) Warnings() []string {
	return warning_producer.warnings
}
