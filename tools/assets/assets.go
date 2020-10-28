package main

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var ro2d = []byte{0x72, 0x6f, 0x32, 0x64}

func main() {
	path, _ := filepath.Abs(os.Args[1])
	path += string(filepath.Separator)

	buf := new(bytes.Buffer)

	writer := zip.NewWriter(buf)

	filepath.Walk(filepath.Join(path, "assets"), func(osFilePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			trunc := strings.ReplaceAll(strings.TrimPrefix(osFilePath, path), "\\", "/")
			log.Println("Packing:", osFilePath)
			fileWriter, err := writer.Create(trunc)
			if err != nil {
				panic(err)
			}

			fileReader, err := os.Open(osFilePath)
			if err != nil {
				panic(err)
			}

			defer fileReader.Close()

			_, err = io.Copy(fileWriter, fileReader)
			if err != nil {
				panic(err)
			}
		}

		return nil
	})

	writer.Close()

	data := buf.Bytes()

	log.Println("XORing the archive")

	for i := 4; i < len(data); i++ {
		data[i] ^= byte(i + i%20)
	}

	copy(data, ro2d)

	log.Println("Saving to:", filepath.Join(path, "assets.dpak"))

	err := ioutil.WriteFile(filepath.Join(path, "assets.dpak"), data, 0644)
	if err != nil {
		panic(err)
	}

	log.Println("Finished")
}
