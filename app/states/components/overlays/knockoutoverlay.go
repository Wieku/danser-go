package overlays

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/dance"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/states/components/common"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"strings"
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
	displayHp float64

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
	deathScale *animation.Glider
}

func newBubble(position vector.Vector2d, time int64, name string, combo int64, lastHit osu.HitResult, lastCombo osu.ComboResult) *bubble {
	bub := new(bubble)
	bub.name = name
	deathShiftX := (rand.Float64() - 0.5) * 10
	deathShiftY := (rand.Float64() - 0.5) * 10
	bub.deathX = float64(position.X) + deathShiftX
	bub.deathSlide = animation.NewGlider(0.0)
	bub.deathFade = animation.NewGlider(0.0)
	bub.deathScale = animation.NewGlider(1)
	bub.deathSlide.SetEasing(easing.OutQuad)
	baseY := position.Y + deathShiftY
	if settings.Knockout.Mode == settings.OneVsOne {
		bub.deathSlide.AddEventS(float64(time), float64(time+2000), baseY, baseY)
		bub.deathFade.AddEventS(float64(time), float64(time+difficulty.ResultFadeIn), 0, 1)
		bub.deathFade.AddEventS(float64(time+difficulty.PostEmpt), float64(time+difficulty.PostEmpt+difficulty.ResultFadeOut), 1, 0)
		bub.deathScale.AddEventSEase(float64(time), float64(time+difficulty.ResultFadeIn*1.2), 0.4, 1, easing.OutElastic)
	} else {
		bub.deathSlide.AddEventS(float64(time), float64(time+2000), baseY, baseY+50)
		bub.deathFade.AddEventS(float64(time), float64(time+200), 0, 1)
		bub.deathFade.AddEventS(float64(time+800), float64(time+1200), 1, 0)
	}
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
	names        map[*graphics.Cursor]string
	lastTime     int64
	//deaths     map[int64]int64
	generator *rand.Rand
	lastObj   int64

	boundaries *common.Boundaries
}

