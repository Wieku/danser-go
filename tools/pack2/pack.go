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
	dstFile, _ := filepath.Abs(os.Args[1])

	file, err := os.Create(dstFile)
	if err != nil {
		panic(err)
	}

	srcPath, _ := filepath.Abs(os.Args[2])
	srcPath += string(os.PathSeparator)

	writer := zip.NewWriter(file)

	filepath.Walk(srcPath, func(osFilePath string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			trunc := strings.ReplaceAll(strings.TrimPrefix(osFilePath, srcPath), "\\", "/")
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
	file.Close()

	log.Println("Finished")
}
