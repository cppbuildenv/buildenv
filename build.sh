#!/bin/bash

export GOPROXY=https://goproxy.io,direct

VERSION=v1.0.0
go build -trimpath -ldflags "-s -w -X buildenv/cmd/cli.Version=${VERSION}"
upx --best buildenv