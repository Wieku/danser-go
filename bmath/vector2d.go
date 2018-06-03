package bmath

import (
	"fmt"
	"math"
)

type Vector2d struct {
	X, Y float64
}

func NewVec2d(x, y float64) Vector2d {
	return Vector2d{x, y}
}

func NewVec2dP(x, y float64) *Vector2d {
	return &Vector2d{x, y}
}

func NewVec2dRad(rad, length float64) Vector2d {
	return Vector2d{math.Cos(rad) * length, math.Sin(rad) * length}
}

func (v Vector2d) X32() float32 {
	return float32(v.X)
}

func (v Vector2d) Y32() float32 {
	return float32(v.Y)
}

func (v *Vector2d) Set(x, y float64) {
	v.X = x
	v.Y = y
}

func (v *Vector2d) SetRad(rad, length float64) {
	v.X = math.Cos(rad) * length
	v.Y = math.Sin(rad) * length
}

func (v Vector2d) printOut() {
	fmt.Println("[", v.X, ":", v.Y, "]")
}

func (v Vector2d) Add(v1 Vector2d) Vector2d {
	return Vector2d{v.X+v1.X, v.Y+v1.Y}
}

func (v Vector2d) AddS(x, y float64) Vector2d {
	return Vector2d{v.X+x, v.Y+y}
}

func (v Vector2d) Sub(v1 Vector2d) Vector2d {
	return Vector2d{v.X-v1.X, v.Y-v1.Y}
}

func (v Vector2d) Mult(v1 Vector2d) Vector2d {
	return Vector2d{v.X*v1.X, v.Y*v1.Y}
}

func (v Vector2d) Mid(v1 Vector2d) Vector2d {
	return Vector2d{(v.X+v1.X)/2, (v.Y+v1.Y)/2}
}

func (v Vector2d) Dot(v1 Vector2d) float64 {
	return v.X*v1.X + v.Y*v1.Y
}

func (v Vector2d) Dst(v1 Vector2d) float64 {
	return math.Sqrt(math.Pow(v1.X-v.X,2) + math.Pow(v1.Y-v.Y,2))
}

func (v Vector2d) DstSq(v1 Vector2d) float64 {
	return math.Pow(v1.X-v.X,2) + math.Pow(v1.Y-v.Y,2)
}

func (v Vector2d) Angle() float64 {
	return v.AngleR() * 180 / math.Pi
}

func (v Vector2d) AngleR() float64 {
	return math.Atan2(v.Y, v.X)
}

func (v Vector2d) Nor() Vector2d {
	len := v.Len()
	return Vector2d{v.X / len, v.Y / len}
}

func (v Vector2d) AngleRV(v1 Vector2d) float64 {
	return math.Atan2(v.Y - v1.Y, v.X - v1.X)
}

func (v Vector2d) Rotate(rad float64) Vector2d {
	cos := math.Cos(rad)
	sin := math.Sin(rad)
	return Vector2d{v.X * cos - v.Y * sin, v.X * sin + v.Y * cos}
}

func (v Vector2d) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vector2d) Scl(mag float64) Vector2d {
	return Vector2d{v.X*mag, v.Y*mag}
}

func (v Vector2d) Copy() Vector2d {
	return Vector2d{v.X, v.Y}
}