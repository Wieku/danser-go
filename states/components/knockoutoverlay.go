package components

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/animation"
	"github.com/wieku/danser-go/animation/easing"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/dance"
	"github.com/wieku/danser-go/discord"
	"github.com/wieku/danser-go/render"
	"github.com/wieku/danser-go/render/batches"
	"github.com/wieku/danser-go/render/font"
	"github.com/wieku/danser-go/rulesets/osu"
	"github.com/wieku/danser-go/settings"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
)

type knockoutPlayer struct {
	fade      *animation.Glider
	slide     *animation.Glider
	height    *animation.Glider
	index     *animation.Glider
	scoreDisp *animation.Glider
	ppDisp    *animation.Glider
	lastCombo int64
	sCombo    int64
	maxCombo  int64
	hasBroken bool
	breakTime int64
	pp        float64
	score     int64
	scores    []int64

	lastHit  osu.HitResult
	fadeHit  *animation.Glider
	scaleHit *animation.Glider

	name         string
	oldIndex     int
	currentIndex int
}

type bubble struct {
	deathFade  *animation.Glider
	deathSlide *animation.Glider
	deathX     float64
	endTime    int64
	name       string
	combo      int64
	lastHit    osu.HitResult
	lastCombo  osu.ComboResult
}

func newBubble(position bmath.Vector2d, time int64, name string, combo int64, lastHit osu.HitResult, lastCombo osu.ComboResult) *bubble {
	bub := new(bubble)
	bub.name = name
	deathShift := (rand.Float64() - 0.5) * 30
	bub.deathX = float64(position.X) + deathShift
	bub.deathSlide = animation.NewGlider(0.0)
	bub.deathFade = animation.NewGlider(0.0)
	bub.deathSlide.SetEasing(easing.OutQuad)
	baseY := position.Y + deathShift
	bub.deathSlide.AddEventS(float64(time), float64(time+2000), baseY, baseY+50)
	bub.deathFade.AddEventS(float64(time), float64(time+200), 0, 1)
	bub.deathFade.AddEventS(float64(time+1800), float64(time+2000), 1, 0)
	bub.endTime = time + 2000
	bub.combo = combo
	bub.lastHit = lastHit
	bub.lastCombo = lastCombo
	return bub
}

type KnockoutOverlay struct {
	controller   *dance.ReplayController
	font         *font.Font
	players      map[string]*knockoutPlayer
	playersArray []*knockoutPlayer
	deathBubbles []*bubble
	names        map[*render.Cursor]string
	lastTime     int64
	//deaths     map[int64]int64
	generator *rand.Rand
	lastObj   int64
}

