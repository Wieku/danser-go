package drawables

import (
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"math/rand"
)

const baseSpeed = 100.0
const separation = 1.4
const decay = 0.5
const minSize = 60.0
const maxSize = 200.0
const bars = 10
const maxTriangles = 40

type Triangle struct {
	*sprite.Sprite
	shade  float32
	cIndex int
}

type Triangles struct {
	manager *sprite.Manager

	lastTime    float64
	velocity    float64
	firstUpdate bool

	colorPalette []color2.Color

	music bass.ITrack

	density float64
	scale   float64
}

func NewTriangles(colors []color2.Color) *Triangles {
	visualiser := &Triangles{velocity: 0}
	visualiser.colorPalette = colors
	visualiser.firstUpdate = true
	visualiser.manager = sprite.NewManager()
	visualiser.scale = 1.0
	visualiser.density = 1.0

	return visualiser
}

func (vis *Triangles) SetTrack(track bass.ITrack) {
	vis.music = track
}

func (vis *Triangles) AddTriangle(onscreen bool) {
	size := (minSize + rand.Float64()*(maxSize-minSize)) * settings.Graphics.GetHeightF() / 768 * vis.scale
	position := vector.NewVec2d((rand.Float64()-0.5)*settings.Graphics.GetWidthF(), settings.Graphics.GetHeightF()/2+size)

	texture := graphics.Triangle
	if settings.Playfield.Background.Triangles.Shadowed {
		texture = graphics.TriangleShadowed
	}

	triangle := &Triangle{
		Sprite: sprite.NewSpriteSingle(texture, -size, position, vector.NewVec2d(0, 0)),
		shade:  rand.Float32() * 0.2,
		cIndex: rand.Int(),
	}

	if vis.colorPalette == nil || len(vis.colorPalette) == 0 {
		triangle.SetColor(color2.NewL(triangle.shade))
	} else {
		triangle.SetColor(vis.colorPalette[triangle.cIndex%len(vis.colorPalette)])
	}

	triangle.SetVFlip(rand.Float64() >= 0.5)
	triangle.SetScale(size / float64(graphics.Triangle.Height))

	if onscreen {
		triangle.SetPosition(vector.NewVec2d(triangle.GetPosition().X, -(rand.Float64()-0.5)*(settings.Graphics.GetHeightF()+size)))
	}

	vis.manager.Add(triangle)
}

func (vis *Triangles) SetColors(colors []color2.Color) {
	vis.colorPalette = colors

	triangles := vis.manager.GetProcessedSprites()

	for i := 0; i < len(triangles); i++ {
		t := triangles[i].(*Triangle)

		if vis.colorPalette == nil || len(vis.colorPalette) == 0 {
			t.SetColor(color2.NewL(t.shade))
		} else {
			t.SetColor(vis.colorPalette[t.cIndex%len(vis.colorPalette)])
		}
	}
}

func (vis *Triangles) Update(time float64) {
	if vis.lastTime == 0 {
		vis.lastTime = time
	}

	delta := time - vis.lastTime

	boost := 0.0

	if vis.music != nil {
		fft := vis.music.GetFFT()

		for i := 0; i < bars; i++ {
			boost += 2 * float64(fft[i]*fft[i]) * float64(bars-i) / float64(bars)
		}
	}

	vis.velocity = math.Max(vis.velocity, math.Min(boost*12, 12))

	vis.velocity *= 1.0 - 0.05*delta/16

	velocity := (vis.velocity + 0.5) * settings.Playfield.Background.Triangles.Speed

	triangles := vis.manager.GetProcessedSprites()
	existingTriangles := len(triangles)

	for i := 0; i < len(triangles); i++ {
		t := triangles[i]
		t.Update(time)

		scale := (t.GetScale().Y * float64(graphics.Triangle.Width)) / maxSize / (settings.Graphics.GetHeightF() / 768) / vis.scale
		t.SetPosition(t.GetPosition().AddS(0, -delta/16*velocity*(0.2+(1.0-scale*0.8)*separation)*settings.Graphics.GetHeightF()/768))

		if t.GetPosition().Y < -settings.Graphics.GetHeightF()/2-t.GetScale().Y*float64(graphics.Triangle.Width)/2 {
			t.ShowForever(false)
			existingTriangles--
		}
	}

	toAdd := int(maxTriangles*vis.density) - existingTriangles

	if toAdd > 0 {
		for i := 0; i < toAdd; i++ {
			vis.AddTriangle(vis.firstUpdate)
		}

		vis.firstUpdate = false
	}

	vis.manager.Update(time)

	vis.lastTime = time
}

func (vis *Triangles) Draw(time float64, batch *batch.QuadBatch) {
	vis.manager.Draw(time, batch)
}

func (vis *Triangles) SetDensity(density float64) {
	vis.density = density
}

func (vis *Triangles) SetScale(scale float64) {
	vis.scale = scale
}
