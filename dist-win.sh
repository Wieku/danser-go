#!/bin/bash
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++

exec=$1
build=$1
if [ $2 != "" ]
then
  exec+='-s'$2
  build+='-snapshot'$2
fi

go run tools/assets/assets.go ./
go build -ldflags "-s -w -X 'github.com/wieku/danser-go/build.VERSION=$build' -X 'github.com/wieku/danser-go/build.Stream=Release'" -o danser-$exec.exe -v -x
go run tools/pack/pack.go danser-$exec-win.zip danser-$exec.exe bass.dll bass_fx.dll assets.dpak libwinpthread-1.dll
rm -f danser-$exec.exe
rm -f assets.dpak