func NewKnockoutOverlay(replayController *dance.ReplayController) *KnockoutOverlay {
	overlay := new(KnockoutOverlay)
	overlay.controller = replayController
	overlay.font = font.GetFont("Exo 2 Bold")
	overlay.players = make(map[string]*knockoutPlayer)
	overlay.playersArray = make([]*knockoutPlayer, 0)
	overlay.deathBubbles = make([]*bubble, 0)
	overlay.names = make(map[*render.Cursor]string)
	overlay.generator = rand.New(rand.NewSource(replayController.GetBeatMap().TimeAdded))
	//overlay.deaths = make(map[int64]int64)

	for i, r := range replayController.GetReplays() {
		overlay.names[replayController.GetCursors()[i]] = r.Name
		overlay.players[r.Name] = &knockoutPlayer{animation.NewGlider(1), animation.NewGlider(0), animation.NewGlider(settings.Graphics.GetHeightF() * 0.9 * 1.04 / (51)), animation.NewGlider(float64(i)), animation.NewGlider(0), animation.NewGlider(0), 0, 0, r.MaxCombo, false, 0, 0.0, 0, make([]int64, len(replayController.GetBeatMap().HitObjects)), osu.HitResults.Hit300, animation.NewGlider(0), animation.NewGlider(0), r.Name, i, i}
		overlay.players[r.Name].index.SetEasing(easing.InOutQuad)
		overlay.playersArray = append(overlay.playersArray, overlay.players[r.Name])
		/*if i == 0 {
			overlay.players[r.Name].fade.SetValue(0)
			overlay.players[r.Name].height.SetValue(0)
			overlay.players[r.Name].fade.AddEvent(105640, 105867, 1)
			overlay.players[r.Name].height.AddEvent(105640, 105867, settings.Graphics.GetHeightF() * 0.9 * 1.04 / (51))
			//overlay.players[r.Name].fade.AddEvent(187000, 189000, 1)
			//overlay.players[r.Name].height.AddEvent(187000, 189000, settings.Graphics.GetHeightF() * 0.9 * 1.04 / (51))
		}*/ /*else if i == 1 {
			overlay.players[r.Name].fade.SetValue(0)
			overlay.players[r.Name].height.SetValue(0)
			overlay.players[r.Name].fade.AddEvent(221432, 226542, 1)
			overlay.players[r.Name].height.AddEvent(221432, 226542, settings.Graphics.GetHeightF() * 0.9 * 1.04 / (51))
		}*/
	}

	rand.Shuffle(len(overlay.playersArray), func(i, j int) {
		overlay.playersArray[i], overlay.playersArray[j] = overlay.playersArray[j], overlay.playersArray[i]
	})

	discord.UpdateKnockout(len(overlay.playersArray), len(overlay.playersArray))

	for i, g := range overlay.playersArray {
		if i != g.currentIndex {
			g.index.Reset()
			g.index.SetValue(float64(i))
			g.currentIndex = i
		}
	}

	replayController.GetRuleset().SetListener(func(cursor *render.Cursor, time int64, number int64, position bmath.Vector2d, result osu.HitResult, comboResult osu.ComboResult, pp float64, score int64) {
		player := overlay.players[overlay.names[cursor]]

		player.score = score
		player.scores[number] = score

		player.pp = pp

		player.scoreDisp.Reset()
		player.scoreDisp.AddEvent(float64(time), float64(time+500), float64(score))

		player.ppDisp.Reset()
		player.ppDisp.AddEvent(float64(time), float64(time+500), pp)

		if comboResult == osu.ComboResults.Increase {
			player.sCombo++
		}

		acceptableHits := result == osu.HitResults.Hit100 || result == osu.HitResults.Hit50 || result == osu.HitResults.Miss
		if acceptableHits {
			player.fadeHit.Reset()
			player.fadeHit.AddEventS(float64(time), float64(time+300), 0.5, 1)
			player.fadeHit.AddEventS(float64(time+600), float64(time+900), 1, 0)
			player.scaleHit.AddEventS(float64(time), float64(time+300), 0.5, 1)
			player.lastHit = result
			if settings.Knockout.Mode == settings.OneVsOne {
				overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, time, overlay.names[cursor], player.sCombo, result, comboResult))
			}
		}

		if comboResult == osu.ComboResults.Reset && number != 0 {

			if !player.hasBroken {
				if settings.Knockout.Mode == settings.XReplays {
					if player.sCombo >= int64(settings.Knockout.BubbleMinimumCombo) {
						overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, time, overlay.names[cursor], player.sCombo, result, comboResult))
						log.Println(overlay.names[cursor], "has broken! Combo:", player.sCombo)
					}
				} else if settings.Knockout.Mode == settings.ComboBreak || (settings.Knockout.Mode == settings.MaxCombo && math.Abs(float64(player.sCombo-player.maxCombo)) < 5) {
					//Fade out player name
					player.hasBroken = true
					player.breakTime = time

					player.fade.AddEvent(float64(time), float64(time+3000), 0)

					player.height.SetEasing(easing.OutQuad)
					player.height.AddEvent(float64(time+2500), float64(time+3000), 0)

					overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, time, overlay.names[cursor], player.sCombo, result, comboResult))

					log.Println(overlay.names[cursor], "has broken! Max combo:", player.sCombo)
				}
			}

			player.sCombo = 0
		}
	})

	sortFunc := func(time int64, number int64, instantSort bool) {
		sort.SliceStable(overlay.playersArray, func(i, j int) bool {
			return (!overlay.playersArray[i].hasBroken && overlay.playersArray[j].hasBroken) || ((!overlay.playersArray[i].hasBroken && !overlay.playersArray[j].hasBroken) && overlay.playersArray[i].score > overlay.playersArray[j].score) || ((overlay.playersArray[i].hasBroken && overlay.playersArray[j].hasBroken) && (overlay.playersArray[i].breakTime > overlay.playersArray[j].breakTime || (overlay.playersArray[i].breakTime == overlay.playersArray[j].breakTime && overlay.playersArray[i].scores[number] > overlay.playersArray[j].scores[number])))
		})
		alive := 0
		for i, g := range overlay.playersArray {
			if !g.hasBroken {
				alive++
			}
			if i != g.currentIndex {
				g.index.Reset()
				animDuration := 0.0
				if !instantSort {
					animDuration = 200 + math.Abs(float64(i-g.currentIndex))*10
				}
				g.index.AddEvent(float64(time), float64(time)+animDuration, float64(i))
				g.currentIndex = i
			}
		}
		discord.UpdateKnockout(alive, len(overlay.playersArray))
	}

	replayController.GetRuleset().SetEndListener(func(time int64, number int64) {
		if number == int64(len(replayController.GetBeatMap().HitObjects)-1) && settings.Knockout.RevivePlayersAtEnd {
			sortFunc(time, number, true)
			for _, player := range overlay.players {
				player.hasBroken = false
				player.breakTime = 0

				player.fade.Reset()
				player.fade.AddEvent(float64(time), float64(time+750), 1)

				player.height.Reset()
				player.height.SetEasing(easing.InQuad)
				player.height.AddEvent(float64(time), float64(time+200), settings.Graphics.GetHeightF()*0.9*1.04/(51))
			}
		} else {
			sortFunc(time, number, false)
		}
	})

	return overlay
}

