#!/bin/bash

export GOPROXY=https://goproxy.io,direct

VERSION=v0.1.0
go build -trimpath -ldflags "-s -w -X buildenv/cmd/cli.Version=${VERSION}"

if [ -x "upx" ]; then
    upx --best buildenv
fi
