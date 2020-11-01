#include "bass_util.h"
#include <string.h>

#ifdef _WIN32
#include <stdlib.h>

wchar_t *convert(char *src) {
	int lengthB = strlen(src);

	wchar_t *dest = (wchar_t *) calloc(lengthB + 1, sizeof(wchar_t));
	int index = 0;

	wchar_t err = 0xFFFD;

	wchar_t w = 0;
	int bytes = 0;

	for (size_t i = 0; i < lengthB; i++) {
		unsigned char c = (unsigned char) src[i];
		if (c <= 0x7f) {//first byte
			if (bytes) {
				dest[index] = err;
				++index;

				bytes = 0;
			}

			dest[index] = c;
			++index;
		} else if (c <= 0xbf) {//second/third/etc byte
			if (bytes) {
				w = ((w << 6) | (c & 0x3f));
				bytes--;
				if (bytes == 0) {
					dest[index] = w;
					++index;
				}
			} else {
				dest[index] = err;
				++index;
			}
		} else if (c <= 0xdf) {//2byte sequence start
			bytes = 1;
			w = c & 0x1f;
		} else if (c <= 0xef) {//3byte sequence start
			bytes = 2;
			w = c & 0x0f;
		} else if (c <= 0xf7) {//3byte sequence start
			bytes = 3;
			w = c & 0x07;
		} else {
			dest[index] = err;
			++index;
			bytes = 0;
		}
	}

	if (bytes) {
		dest[index] = err;
	}

	return dest;
}

#endif

HSTREAM CreateBassStream(char *file, DWORD flags) {
#ifdef _WIN32
	wchar_t *wFile = convert(file);

	HSTREAM stream = BASS_StreamCreateFile(0, wFile, 0, 0, flags | BASS_UNICODE);

	free(wFile);

	return stream;
#else
	return BASS_StreamCreateFile(0, file, 0, 0, flags);
#endif
}

HSAMPLE LoadBassSample(char *file, DWORD max, DWORD flags) {
#ifdef _WIN32
	wchar_t *wFile = convert(file);

	HSAMPLE sample = BASS_SampleLoad(0, wFile, 0, 0, max, flags | BASS_UNICODE);

	free(wFile);

	return sample;
#else
	return BASS_SampleLoad(0, file, 0, 0, max, flags);
#endif
}