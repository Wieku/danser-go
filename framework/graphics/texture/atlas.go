package texture

import (
	"github.com/faiface/mainthread"
	"github.com/go-gl/gl/v3.3-core/gl"
	"log"
	"runtime"
)

type rectangle struct {
	x, y, width, height int
}

func (rect rectangle) area() int {
	return rect.width * rect.height
}

type TextureAtlas struct {
	store       *textureStore
	defRegion   TextureRegion
	min, mag    Filter
	padding     int
	subTextures map[string]*TextureRegion
	emptySpaces map[int32][]rectangle
}

func NewTextureAtlas(size, mipmaps int) *TextureAtlas {
	texture := new(TextureAtlas)
	texture.subTextures = make(map[string]*TextureRegion)
	texture.emptySpaces = make(map[int32][]rectangle)
	texture.emptySpaces[0] = []rectangle{{0, 0, size, size}}

	var siz int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &siz)
	if int(siz) < size {
		log.Printf("WARNING: GPU supports only %dx%d textures\n", siz, siz)
		size = int(siz)
	}

	texture.store = newStore(1, size, size, mipmaps)
	texture.defRegion = TextureRegion{texture, 0, 1, 0, 1, int32(size), int32(size), 0}
	texture.padding = 1 << uint(texture.store.mipmaps)

	runtime.SetFinalizer(texture, (*TextureAtlas).Dispose)

	return texture
}

func (texture *TextureAtlas) AddTexture(name string, width, height int, data []uint8) *TextureRegion {
	texture.Bind(texture.store.binding)
	if len(data) != width*height*4 {
		panic("Wrong number of pixels given!")
	}

	if int(texture.GetWidth()) < width || int(texture.GetHeight()) < height {
		log.Panicf("Texture is too big! Atlas size: %dx%d, texture size: %dx%d", texture.GetWidth(), texture.GetHeight(), width, height)
	}

	imBounds := rectangle{0, 0, width + texture.padding, height + texture.padding}
	for layer := int32(0); layer < texture.store.layers; layer++ {
		j := -1
		smallest := rectangle{0, 0, int(texture.store.width), int(texture.store.height)}

		for i := range texture.emptySpaces[layer] {
			space := texture.emptySpaces[layer][i]
			if imBounds.width <= space.width && imBounds.height <= space.height {
				if space.area() <= smallest.area() {
					j = i
					smallest = space
				}
			}
		}

		if j == -1 {
			if layer == texture.store.layers-1 {
				texture.newLayer()
			}
			continue
		} else {
			dw := smallest.width - imBounds.width
			dh := smallest.height - imBounds.height

			var rect1 rectangle
			var rect2 rectangle

			if dh > dw {
				rect1 = rectangle{smallest.x + imBounds.width, smallest.y, smallest.width - imBounds.width, imBounds.height}
				rect2 = rectangle{smallest.x, smallest.y + imBounds.height, smallest.width, smallest.height - imBounds.height}
			} else {
				rect1 = rectangle{smallest.x + imBounds.width, smallest.y, smallest.width - imBounds.width, smallest.height}
				rect2 = rectangle{smallest.x, smallest.y + imBounds.height, imBounds.width, smallest.height - imBounds.height}
			}

			texture.emptySpaces[layer] = append(texture.emptySpaces[layer][:j], texture.emptySpaces[layer][j+1:]...)
			texture.emptySpaces[layer] = append(texture.emptySpaces[layer], rect1, rect2)

			gl.TexSubImage3D(gl.TEXTURE_2D_ARRAY, 0, int32(smallest.x), int32(smallest.y), int32(layer), int32(width), int32(height), 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
			if texture.store.mipmaps > 1 {
				gl.GenerateMipmap(gl.TEXTURE_2D_ARRAY)
			}

			region := TextureRegion{Texture: texture, Width: int32(width), Height: int32(height), Layer: int32(layer)}
			region.U1 = (float32(smallest.x) + 0.5) / float32(texture.store.width)
			region.V1 = (float32(smallest.y) + 0.5) / float32(texture.store.height)
			region.U2 = region.U1 + float32(width-1)/float32(texture.store.width)
			region.V2 = region.V1 + float32(height-1)/float32(texture.store.height)
			texture.subTextures[name] = &region
			return &region
		}
	}

	return nil
}

func (texture *TextureAtlas) GetTexture(name string) *TextureRegion {
	return texture.subTextures[name]
}

func (texture *TextureAtlas) newLayer() {
	texture.emptySpaces[texture.store.layers] = []rectangle{{0, 0, int(texture.store.width), int(texture.store.height)}}

	layers := texture.store.layers + 1

	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fbo)

	dstStore := newStore(int(layers), int(texture.store.width), int(texture.store.height), int(texture.store.mipmaps))
	dstStore.SetFiltering(texture.min, texture.mag)
	dstStore.Bind(texture.store.binding)

	for layer := int32(0); layer < layers-1; layer++ {
		for level := int32(0); level < texture.store.mipmaps; level++ {
			gl.FramebufferTextureLayer(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, texture.store.id, level, layer)

			div := int32(1 << uint(level))
			gl.CopyTexSubImage3D(gl.TEXTURE_2D_ARRAY, level, 0, 0, layer, 0, 0, texture.store.width/div, texture.store.height/div)
		}
	}

	gl.DeleteFramebuffers(1, &fbo)
	texture.store.Dispose()

	texture.store = dstStore
}

func (texture *TextureAtlas) SetData(x, y, width, height, layer int, data []uint8) {
	if len(data) != width*height*4 {
		panic("Wrong number of pixels given!")
	}

	gl.TexSubImage3D(gl.TEXTURE_2D_ARRAY, 0, int32(x), int32(y), int32(layer), int32(width), int32(height), 1, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data))
	if texture.store.mipmaps > 1 {
		gl.GenerateMipmap(gl.TEXTURE_2D_ARRAY)
	}
}

func (texture *TextureAtlas) GetID() uint32 {
	return texture.store.id
}

func (texture *TextureAtlas) GetWidth() int32 {
	return texture.store.width
}

func (texture *TextureAtlas) GetHeight() int32 {
	return texture.store.height
}

func (texture *TextureAtlas) GetRegion() TextureRegion {
	return texture.defRegion
}

func (texture *TextureAtlas) GetLayers() int32 {
	return texture.store.layers
}

func (texture *TextureAtlas) SetFiltering(min, mag Filter) {
	texture.min, texture.mag = min, mag
	texture.store.SetFiltering(min, mag)
}

func (texture *TextureAtlas) Bind(loc uint) {
	texture.store.Bind(loc)
}

func (texture *TextureAtlas) GetLocation() uint {
	return texture.store.binding
}

func (texture *TextureAtlas) Dispose() {
	mainthread.CallNonBlock(func() {
		texture.store.Dispose()
	})
}
