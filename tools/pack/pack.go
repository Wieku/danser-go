package main

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	path, _ := filepath.Abs(os.Args[1])

	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}

	writer := zip.NewWriter(file)

	for i := 2; i < len(os.Args); i++ {
		osFilePath, _ := filepath.Abs(os.Args[i])

		trunc := strings.ReplaceAll(strings.TrimPrefix(osFilePath, filepath.Dir(path)+string(filepath.Separator)), "\\", "/")

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

	writer.Close()
	file.Close()

	log.Println("Finished")
}
