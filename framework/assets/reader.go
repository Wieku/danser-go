package assets

import "io"

type xorReader struct {
	mainReader io.ReaderAt
}

func (r *xorReader) ReadAt(p []byte, off int64) (n int, err error) {
	n, err = r.mainReader.ReadAt(p, off)

	for i := 0; i < n; i++ {
		iOff := int64(i) + off

		if iOff < 4 {
			p[i] = zipHeader[iOff]
		} else {
			p[i] ^= byte(iOff + iOff%20)
		}
	}

	return
}
