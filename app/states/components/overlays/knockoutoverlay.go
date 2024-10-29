package overlays

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap/difficulty"
	"github.com/wieku/danser-go/app/dance"
	"github.com/wieku/danser-go/app/discord"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/app/rulesets/osu"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/states/components/common"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/texture"
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

type stats struct {
	pp       float64
	score    int64
	accuracy float64
}

type knockoutPlayer struct {
	fade   *animation.Glider
	slide  *animation.Glider
	height *animation.Glider
	index  *animation.Glider

	scoreDisp *animation.TargetGlider
	ppDisp    *animation.TargetGlider
	accDisp   *animation.TargetGlider

	lastCombo int64
	sCombo    int64
	maxCombo  int64
	hasBroken bool
	breakTime int64
	pp        float64
	score     int64

	perObjectStats []stats

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
	endTime    float64
	name       string
	combo      int64
	lastHit    osu.HitResult
	lastCombo  osu.ComboResult
	deathScale *animation.Glider
}

func newBubble(position vector.Vector2f, time float64, name string, combo int64, lastHit osu.HitResult, lastCombo osu.ComboResult) *bubble {
	deathShiftX := (rand.Float64() - 0.5) * 10
	deathShiftY := (rand.Float64() - 0.5) * 10
	baseY := position.Y64() + deathShiftY

	bub := new(bubble)
	bub.name = name
	bub.deathX = position.X64() + deathShiftX
	bub.deathSlide = animation.NewGlider(0.0)
	bub.deathFade = animation.NewGlider(0.0)
	bub.deathScale = animation.NewGlider(1)
	bub.deathSlide.SetEasing(easing.OutQuad)

	if settings.Knockout.Mode == settings.OneVsOne {
		bub.deathSlide.AddEventS(time, time+2000, baseY, baseY)
		bub.deathFade.AddEventS(time, time+difficulty.ResultFadeIn, 0, 1)
		bub.deathFade.AddEventS(time+difficulty.PostEmpt, time+difficulty.PostEmpt+difficulty.ResultFadeOut, 1, 0)
		bub.deathScale.AddEventSEase(time, time+difficulty.ResultFadeIn*1.2, 0.4, 1, easing.OutElastic)
	} else {
		bub.deathSlide.AddEventS(time, time+2000, baseY, baseY+50)
		bub.deathFade.AddEventS(time, time+200, 0, 1)
		bub.deathFade.AddEventS(time+800, time+1200, 1, 0)
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
	generator    *rand.Rand

	audioTime  float64
	normalTime float64

	boundaries *common.Boundaries

	Button        *texture.TextureRegion
	ButtonClicked *texture.TextureRegion

	ScaledHeight float64
	ScaledWidth  float64

	music bass.ITrack

	breakMode bool
	fade      *animation.Glider

	alivePlayers int
}

func NewKnockoutOverlay(replayController *dance.ReplayController) *KnockoutOverlay {
	overlay := new(KnockoutOverlay)
	overlay.controller = replayController

	if font.GetFont("Quicksand Bold") == nil {
		file, _ := assets.Open("assets/fonts/Quicksand-Bold.ttf")
		font.LoadFont(file)
		file.Close()
	}

	overlay.font = font.GetFont("Quicksand Bold")

	overlay.players = make(map[string]*knockoutPlayer)
	overlay.playersArray = make([]*knockoutPlayer, 0)
	overlay.deathBubbles = make([]*bubble, 0)
	overlay.names = make(map[*graphics.Cursor]string)
	overlay.generator = rand.New(rand.NewSource(replayController.GetBeatMap().TimeAdded))

	overlay.ScaledHeight = 1080.0
	overlay.ScaledWidth = overlay.ScaledHeight * settings.Graphics.GetAspectRatio()

	overlay.fade = animation.NewGlider(1)

	for i, r := range replayController.GetReplays() {
		cursor := replayController.GetCursors()[i]
		overlay.names[cursor] = r.Name
		overlay.players[r.Name] = &knockoutPlayer{animation.NewGlider(1), animation.NewGlider(0), animation.NewGlider(overlay.ScaledHeight * 0.9 * 1.04 / (51)), animation.NewGlider(float64(i)), animation.NewTargetGlider(0, 0), animation.NewTargetGlider(0, 2), animation.NewTargetGlider(100, 2), 0, 0, r.MaxCombo, false, 0, 0.0, 0, make([]stats, len(replayController.GetBeatMap().HitObjects)), 0.0, osu.Hit300, animation.NewGlider(0), animation.NewGlider(0), r.Name, i, i}
		overlay.players[r.Name].index.SetEasing(easing.InOutQuad)
		overlay.playersArray = append(overlay.playersArray, overlay.players[r.Name])

		overlay.alivePlayers++
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

	replayController.GetRuleset().SetListener(overlay.hitReceived)

	sortFunc := func(number int64, instantSort bool) {
		alive := 0
		for _, g := range overlay.playersArray {
			if !g.hasBroken {
				alive++
			}
		}

		if settings.Knockout.LiveSort {
			cond := strings.ToLower(settings.Knockout.SortBy)

			sort.SliceStable(overlay.playersArray, func(i, j int) bool {
				mainCond := true
				switch cond {
				case "pp":
					mainCond = overlay.playersArray[i].perObjectStats[number].pp > overlay.playersArray[j].perObjectStats[number].pp
				case "acc", "accuracy":
					mainCond = overlay.playersArray[i].perObjectStats[number].accuracy > overlay.playersArray[j].perObjectStats[number].accuracy
				default:
					mainCond = overlay.playersArray[i].perObjectStats[number].score > overlay.playersArray[j].perObjectStats[number].score
				}

				return (!overlay.playersArray[i].hasBroken && overlay.playersArray[j].hasBroken) || ((!overlay.playersArray[i].hasBroken && !overlay.playersArray[j].hasBroken) && mainCond) || ((overlay.playersArray[i].hasBroken && overlay.playersArray[j].hasBroken) && (overlay.playersArray[i].breakTime > overlay.playersArray[j].breakTime || (overlay.playersArray[i].breakTime == overlay.playersArray[j].breakTime && mainCond)))
			})

			for i, g := range overlay.playersArray {
				if i != g.currentIndex {
					g.index.Reset()

					animDuration := 0.0
					if !instantSort {
						animDuration = 200 + math.Abs(float64(i-g.currentIndex))*10
					}

					g.index.AddEvent(overlay.normalTime, overlay.normalTime+animDuration, float64(i))
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
				player.fade.AddEvent(overlay.normalTime, overlay.normalTime+750, 1)

				player.height.Reset()
				player.height.SetEasing(easing.InQuad)
				player.height.AddEvent(overlay.normalTime, overlay.normalTime+200, overlay.ScaledHeight*0.9*1.04/(51))
			}

			sortFunc(number, true)
		} else {
			sortFunc(number, false)
		}
	})

	overlay.boundaries = common.NewBoundaries()

	overlay.Button = skin.GetTexture("knockout-button")
	overlay.ButtonClicked = skin.GetTexture("knockout-button-active")

	return overlay
}

func (overlay *KnockoutOverlay) hitReceived(cursor *graphics.Cursor, judgementResult osu.JudgementResult, score osu.Score) {
	if judgementResult.HitResult == osu.PositionalMiss {
		return
	}

	player := overlay.players[overlay.names[cursor]]

	if overlay.controller.GetRuleset().GetBeatMap().Diff.Mods.Active(difficulty.HardRock) != overlay.controller.GetReplays()[player.oldIndex].ModsV.Active(difficulty.HardRock) {
		judgementResult.Position.Y = 384 - judgementResult.Position.Y
	}

	player.score = score.Score
	player.pp = score.PP.Total

	player.scoreDisp.SetValue(float64(player.score), false)
	player.ppDisp.SetValue(player.pp, false)

	player.perObjectStats[judgementResult.Number].score = player.score
	player.perObjectStats[judgementResult.Number].pp = player.pp
	player.perObjectStats[judgementResult.Number].accuracy = score.Accuracy

	player.accDisp.SetValue(score.Accuracy*100, false)

	if judgementResult.ComboResult == osu.Increase {
		player.sCombo++
	}

	resultClean := judgementResult.HitResult & osu.BaseHitsM

	acceptableHits := resultClean&(osu.Hit100|osu.Hit50|osu.Miss) > 0
	if acceptableHits {
		player.fadeHit.Reset()
		player.fadeHit.AddEventS(overlay.normalTime, overlay.normalTime+300, 0.5, 1)
		player.fadeHit.AddEventS(overlay.normalTime+600, overlay.normalTime+900, 1, 0)
		player.scaleHit.AddEventS(overlay.normalTime, overlay.normalTime+300, 0.5, 1)
		player.lastHit = judgementResult.HitResult & (osu.HitValues | osu.Miss) //resultClean
		if settings.Knockout.Mode == settings.OneVsOne {
			overlay.deathBubbles = append(overlay.deathBubbles, newBubble(judgementResult.Position, overlay.normalTime, overlay.names[cursor], player.sCombo, resultClean, judgementResult.ComboResult))
		}
	}

	comboBreak := judgementResult.ComboResult == osu.Reset
	if (settings.Knockout.Mode == settings.SSOrQuit && (acceptableHits || comboBreak)) || (comboBreak && judgementResult.Number != 0) {
		if !player.hasBroken {
			if settings.Knockout.Mode == settings.XReplays {
				if player.sCombo >= int64(settings.Knockout.BubbleMinimumCombo) {
					overlay.deathBubbles = append(overlay.deathBubbles, newBubble(judgementResult.Position, overlay.normalTime, overlay.names[cursor], player.sCombo, resultClean, judgementResult.ComboResult))
					log.Println(overlay.names[cursor], "has broken! Combo:", player.sCombo)
				}
			} else if (settings.Knockout.Mode == settings.SSOrQuit ||
				(settings.Knockout.Mode == settings.ComboBreak && judgementResult.Time > int64(settings.Knockout.GraceEndTime*1000)) ||
				(settings.Knockout.Mode == settings.MaxCombo && math.Abs(float64(player.sCombo-player.maxCombo)) < 5)) &&
				overlay.alivePlayers > settings.Knockout.MinPlayers {
				//Fade out player name
				player.hasBroken = true
				player.breakTime = judgementResult.Time

				overlay.alivePlayers--

				player.fade.AddEvent(overlay.normalTime, overlay.normalTime+3000, 0)

				player.height.SetEasing(easing.OutQuad)
				player.height.AddEvent(overlay.normalTime+2500, overlay.normalTime+3000, 0)

				overlay.deathBubbles = append(overlay.deathBubbles, newBubble(judgementResult.Position, overlay.normalTime, overlay.names[cursor], player.sCombo, resultClean, judgementResult.ComboResult))

				log.Println(overlay.names[cursor], "has broken! Max combo:", player.sCombo)
			}
		}
	}

	if comboBreak {
		player.sCombo = 0
	}
}

func (overlay *KnockoutOverlay) Update(time float64) {
	if overlay.audioTime == 0 {
		overlay.audioTime = time
		overlay.normalTime = time
	}

	delta := time - overlay.audioTime

	if overlay.music != nil && overlay.music.GetState() == bass.MusicPlaying {
		delta /= overlay.music.GetSpeed()
	}

	overlay.normalTime += delta

	overlay.audioTime = time

	overlay.updateBreaks(overlay.normalTime)
	overlay.fade.Update(overlay.normalTime)

	for _, r := range overlay.controller.GetReplays() {
		player := overlay.players[r.Name]
		player.height.Update(overlay.normalTime)
		player.fade.Update(overlay.normalTime)
		player.fadeHit.Update(overlay.normalTime)
		player.scaleHit.Update(overlay.normalTime)
		player.index.Update(overlay.normalTime)
		player.scoreDisp.Update(overlay.normalTime)
		player.ppDisp.Update(overlay.normalTime)
		player.accDisp.Update(overlay.normalTime)
		player.lastCombo = r.Combo

		currentHp := overlay.controller.GetRuleset().GetHP(overlay.controller.GetCursors()[player.oldIndex])

		if player.displayHp < currentHp {
			player.displayHp = min(1.0, player.displayHp+math.Abs(currentHp-player.displayHp)/4*delta/16.667)
		} else if player.displayHp > currentHp {
			player.displayHp = max(0.0, player.displayHp-math.Abs(player.displayHp-currentHp)/6*delta/16.667)
		}
	}
}

func (overlay *KnockoutOverlay) SetMusic(music bass.ITrack) {
	overlay.music = music
}

func (overlay *KnockoutOverlay) DrawBackground(batch *batch.QuadBatch, _ []color2.Color, alpha float64) {
	alpha *= overlay.fade.GetValue()
	overlay.boundaries.Draw(batch.Projection, float32(overlay.controller.GetBeatMap().Diff.CircleRadius), float32(alpha))
}

func (overlay *KnockoutOverlay) DrawBeforeObjects(_ *batch.QuadBatch, _ []color2.Color, _ float64) {}

func (overlay *KnockoutOverlay) DrawNormal(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	alpha *= overlay.fade.GetValue()

	scl := 384.0 * (1080.0 / 900.0 * 0.9) / (51)

	alive := 0
	for _, r := range overlay.controller.GetReplays() {
		player := overlay.players[r.Name]
		if !player.hasBroken {
			alive++
		}
	}

	for i := 0; i < len(overlay.deathBubbles); i++ {
		batch.ResetTransform()
		bubble := overlay.deathBubbles[i]
		bubble.deathFade.Update(overlay.normalTime)
		bubble.deathSlide.Update(overlay.normalTime)
		bubble.deathScale.Update(overlay.normalTime)

		if bubble.deathFade.GetValue() >= 0.01 {
			if settings.Knockout.Mode == settings.OneVsOne {
				val := strconv.Itoa(int(bubble.lastHit.ScoreValue()))
				if bubble.lastCombo == osu.Reset {
					val = "X"
				}

				rep := overlay.players[bubble.name]
				batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*bubble.deathFade.GetValue())
				width := overlay.font.GetWidth(scl*bubble.deathScale.GetValue(), val)
				overlay.font.Draw(batch, bubble.deathX-width/2, bubble.deathSlide.GetValue()+scl*bubble.deathScale.GetValue()/3, scl*bubble.deathScale.GetValue(), val)
			} else {
				rep := overlay.players[bubble.name]
				batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*bubble.deathFade.GetValue())
				width := overlay.font.GetWidth(scl, bubble.name)
				overlay.font.Draw(batch, bubble.deathX-width/2, bubble.deathSlide.GetValue()-scl/2, scl, bubble.name)

				batch.SetColor(1, 1, 1, alpha*bubble.deathFade.GetValue())

				if bubble.lastCombo == osu.Reset {
					combo := fmt.Sprintf("%dx", bubble.combo)
					comboWidth := overlay.font.GetWidth(scl*0.8, combo)
					overlay.font.Draw(batch, bubble.deathX-comboWidth/2, bubble.deathSlide.GetValue()+scl*0.8/2, scl*0.8, combo)
				} else {
					switch bubble.lastHit {
					case osu.Hit100:
						batch.SetSubScale(scl*(float64(graphics.Hit100.Width)/float64(graphics.Hit100.Height))/2, scl/2)
						batch.SetTranslation(vector.NewVec2d(bubble.deathX, bubble.deathSlide.GetValue() /*- scl*0.8*/))
						batch.DrawUnit(*graphics.Hit100)
					case osu.Hit50:
						batch.SetSubScale(scl*(float64(graphics.Hit50.Width)/float64(graphics.Hit50.Height))/2, scl/2)
						batch.SetTranslation(vector.NewVec2d(bubble.deathX, bubble.deathSlide.GetValue()-scl*0.8))
						batch.DrawUnit(*graphics.Hit50)
					}
				}
			}
		}

		if bubble.endTime <= overlay.normalTime {
			overlay.deathBubbles = append(overlay.deathBubbles[:i], overlay.deathBubbles[i+1:]...)
			i--
		}
	}

	minSize := settings.Knockout.MinCursorSize
	maxSize := settings.Knockout.MaxCursorSize
	settings.Cursor.CursorSize = minSize + (maxSize-minSize)*math.Pow(1-math.Sin(float64(alive)/max(51, float64(settings.PLAYERS))*math.Pi/2), 3)

	batch.SetScale(1, 1)
}

