#!/bin/bash
mkdir -p go/src/github.com/Originate/
cd go/src/github.com/Originate/
git clone https://github.com/Originate/go_rps.git
export GOPATH=/workspace/go
set PATH $PATH $GOPATH
set PATH $PATH $GOPATH/bin