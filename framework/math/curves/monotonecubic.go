package curves

import (
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/math/vector"
	"sort"
)

// Code taken from https://en.wikipedia.org/wiki/Monotone_cubic_interpolation
// NOTE: It's not meant to be used for path interpolation, should be used only for graphs.

type MonotoneCubic struct {
	Points        []vector.Vector2f
	length        float32
	c1s, c2s, c3s []float32
}

func NewMonotoneCubic(points []vector.Vector2f) *MonotoneCubic {
	bz := &MonotoneCubic{
		Points: make([]vector.Vector2f, len(points)),
	}

	copy(bz.Points, points)

	sort.SliceStable(bz.Points, func(i, j int) bool {
		return bz.Points[i].X < bz.Points[j].X
	})

	ds := make([]vector.Vector2f, len(bz.Points)-1)
	ms := make([]float32, len(bz.Points)-1)

	for i := 0; i < len(bz.Points)-1; i++ {
		d := bz.Points[i+1].Sub(bz.Points[i])

		bz.length += d.X

		ds[i] = d
		ms[i] = d.Y / d.X
	}

	bz.c1s = append(bz.c1s, ms[0])

	for i := 0; i < len(ds)-1; i++ {
		m := ms[i]
		mNext := ms[i+1]

		if m*mNext <= 0 {
			bz.c1s = append(bz.c1s, 0)
		} else {
			dx_ := ds[i].X
			dxNext := ds[i+1].X
			common := dx_ + dxNext
			bz.c1s = append(bz.c1s, 3*common/((common+dxNext)/m+(common+dx_)/mNext))
		}
	}

	bz.c1s = append(bz.c1s, ms[len(ms)-1])

	for i := 0; i < len(bz.c1s)-1; i++ {
		c1 := bz.c1s[i]
		m_ := ms[i]
		invDx := 1 / ds[i].X
		common_ := c1 + bz.c1s[i+1] - m_ - m_
		bz.c2s = append(bz.c2s, (m_-c1-common_)*invDx)
		bz.c3s = append(bz.c3s, common_*invDx*invDx)
	}

	return bz
}

func (bz *MonotoneCubic) PointAt(t float32) vector.Vector2f {
	x := bz.Points[0].X + bz.length*mutils.Clamp(t, 0.0, 1.0)

	i := sort.Search(len(bz.c3s), func(i int) bool {
		return bz.Points[i].X >= x
	})
	i = mutils.Clamp(i-1, 0, len(bz.c3s)-1)

	// Interpolate
	diff := x - bz.Points[i].X
	diffSq := diff * diff

	return vector.NewVec2f(x, bz.Points[i].Y+bz.c1s[i]*diff+bz.c2s[i]*diffSq+bz.c3s[i]*diff*diffSq)
}

func (bz *MonotoneCubic) GetLength() float32 {
	return bz.length
}

func (bz *MonotoneCubic) GetStartAngle() float32 {
	return bz.Points[0].AngleRV(bz.PointAt(1.0 / bz.length))
}

func (bz *MonotoneCubic) GetEndAngle() float32 {
	return bz.Points[len(bz.Points)-1].AngleRV(bz.PointAt(1.0 - 1.0/bz.length))
}
