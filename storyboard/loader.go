package storyboard

import (
	"github.com/wieku/danser/render"
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
}

func getSection(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "[") {
		return strings.TrimRight(strings.TrimLeft(line, "["), "]")
	}
	return ""
}

func NewStoryboard(beatMap *beatmap.BeatMap) *Storyboard {

	fullPath := ""

	filepath.Walk(settings.General.OsuSongsDir+string(os.PathSeparator)+beatMap.Dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".osb") {
			fullPath = path
		}
		return nil
	})

	if fullPath == "" {
		return nil
	}

	storyboard := &Storyboard{zIndex: -1, background: NewStoryboardLayer(), pass: NewStoryboardLayer(), foreground: NewStoryboardLayer(), atlas: texture.NewTextureAtlas(8192, 4)}
	storyboard.atlas.Bind(17)
	storyboard.textures = make(map[string]*texture.TextureRegion)

	file, err := os.Open(fullPath)

	if err != nil {
		log.Println(err)
		return nil
	}

	scanner := bufio.NewScanner(file)

	var currentSection string
	var currentSprite string
	var commands []string

	variables := make(map[string]string)

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
					storyboard.loadSprite(fullPath, currentSprite, commands)
				}

				currentSprite = line
				commands = make([]string, 0)
			} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "_") {
				commands = append(commands, line)
			}
			break
		}
	}

	storyboard.loadSprite(fullPath, currentSprite, commands)

	storyboard.background.FinishLoading()
	storyboard.pass.FinishLoading()
	storyboard.foreground.FinishLoading()

	for k := range storyboard.textures {
		if k == beatMap.Bg {
			storyboard.bgFileUsed = true
		}
	}

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
			texture := storyboard.getTexture(filepath.Dir(path), baseFile+strconv.Itoa(i)+extension)
			if texture != nil {
				textures = append(textures, texture)
			}
		}

	} else {
		textures = append(textures, storyboard.getTexture(filepath.Dir(path), image))
	}

	storyboard.zIndex++

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

func (storyboard *Storyboard) getTexture(path, image string) *texture.TextureRegion {
	var texture *texture.TextureRegion

	if texture = storyboard.textures[image]; texture == nil {
		nrgba, err := utils.LoadImage(path + string(os.PathSeparator) + image)

		if err != nil {
			log.Println(err)
		} else {
			texture = storyboard.atlas.AddTexture(image, nrgba.Bounds().Dx(), nrgba.Bounds().Dy(), nrgba.Pix)
			storyboard.textures[image] = texture
		}
	}

	return texture
}

func (storyboard *Storyboard) Update(time int64) {
	storyboard.background.Update(time)
	storyboard.pass.Update(time)
	storyboard.foreground.Update(time)
}

func (storyboard *Storyboard) Draw(time int64, batch *render.SpriteBatch) {
	storyboard.background.Draw(time, batch)
	storyboard.pass.Draw(time, batch)
	storyboard.foreground.Draw(time, batch)
}

func (storyboard *Storyboard) BGFileUsed() bool {
	return storyboard.bgFileUsed
}
