package common

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/shape"
	color2 "github.com/wieku/danser-go/framework/math/color"
)

type Boundaries struct {
	shapeRenderer *shape.Renderer
}

func NewBoundaries() *Boundaries {
	boundaries := new(Boundaries)
	boundaries.shapeRenderer = shape.NewRendererSize(18)

	return boundaries
}

func (boundaries *Boundaries) Draw(projection mgl32.Mat4, circleRadius, alpha float32) {
	if !settings.Gameplay.Boundaries.Enabled {
		return
	}

	thickness := float32(settings.Gameplay.Boundaries.BorderThickness)
	cs := circleRadius

	boundaries.shapeRenderer.SetCamera(projection)
	boundaries.shapeRenderer.Begin()

	if bAlpha := settings.Gameplay.Boundaries.BackgroundOpacity; bAlpha > 0.001 {
		colHSV := settings.Gameplay.Boundaries.BackgroundColor
		color := color2.NewHSVA(float32(colHSV.Hue), float32(colHSV.Saturation), float32(colHSV.Value), float32(bAlpha)*alpha)

		boundaries.shapeRenderer.SetColorM(color)
		boundaries.shapeRenderer.DrawQuad(-cs, -cs, camera.OsuWidth+cs, -cs, camera.OsuWidth+cs, camera.OsuHeight+cs, -cs, camera.OsuHeight+cs)
	}

	if bAlpha := settings.Gameplay.Boundaries.BorderOpacity; bAlpha > 0.001 {
		colHSV := settings.Gameplay.Boundaries.BorderColor
		color := color2.NewHSVA(float32(colHSV.Hue), float32(colHSV.Saturation), float32(colHSV.Value), float32(bAlpha)*alpha)

		boundaries.shapeRenderer.SetColorM(color)

		half := thickness / 2

		csH := cs + half

		if settings.Gameplay.Boundaries.BorderFill > 0.99 {
			boundaries.shapeRenderer.DrawLine(-csH-half, -csH, camera.OsuWidth+csH+half, -csH, thickness)
			boundaries.shapeRenderer.DrawLine(-csH-half, camera.OsuHeight+csH, camera.OsuWidth+csH+half, camera.OsuHeight+csH, thickness)
		} else {
			dx := (camera.OsuWidth + cs*2) / 2 * float32(settings.Gameplay.Boundaries.BorderFill)

			// top
			boundaries.shapeRenderer.DrawLine(-cs-thickness, -csH, -cs+dx, -csH, thickness)
			boundaries.shapeRenderer.DrawLine(camera.OsuWidth+cs-dx, -csH, camera.OsuWidth+cs+thickness, -csH, thickness)

			// bottom
			boundaries.shapeRenderer.DrawLine(-cs-thickness, camera.OsuHeight+csH, -cs+dx, camera.OsuHeight+csH, thickness)
			boundaries.shapeRenderer.DrawLine(camera.OsuWidth+cs-dx, camera.OsuHeight+csH, camera.OsuWidth+cs+thickness, camera.OsuHeight+csH, thickness)
		}

		if settings.Gameplay.Boundaries.BorderFill > camera.OsuHeight/camera.OsuWidth {
			boundaries.shapeRenderer.DrawLine(-csH, -csH+half, -csH, camera.OsuHeight+csH-half, thickness)
			boundaries.shapeRenderer.DrawLine(camera.OsuWidth+csH, -csH+half, camera.OsuWidth+csH, camera.OsuHeight+csH-half, thickness)
		} else {
			dy := (camera.OsuWidth + cs*2) / 2 * float32(settings.Gameplay.Boundaries.BorderFill)

			// left
			boundaries.shapeRenderer.DrawLine(-csH, -cs, -csH, -cs+dy, thickness)
			boundaries.shapeRenderer.DrawLine(-csH, camera.OsuHeight+cs-dy, -csH, camera.OsuHeight+cs, thickness)

			// right
			boundaries.shapeRenderer.DrawLine(camera.OsuWidth+csH, -cs, camera.OsuWidth+csH, -cs+dy, thickness)
			boundaries.shapeRenderer.DrawLine(camera.OsuWidth+csH, camera.OsuHeight+cs-dy, camera.OsuWidth+csH, camera.OsuHeight+cs, thickness)
		}
	}

	boundaries.shapeRenderer.End()
}
