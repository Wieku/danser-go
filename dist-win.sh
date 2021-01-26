#!/bin/bash
export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
export WINDRESFLAGS="-F pe-x86-64"

exec=$1
build=$1
ver=${1//./,}','

if [[ $2 != "" ]]
then
  exec+='-s'$2
  build+='-snapshot'$2
  ver+=$2
else
  ver+='0'
fi

base='1 VERSIONINFO
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
      VALUE "FileDescription", "3rd party osu! cursordance/replay client"
      VALUE "LegalCopyright", "Wieku 2018-2021"
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

echo $base > danser.rc

windres -l 0 $WINDRESFLAGS -i danser.rc -o danser.syso

go run tools/assets/assets.go ./
go build -ldflags "-s -w -X 'github.com/wieku/danser-go/build.VERSION=$build' -X 'github.com/wieku/danser-go/build.Stream=Release'" -o danser-$exec.exe -v -x
go run tools/pack/pack.go danser-$exec-win.zip danser-$exec.exe bass.dll bass_fx.dll bassenc.dll bassmix.dll assets.dpak libwinpthread-1.dll
rm -f danser-$exec.exe assets.dpak danser.rc danser.syso