func NewKnockoutOverlay(replayController *dance.ReplayController) *KnockoutOverlay {
	overlay := new(KnockoutOverlay)
	overlay.controller = replayController
	overlay.font = font.GetFont("Exo 2 Bold")
	overlay.players = make(map[string]*knockoutPlayer)
	overlay.playersArray = make([]*knockoutPlayer, 0)
	overlay.deathBubbles = make([]*bubble, 0)
	overlay.names = make(map[*graphics.Cursor]string)
	overlay.generator = rand.New(rand.NewSource(replayController.GetBeatMap().TimeAdded))
	//overlay.deaths = make(map[int64]int64)

	for i, r := range replayController.GetReplays() {
		overlay.names[replayController.GetCursors()[i]] = r.Name
		overlay.players[r.Name] = &knockoutPlayer{animation.NewGlider(1), animation.NewGlider(0), animation.NewGlider(settings.Graphics.GetHeightF() * 0.9 * 1.04 / (51)), animation.NewGlider(float64(i)), animation.NewGlider(0), animation.NewGlider(0), 0, 0, r.MaxCombo, false, 0, 0.0, 0, make([]int64, len(replayController.GetBeatMap().HitObjects)), 0.0, osu.Hit300, animation.NewGlider(0), animation.NewGlider(0), r.Name, i, i}
		overlay.players[r.Name].index.SetEasing(easing.InOutQuad)
		overlay.playersArray = append(overlay.playersArray, overlay.players[r.Name])
	}

	if settings.Knockout.LiveSort {
		rand.Shuffle(len(overlay.playersArray), func(i, j int) {
			overlay.playersArray[i], overlay.playersArray[j] = overlay.playersArray[j], overlay.playersArray[i]
		})
	}

	discord.UpdateKnockout(len(overlay.playersArray), len(overlay.playersArray))

	for i, g := range overlay.playersArray {
		if i != g.currentIndex {
			g.index.Reset()
			g.index.SetValue(float64(i))
			g.currentIndex = i
		}
	}

	replayController.GetRuleset().SetListener(func(cursor *graphics.Cursor, time int64, number int64, position vector.Vector2d, result osu.HitResult, comboResult osu.ComboResult, pp float64, score int64) {
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

		resultClean := result & osu.BaseHitsM

		acceptableHits := resultClean&(osu.Hit100|osu.Hit50|osu.Miss) > 0
		if acceptableHits {
			player.fadeHit.Reset()
			player.fadeHit.AddEventS(float64(time), float64(time+300), 0.5, 1)
			player.fadeHit.AddEventS(float64(time+600), float64(time+900), 1, 0)
			player.scaleHit.AddEventS(float64(time), float64(time+300), 0.5, 1)
			player.lastHit = result & (osu.HitValues | osu.Miss) //resultClean
			if settings.Knockout.Mode == settings.OneVsOne {
				overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, time, overlay.names[cursor], player.sCombo, resultClean, comboResult))
			}
		}

		if comboResult == osu.ComboResults.Reset && number != 0 {

			if !player.hasBroken {
				if settings.Knockout.Mode == settings.XReplays {
					if player.sCombo >= int64(settings.Knockout.BubbleMinimumCombo) {
						overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, time, overlay.names[cursor], player.sCombo, resultClean, comboResult))
						log.Println(overlay.names[cursor], "has broken! Combo:", player.sCombo)
					}
				} else if settings.Knockout.Mode == settings.ComboBreak || (settings.Knockout.Mode == settings.MaxCombo && math.Abs(float64(player.sCombo-player.maxCombo)) < 5) {
					//Fade out player name
					player.hasBroken = true
					player.breakTime = time

					player.fade.AddEvent(float64(time), float64(time+3000), 0)

					player.height.SetEasing(easing.OutQuad)
					player.height.AddEvent(float64(time+2500), float64(time+3000), 0)

					overlay.deathBubbles = append(overlay.deathBubbles, newBubble(position, time, overlay.names[cursor], player.sCombo, resultClean, comboResult))

					log.Println(overlay.names[cursor], "has broken! Max combo:", player.sCombo)
				}
			}

			player.sCombo = 0
		}
	})

	sortFunc := func(time int64, number int64, instantSort bool) {
		alive := 0
		for _, g := range overlay.playersArray {
			if !g.hasBroken {
				alive++
			}
		}

		if settings.Knockout.LiveSort {
			sort.SliceStable(overlay.playersArray, func(i, j int) bool {
				return (!overlay.playersArray[i].hasBroken && overlay.playersArray[j].hasBroken) || ((!overlay.playersArray[i].hasBroken && !overlay.playersArray[j].hasBroken) && overlay.playersArray[i].score > overlay.playersArray[j].score) || ((overlay.playersArray[i].hasBroken && overlay.playersArray[j].hasBroken) && (overlay.playersArray[i].breakTime > overlay.playersArray[j].breakTime || (overlay.playersArray[i].breakTime == overlay.playersArray[j].breakTime && overlay.playersArray[i].scores[number] > overlay.playersArray[j].scores[number])))
			})
			for i, g := range overlay.playersArray {
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
		}

		discord.UpdateKnockout(alive, len(overlay.playersArray))
	}

	replayController.GetRuleset().SetEndListener(func(time int64, number int64) {
		if number == int64(len(replayController.GetBeatMap().HitObjects)-1) && settings.Knockout.RevivePlayersAtEnd {
			for _, player := range overlay.players {
				player.hasBroken = false
				player.breakTime = 0

				player.fade.Reset()
				player.fade.AddEvent(float64(time), float64(time+750), 1)

				player.height.Reset()
				player.height.SetEasing(easing.InQuad)
				player.height.AddEvent(float64(time), float64(time+200), settings.Graphics.GetHeightF()*0.9*1.04/(51))
			}
			sortFunc(time, number, true)
		} else {
			sortFunc(time, number, false)
		}
	})

	overlay.boundaries = common.NewBoundaries()

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

			currentHp := overlay.controller.GetRuleset().GetHP(overlay.controller.GetCursors()[player.oldIndex])

			if player.displayHp < currentHp {
				player.displayHp = math.Min(1.0, player.displayHp+math.Abs(currentHp-player.displayHp)/4/16.667)
			} else if player.displayHp > currentHp {
				player.displayHp = math.Max(0.0, player.displayHp-math.Abs(player.displayHp-currentHp)/6/16.667)
			}

		}
	}
	overlay.lastTime = time
}

func (overlay *KnockoutOverlay) DrawBeforeObjects(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	overlay.boundaries.Draw(batch.Projection, float32(overlay.controller.GetBeatMap().Diff.CircleRadius), float32(alpha))
}

