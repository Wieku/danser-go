package play

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/api"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shape"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type StrainGraph struct {
	shapeRenderer *shape.Renderer
	strains       api.StrainPeaks
	baseLine      float64
	maxStrain     float32

	countFromZero bool
	countTrueEnd  bool

	startTime float64
	endTime   float64

	trueStartTime float64
	trueEndTime   float64

	strainStartTime float64
	strainEndTime   float64
	strainLength    float64

	startProgress float64
	endProgress   float64

	fbo *buffer.Framebuffer

	leftSprite   *sprite.Sprite
	centerSprite *sprite.Sprite
	rightSprite  *sprite.Sprite

	screenWidth float64

	size          vector.Vector2d
	drawOutline   bool
	outlineWidth  float64
	innerOpacity  float64
	innerDarkness float64
}

func NewStrainGraph(beatMap *beatmap.BeatMap, peaks api.StrainPeaks, countFromZero, countTrueEnd bool) *StrainGraph {
	graph := &StrainGraph{
		shapeRenderer: shape.NewRenderer(),
		strains:       peaks,

		trueStartTime: beatMap.HitObjects[min(0, len(beatMap.HitObjects)-1)].GetStartTime(),
		trueEndTime:   beatMap.HitObjects[len(beatMap.HitObjects)-1].GetEndTime(),

		strainStartTime: beatMap.HitObjects[min(1, len(beatMap.HitObjects)-1)].GetStartTime(),
		strainEndTime:   beatMap.HitObjects[len(beatMap.HitObjects)-1].GetStartTime(),

		screenWidth:   768 * settings.Graphics.GetAspectRatio(),
		countFromZero: countFromZero,
		countTrueEnd:  countTrueEnd,
	}

	graph.strainLength = graph.strainEndTime - graph.strainStartTime

	graph.startTime = graph.trueStartTime
	if countFromZero {
		graph.startTime = min(graph.startTime, 0)
	}

	graph.endTime = graph.strainEndTime
	if countTrueEnd {
		graph.endTime = graph.trueEndTime
	}

	// Those magic numbers are derived from sr formula with all difficulty values being 0 (e.g. at breaks)
	graph.baseLine = 0.1401973407499798
	if beatMap.Diff.CheckModActive(difficulty.Flashlight) {
		graph.baseLine = 0.14386309174146011
	}

	graph.leftSprite = sprite.NewSpriteSingle(nil, 0, vector.NewVec2d(graph.screenWidth, 728), vector.TopLeft)
	graph.leftSprite.SetCutOrigin(vector.CentreLeft)

	graph.centerSprite = sprite.NewSpriteSingle(nil, 0, vector.NewVec2d(graph.screenWidth, 728), vector.TopLeft)
	graph.centerSprite.SetCutOrigin(vector.CentreLeft)

	graph.rightSprite = sprite.NewSpriteSingle(nil, 0, vector.NewVec2d(graph.screenWidth, 728), vector.TopRight)
	graph.rightSprite.SetCutOrigin(vector.CentreRight)

	return graph
}

func (graph *StrainGraph) SetTimes(start, end float64) {
	graph.startProgress = mutils.Clamp((start-graph.startTime)/max(graph.endTime-graph.startTime, 1), 0, 1)
	graph.endProgress = mutils.Clamp((end-graph.startTime)/max(graph.endTime-graph.startTime, 1), 0, 1)

	graph.leftSprite.SetCutX(0, 1-graph.startProgress)

	graph.centerSprite.SetCutOrigin(vector.CentreLeft.AddS(graph.startProgress*2, 0))
	graph.centerSprite.SetCutX(1, 1-(graph.endProgress-graph.startProgress)/(1-graph.startProgress))

	graph.rightSprite.SetCutX(graph.endProgress, 0)
}

func (graph *StrainGraph) generateCurve() curves.Curve {
	if len(graph.strains.Total) == 0 {
		graph.maxStrain = 1
		return curves.NewLinear(vector.NewVec2f(0, 0), vector.NewVec2f(1, 0))
	}

	// Number of strain sections to merge
	// For example for a 5-minute map we will get 10 sections, so 4s because one section is 400ms
	// It's also scaled with width of the strain graph so wider one shows more detailed graph
	sectSize := max(int((graph.endTime-graph.startTime)/30000*(200/graph.size.X)), 1)

	var points []vector.Vector2f

	if graph.countFromZero && graph.trueStartTime > 0 { // Don't add intro if map starts before music
		points = append(points, vector.NewVec2f(0, 0), vector.NewVec2f(float32(max(0, graph.trueStartTime-400)), 0))
	}

	points = append(points, vector.NewVec2f(float32(graph.trueStartTime-0.001), float32(graph.strains.Total[0]-graph.baseLine))) //slight nudge to the left in case it's a 1 object map

	sections := (len(graph.strains.Total) - 1) / sectSize

	for i := 0; i <= sections; i++ {
		bI := i * sectSize

		maxI := min(len(graph.strains.Total), bI+sectSize)

		var lMaxStrain float32
		for ; bI < maxI; bI++ {
			lMaxStrain = max(lMaxStrain, float32(graph.strains.Total[bI]-graph.baseLine))
		}

		graph.maxStrain = max(graph.maxStrain, lMaxStrain)
		points = append(points, vector.NewVec2f(float32(graph.strainStartTime+(float64(i)/float64(max(1, sections)))*graph.strainLength), lMaxStrain))
	}

	if graph.countTrueEnd && graph.trueEndTime > graph.strainEndTime {
		points = append(points, vector.NewVec2f(float32(graph.trueEndTime), float32(graph.strains.Total[len(graph.strains.Total)-1]-graph.baseLine)))
	}

	return curves.NewMonotoneCubic(points)
}

