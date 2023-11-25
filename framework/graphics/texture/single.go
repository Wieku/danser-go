package texture

import (
	"github.com/wieku/danser-go/framework/goroutines"
	"image"
	"runtime"
)

type TextureSingle struct {
	store     *textureStore
	defRegion TextureRegion
}

func NewTextureSingle(width, height, mipmaps int) *TextureSingle {
	return NewTextureSingleFormat(width, height, RGBA, mipmaps)
}

func NewTextureSingleFormat(width, height int, format Format, mipmaps int) *TextureSingle {
	texture := new(TextureSingle)
	texture.store = newStore(1, width, height, format, mipmaps)
	texture.defRegion = TextureRegion{texture, 0, 1, 0, 1, float32(width), float32(height), 0}

	runtime.SetFinalizer(texture, (*TextureSingle).Dispose)

	return texture
}

func LoadTextureSingle(img *image.RGBA, mipmaps int) *TextureSingle {
	texture := NewTextureSingle(img.Bounds().Dx(), img.Bounds().Dy(), mipmaps)
	texture.SetData(0, 0, img.Bounds().Dx(), img.Bounds().Dy(), img.Pix)
	return texture
}

func (texture *TextureSingle) SetData(x, y, width, height int, data []uint8) {
	texture.store.SetData(x, y, width, height, 0, data, true)
}

func (texture *TextureSingle) GetID() uint32 {
	return texture.store.id
}

func (texture *TextureSingle) GetWidth() int32 {
	return texture.store.width
}

func (texture *TextureSingle) GetHeight() int32 {
	return texture.store.height
}

func (texture *TextureSingle) GetRegion() TextureRegion {
	return texture.defRegion
}

func (texture *TextureSingle) GetLayers() int32 {
	return 1
}

func (texture *TextureSingle) SetFiltering(min, mag Filter) {
	texture.store.SetFiltering(min, mag)
}

func (texture *TextureSingle) Bind(loc uint) {
	texture.store.Bind(loc)
}

func (texture *TextureSingle) GetLocation() uint {
	return texture.store.binding
}

func (texture *TextureSingle) Dispose() {
	goroutines.CallNonBlockMain(func() {
		texture.store.Dispose()
	})
}
