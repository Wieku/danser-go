#!/bin/bash
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1
export CC=gcc
export CXX=g++

exec=$1
build=$1
if [[ $# > 1 ]]
then
  exec+='-s'$2
  build+='-snapshot'$2
fi

go run tools/assets/assets.go ./

go build -trimpath -ldflags "-s -w -X 'github.com/wieku/danser-go/build.VERSION=$build' -X 'github.com/wieku/danser-go/build.Stream=Release'" -buildmode=c-shared -o danser-core.so -v -x

mv danser-core.so libdanser-core.so

gcc -no-pie --verbose -O3 -o danser -I. cmain/main_danser.c -Wl,-rpath,. -L. -ldanser-core

gcc -no-pie --verbose -O3 -D LAUNCHER -o danser-launcher -I. cmain/main_danser.c -Wl,-rpath,. -L. -ldanser-core

go run tools/pack/pack.go danser-$exec-linux.zip libdanser-core.so danser danser-launcher libbass.so libbass_fx.so libbassmix.so libyuv.so assets.dpak

rm -f danser danser-launcher libdanser-core.so danser-core.h assets.dpak