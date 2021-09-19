package sprite

import (
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/math/vector"
)

type TextSprite struct {
	*Sprite

	fnt  *font.Font
	size float64

	text string

	monospaced bool
	overlap    float64
}

func NewTextSprite(text string, fnt *font.Font, depth float64, position vector.Vector2d, origin vector.Vector2d) *TextSprite {
	return NewTextSpriteSize(text, fnt, fnt.GetSize(), depth, position, origin)
}

func NewTextSpriteSize(text string, fnt *font.Font, size float64, depth float64, position vector.Vector2d, origin vector.Vector2d) *TextSprite {
	return &TextSprite{
		Sprite:  NewSpriteSingle(nil, depth, position, origin),
		fnt:     fnt,
		size:    size,
		text:    text,
		overlap: fnt.Overlap,
	}
}

func (sprite *TextSprite) SetText(text string) {
	sprite.text = text
}

func (sprite *TextSprite) SetSize(size float64) {
	sprite.size = size
}

func (sprite *TextSprite) ResetSize() {
	sprite.size = sprite.fnt.GetSize()
}

func (sprite *TextSprite) SetOverlap(overlap float64) {
	sprite.overlap = overlap
}

func (sprite *TextSprite) ResetOverlap() {
	sprite.overlap = sprite.fnt.Overlap
}

func (sprite *TextSprite) SetMonospaced(monospaced bool) {
	sprite.monospaced = monospaced
}

func (sprite *TextSprite) GetWidth() (width float64) {
	prevOverlap := sprite.fnt.Overlap
	sprite.fnt.Overlap = sprite.overlap

	if sprite.monospaced {
		width = sprite.fnt.GetWidthMonospaced(sprite.size*sprite.scale.Y, sprite.text)
	} else {
		width = sprite.fnt.GetWidth(sprite.size*sprite.scale.Y, sprite.text)
	}

	sprite.fnt.Overlap = prevOverlap

	return
}

func (sprite *TextSprite) Draw(_ float64, batch *batch.QuadBatch) {
	prevOverlap := sprite.fnt.Overlap

	sprite.fnt.Overlap = sprite.overlap

	sprite.fnt.DrawOriginRotationV(batch, sprite.position, sprite.origin, sprite.size*sprite.scale.Y, sprite.rotation, sprite.monospaced, sprite.text)

	sprite.fnt.Overlap = prevOverlap
}