func (overlay *KnockoutOverlay) Update(time int64) {
	for sTime := overlay.lastTime + 1; sTime <= time; sTime++ {
		for _, r := range overlay.controller.GetReplays() {
			player := overlay.players[r.Name]
			player.height.Update(float64(sTime))
			player.fade.Update(float64(sTime))
			player.fadeHit.Update(float64(sTime))
			player.scaleHit.Update(float64(sTime))
			player.index.Update(float64(sTime))
			player.scoreDisp.Update(float64(sTime))
			player.ppDisp.Update(float64(sTime))
			player.lastCombo = r.Combo
		}
	}
	overlay.lastTime = time
}

func (overlay *KnockoutOverlay) DrawBeforeObjects(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	cs := overlay.controller.GetBeatMap().Diff.CircleRadius
	sizeX := 512 + cs*2
	sizeY := 384 + cs*2

	batch.SetScale(sizeX/2, sizeY/2)
	batch.SetColor(0, 0, 0, 0.8)
	batch.SetTranslation(bmath.NewVec2d(256, 192)) //bg
	batch.DrawUnit(render.Pixel.GetRegion())

	batch.SetColor(1, 1, 1, 1)
	batch.SetScale(sizeX/2, 0.3)
	batch.SetTranslation(bmath.NewVec2d(256, -cs)) //top line
	batch.DrawUnit(render.Pixel.GetRegion())

	batch.SetTranslation(bmath.NewVec2d(256, 384+cs)) //bottom line
	batch.DrawUnit(render.Pixel.GetRegion())

	batch.SetScale(0.3, sizeY/2)
	batch.SetTranslation(bmath.NewVec2d(-cs, 192)) //left line
	batch.DrawUnit(render.Pixel.GetRegion())
	batch.SetTranslation(bmath.NewVec2d(512+cs, 192)) //right line
	batch.DrawUnit(render.Pixel.GetRegion())
	batch.SetScale(1, 1)
}

