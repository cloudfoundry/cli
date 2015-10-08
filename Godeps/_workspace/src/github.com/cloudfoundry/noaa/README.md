#NOAA

[![Build Status](https://travis-ci.org/cloudfoundry/noaa.svg?branch=master)](https://travis-ci.org/cloudfoundry/noaa)
[![Coverage Status](https://coveralls.io/repos/cloudfoundry/noaa/badge.png)](https://coveralls.io/r/cloudfoundry/noaa)
[![GoDoc](https://godoc.org/github.com/cloudfoundry/noaa?status.png)](https://godoc.org/github.com/cloudfoundry/noaa)

NOAA is a client library to consume metric and log messages from Doppler.

##WARNING

This library does not work with Go 1.3 through 1.3.3, due to a bug in the standard libraries.

##Usage

See the included sample applications. In order to use the samples, you will have to export the following environment variable:

* `CF_ACCESS_TOKEN` - You can get this value by executing (`$ cf oauth-token`). Example: 

```bash
export CF_ACCESS_TOKEN="bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI3YmM2MzllOC0wZGM0LTQ4YzItYTAzYS0xYjkyYzRhMWFlZTIiLCJzdWIiOiI5YTc5MTVkOS04MDc1LTQ3OTUtOTBmOS02MGM0MTU0YTJlMDkiLCJzY29wZSI6WyJzY2ltLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLmFkbWluIiwicGFzc3dvcmQud3JpdGUiLCJzY2ltLndyaXRlIiwib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlci53cml0ZSIsImNsb3VkX2NvbnRyb2xsZXIucmVhZCJdLCJjbGllbnRfaWQiOiJjZiIsImNpZCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjlhNzkxNWQ5LTgwNzUtNDc5NS05MGY5LTYwYzQxNTRhMmUwOSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImlhdCI6MTQwNDg0NzU3NywiZXhwIjoxNDA0ODQ4MTc3LCJpc3MiOiJodHRwczovL3VhYS4xMC4yNDQuMC4zNC54aXAuaW8vb2F1dGgvdG9rZW4iLCJhdWQiOlsic2NpbSIsIm9wZW5pZCIsImNsb3VkX2NvbnRyb2xsZXIiLCJwYXNzd29yZCJdfQ.mAaOJthCotW763lf9fysygqdES_Mz1KFQ3HneKbwY4VJx-ARuxxiLh8l_8Srx7NJBwGlyEtfYOCBcIdvyeDCiQ0wT78Zw7ZJYFjnJ5-ZkDy5NbMqHbImDFkHRnPzKFjJHip39jyjAZpkFcrZ8_pUD8XxZraqJ4zEf6LFdAHKFBM"
```

* `DOPPLER_ADDR` - It is based on your environment. Example:

```bash
export DOPPLER_ADDR="wss://doppler.10.244.0.34.xip.io:443"
```

###Application logs

The `sample/main.go` application streams logs for a particular app. The following environment variable needs to be set:

* `APP_GUID` - You can get this value from running `$ cf app APP --guid`. Example:

```
export APP_GUID=55fdb274-d6c9-4b8c-9b1f-9b7e7f3a346c
```

Then you can run the sample app like this:

```
go build -o bin/sample sample/main.go
bin/sample
```

###Logs and metrics firehose

The `firehose_sample/main.go` application streams metrics data and logs for all apps. 

You can run the firehose sample app like this:

```
go build -o bin/firehose_sample firehose_sample/main.go
bin/firehose_sample
```

Multiple subscribers may connect to the firehose endpoint, each with a unique subscription_id (configurable in `main.go`). Each subscriber (in practice, a pool of clients with a common subscription_id) receives the entire stream. For each subscription_id, all data will be distributed evenly among that subscriber's client pool.


###Container metrics

The `container_metrics_sample/main.go` application streams container metrics for the specified appId.

You can run the container metrics sample app like this:

```
go build -o bin/containermetrics_sample container_metrics_sample/main.go
bin/containermetrics_sample
```

For more information to setup a test environment in order to pull container metrics look at the README.md in the container_metrics_sample.

##Development

Use `go get -d -v -t ./... && ginkgo --race --randomizeAllSpecs --failOnPending --skipMeasurements --cover` to
run the tests.
