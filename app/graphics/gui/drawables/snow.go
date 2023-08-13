package drawables

import (
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/input"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"math/rand"
)

type Snowflake struct {
	*sprite.Sprite

	horizVel float64
	wind     float64
}

type Snow struct {
	manager *sprite.Manager

	lastTime    float64
	firstUpdate bool

	lastMouse float64
}

func NewSnow() *Snow {
	return &Snow{
		firstUpdate: true,
		lastTime:    math.NaN(),
		manager:     sprite.NewManager(),
		lastMouse:   -1,
	}
}

func (vis *Snow) AddSnowflake(onscreen bool) {
	size := (minSize + rand.Float64()*(maxSize-minSize)) * settings.Graphics.GetHeightF() / 768 * 0.15
	position := vector.NewVec2d((rand.Float64()*1.4-0.2)*settings.Graphics.GetWidthF(), -size)

	if onscreen {
		position.Y = rand.Float64() * settings.Graphics.GetHeightF()
	}

	texture := graphics.Snowflakes[rand.Intn(len(graphics.Snowflakes))]

	snowflake := &Snowflake{
		Sprite:   sprite.NewSpriteSingle(texture, -size, position, vector.Centre),
		horizVel: (rand.Float64() - 0.5) / 8,
		wind:     (rand.Float64() - 0.5) / 4000,
	}

	snowflake.SetColor(color2.NewL(1 - rand.Float32()*0.3))
	snowflake.SetRotation(rand.Float64() * math.Pi * 2)
	snowflake.SetScale(size / float64(snowflake.Texture.Height))
	snowflake.SetAdditive(true)
	snowflake.SetAlpha(0.4 + rand.Float32()*0.3)

	vis.manager.Add(snowflake)
}

func (vis *Snow) Update(time float64) {
	if math.IsNaN(vis.lastTime) {
		vis.lastTime = time
	}

	delta := (time - vis.lastTime) / 16

	mX, _ := input.Win.GetCursorPos()
	mX = mutils.Clamp(mX, 0, settings.Graphics.GetWidthF())

	vUpdate := 0.0

	if vis.lastMouse != -1 {
		vUpdate = (mX - vis.lastMouse) / 2000
	}

	vis.lastMouse = mX

	triangles := vis.manager.GetProcessedSprites()
	existingSnow := len(triangles)

	for i := 0; i < len(triangles); i++ {
		t := triangles[i].(*Snowflake)
		t.Update(time)

		t.horizVel += delta*t.wind + vUpdate

		scale := (t.GetScale().Y * float64(t.Texture.Width)) / maxSize / (settings.Graphics.GetHeightF() / 768) / 0.15

		t.SetPosition(t.GetPosition().AddS(delta*t.horizVel, delta*0.5*(1.0-scale*0.25)*settings.Graphics.GetHeightF()/768))
		t.SetRotation(t.GetRotation() + delta*t.horizVel/6)

		if t.GetPosition().Y > settings.Graphics.GetHeightF()+t.GetScale().Y*float64(t.Texture.Width)/2 {
			t.ShowForever(false)
			existingSnow--
		}
	}

	toAdd := maxTriangles*5 - existingSnow

	if toAdd > 0 {
		for i := 0; i < toAdd; i++ {
			vis.AddSnowflake(vis.firstUpdate)
		}

		vis.firstUpdate = false
	}

	vis.manager.Update(time)

	vis.lastTime = time
}

func (vis *Snow) Draw(time float64, batch *batch.QuadBatch) {
	vis.manager.Draw(time, batch)
}
