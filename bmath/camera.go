package bmath

import "github.com/go-gl/mathgl/mgl32"

type Rectangle struct {
	MinX, MinY, MaxX, MaxY float64
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
	camera.screenRect.MinX = -float64(width) / 2
	camera.screenRect.MaxX = float64(width) / 2

	if yDown {
		camera.screenRect.MinY = float64(height) / 2
		camera.screenRect.MaxY = -float64(height) / 2
	} else {
		camera.screenRect.MinY = -float64(height) / 2
		camera.screenRect.MaxY = float64(height) / 2
	}

	if yDown {
		camera.projection = mgl32.Ortho(float32(camera.screenRect.MinX), float32(camera.screenRect.MaxX), float32(camera.screenRect.MinY), float32(camera.screenRect.MaxY), 1, -1)
	} else {
		camera.projection = mgl32.Ortho(float32(camera.screenRect.MinX), float32(camera.screenRect.MaxX), float32(camera.screenRect.MinY), float32(camera.screenRect.MaxY), -1, 1)
	}

	camera.rebuildCache = true
	camera.viewDirty = true
}

func (camera *Camera) SetOsuViewport(width, height int) {
	osuAspect := 512.0 / 384.0
	screenAspect := float64(width) / float64(height)

	if screenAspect > osuAspect {
		sh := (384.0 - float64(384.0)*900.0/1080.0) / 2
		sw := (512.0*screenAspect*900.0/1080.0 - 512.0) / 2
		camera.screenRect.MinX = -sw
		camera.screenRect.MaxX = 512.0 + sw

		camera.screenRect.MinY = 384.0 + sh
		camera.screenRect.MaxY = -sh
	}

	camera.projection = mgl32.Ortho(float32(camera.screenRect.MinX), float32(camera.screenRect.MaxX), float32(camera.screenRect.MinY), float32(camera.screenRect.MaxY), 1, -1)
	camera.rebuildCache = true
	camera.viewDirty = true
}

func (camera *Camera) SetViewportF(x, y, width, height int) {
	camera.screenRect.MinX = float64(x)
	camera.screenRect.MaxX = float64(width)
	camera.screenRect.MinY = float64(y)
	camera.screenRect.MaxY = float64(height)

	camera.projection = mgl32.Ortho(float32(camera.screenRect.MinX), float32(camera.screenRect.MaxX), float32(camera.screenRect.MinY), float32(camera.screenRect.MaxY), 1, -1)
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
	//mgl32.Vec4(2*screenPos.X32()/float32(camera.width)-1, 2*screenPos.Y32()/float32(camera.height)-1, 0, 1)
	return Vector2d{}
}

func (camera Camera) GetWorldRect() Rectangle {
	res := camera.invProjectionView.Mul4x1(mgl32.Vec4{-1.0, 1.0, 0.0, 1.0}) //.Add(mgl32.Vec4{256, 192, 0, 0})
	var rectangle Rectangle
	rectangle.MinX = float64(res[0])
	rectangle.MinY = float64(res[1])
	res = camera.invProjectionView.Mul4x1(mgl32.Vec4{1.0, -1.0, 0.0, 1.0}) //.Add(mgl32.Vec4{256, 192, 0, 0})
	rectangle.MaxX = float64(res[0])
	rectangle.MaxY = float64(res[1])
	if rectangle.MinY > rectangle.MaxY {
		a := rectangle.MinY
		rectangle.MinY, rectangle.MaxY = rectangle.MaxY, a
	}
	return rectangle
}
