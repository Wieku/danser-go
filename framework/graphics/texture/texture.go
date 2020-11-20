package texture

import (
	"github.com/go-gl/gl/v3.3-core/gl"
)

type Filter int32

var Filtering = struct {
	Nearest,
	Linear,
	MipMap,
	MipMapNearestNearest,
	MipMapLinearNearest,
	MipMapNearestLinear,
	MipMapLinearLinear Filter
}{gl.NEAREST, gl.LINEAR, gl.LINEAR_MIPMAP_LINEAR, gl.NEAREST_MIPMAP_NEAREST, gl.LINEAR_MIPMAP_NEAREST, gl.NEAREST_MIPMAP_LINEAR, gl.LINEAR_MIPMAP_LINEAR}

type Texture interface {
	GetID() uint32
	GetWidth() int32
	GetHeight() int32
	GetRegion() TextureRegion
	GetLayers() int32
	SetFiltering(min, mag Filter)
	Bind(loc uint)
	GetLocation() uint
	Dispose()
}

type TextureRegion struct {
	Texture        Texture
	U1, U2, V1, V2 float32
	Width, Height  int32
	Layer          int32
}

type textureStore struct {
	id                             uint32
	binding                        uint
	layers, width, height, mipmaps int32
	format                         Format
	min, mag                       Filter
	disposed                       bool
}

func newStore(layerNum, width, height int, format Format, mipmaps int) *textureStore {
	store := new(textureStore)
	gl.GenTextures(1, &store.id)

	store.layers = int32(layerNum)
	store.width = int32(width)
	store.height = int32(height)
	store.format = format

	if mipmaps < 1 {
		mipmaps = 1
	}
	store.mipmaps = int32(mipmaps)

	store.Bind(0)
	gl.TexStorage3D(gl.TEXTURE_2D_ARRAY, store.mipmaps, format.InternalFormat(), store.width, store.height, store.layers)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_BASE_LEVEL, 0)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAX_LEVEL, store.mipmaps-1)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	if mipmaps > 1 {
		store.SetFiltering(Filtering.MipMap, Filtering.Linear)
	} else {
		store.SetFiltering(Filtering.Linear, Filtering.Linear)
	}

	return store
}

func (store *textureStore) Bind(loc uint) {
	store.binding = loc
	gl.ActiveTexture(gl.TEXTURE0 + uint32(loc))
	gl.BindTexture(gl.TEXTURE_2D_ARRAY, store.id)
}

func (store *textureStore) Clear() {
	gl.ClearTexImage(store.id, 0, store.format.Format(), store.format.Type(), gl.Ptr(nil))
}

func (store *textureStore) SetFiltering(min, mag Filter) {
	store.min = min
	store.mag = mag

	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MIN_FILTER, int32(min))
	gl.TexParameteri(gl.TEXTURE_2D_ARRAY, gl.TEXTURE_MAG_FILTER, int32(mag))
}

func (store *textureStore) Dispose() {
	if !store.disposed {
		gl.DeleteTextures(1, &store.id)
	}

	store.disposed = true
}
