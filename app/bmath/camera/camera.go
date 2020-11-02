package camera

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/math/vector"
)

const OsuWidth = 512.0
const OsuHeight = 384.0

type Rectangle struct {
	MinX, MinY, MaxX, MaxY float32
}

type Camera struct {
	screenRect        Rectangle
	projection        mgl32.Mat4
	view              mgl32.Mat4
	projectionView    mgl32.Mat4
	invProjectionView mgl32.Mat4

	viewDirty bool
	origin    vector.Vector2d
	position  vector.Vector2d
	rotation  float64
	scale     vector.Vector2d

	rebuildCache bool
	cache        []mgl32.Mat4
}

func NewCamera() *Camera {
	return &Camera{scale: vector.NewVec2d(1, 1)}
}

func (camera *Camera) SetViewport(width, height int, yDown bool) {
	camera.screenRect.MinX = -float32(width) / 2
	camera.screenRect.MaxX = float32(width) / 2

	if yDown {
		camera.screenRect.MinY = float32(height) / 2
		camera.screenRect.MaxY = -float32(height) / 2
	} else {
		camera.screenRect.MinY = -float32(height) / 2
		camera.screenRect.MaxY = float32(height) / 2
	}

	if yDown {
		camera.projection = mgl32.Ortho(camera.screenRect.MinX, camera.screenRect.MaxX, camera.screenRect.MinY, camera.screenRect.MaxY, 1, -1)
	} else {
		camera.projection = mgl32.Ortho(camera.screenRect.MinX, camera.screenRect.MaxX, camera.screenRect.MinY, camera.screenRect.MaxY, -1, 1)
	}

	camera.rebuildCache = true
	camera.viewDirty = true
}

func (camera *Camera) SetOsuViewport(width, height int, scale float64, offset bool) {
	baseScale := float64(height) / OsuHeight
	if OsuWidth/OsuHeight > float64(width)/float64(height) {
		baseScale = float64(width) / OsuWidth
	}

	scl := baseScale * 0.8 * scale

	shift := settings.Playfield.ShiftY
	if offset {
		shift = 8
	}

	camera.SetViewport(width, height, true)
	camera.SetOrigin(vector.NewVec2d(OsuWidth/2, OsuHeight/2))
	camera.SetPosition(vector.NewVec2d(settings.Playfield.ShiftX, shift).Scl(scl))
	camera.SetScale(vector.NewVec2d(scl, scl))
	camera.Update()

	camera.rebuildCache = true
	camera.viewDirty = true
}

func (camera *Camera) SetViewportF(x, y, width, height int) {
	camera.screenRect.MinX = float32(x)
	camera.screenRect.MaxX = float32(width)
	camera.screenRect.MinY = float32(y)
	camera.screenRect.MaxY = float32(height)

	camera.projection = mgl32.Ortho(camera.screenRect.MinX, camera.screenRect.MaxX, camera.screenRect.MinY, camera.screenRect.MaxY, 1, -1)
	camera.rebuildCache = true
	camera.viewDirty = true
}

func (camera *Camera) calculateView() {
	camera.view = mgl32.Translate3D(camera.position.X32(), camera.position.Y32(), 0).Mul4(mgl32.HomogRotate3DZ(float32(camera.rotation))).Mul4(mgl32.Scale3D(camera.scale.X32(), camera.scale.Y32(), 1)).Mul4(mgl32.Translate3D(camera.origin.X32(), camera.origin.Y32(), 0))
}

func (camera *Camera) SetPosition(pos vector.Vector2d) {
	camera.position = pos
	camera.viewDirty = true
}

func (camera *Camera) SetOrigin(pos vector.Vector2d) {
	camera.origin = pos.Scl(-1)
	camera.viewDirty = true
}

func (camera *Camera) SetScale(scale vector.Vector2d) {
	camera.scale = scale
	camera.viewDirty = true
}

func (camera *Camera) SetRotation(rad float64) {
	camera.rotation = rad
	camera.viewDirty = true
}

func (camera *Camera) Rotate(rad float64) {
	camera.rotation += rad
	camera.viewDirty = true
}

func (camera *Camera) Translate(pos vector.Vector2d) {
	camera.position = camera.position.Add(pos)
	camera.viewDirty = true
}

func (camera *Camera) Scale(scale vector.Vector2d) {
	camera.scale = camera.scale.Mult(scale)
	camera.viewDirty = true
}

func (camera *Camera) Update() {
	if camera.viewDirty {
		camera.calculateView()
		camera.projectionView = camera.projection.Mul4(camera.view)
		camera.invProjectionView = camera.projectionView.Inv()
		camera.rebuildCache = true
		camera.viewDirty = false
	}
}

func (camera *Camera) GenRotated(rotations int, rotOffset float64) []mgl32.Mat4 {

	if len(camera.cache) != rotations || camera.rebuildCache {
		if len(camera.cache) != rotations {
			camera.cache = make([]mgl32.Mat4, rotations)
		}

		pos := mgl32.Translate3D(camera.position.X32(), camera.position.Y32(), 0)
		view := mgl32.HomogRotate3DZ(float32(camera.rotation)).Mul4(mgl32.Scale3D(camera.scale.X32(), camera.scale.Y32(), 1)).Mul4(mgl32.Translate3D(camera.origin.X32(), camera.origin.Y32(), 0))

		for i := 0; i < rotations; i++ {
			camera.cache[i] = camera.projection.Mul4(pos).Mul4(mgl32.HomogRotate3DZ(float32(i) * float32(rotOffset))).Mul4(view)
		}
		camera.rebuildCache = false
	}

	return camera.cache
}

func (camera Camera) GetProjectionView() mgl32.Mat4 {
	return camera.projectionView
}

func (camera Camera) Unproject(screenPos vector.Vector2d) vector.Vector2d {
	res := camera.invProjectionView.Mul4x1(mgl32.Vec4{(screenPos.X32() + camera.screenRect.MinX) / camera.screenRect.MaxX, -(screenPos.Y32() + camera.screenRect.MaxY) / camera.screenRect.MinY, 0.0, 1.0})
	return vector.NewVec2d(float64(res[0]), float64(res[1]))
}

func (camera Camera) Project(worldPos vector.Vector2d) vector.Vector2d {
	res := camera.projectionView.Mul4x1(mgl32.Vec4{worldPos.X32(), worldPos.Y32(), 0.0, 1.0}).Add(mgl32.Vec4{1, 1, 0, 0}).Mul(0.5)
	//midX := camera.screenRect.MaxX-camera.screenRect.MinX
	return vector.NewVec2f(camera.screenRect.MinX+res[0]*(camera.screenRect.MaxX-camera.screenRect.MinX), camera.screenRect.MinY+res[1]*(camera.screenRect.MaxY-camera.screenRect.MinY)).Copy64()
}

func (camera Camera) GetWorldRect() Rectangle {
	res := camera.invProjectionView.Mul4x1(mgl32.Vec4{-1.0, 1.0, 0.0, 1.0})
	var rectangle Rectangle
	rectangle.MinX = res[0]
	rectangle.MinY = res[1]
	res = camera.invProjectionView.Mul4x1(mgl32.Vec4{1.0, -1.0, 0.0, 1.0})
	rectangle.MaxX = res[0]
	rectangle.MaxY = res[1]
	if rectangle.MinY > rectangle.MaxY {
		a := rectangle.MinY
		rectangle.MinY, rectangle.MaxY = rectangle.MaxY, a
	}
	return rectangle
}
