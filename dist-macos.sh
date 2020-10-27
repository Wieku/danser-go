#!/bin/bash
export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=1
export CC=gcc
export CXX=g++

exec=$1
build=$1
if [ $2 != "" ]
then
  exec+='-s'$2
  build+='-snapshot'$2
fi

go run tools/assets/assets.go ./
go build -ldflags "-s -w -X 'github.com/wieku/danser-go/build.VERSION=$build' -X 'github.com/wieku/danser-go/build.Stream=Release'" -o danser-$exec -v -x
go run tools/pack/pack.go danser-$exec-osx.zip danser-$exec libbass.dylib libbass_fx.dylib assets.dpak
rm -f danser-$exec
rm -f assets.dpak