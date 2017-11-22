#!/bin/bash

dir_resolve()
{
    cd "$1" 2>/dev/null || return $?  # cd to desired directory; if fail, quell any error messages but return exit status
    echo "`pwd -P`" # output full, link-resolved path
}

set -e

TARGET=`dirname $0`
TARGET=`dir_resolve $TARGET`
cd $TARGET

go get github.com/gogo/protobuf/{proto,protoc-gen-gogo,gogoproto}
go get -u github.com/mailru/easyjson/...

rm -rf generate-go-tmp
mkdir -p generate-go-tmp/events

mkdir -p events
for i in $(ls definitions/events/*.proto); do
    cp go/go_preamble.proto generate-go-tmp/events/`basename $i`
    cat $i >> generate-go-tmp/events/`basename $i`
done

pushd generate-go-tmp/events > /dev/null
protoc --plugin=$(which protoc-gen-gogo) --gogo_out=$TARGET/events --proto_path=$GOPATH/src:$GOPATH/src/github.com/gogo/protobuf/protobuf:. *.proto
popd > /dev/null

rm -r generate-go-tmp

# generate easyjson marshalers/unmarshalers
# use -no_std_marshalers so that using the optimized easyjson marshalers is opt-in
easyjson -all -no_std_marshalers -pkg events
