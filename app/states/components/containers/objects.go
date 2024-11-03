package containers

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/beatmap/objects"
	"github.com/wieku/danser-go/app/graphics/sliderrenderer"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/profiler"
	"log"
	"math"
	"sort"
)

type renderableProxy struct {
	renderable   objects.Renderable
	isSliderBody bool
	depth        float64
	endTime      float64
}

type HitObjectContainer struct {
	beatMap        *beatmap.BeatMap
	objectQueue    []objects.IHitObject
	renderables    []*renderableProxy
	spriteManager  *sprite.Manager
	lastTime       float64
	countProcessed int
}

func NewHitObjectContainer(beatMap *beatmap.BeatMap) *HitObjectContainer {
	log.Println("Creating HitObject container...")

	container := &HitObjectContainer{
		beatMap:       beatMap,
		objectQueue:   beatMap.GetObjectsCopy(),
		spriteManager: sprite.NewManager(),
		renderables:   make([]*renderableProxy, 0),
	}

	container.createFollowPoints()

	log.Println("Container created.")

	return container
}

func (container *HitObjectContainer) createFollowPoints() {
	const (
		preEmpt  = 800.0
		lineDist = 32.0
	)

	textures := skin.GetFrames("followpoint", true)

	for i := 1; i < len(container.objectQueue); i++ {
		_, ok1 := container.objectQueue[i-1].(*objects.Spinner)
		_, ok2 := container.objectQueue[i].(*objects.Spinner)
		if ok1 || ok2 || container.objectQueue[i].IsNewCombo() { //nolint:wsl
			continue
		}

		prevTime := container.objectQueue[i-1].GetEndTime()
		prevPos := container.objectQueue[i-1].GetStackedEndPositionMod(container.beatMap.Diff).Copy64()

		nextTime := container.objectQueue[i].GetStartTime()
		nextPos := container.objectQueue[i].GetStackedStartPositionMod(container.beatMap.Diff).Copy64()

		duration := nextTime - prevTime

		vec := nextPos.Sub(prevPos)
		distance := vec.Len()
		rotation := vec.AngleR()

		for progress := max(lineDist*1.5, distance-5_000); progress < distance-lineDist; progress += lineDist { // Limit the points to 5k osu!pixels from the next object
			t := progress / distance

			tStart := prevTime + t*duration - preEmpt
			tEnd := prevTime + t*duration

			pStart := prevPos.Add(vec.Scl(t - 0.1))
			pos := prevPos.Add(vec.Scl(t))

			followPoint := sprite.NewAnimation(textures, skin.GetInfo().GetFrameTime(len(textures)), true, -float64(i), pos, vector.Centre)
			followPoint.SetRotation(rotation)
			followPoint.SetAlpha(0)
			followPoint.ShowForever(false)

			if skin.GetInfo().DefaultSkinFollowpointBehavior {
				followPoint.AddTransform(animation.NewVectorTransformV(animation.Move, easing.OutQuad, tStart, tStart+container.beatMap.Diff.TimeFadeIn, pStart, pos))
				followPoint.AddTransform(animation.NewSingleTransform(animation.Scale, easing.OutQuad, tStart, tStart+container.beatMap.Diff.TimeFadeIn, 1.5, 1))
			}

			followPoint.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, tStart, tStart+container.beatMap.Diff.TimeFadeIn, 0, 1))
			followPoint.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, tEnd, tEnd+container.beatMap.Diff.TimeFadeIn, 1, 0))
			followPoint.AdjustTimesToTransformations()

			container.spriteManager.Add(followPoint)
		}
	}
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
	container.spriteManager.Update(time)

	if time > 0 {
		delta := time - container.lastTime

		settings.Objects.Colors.Color.Update(delta)
		settings.Objects.Colors.Sliders.Border.Color.Update(delta)
		settings.Objects.Colors.Sliders.Body.Color.Update(delta)

		container.lastTime = time
	}
}

func (container *HitObjectContainer) preProcessQueue(time float64) {
	if len(container.objectQueue) > 0 {
		for i := 0; i < len(container.objectQueue); i++ {
			if p := container.objectQueue[i]; p.GetStartTime()-max(15000, container.beatMap.Diff.Preempt) <= time {
				if p := container.objectQueue[i]; p.GetStartTime()-math.Floor(container.beatMap.Diff.Preempt) <= time {
					if _, ok := p.(*objects.Spinner); ok {
						container.addProxy(&renderableProxy{
							renderable:   p.(objects.Renderable),
							isSliderBody: false,
							depth:        math.MaxFloat64,
							endTime:      p.GetEndTime() + difficulty.HitFadeOut,
						})
					} else {
						container.addProxy(&renderableProxy{
							renderable:   p.(objects.Renderable),
							isSliderBody: false,
							depth:        p.GetStartTime(),
							endTime:      p.GetEndTime() + float64(container.beatMap.Diff.Hit50) + difficulty.HitFadeOut,
						})
					}

					if _, ok := p.(*objects.Slider); ok {
						container.addProxy(&renderableProxy{
							renderable:   p.(objects.Renderable),
							isSliderBody: true,
							depth:        p.GetEndTime() + 10,
							endTime:      p.GetEndTime() + difficulty.HitFadeOut,
						})
					}

					container.countProcessed++

					container.objectQueue = container.objectQueue[1:]
					i--
				}
			} else {
				break
			}
		}
	}
}

