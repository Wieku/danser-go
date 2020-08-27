package drawables

import (
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/render"
	"github.com/wieku/danser-go/app/render/batches"
	"github.com/wieku/danser-go/app/render/sprites"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/bass"
	"math"
	"math/rand"
	"sort"
)

const baseSpeed = 100.0
const separation = 1.4
const decay = 0.5
const minSize = 60.0
const maxSize = 200.0
const bars = 20
const maxTriangles = 40

type Triangles struct {
	triangles    []*sprites.Sprite
	lastTime     float64
	counter      float64
	velocity     float64
	fft          []float64
	colorPalette []bmath.Color
	music        *bass.Music
}

func NewTriangles(colors []bmath.Color) *Triangles {
	visualiser := &Triangles{triangles: make([]*sprites.Sprite, 0), velocity: 100}
	visualiser.colorPalette = colors

	for i := 0; i < maxTriangles; i++ {
		visualiser.AddTriangle(false)
	}
	sort.Slice(visualiser.triangles, func(i, j int) bool {
		return visualiser.triangles[i].GetDepth() > visualiser.triangles[j].GetDepth()
	})
	return visualiser
}

func (vis *Triangles) SetTrack(track *bass.Music) {
	vis.music = track
}

func (vis *Triangles) AddTriangle(onscreen bool) {
	size := (minSize + rand.Float64()*(maxSize-minSize)) * settings.Graphics.GetHeightF() / 768
	position := bmath.NewVec2d((rand.Float64()-0.5)*settings.Graphics.GetWidthF(), settings.Graphics.GetHeightF()/2+size)
	sprite := sprites.NewSpriteSingle(render.Triangle, size, position, bmath.NewVec2d(0, 0))
	if vis.colorPalette == nil {
		sprite.SetColor(bmath.Color{rand.Float64(), rand.Float64(), rand.Float64(), 1})
	} else {
		col := vis.colorPalette[rand.Intn(len(vis.colorPalette))]
		sprite.SetColor(col)
	}

	sprite.SetVFlip(rand.Float64() >= 0.5)
	sprite.SetScale(size / float64(render.Triangle.Height))
	sprite.SetAlpha(0.65) //0.5+rand.Float64()*0.5)
	if onscreen {
		sprite.SetPosition(bmath.NewVec2d(sprite.GetPosition().X, -(rand.Float64()-0.5)*(settings.Graphics.GetHeightF()+size)))
		//sprite.AddTransform(animation.NewSingleTransform(animation.MoveY, easing.OutQuad, -2000, -1000, position.Y, -(rand.Float64() - 0.5)*(settings.Graphics.GetHeightF()+size)), false)
	}

	vis.triangles = append(vis.triangles, sprite)
}

func (vis *Triangles) SetColors(colors []bmath.Color) {
	vis.colorPalette = colors
}

func (vis *Triangles) Update(time float64) {
	if vis.lastTime == 0 {
		vis.lastTime = time
	}
	delta := time - vis.lastTime

	boost := 0.0

	fft := vis.music.GetFFT()

	for i := 0; i < bars; i++ {
		boost += 2 * float64(fft[i]) * float64(bars-i) / float64(bars)
	}

	vis.velocity = math.Max(vis.velocity, math.Min(boost*1.5, 6))

	vis.velocity *= 1.0 - 0.05*delta/16

	toAdd := 0

	velocity := vis.velocity + 0.5

	for i := 0; i < len(vis.triangles); i++ {
		t := vis.triangles[i]
		t.Update(int64(time))
		scale := (t.GetScale().Y * float64(render.Triangle.Width)) / maxSize / (settings.Graphics.GetHeightF() / 768)
		t.SetPosition(t.GetPosition().AddS(0, -delta/16*velocity*(0.2+(1.0-scale*0.8)*separation)*settings.Graphics.GetHeightF()/768))
		if t.GetPosition().Y < -settings.Graphics.GetHeightF()/2-t.GetScale().Y*float64(render.Triangle.Width)/2 {
			vis.triangles = append(vis.triangles[:i], vis.triangles[i+1:]...)
			i--
			toAdd++
		}
	}

	if toAdd > 0 {
		for i := 0; i < toAdd; i++ {
			vis.AddTriangle(false)
		}
		sort.Slice(vis.triangles, func(i, j int) bool {
			return vis.triangles[i].GetDepth() > vis.triangles[j].GetDepth()
		})
	}

	vis.lastTime = time
}

func (vis *Triangles) Draw(time float64, batch *batches.SpriteBatch) {
	for _, t := range vis.triangles {
		t.Draw(int64(time), batch)
	}
}
