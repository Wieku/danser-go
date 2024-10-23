package graphics

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/bmath/camera"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/blend"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
	"math/rand"
	"time"
)

type cursorRenderer interface {
	SetPosition(position vector.Vector2f)
	Update(delta float64)
	UpdateRenderer()
	DrawM(scale, expand float64, batch *batch.QuadBatch, color color2.Color, colorGlow color2.Color)
}

var cursorFbo *buffer.Framebuffer = nil
var cursorFBOSprite *sprite.Sprite

var cursorSpaceFbo *buffer.Framebuffer = nil
var cursorSpaceFBOSprite *sprite.Sprite

var fboBatch *batch.QuadBatch

var Camera *camera.Camera
var osuRect camera.Rectangle

var useAdditive = false

func initCursor() {
	if settings.Cursor.TrailStyle < 1 || settings.Cursor.TrailStyle > 4 {
		panic("Wrong cursor trail type")
	}

	cursorFbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	region := cursorFbo.Texture().GetRegion()
	cursorFBOSprite = sprite.NewSpriteSingle(&region, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), vector.Centre)

	cursorSpaceFbo = buffer.NewFrame(int(settings.Graphics.GetWidth()), int(settings.Graphics.GetHeight()), true, false)
	regionSpace := cursorSpaceFbo.Texture().GetRegion()
	cursorSpaceFBOSprite = sprite.NewSpriteSingle(&regionSpace, 0, vector.NewVec2d(settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()/2), vector.Centre)

	fboBatch = batch.NewQuadBatchSize(1)
	fboBatch.SetCamera(mgl32.Ortho(0, float32(settings.Graphics.GetWidth()), 0, float32(settings.Graphics.GetHeight()), -1, 1))

	osuRect = Camera.GetWorldRect()
}

type Cursor struct {
	scale *animation.Glider

	lastLeftState, lastRightState bool

	LeftButton, RightButton bool
	LeftKey, RightKey       bool
	LeftMouse, RightMouse   bool

	IsReplayFrame bool // TODO: temporary hacky solution for spinners
	IsPlayer      bool
	IsAutoplay    bool
	IsReplay      bool

	OldSpinnerScoring bool

	ModifiedMods bool

	LastFrameTime    int64 //
	CurrentFrameTime int64 //
	RawPosition      vector.Vector2f
	InvertDisplay    bool

	Position vector.Vector2f

	Name      string
	ScoreID   int64
	ScoreTime time.Time

	lastSetting bool

	renderer cursorRenderer

	SmokeKey           bool
	lastSmokeKey       bool
	smokePointCount    int
	firstSmokePosition vector.Vector2f
	lastSmokePosition  vector.Vector2f
	smokeTexture       *texture.TextureRegion
	smokeContainer     *sprite.Manager

	rippleContainer *sprite.Manager
	time            float64
}

func NewCursor() *Cursor {
	if cursorFbo == nil {
		initCursor()
	}

	cursor := &Cursor{Position: vector.NewVec2f(256, -500)}
	cursor.scale = animation.NewGlider(1.0)

	cursor.lastSetting = settings.Skin.Cursor.UseSkinCursor

	if cursor.lastSetting {
		cursor.renderer = newOsuRenderer()
	} else {
		cursor.renderer = newDanserRenderer()
	}

	skin.GetTexture("cursor-ripple")

	cursor.smokeTexture = skin.GetTexture("cursor-smoke")
	cursor.smokeContainer = sprite.NewManager()

	cursor.rippleContainer = sprite.NewManager()

	return cursor
}

