package utils

import (
	"os"
	"image"
	_ "image/jpeg"
	"image/draw"
	"github.com/faiface/glhf"
	"log"
	"github.com/go-gl/gl/v3.3-core/gl"
)

func LoadImage(path string) (*image.NRGBA, error) {
	file, err := os.Open(path)
	log.Println("Loading texture: ", file.Name())
	if err != nil {
		log.Println("er1")
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		log.Println("er2")
		return nil, err
	}
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(nrgba, nrgba.Bounds(), img, bounds.Min, draw.Src)
	return nrgba, nil
}

func LoadTexture(path string) (*glhf.Texture, error) {
	img, err := LoadImage(path)
	log.Println(path, img.NRGBAAt(48, 5))
	if err == nil {
		tex := glhf.NewTexture(
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			true,
			img.Pix,
		)
		tex.Begin()
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		tex.End()
		return tex, nil
	}
	return nil, err
}