func (graph *StrainGraph) drawFBO(batch *batch.QuadBatch) {
	const step float32 = 0.5

	splinePoint := func(spline curves.Curve, x float32) float32 {
		y := max(spline.PointAt(x).Y, 0) / graph.maxStrain
		return y * y
	}

	upscale := settings.Graphics.GetHeightF() / 768

	conf := settings.Gameplay.StrainGraph

	graph.size = vector.NewVec2d(conf.Width, conf.Height)
	graph.drawOutline = conf.Outline.Show
	graph.outlineWidth = conf.Outline.Width
	graph.innerOpacity = conf.Outline.InnerOpacity
	graph.innerDarkness = conf.Outline.InnerDarkness

	oWidth := float32(graph.outlineWidth * upscale)
	yOffset := float32(2 * upscale)

	if graph.fbo != nil {
		graph.fbo.Dispose()
	}

	fboWidth := float32(math.Round(graph.size.X * upscale))
	fboHeight := float32(math.Round(graph.size.Y * upscale))

	graph.fbo = buffer.NewFrameMultisample(int(fboWidth), int(fboHeight), 8)

	graph.fbo.Bind()
	graph.fbo.ClearColor(0, 0, 0, 0)

	graph.shapeRenderer.SetCamera(mgl32.Ortho2D(0, fboWidth, fboHeight, 0))

	viewport.Push(int(fboWidth), int(fboHeight))

	graph.shapeRenderer.Begin()

	if graph.drawOutline {
		graph.shapeRenderer.SetColor(1-graph.innerDarkness, 1-graph.innerDarkness, 1-graph.innerDarkness, graph.innerOpacity)

		fboHeight -= oWidth / 2
		yOffset = max(yOffset, oWidth/2)
	} else {
		graph.shapeRenderer.SetColor(1, 1, 1, 1)
	}

	spline := graph.generateCurve()

	strainScale := fboHeight - yOffset

	pY1 := splinePoint(spline, 0)*strainScale + yOffset

	for pX := step; pX <= fboWidth; pX += step {
		pY2 := splinePoint(spline, pX/fboWidth)*strainScale + yOffset

		graph.shapeRenderer.DrawQuad(pX-step, 0, pX-step, pY1, pX, pY2, pX, 0)

		pY1 = pY2
	}

	if graph.drawOutline {
		graph.shapeRenderer.SetColor(1, 1, 1, 1)

		pY1 = splinePoint(spline, 0)*strainScale + yOffset

		graph.shapeRenderer.DrawCircle(vector.NewVec2f(0, pY1), oWidth/2)

		for pX := step; pX <= fboWidth; pX += step {
			pY2 := splinePoint(spline, pX/fboWidth)*strainScale + yOffset

			graph.shapeRenderer.DrawLine(pX-step, pY1, pX, pY2, oWidth)
			graph.shapeRenderer.DrawCircle(vector.NewVec2f(pX, pY2), oWidth/2)

			pY1 = pY2
		}
	}

	graph.shapeRenderer.End()

	graph.fbo.Unbind()

	viewport.Pop()

	batch.ResetTransform()

	region := graph.fbo.Texture().GetRegion()

	// Reestablish scaling using final FBO sizes because 768/screenHeight was causing 1px gaps in some scenarios
	graph.leftSprite.SetScaleV(vector.NewVec2d(graph.size.X/float64(region.Width), graph.size.Y/float64(region.Height)))
	graph.centerSprite.SetScaleV(vector.NewVec2d(graph.size.X/float64(region.Width), graph.size.Y/float64(region.Height)))
	graph.rightSprite.SetScaleV(vector.NewVec2d(graph.size.X/float64(region.Width), graph.size.Y/float64(region.Height)))

	graph.leftSprite.Texture = &region
	graph.centerSprite.Texture = &region
	graph.rightSprite.Texture = &region
}

func (graph *StrainGraph) Draw(batch *batch.QuadBatch, alpha float64) {
	conf := settings.Gameplay.StrainGraph

	sgAlpha := conf.Opacity * alpha

	if sgAlpha < 0.001 || !conf.Show {
		return
	}

	if graph.fbo == nil || graph.size.X != conf.Width || graph.size.Y != conf.Height ||
		graph.drawOutline != conf.Outline.Show || graph.outlineWidth != conf.Outline.Width ||
		graph.innerDarkness != conf.Outline.InnerDarkness || graph.innerOpacity != conf.Outline.InnerOpacity {
		batch.Flush()
		graph.drawFBO(batch)
	}

	batch.ResetTransform()

	batch.SetColor(1, 1, 1, sgAlpha)

	origin := vector.ParseOrigin(conf.Align).AddS(1, 1).Scl(0.5)
	basePos := vector.NewVec2d(conf.XPosition, conf.YPosition)

	pos1 := basePos.Sub(origin.Mult(graph.size))
	pos2 := pos1.AddS(graph.startProgress*graph.size.X, 0)
	pos3 := pos1.AddS(graph.size.X, 0)

	graph.leftSprite.SetPosition(pos1)
	graph.centerSprite.SetPosition(pos2)
	graph.rightSprite.SetPosition(pos3)

	bgParsed := color.NewHSV(float32(conf.BgColor.Hue), float32(conf.BgColor.Saturation), float32(conf.BgColor.Value))
	fgParsed := color.NewHSV(float32(conf.FgColor.Hue), float32(conf.FgColor.Saturation), float32(conf.FgColor.Value))

	graph.leftSprite.SetColor(bgParsed)
	graph.centerSprite.SetColor(fgParsed)
	graph.rightSprite.SetColor(bgParsed)

	graph.leftSprite.Draw(0, batch)
	graph.centerSprite.Draw(0, batch)
	graph.rightSprite.Draw(0, batch)

	batch.ResetTransform()
}
