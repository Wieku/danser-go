package containers

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/graphics/sliderrenderer"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"math"
	"sort"
)

type renderableProxy struct {
	renderable   objects.Renderable
	IsSliderBody bool
	depth        int64
	endTime      int64
}

type HitObjectContainer struct {
	beatMap       *beatmap.BeatMap
	objectQueue   []objects.BaseObject
	renderables   []*renderableProxy
	spriteManager *sprite.SpriteManager
	lastTime      float64
}

func NewHitObjectContainer(beatMap *beatmap.BeatMap) *HitObjectContainer {
	container := new(HitObjectContainer)
	container.beatMap = beatMap
	container.objectQueue = beatMap.GetObjectsCopy()
	container.spriteManager = sprite.NewSpriteManager()
	container.renderables = make([]*renderableProxy, 0)

	prempt := 800.0
	postmt := 240.0
	lineDist := 32.0

	for i := 1; i < len(container.objectQueue); i++ {
		_, ok1 := container.objectQueue[i-1].(*objects.Spinner)
		_, ok2 := container.objectQueue[i].(*objects.Spinner)
		if ok1 || ok2 || container.objectQueue[i].GetBasicData().NewCombo {
			continue
		}

		prevTime := float64(container.objectQueue[i-1].GetBasicData().EndTime)
		prevPos := container.objectQueue[i-1].GetBasicData().EndPos.Copy64()

		nextTime := float64(container.objectQueue[i].GetBasicData().StartTime)
		nextPos := container.objectQueue[i].GetBasicData().StartPos.Copy64()

		vec := nextPos.Sub(prevPos)
		duration := nextTime - prevTime
		distance := vec.Len()
		rotation := vec.AngleR()

		for prog := lineDist * 1.5; prog < distance-lineDist; prog += lineDist {
			t := prog / distance

			tStart := prevTime + t*duration - prempt
			tEnd := prevTime + t*duration

			pos := prevPos.Add(vec.Scl(t))

			textures := skin.GetFrames("followpoint", true)

			sprite := sprite.NewAnimation(textures, 1000.0/float64(len(textures)), true, -float64(i), pos, bmath.Origin.Centre)
			sprite.SetRotation(rotation)
			sprite.ShowForever(false)

			sprite.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, tStart, tStart+postmt, 0, 1))
			sprite.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, tEnd, tEnd+postmt, 1, 0))
			sprite.AdjustTimesToTransformations()
			sprite.SetAlpha(0)

			container.spriteManager.Add(sprite)
		}
	}

	return container
}

func (container *HitObjectContainer) addProxy(proxy *renderableProxy) {
	n := sort.Search(len(container.renderables), func(j int) bool {
		return proxy.depth < container.renderables[j].depth
	})

	container.renderables = append(container.renderables, nil) //allocate bigger array in case when len=cap
	copy(container.renderables[n+1:], container.renderables[n:])

	container.renderables[n] = proxy
}

func (container *HitObjectContainer) Update(time float64) {
	container.spriteManager.Update(int64(time))

	if time > 0 {
		delta := time - container.lastTime
		settings.Objects.Colors.Color.Update(delta)
		settings.Objects.Colors.Sliders.Border.Color.Update(delta)
		settings.Objects.Colors.Sliders.Body.Color.Update(delta)

		container.lastTime = time
	}

}

