package sprite

import (
	"github.com/wieku/danser-go/framework/graphics/batch"
	"math"
	"sort"
	"sync"
)

type SpriteManager struct {
	spriteQueue     []*Sprite
	spriteProcessed []*Sprite
	interArray      []*Sprite
	drawArray       []*Sprite
	visibleObjects  int
	interObjects    int
	allSprites      int

	mutex *sync.Mutex
	dirty bool
}

func NewSpriteManager() *SpriteManager {
	return &SpriteManager{mutex: &sync.Mutex{}}
}

func (layer *SpriteManager) Add(sprite *Sprite) {
	startTime := sprite.GetStartTime()
	if sprite.showForever {
		startTime = -math.MaxFloat64
	}

	n := sort.Search(len(layer.spriteQueue), func(j int) bool {
		return startTime < layer.spriteQueue[j].GetStartTime()
	})

	layer.spriteQueue = append(layer.spriteQueue, nil) //allocate bigger array in case when len=cap
	copy(layer.spriteQueue[n+1:], layer.spriteQueue[n:])

	layer.spriteQueue[n] = sprite
}

func (layer *SpriteManager) Update(time int64) {
	dirtyLocal := false
	toRemove := 0

	for i := 0; i < len(layer.spriteQueue); i++ {
		c := layer.spriteQueue[i]
		if float64(time) < c.GetStartTime() && !c.showForever {
			break
		}

		toRemove++
	}

	if toRemove > 0 {
		for i := 0; i < toRemove; i++ {
			s := layer.spriteQueue[i]

			n := sort.Search(len(layer.spriteProcessed), func(j int) bool {
				return s.GetDepth() < layer.spriteProcessed[j].GetDepth()
			})

			layer.spriteProcessed = append(layer.spriteProcessed, nil) //allocate bigger array in case when len=cap
			copy(layer.spriteProcessed[n+1:], layer.spriteProcessed[n:])

			layer.spriteProcessed[n] = s
		}

		dirtyLocal = true
		layer.spriteQueue = layer.spriteQueue[toRemove:]
	}

	for i := 0; i < len(layer.spriteProcessed); i++ {
		c := layer.spriteProcessed[i]
		c.Update(time)

		if float64(time) >= c.GetEndTime() && !c.showForever {
			copy(layer.spriteProcessed[i:], layer.spriteProcessed[i+1:])
			layer.spriteProcessed = layer.spriteProcessed[:len(layer.spriteProcessed)-1]

			dirtyLocal = true
			i--
		}
	}

	if dirtyLocal {
		layer.mutex.Lock()

		if len(layer.interArray) < len(layer.spriteProcessed) || len(layer.interArray) > len(layer.spriteProcessed)*3 {
			layer.interArray = make([]*Sprite, len(layer.spriteProcessed)*2)
		}

		layer.interObjects = len(layer.spriteProcessed)
		copy(layer.interArray, layer.spriteProcessed)

		layer.dirty = true

		layer.mutex.Unlock()
	}
}

func (layer *SpriteManager) GetNumRendered() (sum int) {
	for i := 0; i < layer.visibleObjects; i++ {
		if layer.drawArray[i] != nil && layer.drawArray[i].GetAlpha() >= 0.01 {
			sum++
		}
	}
	return
}

func (layer *SpriteManager) GetNumInQueue() int {
	return len(layer.spriteQueue)
}

func (layer *SpriteManager) GetNumProcessed() int {
	return len(layer.spriteProcessed)
}

func (layer *SpriteManager) GetLoad() (sum float64) {
	for i := 0; i < layer.visibleObjects; i++ {
		if layer.drawArray[i] != nil && layer.drawArray[i].GetAlpha() >= 0.01 {
			sum += layer.drawArray[i].GetLoad()
		}
	}
	return
}

func (layer *SpriteManager) Draw(time int64, batch *batch.QuadBatch) {
	layer.mutex.Lock()
	if layer.dirty {
		layer.visibleObjects = 0

		if len(layer.interArray) != len(layer.drawArray) {
			layer.drawArray = make([]*Sprite, len(layer.interArray))
		}

		copy(layer.drawArray, layer.interArray[:layer.interObjects])
		layer.visibleObjects = layer.interObjects

		layer.dirty = false
	}
	layer.mutex.Unlock()

	for i := 0; i < layer.visibleObjects; i++ {
		if layer.drawArray[i] != nil {
			layer.drawArray[i].Draw(time, batch)
		}
	}
}
