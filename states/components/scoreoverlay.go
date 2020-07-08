package components

import (
	"github.com/wieku/danser-go/render/batches"
	"github.com/wieku/danser-go/bmath"
	"github.com/wieku/danser-go/render"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/render/font"
	"github.com/wieku/danser-go/animation"
	"github.com/wieku/danser-go/rulesets/osu"
	"github.com/wieku/danser-go/settings"
	"github.com/wieku/danser-go/render/texture"
	"math/rand"
)

const (
	FadeIn   = 120
	FadeOut  = 600
	PostEmpt = 500
)

type Overlay interface {
	Update(int64)
	DrawNormal(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	DrawHUD(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64)
	IsBroken(cursor *render.Cursor) bool
	NormalBeforeCursor() bool
}

type PseudoSprite struct {
	texture   *texture.TextureRegion
	fade      *animation.Glider
	scale     *animation.Glider
	rotate    *animation.Glider
	slideDown *animation.Glider
	toRemove  *animation.Glider
	position  bmath.Vector2d
}

func newSprite(time int64, result osu.HitResult, position bmath.Vector2d) *PseudoSprite {
	sprite := new(PseudoSprite)
	switch result {
	case osu.HitResults.Hit100:
		sprite.texture = render.Hit100
	case osu.HitResults.Hit50:
		sprite.texture = render.Hit50
	case osu.HitResults.Miss:
		sprite.texture = render.Hit0
	default:
		return nil
	}

	sprite.fade = animation.NewGlider(0.0)
	sprite.fade.AddEventS(float64(time), float64(time+FadeIn), 0.0, 1.0)
	sprite.fade.AddEventS(float64(time+PostEmpt), float64(time+PostEmpt+FadeOut), 1.0, 0.0)

	sprite.scale = animation.NewGlider(0.0)
	sprite.scale.AddEventS(float64(time), float64(time+FadeIn*0.8), 0.6, 1.1)
	sprite.scale.AddEventS(float64(time+FadeIn), float64(time+FadeIn*1.2), 1.1, 0.9)
	sprite.scale.AddEventS(float64(time+FadeIn*1.2), float64(time+FadeIn*1.4), 0.9, 1.0)

	sprite.rotate = animation.NewGlider(0.0)

	if result == osu.HitResults.Miss {
		rotation := rand.Float64()*0.3 - 0.15
		sprite.rotate.AddEventS(float64(time), float64(time+FadeIn), 0.0, rotation)
		sprite.rotate.AddEventS(float64(time+FadeIn), float64(time+PostEmpt+FadeOut), rotation, rotation*2)
	}

	sprite.slideDown = animation.NewGlider(0.0)

	if result == osu.HitResults.Miss {
		sprite.slideDown.AddEventS(float64(time), float64(time+PostEmpt+FadeOut), -5, 40)
	}

	sprite.toRemove = animation.NewGlider(0.0)
	sprite.toRemove.AddEventS(float64(time+PostEmpt+FadeOut), float64(time+PostEmpt+FadeOut), 0, 1)
	sprite.position = position
	return sprite
}

func (sprite *PseudoSprite) Update(time int64) {
	sprite.fade.Update(float64(time))
	sprite.scale.Update(float64(time))
	sprite.rotate.Update(float64(time))
	sprite.toRemove.Update(float64(time))
	sprite.slideDown.Update(float64(time))
}

func (sprite *PseudoSprite) Draw(batch *batches.SpriteBatch) bool {
	batch.SetColor(1, 1, 1, sprite.fade.GetValue())
	batch.SetRotation(sprite.rotate.GetValue())
	proportions := float64(sprite.texture.Width) / float64(sprite.texture.Height)
	batch.SetSubScale(sprite.scale.GetValue()*20*proportions, sprite.scale.GetValue()*20)
	batch.SetTranslation(sprite.position.AddS(0, sprite.slideDown.GetValue()))

	batch.DrawUnit(*sprite.texture)

	batch.SetRotation(0)
	batch.SetSubScale(1, 1)
	return sprite.toRemove.GetValue() > 0.5
}

/*type knockoutPlayer struct {
	fade      *animation.Glider
	slide     *animation.Glider
	height    *animation.Glider
	lastCombo int64
	hasBroken bool

	lastHit  osu.HitResult
	fadeHit  *animation.Glider
	scaleHit *animation.Glider

	deathFade  *animation.Glider
	deathSlide *animation.Glider
	deathX     float64
}*/

type ScoreOverlay struct {
	//controller *dance.ReplayController
	font *font.Font
	//players    map[string]*knockoutPlayer
	//names      map[*render.Cursor]string
	lastTime int64
	//deaths     map[int64]int64
	//generator *rand.Rand
	combo         int64
	newCombo      int64
	newComboScale *animation.Glider
	newComboScaleB *animation.Glider
	oldScore int64
	scoreGlider *animation.Glider
	ruleset       *osu.OsuRuleSet
	cursor        *render.Cursor
	sprites       []*PseudoSprite
}

func NewScoreOverlay(ruleset *osu.OsuRuleSet, cursor *render.Cursor) *ScoreOverlay {
	overlay := new(ScoreOverlay)
	//overlay.controller = replayController
	overlay.ruleset = ruleset
	overlay.cursor = cursor
	overlay.font = font.GetFont("Exo 2 Bold")
	overlay.newComboScale = animation.NewGlider(1)
	overlay.newComboScaleB = animation.NewGlider(1)
	overlay.scoreGlider = animation.NewGlider(0)
	/*overlay.players = make(map[string]*knockoutPlayer)
	overlay.names = make(map[*render.Cursor]string)
	overlay.generator = rand.New(rand.NewSource(replayController.GetBeatMap().TimeAdded))*/
	//overlay.deaths = make(map[int64]int64)

	/*for i, r := range replayController.GetReplays() {
		overlay.names[replayController.GetCursors()[i]] = r.Name
		overlay.players[r.Name] = &knockoutPlayer{animation.NewGlider(1), animation.NewGlider(0), animation.NewGlider(settings.Graphics.GetHeightF() * 0.9 * 1.04 / (51)), 0, false, osu.HitResults.Hit300, animation.NewGlider(0), animation.NewGlider(0), animation.NewGlider(0), animation.NewGlider(0), 0}
	}*/
	ruleset.SetListener(func(cursor *render.Cursor, time int64, number int64, position bmath.Vector2d, result osu.HitResult, comboResult osu.ComboResult, pp float64, score1 int64) {
		/*player := overlay.players[overlay.names[cursor]]

		if result == osu.HitResults.Hit100 || result == osu.HitResults.Hit50 || result == osu.HitResults.Miss {
			player.fadeHit.Reset()
			player.fadeHit.AddEventS(float64(time), float64(time+300), 0.5, 1)
			player.fadeHit.AddEventS(float64(time+600), float64(time+900), 1, 0)
			player.scaleHit.AddEventS(float64(time), float64(time+300), 0.5, 1)
			player.lastHit = result
		}*/

		if result == osu.HitResults.Hit100 || result == osu.HitResults.Hit50 || result == osu.HitResults.Miss {
			overlay.sprites = append(overlay.sprites, newSprite(time, result, position))
		}

		if comboResult == osu.ComboResults.Increase {
			overlay.combo = overlay.newCombo
			overlay.newCombo++
			overlay.newComboScaleB.AddEventS(float64(time), float64(time+300), 2, 1.0)
			overlay.newComboScale.AddEventS(float64(time), float64(time+200), 1.1, 1.0)
		} else if comboResult == osu.ComboResults.Reset {
			overlay.newCombo = 0
			overlay.combo = 0
		}

		_, _, score, _ := overlay.ruleset.GetResults(overlay.cursor)

		overlay.scoreGlider.AddEventS(float64(time), float64(time+1000), float64(overlay.oldScore), float64(score))
		overlay.oldScore = score
	})
	return overlay
}

func (overlay *ScoreOverlay) Update(time int64) {
	for sTime := overlay.lastTime + 1; sTime <= time; sTime++ {
		overlay.newComboScale.Update(float64(sTime))
		overlay.newComboScaleB.Update(float64(sTime))
		overlay.scoreGlider.Update(float64(sTime))
	}
	if overlay.combo != overlay.newCombo && overlay.newComboScale.GetValue() < 1.01 {
		overlay.combo = overlay.newCombo
	}

	for i := 0; i < len(overlay.sprites); i++ {
		s := overlay.sprites[i]
		s.Update(time)
	}

	overlay.lastTime = time
}

func (overlay *ScoreOverlay) DrawNormal(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	batch.SetScale(1, 1)
	for i := 0; i < len(overlay.sprites); i++ {
		s := overlay.sprites[i]

		if ok := s.Draw(batch); ok {
			overlay.sprites = append(overlay.sprites[:i], overlay.sprites[i+1:]...)
			i--
		}
	}

	//scl := /*settings.Graphics.GetHeightF() * 0.9*(900.0/1080.0)*/ 384.0*(1080.0/900.0*0.9) / (51)
	//batch.SetScale(1, -1)
	//rescale := /*384.0/512.0 * (1080.0/settings.Graphics.GetHeightF())*/ 1.0
	//for i, r := range overlay.controller.GetReplays() {
	//	player := overlay.players[r.Name]
	//	if player.deathFade.GetValue() >= 0.01 {
	//
	//		batch.SetColor(float64(colors[i].X()), float64(colors[i].Y()), float64(colors[i].Z()), alpha*player.deathFade.GetValue())
	//		width := overlay.font.GetWidth(scl*rescale, r.Name)
	//		overlay.font.Draw(batch, player.deathX-width/2, player.deathSlide.GetValue(), scl*rescale, r.Name)
	//
	//		batch.SetColor(1, 1, 1, alpha*player.deathFade.GetValue())
	//		batch.SetSubScale(scl/2*rescale, scl/2*rescale)
	//		batch.SetTranslation(bmath.NewVec2d(player.deathX+width/2+scl*0.5*rescale, player.deathSlide.GetValue()-scl*0.5*rescale))
	//		batch.DrawUnit(*render.Hit0)
	//	}
	//
	//}
	//batch.SetScale(1, 1)
}

func (overlay *ScoreOverlay) DrawHUD(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	//controller := overlay.controller
	//replays := controller.GetReplays()

	scale := settings.Graphics.GetHeightF() / 1080.0

	batch.SetColor(0.7, 0.7, 0.7, 0.5)
	overlay.font.Draw(batch, 10, 10, scale*64*overlay.newComboScaleB.GetValue(), fmt.Sprintf("%dx", overlay.newCombo))
	batch.SetColor(0.7, 0.7, 0.7, 1)
	overlay.font.Draw(batch, 10, 10, scale*overlay.newComboScale.GetValue()*64, fmt.Sprintf("%dx", overlay.combo))

	acc, _, _, _ := overlay.ruleset.GetResults(overlay.cursor)

	accText := fmt.Sprintf("%0.2f%%", acc)

	scoreText := fmt.Sprintf("%08d", int64(overlay.scoreGlider.GetValue()))

	overlay.font.Draw(batch, settings.Graphics.GetWidthF()-overlay.font.GetWidth(scale*96, scoreText), settings.Graphics.GetHeightF()-96, scale*96, scoreText)
	overlay.font.DrawCentered(batch, settings.Graphics.GetWidthF()/2, settings.Graphics.GetHeightF()-32, scale*32, accText)

	//scl := settings.Graphics.GetHeightF() * 0.9 / 51
	//margin := scl*0.02

	//highestCombo := int64(0)
	//cumulativeHeight := 0.0
	//for _, r := range replays {
	//	cumulativeHeight += overlay.players[r.Name].height.GetValue()
	//	if r.Combo > highestCombo {
	//		highestCombo = r.Combo
	//	}
	//}
	//
	//rowPosY := settings.Graphics.GetHeightF() - (settings.Graphics.GetHeightF()-cumulativeHeight)/2
	//
	//cL := strconv.FormatInt(highestCombo, 10)
	//
	//for i, r := range replays {
	//	player := overlay.players[r.Name]
	//	batch.SetColor(float64(colors[i].X()), float64(colors[i].Y()), float64(colors[i].Z()), alpha*player.fade.GetValue())
	//
	//	rowBaseY := rowPosY - player.height.GetValue()/2 /*+margin*10*/
	//
	//	for j := 0; j < 4; j++ {
	//		if controller.GetClick(i, j) {
	//			batch.SetSubScale(scl*0.9/2, scl*0.9/2)
	//			batch.SetTranslation(bmath.NewVec2d((float64(j)+0.5)*scl, /*rowPosY*/ rowBaseY))
	//			batch.DrawUnit(render.Pixel.GetRegion())
	//		}
	//	}
	//
	//	batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())
	//
	//	accuracy := fmt.Sprintf("%6.2f%% %"+strconv.Itoa(len(cL))+"dx", r.Accuracy, r.Combo)
	//	accuracy1 := "100.00% " + cL + "x "
	//	nWidth := overlay.font.GetWidthMonospaced(scl, accuracy1)
	//
	//	overlay.font.DrawMonospaced(batch, 3*scl, rowBaseY-scl*0.8/2, scl, accuracy)
	//
	//	batch.SetSubScale(scl*0.9/2, -scl*0.9/2)
	//	batch.SetTranslation(bmath.NewVec2d(3*scl+nWidth, rowBaseY))
	//	batch.DrawUnit(*render.GradeTexture[int64(r.Grade)])
	//
	//	batch.SetColor(float64(colors[i].X()), float64(colors[i].Y()), float64(colors[i].Z()), alpha*player.fade.GetValue())
	//	overlay.font.Draw(batch, 4*scl+nWidth, rowBaseY-scl*0.8/2, scl, r.Name)
	//	width := overlay.font.GetWidth(scl, r.Name)
	//
	//	if r.Mods != "" {
	//		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())
	//		overlay.font.Draw(batch, 4*scl+width+nWidth, rowBaseY-scl*0.8/2, scl*0.8, "+"+r.Mods)
	//		width += overlay.font.GetWidth(scl*0.8, "+"+r.Mods)
	//	}
	//
	//	batch.SetColor(1, 1, 1, alpha*player.fade.GetValue()*player.fadeHit.GetValue())
	//	batch.SetSubScale(scl*0.9/2*player.scaleHit.GetValue(), -scl*0.9/2*player.scaleHit.GetValue())
	//	batch.SetTranslation(bmath.NewVec2d(4*scl+width+nWidth+scl*0.5, rowBaseY))
	//
	//	switch player.lastHit {
	//	case osu.HitResults.Hit100:
	//		batch.DrawUnit(*render.Hit100)
	//	case osu.HitResults.Hit50:
	//		batch.DrawUnit(*render.Hit50)
	//	case osu.HitResults.Miss:
	//		batch.DrawUnit(*render.Hit0)
	//	}
	//
	//	rowPosY -= player.height.GetValue()
	//}
}

func (overlay *ScoreOverlay) IsBroken(cursor *render.Cursor) bool {
	return false
}

func (overlay *ScoreOverlay) NormalBeforeCursor() bool {
	return true
}