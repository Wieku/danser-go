package texture

// #cgo LDFLAGS: -lm -static-libgcc -static-libstdc++ -Wl,-Bstatic -lstdc++ -lpthread -Wl,-Bdynamic
// #define STB_IMAGE_IMPLEMENTATION
// #define STBI_FAILURE_USERMSG
// #include "stb_image.h"
import "C"
import (
	"errors"
	"image"
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

	pixmap.arrPointer = C.malloc(C.ulonglong(size))
	pixmap.Data = (*[1 << 30]uint8)(pixmap.arrPointer)[:size:size]

	return pixmap
}

func NewPixmapFile(file *os.File) (*Pixmap, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	filePointer := C.malloc(C.ulonglong(stat.Size()))
	fileData := (*[1 << 30]uint8)(filePointer)[:stat.Size():stat.Size()]

	_, err = file.Read(fileData)
	if err != nil {
		return nil, err
	}

	var x, y C.int
	data := C.stbi_load_from_memory((*C.stbi_uc)(&fileData[0]), C.int(len(fileData)), &x, &y, nil, 4)

	C.free(filePointer)

	if data == nil {
		msg := C.GoString(C.stbi_failure_reason())
		return nil, errors.New(msg)
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
		Stride: 4,
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