func (overlay *KnockoutOverlay) DrawNormal(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
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
		bubble.deathScale.Update(float64(overlay.lastTime))
		if bubble.deathFade.GetValue() >= 0.01 {
			if settings.Knockout.Mode == settings.OneVsOne {
				val := strconv.Itoa(int(bubble.lastHit.ScoreValue()))
				if bubble.lastCombo == osu.ComboResults.Reset {
					val = "X"
				}
				rep := overlay.players[bubble.name]
				batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*bubble.deathFade.GetValue())
				width := overlay.font.GetWidth(scl*rescale*bubble.deathScale.GetValue(), val)
				overlay.font.Draw(batch, bubble.deathX-width/2, bubble.deathSlide.GetValue()+scl*rescale*bubble.deathScale.GetValue()/3, scl*rescale*bubble.deathScale.GetValue(), val)
			} else {
				rep := overlay.players[bubble.name]
				batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*bubble.deathFade.GetValue())
				width := overlay.font.GetWidth(scl*rescale, bubble.name)
				overlay.font.Draw(batch, bubble.deathX-width/2, bubble.deathSlide.GetValue()-scl*rescale/2, scl*rescale, bubble.name)

				batch.SetColor(1, 1, 1, alpha*bubble.deathFade.GetValue())

				if bubble.lastCombo == osu.ComboResults.Reset {
					combo := fmt.Sprintf("%dx", bubble.combo)
					comboWidth := overlay.font.GetWidth(scl*rescale*0.8, combo)
					overlay.font.Draw(batch, bubble.deathX-comboWidth/2, bubble.deathSlide.GetValue()+scl*rescale*0.8/2, scl*rescale*0.8, combo)
				} else {
					switch bubble.lastHit {
					case osu.Hit100:
						batch.SetSubScale(scl*(float64(graphics.Hit100.Width)/float64(graphics.Hit100.Height))/2, -scl/2)
						batch.SetTranslation(vector.NewVec2d(bubble.deathX, bubble.deathSlide.GetValue() /*- scl*rescale*0.8*/))
						batch.DrawUnit(*graphics.Hit100)
					case osu.Hit50:
						batch.SetSubScale(scl*(float64(graphics.Hit50.Width)/float64(graphics.Hit50.Height))/2, -scl/2)
						batch.SetTranslation(vector.NewVec2d(bubble.deathX, bubble.deathSlide.GetValue()-scl*rescale*0.8))
						batch.DrawUnit(*graphics.Hit50)
					}
				}
			}
		}

		if bubble.endTime <= overlay.lastTime {
			overlay.deathBubbles = append(overlay.deathBubbles[:i], overlay.deathBubbles[i+1:]...)
			i--
		}
	}

	minSize := settings.Knockout.MinCursorSize
	maxSize := settings.Knockout.MaxCursorSize
	settings.Cursor.CursorSize = minSize + (maxSize-minSize)*math.Pow(1-math.Sin(float64(alive)/51*math.Pi/2), 3)

	batch.SetScale(1, 1)
}

