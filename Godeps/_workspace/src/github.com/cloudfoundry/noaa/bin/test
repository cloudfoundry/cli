#!/bin/bash

set -e

SCRIPT_DIR=`dirname $0`
cd ${SCRIPT_DIR}/..

echo "Go formatting..."
go fmt ./...

echo "Go vetting..."
go vet ./...

echo "Recursive ginkgo... ${*:+(with parameter(s) }$*${*:+)}"
ginkgo -r --race --randomizeAllSpecs --failOnPending -cover $*
