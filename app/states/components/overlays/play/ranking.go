package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/shape"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	"github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/scaling"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	img1  = 40 / 0.625
	img2  = 240 / 0.625
	text1 = 80 / 0.625
	text2 = 280 / 0.625
	row1  = 160 / 0.625
	row2  = 220 / 0.625
	row3  = 280 / 0.625
	row4  = 320 / 0.625
)

type RankingPanel struct {
	manager *sprite.Manager
	time    float64

	ScaledWidth float64
	cursor      *graphics.Cursor
	ruleset     *osu.OsuRuleSet
	count300    *HitCounter
	count100    *HitCounter
	count50     *HitCounter
	countGeki   *HitCounter
	countKatu   *HitCounter
	countMiss   *HitCounter
	score       string
	maxCombo    string
	accuracy    string
	pp          string

	beatmapName    string
	beatmapCreator string
	playedBy       string
	gradeS         *sprite.Sprite
	hpSections     []vector.Vector2d
	shapeRenderer  *shape.Renderer
	hpGraph        []vector.Vector2d
	stats          []string
	perfect        *sprite.Sprite
}

func NewRankingPanel(cursor *graphics.Cursor, ruleset *osu.OsuRuleSet, hitError *HitErrorMeter, hpGraph []vector.Vector2d) *RankingPanel {
	panel := &RankingPanel{
		manager:     sprite.NewManager(),
		ScaledWidth: settings.Graphics.GetAspectRatio() * 768,
		cursor:      cursor,
		ruleset:     ruleset,
	}

	bg := sprite.NewSpriteSingle(nil, -1, vector.NewVec2d(panel.ScaledWidth, 768).Scl(0.5), vector.Centre)
	bg.SetColor(color.NewL(0.75))

	bgLoadFunc := func() {
		image, err := texture.NewPixmapFileString(filepath.Join(settings.General.GetSongsDir(), ruleset.GetBeatMap().Dir, ruleset.GetBeatMap().Bg))
		if err != nil {
			image, err = assets.GetPixmap("assets/textures/background-1.png")
			if err != nil {
				panic(err)
			}
		}

		if image != nil {
			goroutines.CallNonBlockMain(func() {
				region := texture.LoadTextureSingle(image.RGBA(), 0).GetRegion()
				bg.Texture = &region

				result := scaling.Fill.Apply(region.Width, region.Height, float32(panel.ScaledWidth), 768)

				bg.SetScaleV(result.Mult(vector.NewVec2f(1/region.Width, 1/region.Height)).Copy64())

				image.Dispose()
			})
		}
	}

	if settings.RECORD {
		bgLoadFunc()
	} else {
		go bgLoadFunc()
	}

	panel.loadMods()

	/*
		v1.0: (0,74)
		v2.0+: (0,102)
	*/
	var rPPos vector.Vector2d
	var rAPos vector.Vector2d
	var rCPos vector.Vector2d
	var rGPos vector.Vector2d
	var rRPos vector.Vector2d

	if skin.GetInfo().Version >= 2 {
		rPPos = vector.NewVec2d(0, 102)
		rAPos = vector.NewVec2d(291, 480)
		rCPos = vector.NewVec2d(8, 480)
		rGPos = vector.NewVec2d(256, 608)
		rRPos = vector.NewVec2d(panel.ScaledWidth-192, 320)
	} else {
		rPPos = vector.NewVec2d(0, 74)
		rAPos = vector.NewVec2d(291, 500)
		rCPos = vector.NewVec2d(8, 500)
		rGPos = vector.NewVec2d(256, 576)
		rRPos = vector.NewVec2d(panel.ScaledWidth-192, 272)
	}

	rPanel := sprite.NewSpriteSingle(skin.GetTexture("ranking-panel"), 1, rPPos, vector.TopLeft)
	rGraph := sprite.NewSpriteSingle(skin.GetTexture("ranking-graph"), 2, rGPos, vector.TopLeft)
	rAcc := sprite.NewSpriteSingle(skin.GetTexture("ranking-accuracy"), 3, rAPos, vector.TopLeft)
	rCombo := sprite.NewSpriteSingle(skin.GetTexture("ranking-maxcombo"), 4, rCPos, vector.TopLeft)

	score := panel.ruleset.GetScore(panel.cursor)

	panel.pp = fmt.Sprintf("%."+strconv.Itoa(settings.Gameplay.PPCounter.Decimals)+"fpp", score.PP.Total)

	panel.gradeS = sprite.NewSpriteSingle(skin.GetTexture("ranking-"+score.Grade.TextureName()), 5, rRPos, vector.Centre)

	p := graphics.Pixel.GetRegion()
	rTop := sprite.NewSpriteSingle(&p, 999, vector.NewVec2d(0, 0), vector.TopLeft)
	rTop.SetScaleV(vector.NewVec2d(panel.ScaledWidth, 96))
	rTop.SetColor(color.NewL(0))
	rTop.SetAlpha(0.8)

	rTitle := sprite.NewSpriteSingle(skin.GetTexture("ranking-title"), 1000, vector.NewVec2d(panel.ScaledWidth-32, 0), vector.TopRight)

	panel.manager.Add(bg)
	panel.manager.Add(rPanel)
	panel.manager.Add(rGraph)
	panel.manager.Add(rAcc)
	panel.manager.Add(rCombo)
	panel.manager.Add(panel.gradeS)
	panel.manager.Add(rTop)
	panel.manager.Add(rTitle)

	panel.count300 = NewHitCounter("hit300", fmt.Sprintf("%dx", score.Count300), vector.NewVec2d(img1, row1))
	panel.count100 = NewHitCounter("hit100", fmt.Sprintf("%dx", score.Count100), vector.NewVec2d(img1, row2))
	panel.count50 = NewHitCounter("hit50", fmt.Sprintf("%dx", score.Count50), vector.NewVec2d(img1, row3))
	panel.countGeki = NewHitCounter("hit300g", fmt.Sprintf("%dx", score.CountGeki), vector.NewVec2d(img2, row1))
	panel.countKatu = NewHitCounter("hit100k", fmt.Sprintf("%dx", score.CountKatu), vector.NewVec2d(img2, row2))
	panel.countMiss = NewHitCounter("hit0", fmt.Sprintf("%dx", score.CountMiss), vector.NewVec2d(img2, row3))

	panel.manager.Add(panel.count300)
	panel.manager.Add(panel.count100)
	panel.manager.Add(panel.count50)
	panel.manager.Add(panel.countGeki)
	panel.manager.Add(panel.countKatu)
	panel.manager.Add(panel.countMiss)

	bMap := ruleset.GetBeatMap()

	panel.beatmapName = fmt.Sprintf("%s - %s [%s]", bMap.Artist, bMap.Name, bMap.Difficulty)
	panel.beatmapCreator = fmt.Sprintf("Beatmap by %s", bMap.Creator)

	scoreTime := panel.cursor.ScoreTime
	if settings.Gameplay.ResultsUseLocalTimeZone {
		scoreTime = scoreTime.Local()
	}

	panel.playedBy = fmt.Sprintf("Played by %s on %s", panel.cursor.Name, scoreTime.Format("2006-01-02 15:04:05 MST"))

	panel.score = fmt.Sprintf("%08d", score.Score)
	panel.maxCombo = fmt.Sprintf("%dx", score.Combo)
	panel.accuracy = fmt.Sprintf("%.2f%%", score.Accuracy*100)

	panel.hpGraph = make([]vector.Vector2d, len(hpGraph))
	copy(panel.hpGraph, hpGraph)

	for len(panel.hpGraph) > 100 {
		for i := len(panel.hpGraph) - 1; i >= 0; i -= 2 {
			panel.hpGraph = append(panel.hpGraph[:i], panel.hpGraph[i+1:]...)
		}
	}

	if score.PerfectCombo {
		log.Println("PERFECT")
		pPos := vector.NewVec2d(320, 688)
		if skin.GetInfo().Version >= 2 {
			pPos = vector.NewVec2d(416, 688)
		}

		panel.perfect = sprite.NewSpriteSingle(skin.GetTexture("ranking-perfect"), 0, pPos, vector.Centre)
	}

	stats := fmt.Sprintf("Slider ticks: %d/%d", int64(score.MaxTicks)-int64(score.CountSB), score.MaxTicks)
	stats += fmt.Sprintf("\nSlider ends: %d/%d", score.SliderEnd, score.MaxSliderEnd)

	stats += "\nAccuracy:\n"
	stats += fmt.Sprintf("Error: %.2fms - %.2fms avg", hitError.GetAvgNeg(), hitError.GetAvgPos())

	if panel.ruleset.GetBeatMap().Diff.Speed != 1.0 {
		stats += fmt.Sprintf("\n            (%.2fms - %.2fms)", hitError.GetAvgNegConverted(), hitError.GetAvgPosConverted())
	}

	stats += "\n"
	stats += fmt.Sprintf("Unstable Rate: %.2f", hitError.GetUnstableRate())

	if panel.ruleset.GetBeatMap().Diff.Speed != 1.0 {
		stats += fmt.Sprintf("\n                             (%.2f)", hitError.GetUnstableRateConverted())
	}

	panel.stats = strings.Split(stats, "\n")

	return panel
}

