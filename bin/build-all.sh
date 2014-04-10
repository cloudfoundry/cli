#!/bin/bash

set -e
set -x

OUTDIR=$(dirname $0)/../out

GOARCH=amd64 GOOS=windows $(dirname $0)/build && cp $OUTDIR/cf $OUTDIR/cf-windows-amd64.exe
GOARCH=386 GOOS=windows $(dirname $0)/build && cp $OUTDIR/cf $OUTDIR/cf-windows-386.exe
GOARCH=amd64 GOOS=linux $(dirname $0)/build  && cp $OUTDIR/cf $OUTDIR/cf-linux-amd64
GOARCH=386 GOOS=linux $(dirname $0)/build  && cp $OUTDIR/cf $OUTDIR/cf-linux-386
GOARCH=amd64 GOOS=darwin $(dirname $0)/build  && cp $OUTDIR/cf $OUTDIR/cf-darwin-amd64
