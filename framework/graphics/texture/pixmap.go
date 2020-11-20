package texture

// #cgo LDFLAGS: -lm
// #define STB_IMAGE_IMPLEMENTATION
// #define STBI_FAILURE_USERMSG
// #include "stb_image.h"
// #include <stdint.h>
import "C"
import (
	"errors"
	"image"
	"io"
	"os"
	"unsafe"
)

type Pixmap struct {
	arrPointer unsafe.Pointer
	Data       []uint8

	Width  int
	Height int

	disposed bool
}

func NewPixMap(width, height int) *Pixmap {
	pixmap := new(Pixmap)
	pixmap.Width = width
	pixmap.Height = height

	size := width * height * 4

	pixmap.arrPointer = C.calloc(C.size_t(size), C.size_t(1))
	pixmap.Data = (*[1 << 30]uint8)(pixmap.arrPointer)[:size:size]

	return pixmap
}

func NewPixmapFile(file *os.File) (*Pixmap, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.Size() == 0 {
		return nil, errors.New("empty file")
	}

	return NewPixmapReader(file, stat.Size())
}

func NewPixmapReader(file io.ReadCloser, _size int64) (*Pixmap, error) {
	filePointer := C.stbi__malloc(C.size_t(_size))
	fileData := (*[1 << 30]uint8)(filePointer)[:_size:_size]

	defer C.free(filePointer)

	if _, err := io.ReadFull(file, fileData); err != nil {
		return nil, err
	}

	var x, y C.int
	data := C.stbi_load_from_memory((*C.stbi_uc)(&fileData[0]), C.int(len(fileData)), &x, &y, nil, 4)

	if data == nil {
		return nil, errors.New(C.GoString(C.stbi_failure_reason()))
	}

	pixmap := new(Pixmap)
	pixmap.Width = int(x)
	pixmap.Height = int(y)

	size := x * y * 4

	pixmap.arrPointer = unsafe.Pointer(data)
	pixmap.Data = (*[1 << 30]uint8)(pixmap.arrPointer)[:size:size]

	return pixmap, nil
}

func NewPixmapFileString(path string) (*Pixmap, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return NewPixmapFile(file)
}

func (pixmap *Pixmap) NRGBA() *image.NRGBA {
	return &image.NRGBA{
		Pix:    pixmap.Data,
		Stride: 4 * pixmap.Width,
		Rect:   image.Rect(0, 0, pixmap.Width, pixmap.Height),
	}
}

func (pixmap *Pixmap) RGBA() *image.RGBA {
	return &image.RGBA{
		Pix:    pixmap.Data,
		Stride: 4 * pixmap.Width,
		Rect:   image.Rect(0, 0, pixmap.Width, pixmap.Height),
	}
}

func (pixmap *Pixmap) Dispose() {
	if pixmap.disposed {
		return
	}

	C.stbi_image_free(pixmap.arrPointer)

	pixmap.disposed = true
}
