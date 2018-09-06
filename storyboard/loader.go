package storyboard

import (
	"github.com/wieku/danser/render"
	"bufio"
	"os"
	"log"
	"strings"
	"github.com/wieku/glhf"
	"strconv"
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/utils"
	"path/filepath"
	"github.com/wieku/danser/settings"
	"sort"
)

type Storyboard struct {
	textures            map[string]*glhf.Texture
	BackgroundSprites   []*Sprite
	BackgroundProcessed []*Sprite
	ForegroundSprites   []*Sprite
	ForegroundProcessed []*Sprite
	zIndex              int64
}

func NewStoryboard(path string) *Storyboard {
	path = settings.General.OsuSongsDir + string(os.PathSeparator) + path
	storyboard := &Storyboard{zIndex:-1}

	storyboard.textures = make(map[string]*glhf.Texture)

	file, err := os.Open(path)

	if err != nil {
		log.Println(err)
		return nil
	}

	scanner := bufio.NewScanner(file)

	var currentSprite string
	var commands []string

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "Sprite") || strings.HasPrefix(line, "4") {

			if currentSprite != "" {
				storyboard.loadSprite(path, currentSprite, commands)
			}

			currentSprite = line
			commands = make([]string, 0)
		} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "_") {
			commands = append(commands, line)
		}
	}

	storyboard.loadSprite(path, currentSprite, commands)

	return storyboard
}

func (storyboard *Storyboard) loadSprite(path, currentSprite string, commands []string) {
	spl := strings.Split(currentSprite, ",")

	origin := Origin[spl[2]]

	x, _ := strconv.ParseFloat(spl[4], 64)
	y, _ := strconv.ParseFloat(spl[5], 64)

	pos := bmath.NewVec2d(x, y)

	image := strings.Replace(spl[3], `"`, "", -1)

	if !strings.HasSuffix(image, ".png") && !strings.HasSuffix(image, ".jpg") {
		image += ".png"
	}

	var texture *glhf.Texture

	if texture = storyboard.textures[image]; texture == nil {
		var err error
		texture, err = utils.LoadTextureU(filepath.Dir(path) + string(os.PathSeparator) + image)
		if err != nil {
			log.Println(err)
		}
		storyboard.textures[image] = texture
	}

	storyboard.zIndex++

	sprite := NewSprite(texture, storyboard.zIndex, pos, origin, commands)

	switch spl[1] {
	case "0", "Background":
		storyboard.BackgroundSprites = append(storyboard.BackgroundSprites, sprite)
		break
	case "3", "Foreground":
		storyboard.ForegroundSprites = append(storyboard.ForegroundSprites, sprite)
		break
	}
}

func (storyboard *Storyboard) Update(time int64) {

	added := false

	for i := 0; i < len(storyboard.BackgroundSprites); i++ {
		c := storyboard.BackgroundSprites[i]
		if c.GetStartTime() <= time {
			storyboard.BackgroundProcessed = append(storyboard.BackgroundProcessed, c)
			added = true
			storyboard.BackgroundSprites = append(storyboard.BackgroundSprites[:i], storyboard.BackgroundSprites[i+1:]...)
			i--
		}
	}

	if added {
		sort.Slice(storyboard.BackgroundProcessed, func(i, j int) bool {
			return storyboard.BackgroundProcessed[i].GetZIndex() < storyboard.BackgroundProcessed[j].GetZIndex()
		})
	}

	for i := 0; i < len(storyboard.BackgroundProcessed); i++ {
		c := storyboard.BackgroundProcessed[i]
		c.Update(time)

		if time >= c.GetEndTime() {
			storyboard.BackgroundProcessed = append(storyboard.BackgroundProcessed[:i], storyboard.BackgroundProcessed[i+1:]...)
			i--
		}
	}

	added = false

	for i := 0; i < len(storyboard.ForegroundSprites); i++ {
		c := storyboard.ForegroundSprites[i]
		if c.GetStartTime() <= time {
			storyboard.ForegroundProcessed = append(storyboard.ForegroundProcessed, c)
			added = true
			storyboard.ForegroundSprites = append(storyboard.ForegroundSprites[:i], storyboard.ForegroundSprites[i+1:]...)
			i--
		}
	}

	if added {
		sort.Slice(storyboard.ForegroundProcessed, func(i, j int) bool {
			return storyboard.ForegroundProcessed[i].GetZIndex() < storyboard.ForegroundProcessed[j].GetZIndex()
		})
	}

	for i := 0; i < len(storyboard.ForegroundProcessed); i++ {
		c := storyboard.ForegroundProcessed[i]
		c.Update(time)

		if time >= c.GetEndTime() {
			storyboard.ForegroundProcessed = append(storyboard.ForegroundProcessed[:i], storyboard.ForegroundProcessed[i+1:]...)
			i--
		}
	}
}

func (storyboard *Storyboard) Draw(time int64, batch *render.SpriteBatch) {

	for _, s := range storyboard.BackgroundProcessed {
		s.Draw(time, batch)
	}

	for _, s := range storyboard.ForegroundProcessed {
		s.Draw(time, batch)
	}
}