func (overlay *KnockoutOverlay) DrawHUD(batch *batch.QuadBatch, colors []color2.Color, alpha float64) {
	alpha *= overlay.fade.GetValue()

	batch.ResetTransform()

	controller := overlay.controller
	replays := controller.GetReplays()

	scl := overlay.ScaledHeight * 0.9 / 51
	//margin := scl*0.02

	highestCombo := int64(0)
	highestPP := 0.0
	highestACC := 0.0
	highestScore := int64(0)
	cumulativeHeight := 0.0
	maxPlayerWidth := 0.0

	for _, r := range replays {
		cumulativeHeight += overlay.players[r.Name].height.GetValue()

		highestCombo = max(highestCombo, overlay.players[r.Name].sCombo)
		highestPP = max(highestPP, overlay.players[r.Name].pp)
		highestACC = max(highestACC, r.Accuracy*100)
		highestScore = max(highestScore, overlay.players[r.Name].score)

		pWidth := overlay.font.GetWidth(scl, r.Name)

		if r.Mods != "" {
			pWidth += overlay.font.GetWidth(scl*0.8, "+"+r.Mods)
		}

		maxPlayerWidth = max(maxPlayerWidth, pWidth)
	}

	//cL := strconv.FormatInt(highestCombo, 10)
	cP := strconv.FormatInt(int64(highestPP), 10)
	cA := strconv.FormatInt(int64(highestACC), 10)
	cS := overlay.font.GetWidthMonospaced(scl, utils.Humanize(highestScore))

	accuracy1 := cA + ".00% " + cP + ".00pp"
	nWidth := overlay.font.GetWidthMonospaced(scl, accuracy1)

	maxLength := 3.2*scl + nWidth + maxPlayerWidth

	xSlideLeft := (overlay.fade.GetValue() - 1.0) * maxLength
	xSlideRight := (1.0 - overlay.fade.GetValue()) * (cS + overlay.font.GetWidthMonospaced(scl, fmt.Sprintf("%dx ", highestCombo)) + 0.5*scl)

	rowPosY := max((overlay.ScaledHeight-cumulativeHeight)/2, scl)
	// Draw textures like keys, grade, hit values
	for _, rep := range overlay.playersArray {
		r := replays[rep.oldIndex]
		player := overlay.players[r.Name]

		rowBaseY := rowPosY + rep.index.GetValue()*(overlay.ScaledHeight*0.9*1.04/(51)) + player.height.GetValue()/2 /*+margin*10*/
		rowPosY -= overlay.ScaledHeight*0.9*1.04/(51) - player.height.GetValue()

		//batch.SetColor(0.1, 0.8, 0.4, alpha*player.fade.GetValue()*0.4)
		//add := 0.3 + float64(int(math.Round(rep.index.GetValue()))%2)*0.2
		//batch.SetColor(add, add, add, alpha*player.fade.GetValue()*0.7)
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
			batch.SetTranslation(vector.NewVec2d((float64(j)+0.5)*scl+xSlideLeft, rowBaseY))

			if controller.GetClick(rep.oldIndex, j) || controller.GetClick(rep.oldIndex, j+2) {
				batch.DrawUnit(*overlay.ButtonClicked)
			} else {
				batch.DrawUnit(*overlay.Button)
			}
		}

		width := overlay.font.GetWidth(scl, r.Name)

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		if r.Mods != "" {
			width += overlay.font.GetWidth(scl*0.8, "+"+r.Mods)
		}

		if r.Grade != osu.NONE {
			text := skin.GetTexture("ranking-" + r.Grade.TextureName() + "-small")

			ratio := 1.0 / 44.0 // default skin's grade height
			if text.Height < 44 {
				ratio = 1.0 / float64(text.Height) // if skin's grade is smaller, make it bigger
			}

			batch.SetSubScale(scl*0.9*ratio, scl*0.9*ratio)
			batch.SetTranslation(vector.NewVec2d(2.6*scl+nWidth+xSlideLeft, rowBaseY))

			batch.DrawTexture(*text)
		}

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue()*player.fadeHit.GetValue())
		//batch.SetSubScale(scl*0.9/2*player.scaleHit.GetValue(), scl*0.9/2*player.scaleHit.GetValue())
		//batch.SetTranslation(vector.NewVec2d(3*scl+width+nWidth+scl*0.5, rowBaseY))

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
				hitTexture := skin.GetTexture(tex)
				batch.SetSubScale(scl*0.8/2*player.scaleHit.GetValue()*(float64(hitTexture.Width)/float64(hitTexture.Height)), scl*0.8/2*player.scaleHit.GetValue())
				batch.SetTranslation(vector.NewVec2d(3.2*scl+width+nWidth+scl*(float64(hitTexture.Width)/float64(hitTexture.Height))*0.5+xSlideLeft, rowBaseY))
				batch.DrawUnit(*hitTexture)
			}
		}
	}

	batch.ResetTransform()

	rowPosY = max((overlay.ScaledHeight-cumulativeHeight)/2, scl)
	ascScl := overlay.font.GetAscent() * (scl / overlay.font.GetSize()) / 2

	// Draw texts
	for _, rep := range overlay.playersArray {
		r := replays[rep.oldIndex]
		player := overlay.players[r.Name]

		rowBaseY := rowPosY + rep.index.GetValue()*(overlay.ScaledHeight*0.9*1.04/(51)) + player.height.GetValue()/2 /*+margin*10*/
		rowPosY -= overlay.ScaledHeight*0.9*1.04/(51) - player.height.GetValue()

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		accuracy := fmt.Sprintf("%"+strconv.Itoa(len(cA)+3)+".2f%% %"+strconv.Itoa(len(cP)+3)+".2fpp", overlay.players[r.Name].accDisp.GetValue(), overlay.players[r.Name].ppDisp.GetValue())
		//_ = cL

		overlay.font.DrawOrigin(batch, 2*scl+xSlideLeft, rowBaseY, vector.CentreLeft, scl, true, accuracy)

		scorestr := utils.Humanize(int64(player.scoreDisp.GetValue()))

		sWC := fmt.Sprintf("%dx ", overlay.players[r.Name].sCombo)

		overlay.font.DrawOrigin(batch, overlay.ScaledWidth-cS-0.5*scl+xSlideRight, rowBaseY, vector.CentreRight, scl, true, sWC)
		overlay.font.DrawOrigin(batch, overlay.ScaledWidth-0.5*scl+xSlideRight, rowBaseY, vector.CentreRight, scl, true, scorestr)

		batch.SetColor(float64(colors[rep.oldIndex].R), float64(colors[rep.oldIndex].G), float64(colors[rep.oldIndex].B), alpha*player.fade.GetValue())
		overlay.font.DrawOrigin(batch, 3.2*scl+nWidth+xSlideLeft, rowBaseY, vector.CentreLeft, scl, false, r.Name)
		width := overlay.font.GetWidth(scl, r.Name)

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		if r.Mods != "" {
			overlay.font.DrawOrigin(batch, 3.2*scl+width+nWidth+xSlideLeft, rowBaseY+ascScl, vector.BottomLeft, scl*0.8, false, "+"+r.Mods)
		}
	}
}

func (overlay *KnockoutOverlay) IsBroken(cursor *graphics.Cursor) bool {
	return overlay.players[overlay.names[cursor]].hasBroken
}

func (overlay *KnockoutOverlay) updateBreaks(time float64) {
	inBreak := false

	for _, b := range overlay.controller.GetRuleset().GetBeatMap().Pauses {
		if overlay.audioTime < b.GetStartTime() {
			break
		}

		if b.GetEndTime()-b.GetStartTime() >= 1000 && overlay.audioTime >= b.GetStartTime() && overlay.audioTime <= b.GetEndTime() {
			inBreak = true

			break
		}
	}

	if !overlay.breakMode && inBreak {
		if settings.Knockout.HideOverlayOnBreaks {
			overlay.fade.AddEventEase(time, time+500, 0, easing.OutQuad)
		}
	} else if overlay.breakMode && !inBreak {
		overlay.fade.AddEventEase(time, time+500, 1, easing.OutQuad)
	}

	overlay.breakMode = inBreak
}

func (overlay *KnockoutOverlay) DisableAudioSubmission(_ bool) {}

func (overlay *KnockoutOverlay) ShouldDrawHUDBeforeCursor() bool {
	return false
}
