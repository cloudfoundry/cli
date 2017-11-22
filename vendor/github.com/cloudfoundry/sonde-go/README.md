# sonde-go
Go implementation of Cloud Foundry's Dropsonde Protocol

## Libraries sharing this protocol

* [Dropsonde](https://github.com/cloudfoundry/dropsonde) is a Go library for applications that wish to emit messages in this format.
* [NOAA](https://github.com/cloudfoundry/noaa) is a library (also in Go) for applications that wish to consume messages from the Cloud Foundry [metric system](https://github.com/cloudfoundry/loggregator).

## Generating Code

1. Install [protobuf](https://github.com/google/protobuf) Version 3 because of our [dropsonde-protocol dependency](https://github.com/cloudfoundry/dropsonde-protocol#generating-code).

   ```bash
   # Currently, the command below installs v3.X
   brew install protobuf
   ```

2. Generate go code

   ```bash
   ./generate-go.sh
   ```

Code will be generated within this directory using the proto files from [Dropsonde Protocol](https://github.com/cloudfoundry/dropsonde-protocol). For other languages, message documentation, and communication protocols, reference [Dropsonde Protocol](https://github.com/cloudfoundry/dropsonde-protocol).

In addition, generate-go.sh will also generate [easyjson](https://github.com/mailru/easyjson) [optimized marshalers/unmarshalers](events/events_easyjson.go).
