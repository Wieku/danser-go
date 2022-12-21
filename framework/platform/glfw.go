package platform

import (
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"image"
	"strconv"
)

var iconSizes = []int{128, 48, 24, 16}

func LoadIcons(win *glfw.Window, prefix, suffix string) {
	var pixmaps []*texture.Pixmap
	var images []image.Image

	for _, size := range iconSizes {
		pxMap, _ := assets.GetPixmap("assets/textures/" + prefix + strconv.Itoa(size) + suffix + ".png")

		pixmaps = append(pixmaps, pxMap)
		images = append(images, pxMap.NRGBA())
	}

	win.SetIcon(images)

	for _, pxMap := range pixmaps {
		pxMap.Dispose()
	}
}