func (container *HitObjectContainer) Draw(batch *batch.QuadBatch, cameras []mgl32.Mat4, time float64, scale, alpha float32) {
	divides := len(cameras)

	if len(container.objectQueue) > 0 {
		for i := 0; i < len(container.objectQueue); i++ {
			if p := container.objectQueue[i]; p.GetBasicData().StartTime-15000 <= int64(time) {
				if p := container.objectQueue[i]; p.GetBasicData().StartTime-int64(container.beatMap.Diff.Preempt) <= int64(time) {

					if _, ok := p.(*objects.Spinner); ok {
						container.addProxy(&renderableProxy{
							renderable:   p.(objects.Renderable),
							IsSliderBody: false,
							depth:        math.MaxInt64,
							endTime:      p.GetBasicData().EndTime + difficulty.HitFadeOut,
						})
					} else {
						container.addProxy(&renderableProxy{
							renderable:   p.(objects.Renderable),
							IsSliderBody: false,
							depth:        p.GetBasicData().StartTime,
							endTime:      p.GetBasicData().EndTime + container.beatMap.Diff.Hit50 + difficulty.HitFadeOut,
						})
					}

					if _, ok := p.(*objects.Slider); ok {
						container.addProxy(&renderableProxy{
							renderable:   p.(objects.Renderable),
							IsSliderBody: true,
							depth:        p.GetBasicData().EndTime + 10,
							endTime:      p.GetBasicData().EndTime + difficulty.HitFadeOut,
						})
					}

					container.objectQueue = container.objectQueue[1:]
					i--
				}
			} else {
				break
			}
		}
	}

	if settings.Playfield.DrawObjects {
		objectColors := settings.Objects.Colors.Color.GetColors(divides, float64(scale), float64(scale))
		borderColors := objectColors
		bodyColors := objectColors

		if !settings.Objects.Colors.Sliders.Border.UseHitCircleColor {
			borderColors = settings.Objects.Colors.Sliders.Border.Color.GetColors(divides, float64(scale), float64(scale))
		}

		if !settings.Objects.Colors.Sliders.Body.UseHitCircleColor {
			bodyColors = settings.Objects.Colors.Sliders.Body.Color.GetColors(divides, float64(scale), float64(scale))
		}

		if !settings.Objects.ScaleToTheBeat {
			scale = 1
		}

		batch.Begin()
		batch.ResetTransform()
		batch.SetColor(1, 1, 1, 1)
		batch.SetScale(float64(scale)*container.beatMap.Diff.CircleRadius/64, float64(scale)*container.beatMap.Diff.CircleRadius/64)

		if divides < settings.Objects.Colors.MandalaTexturesTrigger && settings.Objects.DrawFollowPoints {
			for i := 0; i < divides; i++ {
				batch.SetCamera(cameras[i])
				container.spriteManager.Draw(int64(time), batch)
			}
		}

		batch.Flush()
		batch.SetScale(1, 1)

		for i := len(container.renderables) - 1; i >= 0; i-- {
			if s, ok := container.renderables[i].renderable.(*objects.Slider); ok && container.renderables[i].IsSliderBody {
				s.DrawBodyBase(int64(time), cameras[0])
			}
		}

		if settings.Objects.Sliders.SliderMerge {
			enabled := false

			for j := 0; j < divides; j++ {
				ind := j - 1
				if ind < 0 {
					ind = divides - 1
				}

				for i := len(container.renderables) - 1; i >= 0; i-- {
					if s, ok := container.renderables[i].renderable.(*objects.Slider); ok && container.renderables[i].IsSliderBody {
						if !enabled {
							enabled = true
							sliderrenderer.BeginRendererMerge()
						}

						s.DrawBody(int64(time), bodyColors[j], borderColors[j], borderColors[ind], cameras[j], scale)
					}
				}
			}
			if enabled {
				sliderrenderer.EndRendererMerge()
			}
		}

		batch.SetAdditive(divides >= settings.Objects.Colors.MandalaTexturesTrigger)
		batch.SetScale(float64(scale)*container.beatMap.Diff.CircleRadius/64, float64(scale)*container.beatMap.Diff.CircleRadius/64)

		for j := 0; j < divides; j++ {
			batch.SetCamera(cameras[j])

			ind := j - 1
			if ind < 0 {
				ind = divides - 1
			}

			batch.Flush()

			enabled := false

			for i := len(container.renderables) - 1; i >= 0; i-- {
				proxy := container.renderables[i]

				if !proxy.IsSliderBody {
					if enabled && !settings.Objects.Sliders.SliderMerge {
						enabled = false

						sliderrenderer.EndRenderer()
					}

					_, sp := container.renderables[i].renderable.(*objects.Spinner)
					if !sp || j == 0 {
						proxy.renderable.Draw(int64(time), objectColors[j], batch)
					}
				} else if !settings.Objects.Sliders.SliderMerge {
					if !enabled {
						enabled = true

						batch.Flush()

						sliderrenderer.BeginRenderer()
					}

					proxy.renderable.(*objects.Slider).DrawBody(int64(time), bodyColors[j], borderColors[j], borderColors[ind], cameras[j], scale)
				}

				if proxy.endTime <= int64(time) {
					container.renderables = append(container.renderables[:i], container.renderables[(i+1):]...)
				}
			}

			if enabled {
				sliderrenderer.EndRenderer()
			}
		}

		if divides < settings.Objects.Colors.MandalaTexturesTrigger && settings.Objects.DrawApproachCircles {
			for j := 0; j < divides; j++ {
				batch.SetCamera(cameras[j])

				for i := len(container.renderables) - 1; i >= 0; i-- {
					if s := container.renderables[i]; !s.IsSliderBody {
						s.renderable.DrawApproach(int64(time), objectColors[j], batch)
					}
				}
			}
		}

		batch.SetAdditive(false)
		batch.SetScale(1, 1)
		batch.End()
	}
}
