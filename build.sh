#!/bin/bash

function build()
{
	pushd $1
	go install -ldflags '-s -w' || exit 1
	popd
}

build cmd/lgrep
build cmd/lcut
build cmd/lpretty

upx -qq $GOPATH/bin/lgrep $GOPATH/bin/lcut $GOPATH/bin/lpretty || exit 1
