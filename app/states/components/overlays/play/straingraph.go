package play

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/rulesets/osu/performance"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/buffer"
	"github.com/wieku/danser-go/framework/graphics/shape"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/viewport"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/curves"
	"github.com/wieku/danser-go/framework/math/math32"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type StrainGraph struct {
	shapeRenderer *shape.Renderer
	strains       performance.StrainPeaks
	maxStrain     float32
	time          float64

	startTime   float64
	endTime     float64
	progress    float64
	fbo         *buffer.Framebuffer
	leftSprite  *sprite.Sprite
	rightSprite *sprite.Sprite

	screenWidth float64

	size vector.Vector2d
}

func NewStrainGraph(ruleset *osu.OsuRuleSet) *StrainGraph {
	graph := &StrainGraph{
		shapeRenderer: shape.NewRenderer(),
		strains:       performance.CalculateStrainPeaks(ruleset.GetBeatMap().HitObjects, ruleset.GetBeatMap().Diff, settings.Gameplay.UseLazerPP),
		startTime:     ruleset.GetBeatMap().HitObjects[mutils.Min(1, len(ruleset.GetBeatMap().HitObjects)-1)].GetStartTime(),
		endTime:       ruleset.GetBeatMap().HitObjects[len(ruleset.GetBeatMap().HitObjects)-1].GetStartTime(),
		screenWidth:   768 * settings.Graphics.GetAspectRatio(),
	}

	graph.leftSprite = sprite.NewSpriteSingle(nil, 0, vector.NewVec2d(graph.screenWidth, 728), vector.TopLeft)
	graph.leftSprite.SetColor(color.NewIRGB(231, 141, 235))
	graph.leftSprite.SetCutOrigin(vector.CentreLeft)

	graph.rightSprite = sprite.NewSpriteSingle(nil, 0, vector.NewVec2d(graph.screenWidth, 728), vector.TopRight)
	graph.rightSprite.SetColor(color.NewL(0.2))
	graph.rightSprite.SetCutOrigin(vector.CentreRight)

	return graph
}

func (graph *StrainGraph) Update(time float64) {
	graph.time = time
	graph.progress = mutils.ClampF((time-graph.startTime)/(graph.endTime-graph.startTime), 0, 1)
	graph.leftSprite.SetCutX(1 - graph.progress)
	graph.rightSprite.SetCutX(graph.progress)
}

func (graph *StrainGraph) generateCurve() curves.Curve {
	// Number of strain sections to merge
	// For example for a 5-minute map we will get 10 sections, so 4s because one section is 400ms
	// It's also scaled with width of the strain graph so wider one shows more detailed graph
	sectSize := mutils.Max(int((graph.endTime-graph.startTime)/30000*(200/graph.size.X)), 1)

	toM := []vector.Vector2f{vector.NewVec2f(0, 0)}

	for i := 0; i < len(graph.strains.Total); i += sectSize {
		maxI := mutils.Min(len(graph.strains.Total), i+sectSize)

		max := 0.0

		for j := i; j < maxI; j++ {
			max = math.Max(max, graph.strains.Total[j])
		}

		graph.maxStrain = math32.Max(graph.maxStrain, float32(max))
		toM = append(toM, vector.NewVec2f(float32(i/sectSize)+0.5, float32(max)))
	}

	toM = append(toM, vector.NewVec2f(float32(len(toM)-1), 0))

	return curves.NewMonotoneCubic(toM)
}

func (graph *StrainGraph) drawFBO(batch *batch.QuadBatch) {
	batch.Flush()

	graph.size = vector.NewVec2d(settings.Gameplay.StrainGraph.Width, settings.Gameplay.StrainGraph.Height)

	w := graph.size.X * settings.Graphics.GetHeightF() / 768
	h := graph.size.Y * settings.Graphics.GetHeightF() / 768

	if graph.fbo != nil {
		graph.fbo.Dispose()
	}

	graph.fbo = buffer.NewFrameMultisample(int(math.Round(w)), int(math.Round(h)), 8)

	graph.fbo.Bind()
	graph.fbo.ClearColor(1, 1, 1, 0)

	graph.shapeRenderer.SetCamera(mgl32.Ortho2D(0, float32(graph.fbo.GetWidth()), float32(graph.fbo.GetHeight()), 0))

	viewport.Push(graph.fbo.GetWidth(), graph.fbo.GetHeight())

	graph.shapeRenderer.Begin()
	graph.shapeRenderer.SetColor(1, 1, 1, 1)

	lWidth := float32(graph.fbo.GetWidth())
	lHeight := float32(graph.fbo.GetHeight()) - 1

	spline := graph.generateCurve()

	lV := math32.Max(spline.PointAt(0).X, 0)

	step := float32(0.5)

	for i := step; i <= lWidth; i += step {
		v := math32.Max(spline.PointAt(i/lWidth).Y, 0)

		pX := i
		pY1 := lV / graph.maxStrain * lHeight
		pY2 := v / graph.maxStrain * lHeight

		lV = v

		graph.shapeRenderer.DrawQuad(pX-step, 0, pX-step, pY1, pX, pY2, pX, 0)
	}

	graph.shapeRenderer.End()

	graph.fbo.Unbind()

	viewport.Pop()

	batch.ResetTransform()

	region := graph.fbo.Texture().GetRegion()

	// Reestablish scaling using final FBO sizes because 768/screenHeight was causing 1px gaps in some scenarios
	graph.leftSprite.SetScaleV(vector.NewVec2d(graph.size.X/float64(region.Width), graph.size.Y/float64(region.Height)))
	graph.rightSprite.SetScaleV(vector.NewVec2d(graph.size.X/float64(region.Width), graph.size.Y/float64(region.Height)))

	graph.leftSprite.Texture = &region
	graph.rightSprite.Texture = &region
}

func (graph *StrainGraph) Draw(batch *batch.QuadBatch, alpha float64) {
	conf := settings.Gameplay.StrainGraph

	sgAlpha := conf.Opacity * alpha

	if sgAlpha < 0.001 || !conf.Show {
		return
	}

	if graph.fbo == nil || graph.size.X != conf.Width || graph.size.Y != conf.Height {
		graph.drawFBO(batch)
	}

	batch.ResetTransform()

	batch.SetColor(1, 1, 1, sgAlpha)

	origin := vector.ParseOrigin(conf.Align).AddS(1, 1).Scl(0.5)
	basePos := vector.NewVec2d(conf.XPosition, conf.YPosition)

	pos1 := basePos.Sub(origin.Mult(graph.size))
	pos2 := pos1.AddS(graph.size.X, 0)

	graph.leftSprite.SetPosition(pos1)
	graph.rightSprite.SetPosition(pos2)

	graph.leftSprite.SetColor(color.NewHSV(float32(conf.FgColor.Hue), float32(conf.FgColor.Saturation), float32(conf.FgColor.Value)))
	graph.rightSprite.SetColor(color.NewHSV(float32(conf.BgColor.Hue), float32(conf.BgColor.Saturation), float32(conf.BgColor.Value)))
	graph.leftSprite.Draw(0, batch)
	graph.rightSprite.Draw(0, batch)

	batch.ResetTransform()
}
