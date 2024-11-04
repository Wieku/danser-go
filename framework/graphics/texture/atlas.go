package texture

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	color2 "github.com/wieku/danser-go/framework/math/color"
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
	*TextureMultiLayer
	padding     int
	subTextures map[string]*TextureRegion
	emptySpaces map[int][]rectangle
}

func NewTextureAtlas(size, mipmaps int) *TextureAtlas {
	return NewTextureAtlasFormat(size, RGBA, mipmaps, 1)
}

func NewTextureAtlasCC(size, mipmaps int, clearColor color2.Color) *TextureAtlas {
	return NewTextureAtlasFormatCC(size, RGBA, mipmaps, 1, clearColor)
}

func NewTextureAtlasFormat(size int, format Format, mipmaps int, layers int) *TextureAtlas {
	return NewTextureAtlasFormatCC(size, format, mipmaps, layers, color2.NewLA(0, 0))
}

func NewTextureAtlasFormatCC(size int, format Format, mipmaps int, layers int, clearColor color2.Color) *TextureAtlas {
	texture := new(TextureAtlas)
	texture.subTextures = make(map[string]*TextureRegion)
	texture.emptySpaces = make(map[int][]rectangle)

	for i := 0; i < layers; i++ {
		texture.emptySpaces[i] = []rectangle{{0, 0, size, size}}
	}

	var siz int32
	gl.GetIntegerv(gl.MAX_TEXTURE_SIZE, &siz)

	if int(siz) < size {
		log.Printf("WARNING: GPU supports only %dx%d textures\n", siz, siz)
		size = int(siz)
	}

	texture.TextureMultiLayer = NewTextureMultiLayerFormatCC(size, size, format, mipmaps, layers, clearColor)

	texture.defRegion = TextureRegion{texture, 0, 1, 0, 1, float32(size), float32(size), 0}
	texture.padding = 1 << uint(texture.store.mipmaps)

	runtime.SetFinalizer(texture, (*TextureAtlas).Dispose)

	return texture
}

func (texture *TextureAtlas) AddTexture(name string, width, height int, data []uint8) *TextureRegion {
	if len(data) != width*height*texture.store.format.Size() {
		panic("Wrong number of pixels given!")
	}

	if int(texture.GetWidth()) <= width || int(texture.GetHeight()) <= height {
		log.Printf("Texture is too big! Atlas size: %dx%d, texture size: %dx%d", texture.GetWidth(), texture.GetHeight(), width, height)
		return nil
	}

	imBounds := rectangle{0, 0, width + texture.padding, height + texture.padding}

	for layer := 0; layer < int(texture.store.layers); layer++ {
		spaceIndex := -1
		smallest := rectangle{0, 0, int(texture.store.width), int(texture.store.height)}

		for i, space := range texture.emptySpaces[layer] {
			if imBounds.width <= space.width && imBounds.height <= space.height {
				if space.area() <= smallest.area() {
					spaceIndex = i
					smallest = space
				}
			}
		}

		if spaceIndex >= 0 {
			dw := smallest.width - imBounds.width
			dh := smallest.height - imBounds.height

			var rect1, rect2 rectangle

			if dh > dw {
				rect1 = rectangle{smallest.x + imBounds.width, smallest.y, smallest.width - imBounds.width, imBounds.height}
				rect2 = rectangle{smallest.x, smallest.y + imBounds.height, smallest.width, smallest.height - imBounds.height}
			} else {
				rect1 = rectangle{smallest.x + imBounds.width, smallest.y, smallest.width - imBounds.width, smallest.height}
				rect2 = rectangle{smallest.x, smallest.y + imBounds.height, imBounds.width, smallest.height - imBounds.height}
			}

			texture.emptySpaces[layer][spaceIndex] = rect1
			texture.emptySpaces[layer] = append(texture.emptySpaces[layer], rect2)

			texture.SetData(smallest.x, smallest.y, width, height, layer, data)

			region := TextureRegion{Texture: texture, Width: float32(width), Height: float32(height), Layer: int32(layer)}
			region.U1 = (float32(smallest.x) + 0.5) / float32(texture.store.width)
			region.V1 = (float32(smallest.y) + 0.5) / float32(texture.store.height)
			region.U2 = region.U1 + float32(width-1)/float32(texture.store.width)
			region.V2 = region.V1 + float32(height-1)/float32(texture.store.height)

			texture.subTextures[name] = &region
			return &region
		} else if layer == int(texture.store.layers)-1 {
			texture.NewLayer()
		}
	}

	return nil
}

func (texture *TextureAtlas) GetTexture(name string) *TextureRegion {
	return texture.subTextures[name]
}

func (texture *TextureAtlas) NewLayer() {
	texture.emptySpaces[int(texture.store.layers)] = []rectangle{{0, 0, int(texture.store.width), int(texture.store.height)}}
	texture.TextureMultiLayer.NewLayer()
}