func (overlay *KnockoutOverlay) DrawHUD(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
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

		highestCombo = bmath.MaxI64(highestCombo, overlay.players[r.Name].sCombo)
		highestPP = math.Max(highestPP, overlay.players[r.Name].pp)
		highestACC = math.Max(highestACC, r.Accuracy)
		highestScore = bmath.MaxI64(highestScore, overlay.players[r.Name].score)
	}

	//cL := strconv.FormatInt(highestCombo, 10)
	cP := strconv.FormatInt(int64(highestPP), 10)
	cA := strconv.FormatInt(int64(highestACC), 10)
	cS := overlay.font.GetWidthMonospaced(scl, humanize(highestScore))

	rowPosY := settings.Graphics.GetHeightF() - (settings.Graphics.GetHeightF()-cumulativeHeight)/2
	// Draw textures like keys, grade, hit values
	for _, rep := range overlay.playersArray {
		r := replays[rep.oldIndex]
		player := overlay.players[r.Name]

		rowBaseY := rowPosY - rep.index.GetValue()*(settings.Graphics.GetHeightF()*0.9*1.04/(51)) - player.height.GetValue()/2 /*+margin*10*/
		rowPosY += settings.Graphics.GetHeightF()*0.9*1.04/(51) - player.height.GetValue()

		//batch.SetColor(0.1, 0.8, 0.4, alpha*player.fade.GetValue()*0.4)
		add := 0.3 + float64(int(math.Round(rep.index.GetValue()))%2)*0.2
		batch.SetColor(add, add, add, alpha*player.fade.GetValue()*0.7)

		//batch.SetAdditive(true)
		//batch.SetSubScale(player.displayHp*30.5*scl*0.9/2, scl*0.9/2)
		//batch.SetTranslation(vector.NewVec2d(player.displayHp*30.5/2*scl*0.9/2 /*rowPosY*/, rowBaseY))
		//batch.DrawUnit(graphics.Pixel.GetRegion())
		//batch.SetSubScale(16.5*scl*0.9/2, scl*0.9/2)
		//batch.SetTranslation(vector.NewVec2d(settings.Graphics.GetWidthF()-16.5/2*scl*0.9/2 /*rowPosY*/, rowBaseY))
		//batch.DrawUnit(graphics.Pixel.GetRegion())
		//batch.SetAdditive(false)

		batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*player.fade.GetValue())

		for j := 0; j < 2; j++ {
			batch.SetSubScale(scl*0.8/2, scl*0.8/2)
			batch.SetTranslation(vector.NewVec2d((float64(j)+0.5)*scl /*rowPosY*/, rowBaseY))
			batch.DrawUnit(*graphics.OvButtonE)
			if controller.GetClick(rep.oldIndex, j) || controller.GetClick(rep.oldIndex, j+2) {
				batch.DrawUnit(*graphics.OvButton)
			}
		}

		accuracy1 := cA + ".00% " + cP + ".00pp "
		nWidth := overlay.font.GetWidthMonospaced(scl, accuracy1)

		width := overlay.font.GetWidth(scl, r.Name)

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		if r.Mods != "" {
			width += overlay.font.GetWidth(scl*0.8, "+"+r.Mods)
		}

		batch.SetSubScale(scl*0.85/2, -scl*0.85/2)
		batch.SetTranslation(vector.NewVec2d(2*scl+scl*0.1+nWidth, rowBaseY))
		if r.Grade != osu.NONE {
			gText := strings.ToLower(strings.ReplaceAll(osu.GradesText[r.Grade], "SS", "X"))
			text := skin.GetTexture("ranking-" + gText + "-small")
			batch.DrawUnit(*text)
		}

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue()*player.fadeHit.GetValue())
		batch.SetSubScale(scl*0.9/2*player.scaleHit.GetValue(), -scl*0.9/2*player.scaleHit.GetValue())
		batch.SetTranslation(vector.NewVec2d(3*scl+width+nWidth+scl*0.5, rowBaseY))

		if player.lastHit != 0 {
			tex := ""
			switch player.lastHit & osu.BaseHitsM {
			case osu.Hit300:
				tex = "hit300"
			case osu.Hit100:
				tex = "hit100"
			case osu.Hit50:
				tex = "hit50"
			case osu.Miss:
				tex = "hit0"
			}

			switch player.lastHit & osu.Additions {
			case osu.KatuAddition:
				tex += "k"
			case osu.GekiAddition:
				tex += "g"
			}

			if tex != "" {
				texture := skin.GetTexture(tex)
				batch.SetSubScale(scl*0.8/2*player.scaleHit.GetValue()*(float64(texture.Width)/float64(texture.Height)), -scl*0.8/2*player.scaleHit.GetValue())
				batch.SetTranslation(vector.NewVec2d(3*scl+width+nWidth+scl*(float64(texture.Width)/float64(texture.Height))*0.5, rowBaseY))
				batch.DrawUnit(*texture)
			}

		}
	}

	rowPosY = settings.Graphics.GetHeightF() - (settings.Graphics.GetHeightF()-cumulativeHeight)/2
	// Draw texts
	for _, rep := range overlay.playersArray {
		r := replays[rep.oldIndex]
		player := overlay.players[r.Name]

		rowBaseY := rowPosY - rep.index.GetValue()*(settings.Graphics.GetHeightF()*0.9*1.04/(51)) - player.height.GetValue()/2 /*+margin*10*/
		rowPosY += settings.Graphics.GetHeightF()*0.9*1.04/(51) - player.height.GetValue()

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		accuracy := fmt.Sprintf("%"+strconv.Itoa(len(cA)+3)+".2f%% %"+strconv.Itoa(len(cP)+3)+".2fpp", r.Accuracy /*r.Combo*/, overlay.players[r.Name].ppDisp.GetValue())
		//_ = cL
		accuracy1 := cA + ".00% " + cP + ".00pp "
		nWidth := overlay.font.GetWidthMonospaced(scl, accuracy1)

		overlay.font.DrawMonospaced(batch, 2*scl, rowBaseY-scl*1/3, scl, accuracy)

		scorestr := humanize(int64(player.scoreDisp.GetValue()))

		sWC := fmt.Sprintf("%dx ", overlay.players[r.Name].sCombo)

		overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-cS-overlay.font.GetWidthMonospaced(scl, sWC)-0.5*scl, rowBaseY-scl*1/3, scl, sWC)
		overlay.font.DrawMonospaced(batch, settings.Graphics.GetWidthF()-overlay.font.GetWidthMonospaced(scl, scorestr)-0.5*scl, rowBaseY-scl*1/3, scl, scorestr)

		batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*player.fade.GetValue())
		overlay.font.Draw(batch, 3*scl+nWidth, rowBaseY-scl*1/3, scl, r.Name)
		width := overlay.font.GetWidth(scl, r.Name)

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		if r.Mods != "" {
			overlay.font.Draw(batch, 3*scl+width+nWidth, rowBaseY-scl*1/3, scl*0.8, "+"+r.Mods)
		}
	}
}

func (overlay *KnockoutOverlay) IsBroken(cursor *graphics.Cursor) bool {
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
