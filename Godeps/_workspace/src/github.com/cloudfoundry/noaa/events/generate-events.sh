#!/bin/bash -x

OUT_DIR=`dirname $0`
OUT_DIR=`cd $OUT_DIR && pwd`

PROTO_DIR=$OUT_DIR/../dropsonde-protocol

cd $PROTO_DIR
./generate-go.sh $OUT_DIR