package drawables

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/graphics"
	"github.com/wieku/danser-go/framework/bass"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/math/vector"
	"math"
)

type Visualiser struct {
	jumpSize      int
	jumpCounter   int
	bars          int
	updateDelay   float64
	decayValue    float64
	barLength     float64
	Position      vector.Vector2d
	startDistance float64
	lastTime      float64
	counter       float64
	fft           []float64
	music         *bass.Track
}

func NewVisualiser(startDistance float64, barLength float64, position vector.Vector2d) *Visualiser {
	visualiser := &Visualiser{jumpSize: 5, bars: 200, updateDelay: 50, decayValue: 0.0024, barLength: barLength, Position: position, startDistance: startDistance}
	visualiser.fft = make([]float64, visualiser.bars)
	return visualiser
}

func (vis *Visualiser) SetStartDistance(distance float64) {
	vis.startDistance = distance
}

func (vis *Visualiser) SetTrack(track *bass.Track) {
	vis.music = track
}

func (vis *Visualiser) Update(time float64) {
	delta := time - vis.lastTime

	vis.counter += delta

	decay := delta * vis.decayValue

	if vis.counter >= vis.updateDelay {
		if vis.music != nil {
			fft := vis.music.GetFFT()

			for i := 0; i < vis.bars; i++ {
				value := float64(fft[(i+vis.jumpCounter)%vis.bars]) * 0.5 // * math.Pow(float64((i+vis.jumpCounter)%vis.bars+1), 0.08)
				if value > vis.fft[i] {
					vis.fft[i] = value
				}
			}

		}

		decay = 0
		vis.jumpCounter = (vis.jumpCounter + vis.jumpSize) % vis.bars
		vis.counter -= vis.updateDelay
	}

	for i := 0; i < vis.bars; i++ {
		vis.fft[i] -= (vis.fft[i] + 0.03) * decay
		if vis.fft[i] < 0 {
			vis.fft[i] = 0
		}
	}

	vis.lastTime = time
}

func (vis *Visualiser) Draw(time float64, batch *batch.QuadBatch) {
	origin := vector.NewVec2d(-1, 0)

	color := mgl32.Vec4{1, 1, 1, 0.3}
	region := graphics.Pixel.GetRegion()
	for i := 0; i < 5; i++ {
		for j, v := range vis.fft {
			rotation := (float64(i)/5 + float64(j)/float64(vis.bars)) * 2 * math.Pi
			position := vector.NewVec2dRad(rotation, vis.startDistance).Add(vis.Position)
			scale := vector.NewVec2d(vis.barLength*v, float64(2*math.Pi*vis.startDistance)/float64(vis.bars))
			batch.DrawStObject(position, origin, scale, false, false, rotation, color, false, region)
		}
	}
}
