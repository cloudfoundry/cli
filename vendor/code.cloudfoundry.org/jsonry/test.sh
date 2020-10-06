#!/usr/bin/env sh

set -e

go run honnef.co/go/tools/cmd/staticcheck ./...
go run github.com/onsi/ginkgo/ginkgo -r
