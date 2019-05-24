package storyboard

import (
	"bufio"
	"os"
	"log"
	"strings"
	"strconv"
	"github.com/wieku/danser/bmath"
	"github.com/wieku/danser/utils"
	"path/filepath"
	"github.com/wieku/danser/settings"
	"github.com/wieku/danser/beatmap"
	"github.com/wieku/danser/render/texture"
	"fmt"
	"github.com/wieku/danser/render/batches"
)

type Storyboard struct {
	textures   map[string]*texture.TextureRegion
	atlas      *texture.TextureAtlas
	background *StoryboardLayer
	pass       *StoryboardLayer
	foreground *StoryboardLayer
	zIndex     int64
	bgFile     string
	bgFileUsed bool
	widescreen bool
}

func getSection(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "[") {
		return strings.TrimRight(strings.TrimLeft(line, "["), "]")
	}
	return ""
}

func NewStoryboard(beatMap *beatmap.BeatMap) *Storyboard {
	path := filepath.Join(settings.General.OsuSongsDir, beatMap.Dir)

	replacer := strings.NewReplacer("\\", "",
		"/", "",
		"<", "",
		">", "",
		"|", "",
		"?", "",
		"*", "",
		":", "",
		"\"", "",)

	fix := func(el string) string {
		return replacer.Replace(el)
	}

	files := []string{filepath.Join(path, beatMap.File), filepath.Join(path, fmt.Sprintf("%s - %s (%s).osb", fix(beatMap.Artist), fix(beatMap.Name), fix(beatMap.Creator)))}

	storyboard := &Storyboard{zIndex: -1, background: NewStoryboardLayer(), pass: NewStoryboardLayer(), foreground: NewStoryboardLayer(), atlas: nil}
	storyboard.textures = make(map[string]*texture.TextureRegion)

	var currentSection string
	var currentSprite string
	var commands []string

	variables := make(map[string]string)
	counter := 0

	for _, fS := range files {
		file, err := os.Open(fS)

		if err != nil {
			log.Println(err)
			continue
		}
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "//") || strings.TrimSpace(line) == "" {
				continue
			}

			section := getSection(line)
			if section != "" {
				currentSection = section
				continue
			}

			switch currentSection {
			case "General":
				split := strings.Split(line, ":")
				if strings.TrimSpace(split[0]) == "WidescreenStoryboard" && strings.TrimSpace(split[1]) == "1" {
					storyboard.widescreen = true
				}
			case "256", "Variables":
				split := strings.Split(line, "=")
				variables[split[0]] = split[1]
				break
			case "32", "Events":
				if strings.ContainsRune(line, '$') {
					for k, v := range variables {
						if strings.Contains(line, k) {
							line = strings.Replace(line, k, v, -1)
						}
					}
				}

				if strings.HasPrefix(line, "Sprite") || strings.HasPrefix(line, "4") || strings.HasPrefix(line, "Animation") || strings.HasPrefix(line, "6") {

					if currentSprite != "" {
						counter++
						storyboard.loadSprite(path, currentSprite, commands)
					}

					currentSprite = line
					commands = make([]string, 0)
				} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "_") {
					commands = append(commands, line)
				}
				break
			}
		}

		if currentSprite != "" {
			counter++
			storyboard.loadSprite(path, currentSprite, commands)
		}

		file.Close()
	}

	if counter == 0 {
		if storyboard.atlas != nil {
			storyboard.atlas.Dispose()
		}
		return nil
	}

	storyboard.background.FinishLoading()
	storyboard.pass.FinishLoading()
	storyboard.foreground.FinishLoading()

	for k := range storyboard.textures {
		if k == beatMap.Bg {
			storyboard.bgFileUsed = true
		}
	}

	log.Println("Storyboard loaded")

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

	textures := make([]*texture.TextureRegion, 0)
	frameDelay := 0.0
	loopForever := true

	if spl[0] == "Animation" || spl[0] == "6" {
		frames, _ := strconv.ParseInt(spl[6], 10, 32)
		frameDelay, _ = strconv.ParseFloat(spl[7], 64)

		if len(spl) > 8 && spl[8] == "LoopOnce" {
			loopForever = false
		}
		extension := filepath.Ext(image)
		baseFile := strings.TrimSuffix(image, extension)

		for i := 0; i < int(frames); i++ {
			texture := storyboard.getTexture(path, baseFile+strconv.Itoa(i)+extension)
			if texture != nil {
				textures = append(textures, texture)
			}
		}

	} else {
		texture := storyboard.getTexture(path, image)
		if texture != nil {
			textures = append(textures, texture)
		}
	}

	storyboard.zIndex++

	if len(textures) != 0 {
		sprite := NewSprite(textures, frameDelay, loopForever, storyboard.zIndex, pos, origin, commands)

		switch spl[1] {
		case "0", "Background":
			storyboard.background.Add(sprite)
			break
		case "2", "Pass":
			storyboard.pass.Add(sprite)
			break
		case "3", "Foreground":
			storyboard.foreground.Add(sprite)
			break
		}
	}
}

func (storyboard *Storyboard) getTexture(path, image string) *texture.TextureRegion {
	var texture1 *texture.TextureRegion

	if texture1 = storyboard.textures[image]; texture1 == nil {
		nrgba, err := utils.LoadImage(path + string(os.PathSeparator) + image)

		if err != nil {
			log.Println(err)
		} else {
			if storyboard.atlas == nil {
				storyboard.atlas = texture.NewTextureAtlas(8192, 4)
				storyboard.atlas.Bind(17)
			}
			texture1 = storyboard.atlas.AddTexture(image, nrgba.Bounds().Dx(), nrgba.Bounds().Dy(), nrgba.Pix)
			storyboard.textures[image] = texture1
		}
	}

	return texture1
}

func (storyboard *Storyboard) Update(time int64) {
	storyboard.background.Update(time)
	storyboard.pass.Update(time)
	storyboard.foreground.Update(time)
}

func (storyboard *Storyboard) Draw(time int64, batch *batches.SpriteBatch) {
	storyboard.background.Draw(time, batch)
	storyboard.pass.Draw(time, batch)
	storyboard.foreground.Draw(time, batch)
}

func (storyboard *Storyboard) GetProcessedSprites() int {
	return storyboard.background.visibleObjects + storyboard.pass.visibleObjects + storyboard.foreground.visibleObjects
}

func (storyboard *Storyboard) GetQueueSprites() int {
	return len(storyboard.background.spriteQueue) + len(storyboard.pass.spriteQueue) + len(storyboard.foreground.spriteQueue)
}

func (storyboard *Storyboard) GetTotalSprites() int {
	return storyboard.background.allSprites + storyboard.pass.allSprites + storyboard.foreground.allSprites
}

func (storyboard *Storyboard) GetLoad() float64 {
	return storyboard.background.GetLoad() + storyboard.pass.GetLoad() + storyboard.foreground.GetLoad()
}

func (storyboard *Storyboard) BGFileUsed() bool {
	return storyboard.bgFileUsed
}

func (storyboard *Storyboard) IsWideScreen() bool {
	return storyboard.widescreen
}
