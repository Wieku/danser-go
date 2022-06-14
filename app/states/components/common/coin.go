package common

import (
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/graphics/gui/drawables"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
)

type DanserCoin struct {
	*BeatSynced

	coinBottom *sprite.Sprite
	coinTop    *sprite.Sprite
	vis        *drawables.Visualiser

	drawVisualiser bool
}

func NewDanserCoin() *DanserCoin {
	coin := new(DanserCoin)

	coin.BeatSynced = NewBeatSynced()

	coin.SetColor(color2.NewL(1))
	coin.SetAlpha(1)

	coin.vis = drawables.NewVisualiser(1.0, 400.0, vector.NewVec2d(0, 0))

	pixmap, err := assets.GetPixmap("assets/textures/coinbig.png")
	if err != nil {
		panic(err)
	}

	LogoT := texture.LoadTextureSingle(pixmap.RGBA(), 4)

	pixmap.Dispose()

	rg := LogoT.GetRegion()

	coin.coinBottom = sprite.NewSpriteSingle(&rg, 0, vector.NewVec2d(0, 0), vector.NewVec2d(0, 0))
	coin.coinTop = sprite.NewSpriteSingle(&rg, 0, vector.NewVec2d(0, 0), vector.NewVec2d(0, 0))

	coin.Texture = &rg

	return coin
}

func (coin *DanserCoin) SetMap(bMap *beatmap.BeatMap, track bass.ITrack) {
	coin.BeatSynced.SetMap(bMap, track)
	coin.vis.SetTrack(track)
}

func (coin *DanserCoin) Update(time float64) {
	coin.BeatSynced.Update(time)

	innerCircleScale := 1.05 - easing.OutQuad(coin.Beat)*0.05
	outerCircleScale := 1.05 + easing.OutQuad(coin.Beat)*0.03

	scl := (1.0 / float64(coin.coinBottom.Texture.Width)) * 1.05

	nScl := coin.GetScale().Scl(scl * 2)

	bScl := nScl.Scl(innerCircleScale)

	coin.vis.SetStartDistance(coin.GetScale().X * innerCircleScale)

	coin.coinBottom.SetScaleV(bScl)
	coin.coinTop.SetScaleV(nScl.Scl(outerCircleScale))

	alpha := 0.3
	if coin.Kiai {
		alpha = 0.12
	}

	coin.vis.SetKiai(coin.Kiai)

	coin.coinTop.SetAlpha(float32(alpha * (1 - easing.OutQuad(coin.Beat)) * coin.GetAlpha()))
	coin.coinBottom.SetAlpha(coin.GetAlpha32())

	coin.vis.Position = coin.GetPosition()
	coin.coinBottom.SetPosition(coin.GetPosition())
	coin.coinTop.SetPosition(coin.GetPosition())
	coin.coinBottom.SetRotation(coin.GetRotation())
	coin.coinTop.SetRotation(coin.GetRotation())

	coin.coinBottom.Update(time)
	coin.coinTop.Update(time)
	coin.vis.Update(time)
}

func (coin *DanserCoin) Draw(time float64, batch *batch.QuadBatch) {
	if coin.drawVisualiser {
		coin.vis.Draw(time, batch)
	}

	coin.coinBottom.Draw(time, batch)
	coin.coinTop.Draw(time, batch)
}

func (coin *DanserCoin) DrawVisualiser(value bool) {
	coin.drawVisualiser = value
}
