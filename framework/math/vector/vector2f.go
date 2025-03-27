package vector

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/framework/math/math32"
	. "github.com/wieku/danser-go/framework/math/math87"
)

const epsilon = 0.00001

type Vector2f struct {
	X, Y float32
}

func NewVec2f(x, y float32) Vector2f {
	return Vector2f{x, y}
}

func NewVec2fRad(rad, length float32) Vector2f {
	return Vector2f{math32.Cos(rad) * length, math32.Sin(rad) * length}
}

func (v Vector2f) X64() float64 {
	return float64(v.X)
}

func (v Vector2f) Y64() float64 {
	return float64(v.Y)
}

func (v Vector2f) AsVec3() mgl32.Vec3 {
	return mgl32.Vec3{v.X, v.Y, 0}
}

func (v Vector2f) AsVec4() mgl32.Vec4 {
	return mgl32.Vec4{v.X, v.Y, 0, 1}
}

func (v Vector2f) String() string {
	return fmt.Sprintf("%fx%f", v.X, v.Y)
}

func (v Vector2f) Add(v1 Vector2f) Vector2f {
	return Vector2f{v.X + v1.X, v.Y + v1.Y}
}

func (v Vector2f) AddS(x, y float32) Vector2f {
	return Vector2f{v.X + x, v.Y + y}
}

func (v Vector2f) Sub(v1 Vector2f) Vector2f {
	return Vector2f{v.X - v1.X, v.Y - v1.Y}
}

func (v Vector2f) SubS(x, y float32) Vector2f {
	return Vector2f{v.X - x, v.Y - y}
}

func (v Vector2f) Mult(v1 Vector2f) Vector2f {
	return Vector2f{v.X * v1.X, v.Y * v1.Y}
}

func (v Vector2f) Mid(v1 Vector2f) Vector2f {
	return Vector2f{(v.X + v1.X) / 2, (v.Y + v1.Y) / 2}
}

func (v Vector2f) Dot(v1 Vector2f) float32 {
	return v.X*v1.X + v.Y*v1.Y
}

func (v Vector2f) Dst(v1 Vector2f) float32 {
	x := v1.X - v.X
	y := v1.Y - v.Y

	return math32.Sqrt(x*x + y*y)
}

// Dst87 is Dst but follows x87 promotion to double
func (v Vector2f) Dst87(v1 Vector2f) float32 { // dotnet framework why
	x := float64(v1.X - v.X)
	y := float64(v1.Y - v.Y)

	return math32.Sqrt(float32(x*x + y*y))
}

func (v Vector2f) DstSq(v1 Vector2f) float32 {
	x := v1.X - v.X
	y := v1.Y - v.Y

	return x*x + y*y
}

// DstSq87 is DstSq but follows x87 promotion to double
func (v Vector2f) DstSq87(v1 Vector2f) float32 {
	x := float64(v1.X - v.X)
	y := float64(v1.Y - v.Y)

	return float32(x*x + y*y)
}

func (v Vector2f) Angle() float32 {
	return v.AngleR() * 180 / math32.Pi
}

func (v Vector2f) AngleR() float32 {
	return math32.Atan2(v.Y, v.X)
}

// Nor - It could be X / sqrt but we need to introduce floating point errors from osu
func (v Vector2f) Nor() Vector2f {
	length := v.LenSq()

	if length < epsilon {
		return v
	}

	scale := 1.0 / math32.Sqrt(length)

	return Vector2f{v.X * scale, v.Y * scale}
}

// Nor87 is Nor but follows x87 promotion to double
func (v Vector2f) Nor87() Vector2f {
	length := v.LenSq87()

	if length < epsilon {
		return v
	}

	scale := Div87(1.0, math32.Sqrt(length))

	return Vector2f{Mul87(v.X, scale), Mul87(v.Y, scale)}
}

func (v Vector2f) AngleRV(v1 Vector2f) float32 {
	return math32.Atan2(v.Y-v1.Y, v.X-v1.X)
}

func (v Vector2f) Lerp(v1 Vector2f, t float32) Vector2f {
	return Vector2f{
		(v1.X-v.X)*t + v.X,
		(v1.Y-v.Y)*t + v.Y,
	}
}

func (v Vector2f) Rotate(rad float32) Vector2f {
	cos := math32.Cos(rad)
	sin := math32.Sin(rad)

	return Vector2f{
		v.X*cos - v.Y*sin,
		v.X*sin + v.Y*cos,
	}
}

func (v Vector2f) Len() float32 {
	return math32.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2f) LenSq() float32 {
	return v.X*v.X + v.Y*v.Y
}

// LenSq87 is LenSq but follows x87 promotion to double
func (v Vector2f) LenSq87() float32 {
	pX := float64(v.X)
	pY := float64(v.Y)
	return float32(pX*pX + pY*pY)
}

func (v Vector2f) Scl(mag float32) Vector2f {
	return Vector2f{v.X * mag, v.Y * mag}
}

// Scl87 is Scl but follows x87 promotion to double
func (v Vector2f) Scl87(mag float32) Vector2f {
	return Vector2f{Mul87(v.X, mag), Mul87(v.Y, mag)}
}

func (v Vector2f) Abs() Vector2f {
	return Vector2f{math32.Abs(v.X), math32.Abs(v.Y)}
}

func (v Vector2f) Copy() Vector2f {
	return Vector2f{v.X, v.Y}
}

func (v Vector2f) Copy64() Vector2d {
	return Vector2d{float64(v.X), float64(v.Y)}
}

func IsStraightLine32(a, b, c Vector2f) bool {
	return math32.Abs((b.Y-a.Y)*(c.X-a.X)-(b.X-a.X)*(c.Y-a.Y)) < 0.001
}

func AngleBetween32(centre, p1, p2 Vector2f) float32 { //nolint:misspell
	a := centre.Dst(p1)
	b := centre.Dst(p2)
	c := p1.Dst(p2)

	return math32.Acos((a*a + b*b - c*c) / (2 * a * b))
}
