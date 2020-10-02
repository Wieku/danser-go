package utils

// #cgo LDFLAGS: -lm
// #define STB_IMAGE_IMPLEMENTATION
// #define STBI_FAILURE_USERMSG
// #include "stb_image.h"
import "C"

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/wieku/danser-go/framework/graphics/texture"
	_ "golang.org/x/image/bmp"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unsafe"
)

func LoadFile(f *os.File) (*image.RGBA, error) {
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var x, y C.int
	data := C.stbi_load_from_memory((*C.stbi_uc)(&bytes[0]), C.int(len(bytes)), &x, &y, nil, 4)

	if data == nil {
		msg := C.GoString(C.stbi_failure_reason())
		return nil, errors.New(msg)
	}

	defer C.stbi_image_free(unsafe.Pointer(data))

	return &image.RGBA{
		Pix:    C.GoBytes(unsafe.Pointer(data), y*x*4),
		Stride: 4,
		Rect:   image.Rect(0, 0, int(x), int(y)),
	}, nil
}

func LoadImage(path string) (*image.RGBA, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	rgba, err := LoadFile(file)
	if err != nil {
		return nil, err
	}

	return rgba, nil
}

func LoadTexture(path string) (*texture.TextureSingle, error) {
	log.Println("Loading texture:", path)
	img, err := LoadImage(path)
	if err == nil {
		log.Println("Loading texture:", path)
		tex := texture.LoadTextureSingle(img, 4)

		return tex, nil
	}
	log.Println("Failed to read a texture: ", err)
	return nil, err
}

func LoadTextureToAtlas(atlas *texture.TextureAtlas, path string) (*texture.TextureRegion, error) {
	log.Println("Loading texture into atlas:", path)
	img, err := LoadImage(path)
	if err == nil {
		return atlas.AddTexture(path, img.Bounds().Dx(), img.Bounds().Dy(), img.Pix), nil
	}
	log.Println("Failed to read a texture: ", err)
	return nil, err
}

// Unzip will decompress a zip archive, moving all files and folders
// within the zip file (parameter 1) to an output directory (parameter 2).
func Unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
