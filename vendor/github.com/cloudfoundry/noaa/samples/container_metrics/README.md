# Container Metrics Sample

## Overview

We can use Dropsonde to send container metrics to metron which will emit them
to Doppler; which is then polled by trafficcontroller. The endpoint on the
trafficcontroller should report the latest container metric for each instance
of the specified app.

`samples/container_metrics/consumer/main.go` connects to the trafficcontroller
and polls the container metrics endpoint.

`samples/container_metrics/emitter/main.go` is a sample app that emits container
metrics to metron using the dropsonde library.

## To see containter metrics:

1. Push an app:

        cf push some-app

1. Export the app's guid as an environment variable:

        export APP_GUID=$(cf app some-app --guid)

1. Take note of the app GUID, you will need it later:

        echo $APP_GUID

1. Set the doppler address as an environment variable:

        export DOPPLER_ADDR="wss://your-doppler.example.com/"

1. Export your access token:

        export CF_ACCESS_TOKEN=$(cf oauth-token | grep bear)

1. Start the consumer:

        go run samples/container_metrics/consumer/main.go

1. Now build the emitter in a new terminal window. Make sure your $GOPATH is
   set and run:

        GOOS=linux go build -o bin/emitter samples/container_metrics/emitter/main.go

1. Move the emitter executable onto a machine with metron running inside your
   cf deployment:

        scp bin/emitter vcap@bosh.example.com:emitter
        ssh -A vcap@bosh.example.com

1. You can get a VM IP by doing `bosh vms` and selecting an IP:

        scp emitter vcap@some.vm.ip:emitter
        ssh vcap@some.vm.ip
        export APP_GUID="YOUR-APP-GUID"
        ./emitter

1. You should now see metrics appearing in the listener window.

### Things to look for:

1. The diskBytes value should be increasing.
1. The diskBytes value should be skipping some numbers, this is b/c we are
   listening on a three second window, but publishing on a one second window
   and are only returning the most recent result.
1. There should be only entries for the applicationId you entered with multiple
   instance indexes.
