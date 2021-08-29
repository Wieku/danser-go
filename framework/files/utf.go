package files

import (
	"bufio"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
)

func NewUnicodeReader(r io.Reader) io.Reader {
	return transform.NewReader(r, unicode.BOMOverride(unicode.UTF8.NewDecoder()))
}

func NewScanner(r io.Reader) *bufio.Scanner {
	return bufio.NewScanner(NewUnicodeReader(r))
}

func NewScannerBuf(r io.Reader, bufSize int) *bufio.Scanner {
	scanner := NewScanner(r)

	buf := make([]byte, 0, bufSize)
	scanner.Buffer(buf, cap(buf))

	return scanner
}
