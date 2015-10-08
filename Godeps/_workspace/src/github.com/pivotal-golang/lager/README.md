lager
=====

[![Build Status](https://travis-ci.org/pivotal-golang/lager.svg?branch=master)](https://travis-ci.org/pivotal-golang/lager)

Lager is a logging library for go.

## Usage

Instantiate a logger with the name of your component.

```go
import (
  "github.com/pivotal-golang/lager"
)

logger := logger.New("my-app")
```

### Sinks

Lager can write logs to a variety of destinations. You can specify the destinations
using Lager sinks:

To write to an arbitrary `Writer` object:

```go
logger.RegisterSink(lager.NewWriterSink(myWriter, lager.INFO))
```

### Emitting logs

Lager supports the usual level-based logging, with an optional argument for arbitrary key-value data.

```go
logger.Info("doing-stuff", logger.Data{
  "informative": true,
})
```

output:
```json
{ "source": "my-app", "message": "doing-stuff", "data": { "informative": true }, "timestamp": 1232345, "log_level": 1 }
```

Error messages also take an `Error` object:

```go
logger.Error("failed-to-do-stuff", errors.New("Something went wrong"))
```

output:
```json
{ "source": "my-app", "message": "failed-to-do-stuff", "data": { "error": "Something went wrong" }, "timestamp": 1232345, "log_level": 1 }
```

### Sessions

You can avoid repetition of contextual data using 'Sessions':

```go

contextualLogger := logger.Session("my-task", logger.Data{
  "request-id": 5,
})

contextualLogger.Info("my-action")
```

output:

```json
{ "source": "my-app", "message": "my-task.my-action", "data": { "request-id": 5 }, "timestamp": 1232345, "log_level": 1 }
```

## License

Lager is [Apache 2.0](https://github.com/pivotal-golang/lager/blob/master/LICENSE) licensed.
