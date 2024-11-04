package texture

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/wieku/danser-go/framework/goroutines"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"runtime"
)

type TextureMultiLayer struct {
	store         *textureStore
	defRegion     TextureRegion
	manualMipmaps bool
	clearColor    color2.Color
}

func NewTextureMultiLayer(width, height, mipmaps, layers int) *TextureMultiLayer {
	return NewTextureMultiLayerFormat(width, height, RGBA, mipmaps, layers)
}

func NewTextureMultiLayerCC(width, height, mipmaps, layers int, clearColor color2.Color) *TextureMultiLayer {
	return NewTextureMultiLayerFormatCC(width, height, RGBA, mipmaps, layers, clearColor)
}

func NewTextureMultiLayerFormat(width, height int, format Format, mipmaps int, layers int) *TextureMultiLayer {
	return NewTextureMultiLayerFormatCC(width, height, format, mipmaps, layers, color2.NewLA(0, 0))
}

func NewTextureMultiLayerFormatCC(width, height int, format Format, mipmaps int, layers int, clearColor color2.Color) *TextureMultiLayer {
	texture := new(TextureMultiLayer)
	texture.clearColor = clearColor

	texture.store = newStore(layers, width, height, format, mipmaps)
	texture.store.ClearColor(texture.clearColor)

	texture.defRegion = TextureRegion{texture, 0, 1, 0, 1, float32(width), float32(height), 0}

	runtime.SetFinalizer(texture, (*TextureMultiLayer).Dispose)

	return texture
}

func (texture *TextureMultiLayer) NewLayer() {
	layers := texture.store.layers + 1

	dstStore := newStore(int(layers), int(texture.store.width), int(texture.store.height), texture.store.format, int(texture.store.mipmaps))
	dstStore.ClearColor(texture.clearColor)
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
	texture.store.SetData(x, y, width, height, layer, data, !texture.manualMipmaps)
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
	goroutines.CallNonBlockMain(func() {
		texture.store.Dispose()
	})
}