func (container *HitObjectContainer) Draw(batch *batch.QuadBatch, baseCamera mgl32.Mat4, cameras []mgl32.Mat4, time float64, scale, alpha float32) {
	profiler.StartGroup("HitObjectContainer.Draw", profiler.PDraw)

	divides := len(cameras)

	container.preProcessQueue(time)

	if settings.Playfield.DrawObjects {
		objectColors := settings.Objects.Colors.Color.GetColors(divides, float64(scale), float64(alpha))
		borderColors := objectColors
		bodyColors := objectColors

		if !settings.Objects.Colors.Sliders.Border.UseHitCircleColor {
			borderColors = settings.Objects.Colors.Sliders.Border.Color.GetColors(divides, float64(scale), float64(alpha))
		}

		if !settings.Objects.Colors.Sliders.Body.UseHitCircleColor {
			bodyColors = settings.Objects.Colors.Sliders.Body.Color.GetColors(divides, float64(scale), float64(alpha))
		}

		if !settings.Objects.ScaleToTheBeat {
			scale = 1
		}

		batch.Begin()
		batch.ResetTransform()
		batch.SetColor(1, 1, 1, float64(alpha))
		batch.SetScale(float64(scale)*container.beatMap.Diff.CircleRadius/64, float64(scale)*container.beatMap.Diff.CircleRadius/64)

		if divides < settings.Objects.Colors.MandalaTexturesTrigger && settings.Objects.DrawFollowPoints {
			for i := 0; i < divides; i++ {
				batch.SetCamera(cameras[i])
				container.spriteManager.Draw(time, batch)
			}
		}

		batch.Flush()
		batch.SetColor(1, 1, 1, 1)
		batch.SetScale(1, 1)

		profiler.StartGroup("HitObjectContainer.SliderPreDraw", profiler.PDraw)
		for i := len(container.renderables) - 1; i >= 0; i-- {
			if s, ok := container.renderables[i].renderable.(*objects.Slider); ok && container.renderables[i].isSliderBody {
				s.DrawBodyBase(time, baseCamera)
			}
		}
		profiler.EndGroup()

		slidersRendered := false

		if settings.Objects.Sliders.SliderMerge {
			enabled := false

			for j := 0; j < divides; j++ {
				ind := j - 1
				if ind < 0 {
					ind = divides - 1
				}

				for i := len(container.renderables) - 1; i >= 0; i-- {
					if s, ok := container.renderables[i].renderable.(*objects.Slider); ok && container.renderables[i].isSliderBody {
						if !enabled {
							enabled = true
							profiler.StartGroup("HitObjectContainer.SliderDraw", profiler.PDraw)
							sliderrenderer.BeginRendererMerge()
						}

						slidersRendered = true
						s.DrawBody(time, bodyColors[j], borderColors[j], borderColors[ind], cameras[j], scale)
					}
				}
			}

			if enabled {
				sliderrenderer.EndRendererMerge()
				profiler.EndGroup()
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

				if !proxy.isSliderBody {
					if enabled && !settings.Objects.Sliders.SliderMerge {
						enabled = false
						sliderrenderer.EndRenderer()
						profiler.EndGroup()
					}

					_, sp := container.renderables[i].renderable.(*objects.Spinner)
					if !sp || j == 0 {
						proxy.renderable.Draw(time, objectColors[j], batch)
					}
				} else if !settings.Objects.Sliders.SliderMerge {
					if !enabled {
						enabled = true

						batch.Flush()

						profiler.StartGroup("HitObjectContainer.SliderDraw", profiler.PDraw)
						sliderrenderer.BeginRenderer()
					}

					slidersRendered = true
					proxy.renderable.(*objects.Slider).DrawBody(time, bodyColors[j], borderColors[j], borderColors[ind], cameras[j], scale)
				}

				if proxy.endTime <= time {
					if !proxy.isSliderBody {
						container.countProcessed--
					}

					container.renderables = append(container.renderables[:i], container.renderables[(i+1):]...)
				}
			}

			if enabled {
				sliderrenderer.EndRenderer()
				profiler.EndGroup()
			}
		}

		if !slidersRendered { //stub to have it on the graph
			profiler.StartGroup("HitObjectContainer.SliderDraw", profiler.PDraw)
			profiler.EndGroup()
		}

		if divides < settings.Objects.Colors.MandalaTexturesTrigger && settings.Objects.DrawApproachCircles {
			for j := 0; j < divides; j++ {
				batch.SetCamera(cameras[j])

				for i := len(container.renderables) - 1; i >= 0; i-- {
					if s := container.renderables[i]; !s.isSliderBody {
						s.renderable.DrawApproach(time, objectColors[j], batch)
					}
				}
			}
		}

		batch.SetAdditive(false)
		batch.SetScale(1, 1)
		batch.End()
	}

	profiler.EndGroup()
}

func (container *HitObjectContainer) GetNumProcessed() int {
	return container.countProcessed
}
