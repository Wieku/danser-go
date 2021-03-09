package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
)

type ScoreboardEntry struct {
	*sprite.Sprite

	name    string
	score   int64
	combo   int64
	rank    int
	visible bool

	scoreHumanized string
	comboHumanized string
	rankHumanized  string
}

func NewScoreboardEntry(name string, score int64, combo int64, rank int, isPlayer bool) *ScoreboardEntry {
	bg := skin.GetTexture("menu-button-background")
	entry := &ScoreboardEntry{
		Sprite: sprite.NewSpriteSingle(bg, 0, vector.NewVec2d(0, 0), vector.NewVec2d(float64(1-2*(bg.Width-470)/bg.Width), 0)),
		name:   name,
		score:  score,
		combo:  combo,
		rank:   rank,
	}

	entry.Sprite.SetScale(0.625)
	entry.SetAlpha(0)

	if isPlayer {
		entry.SetColor(color2.NewL(0.9))
	} else {
		entry.SetColor(color2.NewIRGB(31, 115, 153))
	}

	fnt := font.GetFont("Ubuntu Regular")
	fnt.Overlap = 2.5

	addDots := false
	for fnt.GetWidth(20, entry.name) > 130 {
		addDots = true
		entry.name = entry.name[:len(entry.name)-1]
	}

	if addDots {
		entry.name += "..."
	}

	fnt.Overlap = 0

	entry.UpdateData()

	return entry
}

func (entry *ScoreboardEntry) UpdateData() {
	entry.scoreHumanized = utils.Humanize(entry.score)
	entry.comboHumanized = utils.Humanize(entry.combo) + "x"
	entry.rankHumanized = fmt.Sprintf("%d", entry.rank)
}

func (entry *ScoreboardEntry) Draw(time float64, batch *batch.QuadBatch, alpha float64) {
	a := entry.Sprite.GetAlpha() * alpha

	if a < 0.01 {
		return
	}

	batch.SetColor(1, 1, 1, 0.6*alpha)

	entry.Sprite.Draw(time, batch)

	batch.SetColor(1, 1, 1, a)

	entryPos := entry.GetPosition()

	fnt := skin.GetFont("scoreentry")

	fnt.Overlap = 2.5
	fnt.DrawOrigin(batch, entryPos.X+3.2, entryPos.Y+8.8, bmath.Origin.TopLeft, fnt.GetSize(), true, entry.scoreHumanized)

	if entry.rank <= 50 {
		batch.SetColor(1, 1, 1, a*0.32)

		fnt.Overlap = 3
		fnt.DrawOrigin(batch, entryPos.X+padding-10, entryPos.Y-22, bmath.Origin.TopRight, fnt.GetSize()*2.2, true, entry.rankHumanized)
	}

	batch.SetColor(0.6, 0.98, 1, a)

	fnt.Overlap = 2.5
	fnt.DrawOrigin(batch, entryPos.X+padding-10, entryPos.Y+8.8, bmath.Origin.TopRight, fnt.GetSize(), true, entry.comboHumanized)

	ubu := font.GetFont("Ubuntu Regular")
	ubu.Overlap = 2.5

	batch.SetScale(1, -1)

	batch.SetColor(0.1, 0.1, 0.1, a*0.8)
	ubu.Draw(batch, entryPos.X+3.5, entryPos.Y-4.5, 20, entry.name)

	batch.SetColor(1, 1, 1, a)
	ubu.Draw(batch, entryPos.X+3, entryPos.Y-5, 20, entry.name)

	ubu.Overlap = 0

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, 1)
}
