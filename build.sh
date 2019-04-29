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
build cmd/lsort

upx -qq $GOPATH/bin/lgrep $GOPATH/bin/lcut $GOPATH/bin/lpretty $GOPATH/bin/lsort || exit 1
