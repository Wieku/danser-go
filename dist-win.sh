#!/bin/bash
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export CGO_LDFLAGS="-static-libstdc++ -static-libgcc -Wl,-Bstatic -lstdc++ -lpthread -Wl,-Bdynamic"
export WINDRESFLAGS="-F pe-x86-64"

exec=$1
build=$1
ver=${1//./,}','

if [[ $# > 1 ]]
then
  exec+='-s'$2
  build+='-snapshot'$2
  ver+=$2
else
  ver+='0'
fi

preRC='#include "winuser.h"
          1 VERSIONINFO
          FILEVERSION '$ver'
          FILEFLAGSMASK 0x3fL
          FILEOS 0x40004L
          FILETYPE 0x1L
          FILESUBTYPE 0x0L
          BEGIN
            BLOCK "StringFileInfo"
            BEGIN
              BLOCK "040904b0"
              BEGIN
                VALUE "CompanyName", "Wieku"
                VALUE "FileDescription", "danser'

postRC='"
      VALUE "LegalCopyright", "Wieku 2018-2022"
      VALUE "ProductName", "danser"
      VALUE "ProductVersion", "'$build'"
    END
    END
    BLOCK "VarFileInfo"
    BEGIN
      VALUE "Translation", 0x409, 1200
    END
END
2 ICON assets/textures/favicon.ico
'

resgen='windres -l 0 '$WINDRESFLAGS' -o danser.syso'

resCore=$preRC'-core.dll'$postRC
resDanser=$preRC''$postRC
resLauncher=$preRC' launcher'$postRC

$resgen <<< $resCore

go run tools/assets/assets.go ./

go build -trimpath -ldflags "-s -w -X 'github.com/wieku/danser-go/build.VERSION=$build' -X 'github.com/wieku/danser-go/build.Stream=Release'" -buildmode=c-shared -o danser-core.dll -v -x

$resgen <<< $resDanser

gcc --verbose -O3 -o danser.exe -I. cmain/main_danser.c -L. -ldanser-core danser.syso -municode

$resgen <<< $resLauncher

gcc --verbose -O3 -D LAUNCHER -o danser-launcher.exe -I. cmain/main_danser.c -L. -ldanser-core danser.syso -municode

go run tools/pack/pack.go danser-$exec-win.zip danser-core.dll danser.exe danser-launcher.exe bass.dll bass_fx.dll bassmix.dll libyuv.dll assets.dpak

rm -f danser.exe danser-launcher.exe danser-core.dll danser-core.h assets.dpak danser.syso