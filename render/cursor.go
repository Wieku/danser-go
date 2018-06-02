package render

import (
	"danser/bmath"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/go-gl/gl/v3.3-core/gl"
	"math"
	"log"
)

type Cursor struct {
	Points []bmath.Vector2d
	Max int
	Position bmath.Vector2d
	LastPos bmath.Vector2d
	time float64
}

func NewCursor() *Cursor {
	return &Cursor{Max:1000, Position: bmath.NewVec2d(100, 100)}
}


func (cr *Cursor) SetPos(pt bmath.Vector2d) {
	cr.Position = pt
}

func (cr *Cursor) Update(tim float64) {
	points := int(cr.Position.Dst(cr.LastPos))*2

	cr.Points = append(cr.Points, cr.LastPos)

	for i:=1; i <= points; i++ {
		cr.Points = append(cr.Points, cr.Position.Sub(cr.LastPos).Scl(float64(i)/float64(points)).Add(cr.LastPos))
	}

	//cr.Points = append(cr.Points, cr.Position)

	cr.LastPos = cr.Position

	times := int64(float64(len(cr.Points))/(6*(60.0/tim))) + 1

	if len(cr.Points) > 0 {
		if int(times) < len(cr.Points) {
			cr.Points = cr.Points[times:]
		} else {
			cr.Points = cr.Points[len(cr.Points):]
		}
	}
}

func (cursor *Cursor) Draw(scale float64, batch *SpriteBatch, color mgl32.Vec4) {
	gl.Disable(gl.DEPTH_TEST)

	gl.ActiveTexture(gl.TEXTURE0)
	CursorTex.Begin()
	gl.ActiveTexture(gl.TEXTURE1)
	CursorTrail.Begin()
	gl.ActiveTexture(gl.TEXTURE2)
	CursorTop.Begin()

	batch.Begin()
	batch.SetTranslation(bmath.NewVec2d(0, 0))
	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), 0.05*float64(color[3]))
	for i, sl := range cursor.Points {

		batch.SetScale(scale*25 * (0.5+float64(i)/float64(len(cursor.Points))*0.4), scale*25 * (0.5+float64(i)/float64(len(cursor.Points))*0.4))
		batch.DrawUnit(sl, 1)

	}
	//color[3] = 1

	batch.SetScale(scale*27, scale*27)

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), float64(color[3]))
	batch.DrawUnit(cursor.Position, 0)
	batch.SetColor(1, 1, 1, math.Sqrt(float64(color[3])))
	batch.DrawUnit(cursor.Position, 2)

	batch.End()

	CursorTrail.End()
	CursorTex.End()
	CursorTop.End()

}

func (cursor *Cursor) DrawM(scale float64, batch *SpriteBatch, prevColor, color mgl32.Vec4) {
	gl.Disable(gl.DEPTH_TEST)

	gl.ActiveTexture(gl.TEXTURE0)
	CursorTex.Begin()
	gl.ActiveTexture(gl.TEXTURE1)
	CursorTrail.Begin()
	gl.ActiveTexture(gl.TEXTURE2)
	CursorTop.Begin()

	batch.Begin()
	batch.SetTranslation(bmath.NewVec2d(0, 0))
	for i, sl := range cursor.Points {
		batch.SetColor(float64(prevColor[0]), float64(prevColor[1]), float64(prevColor[2]), 0.05*float64(prevColor[3]))
		batch.SetScale(scale*28 * (0.5+float64(i)/float64(len(cursor.Points))*0.4), scale*28 * (0.5+float64(i)/float64(len(cursor.Points))*0.4))
		batch.DrawUnit(sl, 1)
		batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), 0.05*float64(color[3]))
		batch.SetScale(scale*24 * (0.5+float64(i)/float64(len(cursor.Points))*0.4), scale*24 * (0.5+float64(i)/float64(len(cursor.Points))*0.4))
		batch.DrawUnit(sl, 1)

	}
	//color[3] = 1

	batch.SetScale(scale*27, scale*27)

	batch.SetColor(float64(color[0]), float64(color[1]), float64(color[2]), float64(color[3]))
	batch.DrawUnit(cursor.Position, 0)
	batch.SetColor(1, 1, 1, math.Sqrt(float64(color[3])))
	batch.DrawUnit(cursor.Position, 2)

	batch.End()

	CursorTrail.End()
	CursorTex.End()
	CursorTop.End()

}