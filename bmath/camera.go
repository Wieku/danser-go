package bmath

import (
	"github.com/go-gl/mathgl/mgl32"
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
	origin    Vector2d
	position  Vector2d
	rotation  float64
	scale     Vector2d

	rebuildCache bool
	cache        []mgl32.Mat4
}

func NewCamera() *Camera {
	return &Camera{scale: NewVec2d(1, 1)}
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

	shift := 0.0
	if offset {
		shift = 8
	}

	camera.SetViewport(width, height, true)
	camera.SetOrigin(NewVec2d(512.0/2, 384.0/2-shift))
	camera.SetScale(NewVec2d(scl, scl))
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

func (camera *Camera) SetPosition(pos Vector2d) {
	camera.position = pos
	camera.viewDirty = true
}

func (camera *Camera) SetOrigin(pos Vector2d) {
	camera.origin = pos.Scl(-1)
	camera.viewDirty = true
}

func (camera *Camera) SetScale(scale Vector2d) {
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

func (camera *Camera) Translate(pos Vector2d) {
	camera.position = camera.position.Add(pos)
	camera.viewDirty = true
}

func (camera *Camera) Scale(scale Vector2d) {
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

		for i := 0; i < rotations; i++ {
			camera.cache[i] = camera.projection.Mul4(mgl32.HomogRotate3DZ(float32(i) * float32(rotOffset))).Mul4(camera.view)
		}
		camera.rebuildCache = false
	}

	return camera.cache
}

func (camera Camera) GetProjectionView() mgl32.Mat4 {
	return camera.projectionView
}

func (camera Camera) Unproject(screenPos Vector2d) Vector2d {
	res := camera.invProjectionView.Mul4x1(mgl32.Vec4{(screenPos.X32() + camera.screenRect.MinX) / camera.screenRect.MaxX, -(screenPos.Y32() + camera.screenRect.MaxY) / camera.screenRect.MinY, 0.0, 1.0})
	return NewVec2d(float64(res[0]), float64(res[1]))
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
