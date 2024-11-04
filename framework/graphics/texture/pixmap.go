package texture

// #cgo LDFLAGS: -lm
// #define STB_IMAGE_IMPLEMENTATION
// #define STB_IMAGE_WRITE_IMPLEMENTATION
// #define STBI_FAILURE_USERMSG
// #include "stb_image.h"
// #include "stb_image_write.h"
// #include <stdint.h>
import "C"
import (
	"errors"
	"image"
	"io"
	"io/ioutil"
	"os"
	"unsafe"
)

type Pixmap struct {
	RawPointer unsafe.Pointer
	Data       []uint8

	Width      int
	Height     int
	Components int

	disposed bool
}

func NewPixMap(width, height int) *Pixmap {
	return NewPixMapC(width, height, 4)
}

func NewPixMapW(width, height int) *Pixmap {
	pixmap := NewPixMapC(width, height, 4)

	for i := 0; i < width*height; i++ {
		pixmap.Data[i*4] = 255
		pixmap.Data[i*4+1] = 255
		pixmap.Data[i*4+2] = 255
		pixmap.Data[i*4+3] = 0
	}

	return pixmap
}

func NewPixMapC(width, height, components int) *Pixmap {
	pixmap := new(Pixmap)
	pixmap.Width = width
	pixmap.Height = height
	pixmap.Components = components

	size := width * height * components

	pixmap.RawPointer = C.calloc(C.size_t(size), C.size_t(1))
	pixmap.Data = (*[1 << 30]uint8)(pixmap.RawPointer)[:size:size]

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

	return NewPixmapFromBytes(fileData)
}

func NewPixmapFromBytes(bytes []byte) (*Pixmap, error) {
	if bytes == nil || len(bytes) == 0 {
		return nil, errors.New("empty bytes")
	}

	var x, y C.int
	data := C.stbi_load_from_memory((*C.stbi_uc)(&bytes[0]), C.int(len(bytes)), &x, &y, nil, 4)

	if data == nil {
		return nil, errors.New(C.GoString(C.stbi_failure_reason()))
	}

	pixmap := new(Pixmap)
	pixmap.Width = int(x)
	pixmap.Height = int(y)
	pixmap.Components = 4

	size := x * y * 4

	pixmap.RawPointer = unsafe.Pointer(data)
	pixmap.Data = (*[1 << 30]uint8)(pixmap.RawPointer)[:size:size]

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
	if pixmap.Components < 4 {
		panic("Can't create NRGBA with RGB pixmap")
	}

	return &image.NRGBA{
		Pix:    pixmap.Data,
		Stride: 4 * pixmap.Width,
		Rect:   image.Rect(0, 0, pixmap.Width, pixmap.Height),
	}
}

func (pixmap *Pixmap) RGBA() *image.RGBA {
	if pixmap.Components < 4 {
		panic("Can't create RGBA with RGB pixmap")
	}

	return &image.RGBA{
		Pix:    pixmap.Data,
		Stride: 4 * pixmap.Width,
		Rect:   image.Rect(0, 0, pixmap.Width, pixmap.Height),
	}
}

func (pixmap *Pixmap) WritePng(destination string, flip bool) error {
	if flip {
		C.stbi_flip_vertically_on_write(1)
	} else {
		C.stbi_flip_vertically_on_write(0)
	}

	var length C.int
	memPointer := C.stbi_write_png_to_mem((*C.uchar)(&pixmap.Data[0]), 0, C.int(pixmap.Width), C.int(pixmap.Height), C.int(pixmap.Components), &length)

	actPointer := unsafe.Pointer(memPointer)

	data := (*[1 << 30]byte)(actPointer)[:length:length]

	err := ioutil.WriteFile(destination, data, 0644)

	C.free(actPointer)

	return err
}

func (pixmap *Pixmap) Dispose() {
	if pixmap.disposed {
		return
	}

	C.stbi_image_free(pixmap.RawPointer)

	pixmap.disposed = true
}
