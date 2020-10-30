package assets

import (
	"archive/zip"
	"bytes"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

var zipHeader = []byte{0x50, 0x4b, 0x03, 0x04}

var initialized bool

var local = true

var zipFile *zip.Reader
var files map[string]*zip.File

func Init(_local bool) {
	initialized = true
	local = _local

	if !local {
		file, err := os.Open("assets.dpak")
		if err != nil {
			log.Println("Failed to open assets package")
			panic(err)
		}

		data, err1 := ioutil.ReadAll(file)
		if err1 != nil || data[0] != 'r' || data[1] != 'o' || data[2] != '2' || data[3] != 'd' {
			panic("Assets package is corrupted")
		}

		replaced := make([]byte, len(data))
		copy(replaced, zipHeader)
		copy(replaced[4:], data[4:])

		for i := 4; i < len(replaced); i++ {
			replaced[i] ^= byte(i + i%20)
		}

		zipFile, err = zip.NewReader(bytes.NewReader(replaced), int64(len(replaced)))
		if err != nil {
			panic(err)
		}

		files = make(map[string]*zip.File)

		for _, f := range zipFile.File {
			files[f.Name] = f
		}
	}
}

func getFile(path string) (io.ReadCloser, int64, error) {
	if !initialized {
		panic("Asset Manager is not initialized!")
	}

	if local {
		fS, err := os.Open(path)
		if err != nil {
			return nil, 0, err
		}

		stat, err := fS.Stat()
		if err != nil {
			return nil, 0, err
		}

		return fS, stat.Size(), err
	}

	path = strings.ReplaceAll(path, "\\", "/")

	if f, exists := files[path]; exists {
		fS, err := f.Open()
		return fS, int64(f.UncompressedSize64), err
	}

	return nil, 0, os.ErrNotExist
}

func Open(path string) (io.ReadCloser, error) {
	file, _, err := getFile(path)
	return file, err
}

func GetBytes(path string) ([]byte, error) {
	file, _, err := getFile(path)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(file)
}

func GetString(path string) (string, error) {
	data, err := GetBytes(path)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func GetPixmap(path string) (*texture.Pixmap, error) {
	data, size, err := getFile(path)
	if err != nil {
		return nil, err
	}

	return texture.NewPixmapReader(data, size)
}
