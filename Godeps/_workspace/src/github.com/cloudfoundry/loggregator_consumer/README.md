#loggregator_consumer 

[![Build Status](https://travis-ci.org/cloudfoundry/loggregator_consumer.svg?branch=master)](https://travis-ci.org/cloudfoundry/loggregator_consumer) [![GoDoc](https://godoc.org/github.com/cloudfoundry/loggregator_consumer?status.png)](https://godoc.org/github.com/cloudfoundry/loggregator_consumer) [![Coverage Status](https://coveralls.io/repos/cloudfoundry/loggregator_consumer/badge.png)](https://coveralls.io/r/cloudfoundry/loggregator_consumer)

Loggregator consumer is a library that allows an application developer to set up
a connection to a loggregator server, and begin receiving log messages from it.
It includes the ability to tail logs as well as get the recent logs.

#WARNING
This library does not work with Go 1.3 through 1.3.2, due to a bug in the standard libraries.

Usage
------------------
See the included sample application. In order to use the sample, you will have to export the following environment variables:

* CF_ACCESS_TOKEN - You can get this value from reading the AccessToken looking at your cf configuration file (`$ cat ~/.cf/config.json`). Example: 

  ```
export CF_ACCESS_TOKEN="bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiI3YmM2MzllOC0wZGM0LTQ4YzItYTAzYS0xYjkyYzRhMWFlZTIiLCJzdWIiOiI5YTc5MTVkOS04MDc1LTQ3OTUtOTBmOS02MGM0MTU0YTJlMDkiLCJzY29wZSI6WyJzY2ltLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLmFkbWluIiwicGFzc3dvcmQud3JpdGUiLCJzY2ltLndyaXRlIiwib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlci53cml0ZSIsImNsb3VkX2NvbnRyb2xsZXIucmVhZCJdLCJjbGllbnRfaWQiOiJjZiIsImNpZCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6IjlhNzkxNWQ5LTgwNzUtNDc5NS05MGY5LTYwYzQxNTRhMmUwOSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImlhdCI6MTQwNDg0NzU3NywiZXhwIjoxNDA0ODQ4MTc3LCJpc3MiOiJodHRwczovL3VhYS4xMC4yNDQuMC4zNC54aXAuaW8vb2F1dGgvdG9rZW4iLCJhdWQiOlsic2NpbSIsIm9wZW5pZCIsImNsb3VkX2NvbnRyb2xsZXIiLCJwYXNzd29yZCJdfQ.mAaOJthCotW763lf9fysygqdES_Mz1KFQ3HneKbwY4VJx-ARuxxiLh8l_8Srx7NJBwGlyEtfYOCBcIdvyeDCiQ0wT78Zw7ZJYFjnJ5-ZkDy5NbMqHbImDFkHRnPzKFjJHip39jyjAZpkFcrZ8_pUD8XxZraqJ4zEf6LFdAHKFBM"
```
* APP_GUID - You can get this value from running `$ CF_TRACE=true cf app dora` and then extracting the app guid from the request URL. Example:

```
export APP_GUID=55fdb274-d6c9-4b8c-9b1f-9b7e7f3a346c
```

Then you can run the sample app like this:

```
export GOPATH=`pwd`
export PATH=$PATH:$GOPATH/bin
go get github.com/cloudfoundry/loggregator_consumer/sample_consumer
sample_consumer
```

Development
-----------------

Use `go get -d -v -t ./... && ginkgo --race --randomizeAllSpecs --failOnPending --skipMeasurements --cover` to
run the tests.