func (overlay *KnockoutOverlay) DrawNormal(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	scl := /*settings.Graphics.GetHeightF() * 0.9*(900.0/1080.0)*/ 384.0 * (1080.0 / 900.0 * 0.9) / (51)
	batch.SetScale(1, -1)
	rescale := /*384.0/512.0 * (1080.0/settings.Graphics.GetHeightF())*/ 1.0

	alive := 0
	for _, r := range overlay.controller.GetReplays() {
		player := overlay.players[r.Name]
		if !player.hasBroken {
			alive++
		}
	}

	for i := 0; i < len(overlay.deathBubbles); i++ {
		bubble := overlay.deathBubbles[i]
		bubble.deathFade.Update(float64(overlay.lastTime))
		bubble.deathSlide.Update(float64(overlay.lastTime))
		if bubble.deathFade.GetValue() >= 0.01 {

			rep := overlay.players[bubble.name]
			batch.SetColor(float64(colors[rep.oldIndex].X()), float64(colors[rep.oldIndex].Y()), float64(colors[rep.oldIndex].Z()), alpha*bubble.deathFade.GetValue())
			width := overlay.font.GetWidth(scl*rescale, bubble.name)
			overlay.font.Draw(batch, bubble.deathX-width/2, bubble.deathSlide.GetValue()-scl*rescale/2, scl*rescale, bubble.name)

			batch.SetColor(1, 1, 1, alpha*bubble.deathFade.GetValue())

			if bubble.lastCombo == osu.ComboResults.Reset {
				combo := fmt.Sprintf("%dx", bubble.combo)
				comboWidth := overlay.font.GetWidth(scl*rescale*0.8, combo)
				overlay.font.Draw(batch, bubble.deathX-comboWidth/2, bubble.deathSlide.GetValue()+scl*rescale*0.8/2, scl*rescale*0.8, combo)
			} else {
				switch bubble.lastHit {
				case osu.HitResults.Hit100:
					batch.SetSubScale(scl*(float64(render.Hit100.Width)/float64(render.Hit100.Height))/2, -scl/2)
					batch.SetTranslation(bmath.NewVec2d(bubble.deathX, bubble.deathSlide.GetValue() /*- scl*rescale*0.8*/))
					batch.DrawUnit(*render.Hit100)
				case osu.HitResults.Hit50:
					batch.SetSubScale(scl*(float64(render.Hit50.Width)/float64(render.Hit50.Height))/2, -scl/2)
					batch.SetTranslation(bmath.NewVec2d(bubble.deathX, bubble.deathSlide.GetValue()-scl*rescale*0.8))
					batch.DrawUnit(*render.Hit50)
				}
			}
		}

		if bubble.endTime <= overlay.lastTime {
			overlay.deathBubbles = append(overlay.deathBubbles[:i], overlay.deathBubbles[i+1:]...)
			i--
		}

	}
	settings.Cursor.CursorSize = 3.0 + (7-3)*math.Pow(1-math.Sin(float64(alive)/51*math.Pi/2), 3)
	batch.SetScale(1, 1)
}

