package components

import (
	"github.com/wieku/danser/dance"
	"github.com/wieku/danser/render/batches"
	"github.com/wieku/danser/settings"
	"strconv"
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/render"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser/render/font"
	"github.com/wieku/danser/animation"
)

type knockoutPlayer struct {
	fade *animation.Glider
	slide *animation.Glider
	height *animation.Glider
	lastCombo int64
	hasBroken bool
}

type KnockoutOverlay struct {
	controller *dance.ReplayController
	font *font.Font
	players map[string]*knockoutPlayer
	names map[*render.Cursor]string
	lastTime int64
}

func NewKnockoutOverlay(replayController *dance.ReplayController) *KnockoutOverlay {
	overlay := new(KnockoutOverlay)
	overlay.controller = replayController
	overlay.font = font.GetFont("Roboto Bold")
	overlay.players = make(map[string]*knockoutPlayer)
	overlay.names = make(map[*render.Cursor]string)

	for i, r := range replayController.GetReplays() {
		overlay.names[replayController.GetCursors()[i]] = r.Name
		overlay.players[r.Name] = &knockoutPlayer{animation.NewGlider(1), animation.NewGlider(0), animation.NewGlider(settings.Graphics.GetHeightF() * 0.9 * 1.04 / (51)), 0, false}
	}
	return overlay
}

func (overlay *KnockoutOverlay) Update(time int64) {

	for sTime:=overlay.lastTime; sTime <= time; sTime++ {
		for _, r := range overlay.controller.GetReplays() {
			player := overlay.players[r.Name]
			if r.Combo < player.lastCombo {
				player.fade.AddEvent(float64(sTime), float64(sTime+3000), 0)
				player.height.AddEvent(float64(sTime+2500), float64(sTime+3000), 0)
				player.hasBroken = true
			}
			player.height.Update(float64(sTime))
			player.fade.Update(float64(sTime))
			player.lastCombo = r.Combo
		}
	}
	overlay.lastTime = time
}

func (overlay *KnockoutOverlay) DrawHUD(batch *batches.SpriteBatch, colors []mgl32.Vec4, alpha float64) {
	controller := overlay.controller

	rpls := controller.GetReplays()

	scl := settings.Graphics.GetHeightF() * 0.9 / ( /*4**/ 51 /*/3*/)
	//margin := scl*0.02

	highest := int64(0)
	cumulativeHeight := 0.0
	for _, r := range rpls {
		cumulativeHeight += overlay.players[r.Name].height.GetValue()
		if r.Combo > highest {
			highest = r.Combo
		}
	}

	yHeight := settings.Graphics.GetHeightF() - (settings.Graphics.GetHeightF() - cumulativeHeight)/2

	cL := strconv.FormatInt(highest, 10)

	for i, r := range rpls {
		player := overlay.players[r.Name]
		batch.SetColor(float64(colors[i].X()), float64(colors[i].Y()), float64(colors[i].Z()), alpha*player.fade.GetValue())

		subY := yHeight-player.height.GetValue()/2/*+margin*10*/

		for j := 0; j < 4; j++ {
			if controller.GetClick(i, j) {
				batch.SetSubScale(scl*0.9/2, scl*0.9/2)
				batch.SetTranslation(bmath.NewVec2d((float64(j)+0.5)*scl, /*yHeight*/subY))
				batch.DrawUnit(render.Pixel.GetRegion())
			}
		}

		batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())

		accuracy := fmt.Sprintf("%6.2f%% %"+strconv.Itoa(len(cL))+"dx", r.Accuracy, r.Combo)
		accuracy1 := "100.00% "+cL+"x "
		nWidth := overlay.font.GetWidthMonospaced(scl, accuracy1)

		overlay.font.DrawMonospaced(batch, 3*scl, subY-scl*0.8/2, scl, accuracy)

		batch.SetSubScale(scl*0.9/2, -scl*0.9/2)
		batch.SetTranslation(bmath.NewVec2d(3*scl+nWidth, /*subY+scl/2*/subY))
		batch.DrawUnit(*render.GradeTexture[int64(r.Grade)])

		batch.SetColor(float64(colors[i].X()), float64(colors[i].Y()), float64(colors[i].Z()), alpha*player.fade.GetValue())
		overlay.font.Draw(batch, 4*scl+nWidth, subY-scl*0.8/2, scl, r.Name)
		if r.Mods != "" {
			width := overlay.font.GetWidth(scl, r.Name)
			batch.SetColor(1, 1, 1, alpha*player.fade.GetValue())
			overlay.font.Draw(batch, 4*scl+width+nWidth, subY-scl*0.8/2, scl*0.8, "+"+r.Mods)
		}

		yHeight-=player.height.GetValue()
	}
}

func (overlay *KnockoutOverlay) IsBroken(cursor *render.Cursor) bool {
	return overlay.players[overlay.names[cursor]].hasBroken
}