func (cursor *Cursor) SetPos(pt vector.Vector2f) {
	cursor.RawPosition = pt
	tmp := pt

	if cursor.InvertDisplay {
		tmp.Y = 384 - tmp.Y
	}

	if settings.Cursor.BounceOnEdges && settings.DIVIDES <= 2 {
		tmp.X -= osuRect.MinX
		tmp.Y -= osuRect.MinY
		tmp.X = math32.Mod(tmp.X, 2*(osuRect.MaxX-osuRect.MinX))
		tmp.Y = math32.Mod(tmp.Y, 2*(osuRect.MaxY-osuRect.MinY))
		tmp.X += osuRect.MinX
		tmp.Y += osuRect.MinY

		for {
			ok1, ok2 := false, false

			if tmp.X < osuRect.MinX {
				tmp.X = 2*osuRect.MinX - tmp.X
			} else if tmp.X > osuRect.MaxX {
				tmp.X = 2*osuRect.MaxX - tmp.X
			} else {
				ok1 = true
			}

			if tmp.Y < osuRect.MinY {
				tmp.Y = 2*osuRect.MinY - tmp.Y
			} else if tmp.Y > osuRect.MaxY {
				tmp.Y = 2*osuRect.MaxY - tmp.Y
			} else {
				ok2 = true
			}

			if ok1 && ok2 {
				break
			}
		}
	}

	cursor.Position = tmp
	cursor.renderer.SetPosition(cursor.Position)
}

func (cursor *Cursor) SetScreenPos(pt vector.Vector2f) {
	cursor.SetPos(Camera.Unproject(pt.Copy64()).Copy32())
}

func (cursor *Cursor) Update(delta float64) {
	delta = math.Abs(delta)
	cursor.time += delta

	leftState := cursor.LeftKey || cursor.LeftMouse
	rightState := cursor.RightKey || cursor.RightMouse

	if settings.Cursor.CursorRipples && settings.PLAYERS == 1 && ((!cursor.lastLeftState && leftState) || (!cursor.lastRightState && rightState)) {
		spr := sprite.NewSpriteSingle(skin.GetTexture("cursor-ripple"), cursor.time, cursor.Position.Copy64(), vector.Centre)
		spr.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, cursor.time, cursor.time+700, 0.3, 0.0))
		spr.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, cursor.time, cursor.time+700, 0.05, 0.5))
		spr.ResetValuesToTransforms()
		spr.AdjustTimesToTransformations()
		spr.ShowForever(false)

		cursor.rippleContainer.Add(spr)
	}

	if cursor.lastLeftState != leftState || cursor.lastRightState != rightState {
		if leftState || rightState {
			cursor.scale.AddEventS(cursor.scale.GetTime(), cursor.scale.GetTime()+100, 1.0, 1.3)
		} else {
			cursor.scale.AddEventS(cursor.scale.GetTime(), cursor.scale.GetTime()+100, cursor.scale.GetValue(), 1.0)
		}

		cursor.lastLeftState = leftState
		cursor.lastRightState = rightState
	}

	cursor.smokeUpdate()

	cursor.scale.UpdateD(delta)

	cursor.renderer.Update(delta)

	cursor.rippleContainer.Update(cursor.time)
}

func (cursor *Cursor) smokeUpdate() {
	if (!settings.Cursor.SmokeEnabled && settings.PLAYERS == 1) || (!settings.Knockout.SmokeEnabled && settings.PLAYERS != 1) {
		return
	}

	if cursor.SmokeKey {
		if !cursor.lastSmokeKey {
			cursor.lastSmokePosition = cursor.Position
			cursor.firstSmokePosition = cursor.Position
		}

		distance := max(2*scaling, cursor.smokeTexture.Width*scaling/4)
		points := cursor.Position.Dst(cursor.lastSmokePosition)

		if int(points/distance) > 0 {
			temp := cursor.lastSmokePosition
			for i := distance; i < points; i += distance {
				temp = cursor.Position.Sub(cursor.lastSmokePosition).Scl(i / points).Add(cursor.lastSmokePosition)

				smoke := sprite.NewSpriteSingle(cursor.smokeTexture, cursor.time*1000+float64(i), temp.Copy64(), vector.Centre)
				smoke.SetAdditive(settings.PLAYERS == 1)
				smoke.SetRotation(rand.Float64() * 2 * math.Pi)
				smoke.SetScale(0.5 / scaling)
				smoke.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, cursor.time, cursor.time+(4000/settings.Cursor.SmokeRemoveSpeed), 0.6, 0.0))
				smoke.ResetValuesToTransforms()
				smoke.AdjustTimesToTransformations()
				smoke.ShowForever(false)

				cursor.smokePointCount++

				cursor.smokeContainer.Add(smoke)
			}

			cursor.lastSmokePosition = temp

			if cursor.smokePointCount > 30 && cursor.Position.Dst(cursor.firstSmokePosition) < 10*scaling {
				cursor.smokeBrighten()
			}
		}
	} else if cursor.lastSmokeKey {
		cursor.smokeBrighten()
	}

	cursor.lastSmokeKey = cursor.SmokeKey

	cursor.smokeContainer.Update(cursor.time)
}

