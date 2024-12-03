#!/bin/bash

export GOPROXY=https://goproxy.io,direct

VERSION=1.0.0
go build -trimpath -ldflags "-s -w -X buildenv/menu/cli.Version=${VERSION}"
upx --best buildenv