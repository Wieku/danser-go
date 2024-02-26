package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	targetDir, _ := filepath.Abs(os.Args[1])

	archiveName := "ffmpeg-n6.1-latest-linux64-gpl-shared-6.1"
	if runtime.GOOS == "windows" {
		archiveName = "ffmpeg-n6.1-latest-win64-gpl-shared-6.1"
	}

	downloadUrl := "https://github.com/Wieku/FFmpeg-Builds/releases/download/latest/" + archiveName + ".zip"

	file := download(downloadUrl)

	info, err := file.Stat()
	if err != nil {
		panic(err)
	}

	zipFile, err := zip.NewReader(file, info.Size())
	if err != nil {
		panic(err)
	}

	os.MkdirAll(filepath.Join(targetDir, "ffmpeg"), 0755)

	log.Println("Starting unpacking...")

	for _, f := range zipFile.File {
		strName := strings.TrimPrefix(f.Name, archiveName+"/")

		if strings.HasPrefix(strName, "bin/") && !strings.HasSuffix(strName, "bin/") {
			unpack(f, filepath.Join(targetDir, "ffmpeg", strings.TrimPrefix(strName, "bin/")))
		}

		if runtime.GOOS == "linux" {
			if strings.HasPrefix(strName, "lib/lib") {
				matches := len(strings.Split(strName, ".")) == 3 // matching lib***.so.version

				if matches {
					unpack(f, filepath.Join(targetDir, "ffmpeg", strings.TrimPrefix(strName, "lib/")))
				}
			}
		}
	}

	file.Close()

	os.Remove(file.Name())

	log.Println("Finished.")
}

func unpack(f *zip.File, s string) {
	log.Println(fmt.Sprintf("Unpacking \"%s\" to \"%s\"...", f.Name, s))
	out, err := os.OpenFile(s, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		panic(err)
	}

	defer out.Close()

	zEntry, err := f.Open()
	if err != nil {
		panic(err)
	}

	defer zEntry.Close()

	_, err = io.Copy(out, zEntry)
	if err != nil {
		panic(err)
	}
}

func download(url string) *os.File {
	log.Println("Downloading ffmpeg...")

	out, err := os.CreateTemp("", "ffmpeg-dist")
	if err != nil {
		panic(err)
	}

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}

	_, err = out.Seek(0, 0)
	if err != nil {
		panic(err)
	}

	log.Println("Download complete.")

	return out
}
