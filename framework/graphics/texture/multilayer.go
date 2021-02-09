package texture

import (
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"runtime"
)

type TextureMultiLayer struct {
	store         *textureStore
	defRegion     TextureRegion
	manualMipmaps bool
}

func NewTextureMultiLayer(width, height, mipmaps, layers int) *TextureMultiLayer {
	return NewTextureMultiLayerFormat(width, height, RGBA, mipmaps, layers)
}

func NewTextureMultiLayerFormat(width, height int, format Format, mipmaps int, layers int) *TextureMultiLayer {
	texture := new(TextureMultiLayer)

	texture.store = newStore(layers, width, height, format, mipmaps)
	texture.store.Clear()

	texture.defRegion = TextureRegion{texture, 0, 1, 0, 1, float32(width), float32(height), 0}

	runtime.SetFinalizer(texture, (*TextureMultiLayer).Dispose)

	return texture
}

func (texture *TextureMultiLayer) NewLayer() {
	layers := texture.store.layers + 1

	dstStore := newStore(int(layers), int(texture.store.width), int(texture.store.height), texture.store.format, int(texture.store.mipmaps))
	dstStore.SetFiltering(texture.store.min, texture.store.mag)
	dstStore.Bind(texture.store.binding)

	mMaps := int32(1)
	if !texture.manualMipmaps {
		mMaps = texture.store.mipmaps
	}

	for level := int32(0); level < mMaps; level++ {
		div := int32(1 << uint(level))
		gl.CopyImageSubData(texture.store.id, gl.TEXTURE_2D_ARRAY, level, 0, 0, 0, dstStore.id, gl.TEXTURE_2D_ARRAY, level, 0, 0, 0, dstStore.width/div, dstStore.height/div, layers-1)
	}

	texture.store.Dispose()

	texture.store = dstStore
}

func (texture *TextureMultiLayer) SetData(x, y, width, height, layer int, data []uint8) {
	if len(data) != width*height*texture.store.format.Size() {
		panic("Wrong number of pixels given!")
	}

	gl.TextureSubImage3D(texture.store.id, 0, int32(x), int32(y), int32(layer), int32(width), int32(height), 1, texture.store.format.Format(), texture.store.format.Type(), gl.Ptr(data))

	if texture.store.mipmaps > 1 && !texture.manualMipmaps {
		gl.GenerateTextureMipmap(texture.store.id)
	}
}

func (texture *TextureMultiLayer) GenerateMipmaps() {
	if texture.store.mipmaps > 1 {
		gl.GenerateTextureMipmap(texture.store.id)
	}
}

func (texture *TextureMultiLayer) GetID() uint32 {
	return texture.store.id
}

func (texture *TextureMultiLayer) GetWidth() int32 {
	return texture.store.width
}

func (texture *TextureMultiLayer) GetHeight() int32 {
	return texture.store.height
}

func (texture *TextureMultiLayer) GetRegion() TextureRegion {
	return texture.defRegion
}

func (texture *TextureMultiLayer) GetLayers() int32 {
	return texture.store.layers
}

func (texture *TextureMultiLayer) SetFiltering(min, mag Filter) {
	texture.store.SetFiltering(min, mag)
}

func (texture *TextureMultiLayer) SetManualMipmapping(value bool) {
	texture.manualMipmaps = value
}

func (texture *TextureMultiLayer) Bind(loc uint) {
	texture.store.Bind(loc)
}

func (texture *TextureMultiLayer) GetLocation() uint {
	return texture.store.binding
}

func (texture *TextureMultiLayer) Dispose() {
	mainthread.CallNonBlock(func() {
		texture.store.Dispose()
	})
}
