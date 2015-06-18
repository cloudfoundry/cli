#Container Metrics Sample

##Overview
We can use Dropsonde to send container metrics to metron which will emit them to Doppler; which is then polled by traffic controller. The endpoint on the traffic controller should report the latest container metric for each instance of the specified app.

`main.go` connects to the traffic controller and polls the container metrics endpoint.

`consumer_metrics_sample/emitter/main.go` is a sample app that emits container metrics to metron using the dropsonde library.

##To see containter metrics:
1. Run

        cf api api.the-env
        cf login admin
        cf push some-app
        cf app some-app --guid

1. Copy out the guid value from the last bash line
1. Paste the guid into the container_metrics_emitter.go and main.go as the appId value at the top of the files
1. In the main.go update the DopplerAddress value by replacing '10.244.0.34.xip.io' with the value for the environment you are testing.
1. Start the listener

        export CF_ACCESS_TOKEN=`cf oauth-token | tail -n 1`
        go run consumer/main.go

1. Set the appId in `emitter/main.go` to the correct appId.

1. Now build the container_metrics_emitter.go in a new bash window with:

        GOPATH=~/go GOOS=linux go build emitter/main.go

1. Move the container_metrics_emitter executable onto a machine with metron running inside your cf deployment

        ssh-add keyfile
        scp container_metrics_emitter vcap@bosh.the-env:container_metrics_emitter
        ssh -A vcap@bosh.the-env

1. You can get a vm ip by doing bosh vms and selecting an ip

        scp container_metrics_emitter vcap@some.vm.ip:container_metrics_emitter
        ssh vcap@some.vm.ip
        ./container_metrics_emitter

1. You should now see metrics appearing in the listener window

###Things to look for:
1. the diskBytes value should be increasing.
1. the diskBytes value should be skipping some numbers, this is b/c we are listening on a three second window, but publishing on a one second window and are only returning the most recent result
1. there should be only entries for the applicationId you entered with multiple instance indexes
