package sprite

import (
	"github.com/wieku/danser-go/framework/graphics/batch"
	"math"
	"sort"
	"sync"
)

type Manager struct {
	spriteQueue     []ISprite
	spriteProcessed []ISprite
	interArray      []ISprite
	drawArray       []ISprite
	visibleObjects  int
	interObjects    int

	mutex *sync.Mutex
	dirty bool
}

func NewManager() *Manager {
	return &Manager{mutex: &sync.Mutex{}}
}

func (manager *Manager) Add(sprite ISprite) {
	startTime := sprite.GetStartTime()
	if sprite.IsAlwaysVisible() {
		startTime = -math.MaxFloat64
	}

	n := sort.Search(len(manager.spriteQueue), func(j int) bool {
		return startTime < manager.spriteQueue[j].GetStartTime()
	})

	manager.spriteQueue = append(manager.spriteQueue, nil) //allocate bigger array in case when len=cap
	copy(manager.spriteQueue[n+1:], manager.spriteQueue[n:])

	manager.spriteQueue[n] = sprite
}

func (manager *Manager) Update(time float64) {
	dirtyLocal := false
	toRemove := 0

	for i := 0; i < len(manager.spriteQueue); i++ {
		c := manager.spriteQueue[i]
		if time < c.GetStartTime() && !c.IsAlwaysVisible() {
			break
		}

		toRemove++
	}

	if toRemove > 0 {
		for i := 0; i < toRemove; i++ {
			s := manager.spriteQueue[i]

			n := sort.Search(len(manager.spriteProcessed), func(j int) bool {
				return s.GetDepth() < manager.spriteProcessed[j].GetDepth()
			})

			manager.spriteProcessed = append(manager.spriteProcessed, nil) //allocate bigger array in case when len=cap
			copy(manager.spriteProcessed[n+1:], manager.spriteProcessed[n:])

			manager.spriteProcessed[n] = s
		}

		dirtyLocal = true
		manager.spriteQueue = manager.spriteQueue[toRemove:]
	}

	for i := 0; i < len(manager.spriteProcessed); i++ {
		c := manager.spriteProcessed[i]
		c.Update(time)

		if time >= c.GetEndTime() && !c.IsAlwaysVisible() {
			copy(manager.spriteProcessed[i:], manager.spriteProcessed[i+1:])
			manager.spriteProcessed = manager.spriteProcessed[:len(manager.spriteProcessed)-1]

			dirtyLocal = true
			i--
		}
	}

	if dirtyLocal {
		manager.mutex.Lock()

		if len(manager.interArray) < len(manager.spriteProcessed) || len(manager.interArray) > len(manager.spriteProcessed)*3 {
			manager.interArray = make([]ISprite, len(manager.spriteProcessed)*2)
		}

		manager.interObjects = len(manager.spriteProcessed)
		copy(manager.interArray, manager.spriteProcessed)

		manager.dirty = true

		manager.mutex.Unlock()
	}
}

func (manager *Manager) GetNumRendered() (sum int) {
	for i := 0; i < manager.visibleObjects; i++ {
		if manager.drawArray[i] != nil && manager.drawArray[i].GetAlpha() >= 0.01 {
			sum++
		}
	}

	return
}

func (manager *Manager) GetNumInQueue() int {
	return len(manager.spriteQueue)
}

func (manager *Manager) GetNumProcessed() int {
	return len(manager.spriteProcessed)
}

func (manager *Manager) GetProcessedSprites() []ISprite {
	return manager.spriteProcessed
}

func (manager *Manager) Draw(time float64, batch *batch.QuadBatch) {
	manager.mutex.Lock()

	if manager.dirty {
		manager.visibleObjects = 0

		if len(manager.interArray) != len(manager.drawArray) {
			manager.drawArray = make([]ISprite, len(manager.interArray))
		}

		copy(manager.drawArray, manager.interArray[:manager.interObjects])
		manager.visibleObjects = manager.interObjects

		manager.dirty = false
	}

	manager.mutex.Unlock()

	for i := 0; i < manager.visibleObjects; i++ {
		if manager.drawArray[i] != nil {
			manager.drawArray[i].Draw(time, batch)
		}
	}
}