func (panel *RankingPanel) loadMods() {
	mods := panel.ruleset.GetBeatMap().Diff.Mods.StringFull()

	offset := -64.0
	for i, s := range mods {
		modSpriteName := "selection-mod-" + strings.ToLower(s)

		mod := sprite.NewSpriteSingle(skin.GetTexture(modSpriteName), 6+float64(i), vector.NewVec2d(panel.ScaledWidth+offset, 416), vector.Centre)

		panel.manager.Add(mod)

		offset -= 32
	}
}

func (panel *RankingPanel) Add(time, hp float64) {
	panel.hpSections = append(panel.hpSections, vector.NewVec2d(time, hp))
}

func (panel *RankingPanel) Update(time float64) {
	panel.time = time
	panel.manager.Update(time)

	if panel.perfect != nil {
		panel.perfect.Update(time)
	}
}

func (panel *RankingPanel) Draw(batch *batch.QuadBatch, alpha float64) {
	batch.SetColor(1, 1, 1, alpha)
	batch.ResetTransform()

	panel.manager.Draw(panel.time, batch)

	fnt := skin.GetFont("score")

	prevOverlap := fnt.Overlap
	fnt.Overlap = -2
	fnt.DrawOrigin(batch, 220/0.625, 94/0.625, vector.Centre, fnt.GetSize()*1.3, true, panel.score)

	fnt.Overlap = prevOverlap

	fnt.DrawOrigin(batch, text1-65/0.625, row4+10/0.625, vector.TopLeft, fnt.GetSize()*1.12, false, panel.maxCombo)
	fnt.DrawOrigin(batch, text2-86/0.625, row4+10/0.625, vector.TopLeft, fnt.GetSize()*1.12, false, panel.accuracy)

	ubuFont := font.GetFont("Ubuntu Regular")
	fnt2 := font.GetFont("SBFont")

	fnt2.Overlap = 0

	if fnt2 == ubuFont {
		fnt2.Overlap = 0.7
	}

	fnt2.Draw(batch, 5, 30-3, 30, panel.beatmapName)

	if fnt2 == ubuFont {
		fnt2.Overlap = 1
	}

	fnt2.Draw(batch, 5, 30+22, 22, panel.beatmapCreator)

	fnt2.Overlap = 0

	fnt2.Draw(batch, 5, 30+22+22, 22, panel.playedBy)

	if settings.Gameplay.PPCounter.ShowInResults {
		fnt3 := font.GetFont("HUDFont")

		batch.SetColor(0, 0, 0, alpha*0.5)
		fnt3.DrawOrigin(batch, panel.ScaledWidth-204, 576+62, vector.Centre, 61, false, panel.pp)
		batch.SetColor(1, 1, 1, alpha)
		fnt3.DrawOrigin(batch, panel.ScaledWidth-205, 576+61, vector.Centre, 61, false, panel.pp)
	}

	if panel.shapeRenderer == nil {
		panel.shapeRenderer = shape.NewRenderer()
	}

	batch.Flush()

	panel.shapeRenderer.SetCamera(batch.Projection)
	panel.shapeRenderer.Begin()

	begin := panel.hpGraph[0].X
	end := panel.hpGraph[len(panel.hpGraph)-1].X

	for i := 0; i < len(panel.hpGraph)-1; i++ {
		p1 := panel.hpGraph[i]
		p2 := panel.hpGraph[i+1]

		p1X := 256 + 8 + 298*(p1.X-begin)/(end-begin)
		p1Y := 608 + 8 + 137.6*(1-p1.Y)
		p2X := 256 + 8 + 298*(p2.X-begin)/(end-begin)
		p2Y := 608 + 8 + 137.6*(1-p2.Y)

		meanHp := (p1.Y + p2.Y) / 2

		if meanHp > 0.5 {
			panel.shapeRenderer.SetColor(0.2, 1, 0.2, alpha)
		} else {
			panel.shapeRenderer.SetColor(1, 0.2, 0.2, alpha)
		}

		panel.shapeRenderer.DrawLine(float32(p1X), float32(p1Y), float32(p2X), float32(p2Y), 2)
	}

	panel.shapeRenderer.End()

	if panel.perfect != nil {
		batch.ResetTransform()
		panel.perfect.Draw(panel.time, batch)
		batch.Flush()
	}

	panel.shapeRenderer.Begin()

	panel.shapeRenderer.SetColor(0, 0, 0, alpha*0.8)

	sX := float32(256 + 28 + 298.0)
	sY := float32(608 + 8 + 137.6*2/4)

	sWidth := float32(0.0)
	sHeight := 12.0*float32(len(panel.stats)) + 10

	for _, s := range panel.stats {
		sWidth = max(sWidth, float32(fnt2.GetWidth(12, s)+10))
	}

	sY -= sHeight / 2

	panel.shapeRenderer.DrawQuad(sX, sY, sX+sWidth, sY, sX+sWidth, sY+sHeight, sX, sY+sHeight)

	panel.shapeRenderer.SetColor(1, 1, 1, alpha)

	panel.shapeRenderer.DrawLine(sX+1, sY, sX+sWidth-1, sY, 2)
	panel.shapeRenderer.DrawLine(sX+1, sY+sHeight, sX+sWidth-1, sY+sHeight, 2)

	panel.shapeRenderer.DrawLine(sX+sWidth, sY-1, sX+sWidth, sY+sHeight+1, 2)
	panel.shapeRenderer.DrawLine(sX, sY-1, sX, sY+sHeight+1, 2)

	panel.shapeRenderer.End()

	for i, s := range panel.stats {
		fnt2.DrawOrigin(batch, float64(sX)+5, float64(sY)+float64(i)*12+6, vector.TopLeft, 12, false, s)
	}
}
