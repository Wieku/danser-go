package vector

import (
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/framework/math/math32"
)

type Vector2f struct {
	X, Y float32
}

func NewVec2f(x, y float32) Vector2f {
	return Vector2f{x, y}
}

func NewVec2fP(x, y float32) *Vector2f {
	return &Vector2f{x, y}
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
	return mgl32.Vec3{float32(v.X), float32(v.Y), 0}
}

func (v Vector2f) AsVec4() mgl32.Vec4 {
	return mgl32.Vec4{float32(v.X), float32(v.Y), 0, 1}
}

func (v *Vector2f) Set(x, y float32) {
	v.X = x
	v.Y = y
}

func (v *Vector2f) SetRad(rad, length float32) {
	v.X = math32.Cos(rad) * length
	v.Y = math32.Sin(rad) * length
}

func (v Vector2f) printOut() {
	fmt.Println("[", v.X, ":", v.Y, "]")
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
	return math32.Sqrt(math32.Pow(v1.X-v.X, 2) + math32.Pow(v1.Y-v.Y, 2))
}

func (v Vector2f) DstSq(v1 Vector2f) float32 {
	return math32.Pow(v1.X-v.X, 2) + math32.Pow(v1.Y-v.Y, 2)
}

func (v Vector2f) Angle() float32 {
	return v.AngleR() * 180 / math32.Pi
}

func (v Vector2f) AngleR() float32 {
	return math32.Atan2(v.Y, v.X)
}

func (v Vector2f) Nor() Vector2f {
	length := v.Len()

	if length == 0 {
		return Vector2f{v.X, v.Y}
	}

	return Vector2f{v.X / length, v.Y / length}
}

func (v Vector2f) AngleRV(v1 Vector2f) float32 {
	return math32.Atan2(v.Y-v1.Y, v.X-v1.X)
}

func (v Vector2f) Rotate(rad float32) Vector2f {
	cos := math32.Cos(rad)
	sin := math32.Sin(rad)
	return Vector2f{v.X*cos - v.Y*sin, v.X*sin + v.Y*cos}
}

func (v Vector2f) Len() float32 {
	return math32.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2f) LenSq() float32 {
	return v.X*v.X + v.Y*v.Y
}

func (v Vector2f) Scl(mag float32) Vector2f {
	return Vector2f{v.X * mag, v.Y * mag}
}

func (v Vector2f) SclOrDenorm(mag float32) Vector2f {
	v1 := v
	if mag > 1.0 || mag < -1.0 {
		v1 = v1.Nor()
	}
	return Vector2f{v1.X * mag, v1.Y * mag}
}

func (v Vector2f) Abs() Vector2f {
	return NewVec2f(math32.Abs(v.X), math32.Abs(v.Y))
}

func (v Vector2f) Copy() Vector2f {
	return Vector2f{v.X, v.Y}
}

func (v Vector2f) Copy64() Vector2d {
	return Vector2d{float64(v.X), float64(v.Y)}
}
