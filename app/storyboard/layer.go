package storyboard

import (
	"github.com/wieku/danser-go/app/render/batches"
	"sort"
	"sync"
)

type StoryboardLayer struct {
	spriteQueue     []Object
	spriteProcessed []Object
	drawArray       []Object
	interArray      []Object
	visibleObjects  int
	interObjects    int
	allSprites      int
	mutex           *sync.Mutex
}

func NewStoryboardLayer() *StoryboardLayer {
	return &StoryboardLayer{mutex: &sync.Mutex{}}
}

func (layer *StoryboardLayer) Add(object Object) {
	layer.spriteQueue = append(layer.spriteQueue, object)
}

func (layer *StoryboardLayer) FinishLoading() {
	sort.Slice(layer.spriteQueue, func(i, j int) bool {
		return layer.spriteQueue[i].GetStartTime() < layer.spriteQueue[j].GetStartTime()
	})
	layer.allSprites = len(layer.spriteQueue)
	layer.drawArray = make([]Object, len(layer.spriteQueue))
	layer.interArray = make([]Object, len(layer.spriteQueue))
}

func (layer *StoryboardLayer) Update(time int64) {
	toRemove := 0

	for i := 0; i < len(layer.spriteQueue); i++ {
		c := layer.spriteQueue[i]
		if time < c.GetStartTime() {
			break
		}

		toRemove++
	}

	if toRemove > 0 {
		for i := 0; i < toRemove; i++ {
			s := layer.spriteQueue[i]

			n := sort.Search(len(layer.spriteProcessed), func(j int) bool {
				return s.GetZIndex() < layer.spriteProcessed[j].GetZIndex()
			})

			layer.spriteProcessed = append(layer.spriteProcessed, nil) //allocate bigger array in case when len=cap
			copy(layer.spriteProcessed[n+1:], layer.spriteProcessed[n:])

			layer.spriteProcessed[n] = s
		}

		layer.spriteQueue = layer.spriteQueue[toRemove:]
	}

	for i := 0; i < len(layer.spriteProcessed); i++ {
		c := layer.spriteProcessed[i]
		c.Update(time)

		if time >= c.GetEndTime() {
			copy(layer.spriteProcessed[i:], layer.spriteProcessed[i+1:])
			layer.spriteProcessed = layer.spriteProcessed[:len(layer.spriteProcessed)-1]
			i--
		}
	}

	layer.mutex.Lock()

	layer.interObjects = len(layer.spriteProcessed)
	copy(layer.interArray, layer.spriteProcessed)

	layer.mutex.Unlock()
}

func (layer *StoryboardLayer) GetLoad() (sum float64) {
	for i := 0; i < layer.visibleObjects; i++ {
		if layer.drawArray[i] != nil {
			sum += layer.drawArray[i].GetLoad()
		}
	}
	return
}

func (layer *StoryboardLayer) Draw(time int64, batch *batches.SpriteBatch) {
	layer.mutex.Lock()

	layer.visibleObjects = layer.interObjects
	copy(layer.drawArray, layer.interArray[:layer.interObjects])

	layer.mutex.Unlock()

	for i := 0; i < layer.visibleObjects; i++ {
		if layer.drawArray[i] != nil {
			layer.drawArray[i].Draw(time, batch)
		}
	}
}
