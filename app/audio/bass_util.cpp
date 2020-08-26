#include "bass_util.hpp"
#include <string>

wchar_t* convert(char* text){
    std::string src(text);
    std::wstring dest;
	dest.clear();
	wchar_t w = 0;
	int bytes = 0;
	wchar_t err = L'�';
	for (size_t i = 0; i < src.size(); i++){
		unsigned char c = (unsigned char)src[i];
		if (c <= 0x7f){//first byte
			if (bytes){
				dest.push_back(err);
				bytes = 0;
			}
			dest.push_back((wchar_t)c);
		}
		else if (c <= 0xbf){//second/third/etc byte
			if (bytes){
				w = ((w << 6)|(c & 0x3f));
				bytes--;
				if (bytes == 0)
					dest.push_back(w);
			}
			else
				dest.push_back(err);
		}
		else if (c <= 0xdf){//2byte sequence start
			bytes = 1;
			w = c & 0x1f;
		}
		else if (c <= 0xef){//3byte sequence start
			bytes = 2;
			w = c & 0x0f;
		}
		else if (c <= 0xf7){//3byte sequence start
			bytes = 3;
			w = c & 0x07;
		}
		else{
			dest.push_back(err);
			bytes = 0;
		}
	}
	if (bytes)
		dest.push_back(err);

    const wchar_t* srcS = dest.c_str();

    size_t size = dest.size() + 1;

    wchar_t* dd = new wchar_t[size];

    memcpy(dd, srcS, size*sizeof(wchar_t));

	return dd;
}

HSTREAM CreateBassStream(char* file, DWORD flags) {
    #ifdef _WIN32

    return BASS_StreamCreateFile(0, convert(file), 0, 0, flags | BASS_UNICODE);

    #else

    return BASS_StreamCreateFile(0, file, 0, 0, flags);

    #endif
}

HSAMPLE LoadBassSample(char* file, DWORD max, DWORD flags) {
    #ifdef _WIN32

    return BASS_SampleLoad(0, convert(file), 0, 0, max, flags | BASS_UNICODE);

    #else

    return BASS_SampleLoad(0, file, 0, 0, max, flags);

    #endif
}