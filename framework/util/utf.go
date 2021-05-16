package util

import (
	"bufio"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"io"
)


func NewScanner(r io.Reader) *bufio.Scanner {
	reader := transform.NewReader(r, unicode.BOMOverride(unicode.UTF8.NewDecoder()))
	return bufio.NewScanner(reader)
}

func NewScannerBuf(r io.Reader, bufSize int) *bufio.Scanner {
	scanner := NewScanner(r)

	buf := make([]byte, 0, bufSize)
	scanner.Buffer(buf, cap(buf))

	return scanner
}