func (overlay *KnockoutOverlay) DrawHUD(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	controller := overlay.controller
	replays := controller.GetReplays()

	scl := settings.Graphics.GetHeightF() * 0.9 / 51
	//margin := scl*0.02

	highestCombo := int64(0)
	highestPP := 0.0
	highestACC := 0.0
	highestScore := int64(0)
	cumulativeHeight := 0.0
	for _, r := range replays {
		cumulativeHeight += overlay.players[r.Name].height.GetValue()
		if overlay.players[r.Name].sCombo > highestCombo {
			highestCombo = overlay.players[r.Name].sCombo
		}
		if overlay.players[r.Name].pp > highestPP {
			highestPP = overlay.players[r.Name].pp
		}
		if r.Accuracy > highestACC {
			highestACC = r.Accuracy
		}
		if overlay.players[r.Name].score > highestScore {
			highestScore = overlay.players[r.Name].score
		}
	}

	rowPosY := settings.Graphics.GetHeightF() - (settings.Graphics.GetHeightF()-cumulativeHeight)/2

	//cL := strconv.FormatInt(highestCombo, 10)
	cP := strconv.FormatInt(int64(highestPP), 10)
	cA := strconv.FormatInt(int64(highestACC), 10)
	cS := overlay.font.GetWidthMonospaced(scl, humanize(highestScore))

	for _, rep := range overlay.playersArray {
		r := replays[rep.oldIndex]
		player := overlay.players[r.Name]
		batch.SetColor(float64(colors[rep.oldIndex].X()), float64(colors[rep.oldIndex].Y()), float64(colors[rep.oldIndex].Z()), alpha*player.fade.GetValue())

		rowBaseY := rowPosY - rep.index.GetValue()*(settings.Graphics.GetHeightF()*0.9*1.04/(51)) - player.height.GetValue()/2 /*+margin*10*/
		rowPosY += settings.Graphics.GetHeightF()*0.9*1.04/(51) - player.height.GetValue()
		for j := 0; j < 2; j++ {
			batch.SetSubScale(scl*0.9/2, scl*0.9/2)
			batch.SetTranslation(bmath.NewVec2d((float64(j)+0.5)*scl /*rowPosY*/, rowBaseY))
			batch.DrawUnit(*render.OvButtonE)
			if controller.GetClick(rep.oldIndex, j) || controller.GetClick(rep.oldIndex, j+2) {
				batch.DrawUnit(*render.OvButton)
			}
		}

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		accuracy := fmt.Sprintf("%"+strconv.Itoa(len(cA)+3)+".2f%% %"+strconv.Itoa(len(cP)+3)+".2fpp", r.Accuracy /*r.Combo*/, overlay.players[r.Name].ppDisp.GetValue())
		//_ = cL
		accuracy1 := cA + ".00% " + cP + ".00pp "
		nWidth := overlay.font.GetWidthMonospaced(scl, accuracy1)

		overlay.font.DrawMonospaced(batch, 2*scl, rowBaseY-scl*0.8/2, scl, accuracy)

		scorestr := humanize(int64(player.scoreDisp.GetValue()))

		sWC := fmt.Sprintf("%dx ", overlay.players[r.Name].sCombo)

		overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-cS-overlay.font.GetWidthMonospaced(scl, sWC)-0.5*scl, rowBaseY-scl*0.8/2, scl, sWC)
		overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-overlay.font.GetWidthMonospaced(scl, scorestr)-0.5*scl, rowBaseY-scl*0.8/2, scl, scorestr)

		batch.SetColor(float64(colors[rep.oldIndex].X()), float64(colors[rep.oldIndex].Y()), float64(colors[rep.oldIndex].Z()), alpha*player.fade.GetValue())
		overlay.font.Draw(batch, 3*scl+nWidth, rowBaseY-scl*0.8/2, scl, r.Name)
		width := overlay.font.GetWidth(scl, r.Name)

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		if r.Mods != "" {
			overlay.font.Draw(batch, 3*scl+width+nWidth, rowBaseY-scl*0.8/2, scl*0.8, "+"+r.Mods)
			width += overlay.font.GetWidth(scl*0.8, "+"+r.Mods)
		}

		batch.SetSubScale(scl*0.9/2, -scl*0.9/2)
		batch.SetTranslation(bmath.NewVec2d(2*scl+scl*0.1+nWidth, rowBaseY))
		if r.Grade != osu.NONE {
			batch.DrawUnit(*render.GradeTexture[int64(r.Grade)])
		}

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue()*player.fadeHit.GetValue())
		batch.SetSubScale(scl*0.9/2*player.scaleHit.GetValue(), -scl*0.9/2*player.scaleHit.GetValue())
		batch.SetTranslation(bmath.NewVec2d(3*scl+width+nWidth+scl*0.5, rowBaseY))

		switch player.lastHit {
		case osu.HitResults.Hit100:
			batch.SetSubScale(scl*0.9/2*player.scaleHit.GetValue()*(float64(render.Hit100.Width)/float64(render.Hit100.Height)), -scl*0.9/2*player.scaleHit.GetValue())
			batch.SetTranslation(bmath.NewVec2d(3*scl+width+nWidth+scl*(float64(render.Hit100.Width)/float64(render.Hit100.Height))*0.5, rowBaseY))
			batch.DrawUnit(*render.Hit100)
		case osu.HitResults.Hit50:
			batch.SetSubScale(scl*0.9/2*player.scaleHit.GetValue()*(float64(render.Hit50.Width)/float64(render.Hit50.Height)), -scl*0.9/2*player.scaleHit.GetValue())
			batch.SetTranslation(bmath.NewVec2d(3*scl+width+nWidth+scl*(float64(render.Hit50.Width)/float64(render.Hit50.Height))*0.5, rowBaseY))
			batch.DrawUnit(*render.Hit50)
		case osu.HitResults.Miss:
			batch.DrawUnit(*render.Hit0)
		}

		//rowPosY -= player.height.GetValue()
	}
}

func (overlay *KnockoutOverlay) IsBroken(cursor *render.Cursor) bool {
	return overlay.players[overlay.names[cursor]].hasBroken
}

func (overlay *KnockoutOverlay) NormalBeforeCursor() bool {
	return true
}

func humanize(number int64) string {
	stringified := strconv.FormatInt(number, 10)

	a := len(stringified) % 3
	if a == 0 {
		a = 3
	}

	humanized := stringified[0:a]

	for i := a; i < len(stringified); i += 3 {
		humanized += "," + stringified[i:i+3]
	}

	return humanized
}
