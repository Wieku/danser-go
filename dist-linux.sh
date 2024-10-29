#!/bin/bash
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1
export CC=gcc
export CXX=g++
export BUILD_DIR=./dist/build-linux
export TARGET_DIR=./dist/artifacts

exec=$1
build=$1
if [[ $# > 1 ]]
then
  exec+='-s'$2
  build+='-snapshot'$2
fi

mkdir -p $BUILD_DIR

go run tools/assets/assets.go ./ $BUILD_DIR/

go build -trimpath -ldflags "-s -w -X 'github.com/wieku/danser-go/build.VERSION=$build' -X 'github.com/wieku/danser-go/build.Stream=Release'" -buildmode=c-shared -o $BUILD_DIR/danser-core.so -v -x -tags "exclude_cimgui_glfw exclude_cimgui_sdli"

mv $BUILD_DIR/danser-core.so $BUILD_DIR/libdanser-core.so
cp {libbass.so,libbass_fx.so,libbassmix.so,libyuv.so} $BUILD_DIR/

gcc -no-pie --verbose -O3 -o $BUILD_DIR/danser-cli -I. cmain/main_danser.c -I$BUILD_DIR/ -Wl,-rpath,'$ORIGIN' -L$BUILD_DIR/ -ldanser-core

gcc -no-pie --verbose -O3 -D LAUNCHER -o $BUILD_DIR/danser -I. cmain/main_danser.c -I$BUILD_DIR/ -Wl,-rpath,'$ORIGIN' -L$BUILD_DIR/ -ldanser-core

rm $BUILD_DIR/danser-core.h

go run tools/ffmpeg/ffmpeg.go $BUILD_DIR/

mkdir -p $TARGET_DIR

go run tools/pack2/pack.go $TARGET_DIR/danser-$exec-linux.zip $BUILD_DIR/

rm -rf $BUILD_DIR