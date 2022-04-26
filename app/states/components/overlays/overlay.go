package overlays

import (
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	color2 "github.com/wieku/danser-go/framework/math/color"
)

type Overlay interface {
	Update(float64)
	SetMusic(bass.ITrack)
	DrawBackground(batch *batch.QuadBatch, colors []color2.Color, alpha float64)
	DrawBeforeObjects(batch *batch.QuadBatch, colors []color2.Color, alpha float64)
	DrawNormal(batch *batch.QuadBatch, colors []color2.Color, alpha float64)
	DrawHUD(batch *batch.QuadBatch, colors []color2.Color, alpha float64)
	IsBroken(cursor *graphics.Cursor) bool
	DisableAudioSubmission(b bool)
	ShouldDrawHUDBeforeCursor() bool
}
