package schedulers

import (
	"danser/beatmap/objects"
	"danser/bmath"
	"danser/render"
)

type Scheduler interface {
	Init(objects []objects.BaseObject, cursor *render.Cursor)
	//Update(time int64)

	/////////////////////////////////////////////////////////////////////////////////////////////////////
	// 添加更多参数
	Update(time int64, position bmath.Vector2d)
	/////////////////////////////////////////////////////////////////////////////////////////////////////
}