func (cursor *Cursor) smokeBrighten() {
	smokes := cursor.smokeContainer.GetProcessedSprites()

	delay := 0.0

	for _, s := range smokes {
		if (s.GetEndTime() - s.GetStartTime()) < 5000 {
			s.ClearTransformations()
			duration := 8000 / settings.Cursor.SmokeRemoveSpeed
			s.AddTransform(animation.NewSingleTransform(animation.Fade, easing.InQuad, cursor.time+delay, cursor.time+delay+duration, 1.0, 0.0))
			s.SetEndTime(cursor.time + delay + duration)

			delay += 2.0
		}
	}

	cursor.smokePointCount = 0
}

func (cursor *Cursor) UpdateRenderer() {
	newSettings := settings.Skin.Cursor.UseSkinCursor

	if newSettings != cursor.lastSetting {
		cursor.lastSetting = newSettings
		if cursor.lastSetting {
			cursor.renderer = newOsuRenderer()
		} else {
			cursor.renderer = newDanserRenderer()
		}
	}

	cursor.renderer.UpdateRenderer()
}

func BeginCursorRender() {
	useAdditive = settings.Cursor.AdditiveBlending && (settings.PLAYERS > 1 || settings.DIVIDES > 1 || settings.TAG > 1) && !settings.Skin.Cursor.UseSkinCursor

	if useAdditive {
		cursorSpaceFbo.Bind()
		cursorSpaceFbo.ClearColor(0.0, 0.0, 0.0, 0.0)
	}

	blend.Push()
	blend.Enable()
	blend.SetFunctionSeparate(blend.SrcAlpha, blend.OneMinusSrcAlpha, blend.One, blend.OneMinusSrcAlpha)
}

func EndCursorRender() {
	if useAdditive {
		cursorSpaceFbo.Unbind()

		fboBatch.Begin()
		cursorSpaceFBOSprite.Draw(0, fboBatch)
		fboBatch.End()
	}

	blend.Pop()
}

func (cursor *Cursor) Draw(scale float64, batch *batch.QuadBatch, color color2.Color) {
	cursor.DrawM(scale, batch, color, color)
}

func (cursor *Cursor) DrawM(scale float64, batch *batch.QuadBatch, color color2.Color, colorGlow color2.Color) {
	if cursor.rippleContainer.GetNumProcessed() > 0 || cursor.smokeContainer.GetNumProcessed() > 0 {
		batch.Begin()
		batch.SetAdditive(false)
		batch.ResetTransform()
		batch.SetColor(1, 1, 1, float64(color.A))
		batch.SetScale(scaling*scaling, scaling*scaling)
		batch.SetSubScale(1, 1)

		if settings.PLAYERS > 1 {
			batch.SetColor32(color.R, color.G, color.B, color.A)
		}

		cursor.smokeContainer.Draw(cursor.time, batch)
		batch.SetColor(1, 1, 1, float64(color.A))
		cursor.rippleContainer.Draw(cursor.time, batch)

		batch.End()
	}

	if useAdditive {
		cursorFbo.Bind()
		cursorFbo.ClearColor(0.0, 0.0, 0.0, 0.0)
	}

	cursor.renderer.DrawM(scale, cursor.scale.GetValue(), batch, color, colorGlow)

	if useAdditive {
		cursorFbo.Unbind()

		fboBatch.Begin()

		blend.Push()
		blend.SetFunction(blend.SrcAlpha, blend.One)

		cursorFBOSprite.Draw(0, fboBatch)
		fboBatch.Flush()

		blend.Pop()

		fboBatch.End()
	}
}
