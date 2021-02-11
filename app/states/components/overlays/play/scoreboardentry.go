package play

import (
	"fmt"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/graphics/font"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"strconv"
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
	entry := &ScoreboardEntry{
		Sprite: sprite.NewSpriteSingle(skin.GetTexture("menu-button-background"), 0, vector.NewVec2d(0, 0), bmath.Origin.CentreRight),
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
	entry.scoreHumanized = humanize(entry.score)
	entry.comboHumanized = fmt.Sprintf("%dx", entry.combo)
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

	fnt.Overlap = 4
	fnt.DrawMonospaced(batch, entryPos.X-padding+3.2, entryPos.Y+16, fnt.GetSize(), entry.scoreHumanized)

	if entry.rank <= 50 {
		batch.SetColor(1, 1, 1, a*0.32)

		fnt.Overlap = 4.8
		fnt.DrawMonospaced(batch, entryPos.X-10-fnt.GetWidthMonospaced(fnt.GetSize()*2.2, entry.rankHumanized), entryPos.Y-8, fnt.GetSize()*2.2, entry.rankHumanized)
	}

	batch.SetColor(0.6, 0.98, 1, a)

	fnt.Overlap = 4
	fnt.DrawMonospaced(batch, entryPos.X-10-fnt.GetWidthMonospaced(fnt.GetSize(), entry.comboHumanized), entryPos.Y+16, fnt.GetSize(), entry.comboHumanized)

	ubu := font.GetFont("Ubuntu Regular")
	ubu.Overlap = 2.5

	batch.SetScale(1, -1)

	batch.SetColor(0.1, 0.1, 0.1, a*0.8)
	ubu.Draw(batch, entryPos.X-padding+3.5, entryPos.Y-4.5, 20, entry.name)

	batch.SetColor(1, 1, 1, a)
	ubu.Draw(batch, entryPos.X-padding+3, entryPos.Y-5, 20, entry.name)

	ubu.Overlap = 0

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, 1)
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