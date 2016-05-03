package gosteno

type Config struct {
	Sinks     []Sink
	Level     LogLevel
	Codec     Codec
	EnableLOC bool
}
