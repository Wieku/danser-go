package pixconv

/*
#cgo CFLAGS: -I/usr/include -I.
#cgo LDFLAGS: -Wl,-rpath,$ORIGIN -L${SRCDIR} -L${SRCDIR}/../../../ -L/usr/lib -lyuv
#include "libyuv.h"
*/
import "C"
import "fmt"

type PixFmt int

const (
	ARGB PixFmt = iota
	I420
	I422
	I444
	NV12
	NV21
)

func (t PixFmt) String() string {
	switch t {
	case ARGB:
		return "ARGB"
	case I420:
		return "I420"
	case NV12:
		return "NV12"
	case NV21:
		return "NV21"
	case I422:
		return "I422"
	case I444:
		return "I444"
	}

	return "unknown"
}

func GetRequiredBufferSize(format PixFmt, w, h int) int {
	switch format {
	case ARGB:
		return w * h * 4
	case I420, NV12, NV21:
		return w * h * 3 / 2
	case I422:
		return w * h * 2
	case I444:
		return w * h * 3
	}

	panic(fmt.Sprintf("Invalid pixel format: %s (%d)", format.String(), format))
}

func Convert(input []byte, inputFormat PixFmt, output []byte, outputFormat PixFmt, w, h int) { //nolint:gocyclo
	switch inputFormat {
	case ARGB:
		switch outputFormat {
		case I420:
			ConvertARGBToI420(input, output, w, h)
		case I422:
			ConvertARGBToI422(input, output, w, h)
		case I444:
			ConvertARGBToI444(input, output, w, h)
		case NV12:
			ConvertARGBToNV12(input, output, w, h)
		case NV21:
			ConvertARGBToNV21(input, output, w, h)
		default:
			panic(fmt.Sprintf("Invalid output format: %s (%d)", outputFormat.String(), outputFormat))
		}
	case I420, I422, I444, NV12, NV21:
		if outputFormat != ARGB {
			panic(fmt.Sprintf("Invalid output format: %s (%d)", outputFormat.String(), outputFormat))
		}

		switch inputFormat {
		case I420:
			ConvertI420ToARGB(input, output, w, h)
		case I422:
			ConvertI422ToARGB(input, output, w, h)
		case I444:
			ConvertI444ToARGB(input, output, w, h)
		case NV12:
			ConvertNV12ToARGB(input, output, w, h)
		case NV21:
			ConvertNV21ToARGB(input, output, w, h)
		}
	default:
		panic(fmt.Sprintf("Invalid input format: %s (%d)", outputFormat.String(), outputFormat))
	}
}

func ConvertARGBToI420(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*4, w*h*3/2)

	C.ARGBToI420((*C.uint8_t)(&input[0]), C.int(w*4), (*C.uint8_t)(&output[0]), C.int(w), (*C.uint8_t)(&output[w*h]), C.int(w/2), (*C.uint8_t)(&output[w*h*5/4]), C.int(w/2), C.int(w), C.int(h))
}

func ConvertARGBToI422(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*4, w*h*2)

	C.ARGBToI422((*C.uint8_t)(&input[0]), C.int(w*4), (*C.uint8_t)(&output[0]), C.int(w), (*C.uint8_t)(&output[w*h]), C.int(w/2), (*C.uint8_t)(&output[w*h*3/2]), C.int(w/2), C.int(w), C.int(h))
}

func ConvertARGBToI444(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*4, w*h*3)

	C.ARGBToI444((*C.uint8_t)(&input[0]), C.int(w*4), (*C.uint8_t)(&output[0]), C.int(w), (*C.uint8_t)(&output[w*h]), C.int(w), (*C.uint8_t)(&output[w*h*2]), C.int(w), C.int(w), C.int(h))
}

func ConvertARGBToNV12(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*4, w*h*3/2)

	C.ARGBToNV12((*C.uint8_t)(&input[0]), C.int(w*4), (*C.uint8_t)(&output[0]), C.int(w), (*C.uint8_t)(&output[w*h]), C.int(w), C.int(w), C.int(h))
}

func ConvertARGBToNV21(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*4, w*h*3/2)

	C.ARGBToNV21((*C.uint8_t)(&input[0]), C.int(w*4), (*C.uint8_t)(&output[0]), C.int(w), (*C.uint8_t)(&output[w*h]), C.int(w), C.int(w), C.int(h))
}

func ConvertI420ToARGB(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*3/2, w*h*4)

	C.I420ToARGB((*C.uint8_t)(&input[0]), C.int(w), (*C.uint8_t)(&input[w*h]), C.int(w/2), (*C.uint8_t)(&input[w*h*5/4]), C.int(w/2), (*C.uint8_t)(&output[0]), C.int(w*4), C.int(w), C.int(h))
}

func ConvertI422ToARGB(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*2, w*h*4)

	C.I422ToARGB((*C.uint8_t)(&input[0]), C.int(w), (*C.uint8_t)(&input[w*h]), C.int(w/2), (*C.uint8_t)(&input[w*h*3/2]), C.int(w/2), (*C.uint8_t)(&output[0]), C.int(w*4), C.int(w), C.int(h))
}

func ConvertI444ToARGB(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*3, w*h*4)

	C.I444ToARGB((*C.uint8_t)(&input[0]), C.int(w), (*C.uint8_t)(&input[w*h]), C.int(w), (*C.uint8_t)(&input[w*h*2]), C.int(w), (*C.uint8_t)(&output[0]), C.int(w*4), C.int(w), C.int(h))
}

func ConvertNV12ToARGB(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*3/2, w*h*4)

	C.NV12ToARGB((*C.uint8_t)(&input[0]), C.int(w), (*C.uint8_t)(&input[w*h]), C.int(w), (*C.uint8_t)(&output[0]), C.int(w*4), C.int(w), C.int(h))
}

func ConvertNV21ToARGB(input []byte, output []byte, w, h int) {
	checkDimensions(input, output, w*h*3/2, w*h*4)

	C.NV21ToARGB((*C.uint8_t)(&input[0]), C.int(w), (*C.uint8_t)(&input[w*h]), C.int(w), (*C.uint8_t)(&output[0]), C.int(w*4), C.int(w), C.int(h))
}

func checkDimensions(input []byte, output []byte, expectedInput int, expectedOutput int) {
	if len(input) < expectedInput {
		panic(fmt.Sprintf("input buffer is smaller than required, expected: %d, actual: %d", expectedInput, len(input)))
	}

	if len(output) < expectedOutput {
		panic(fmt.Sprintf("output buffer is smaller than required, expected: %d, actual: %d", expectedOutput, len(output)))
	}
}
