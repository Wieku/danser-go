package storyboard

import (
	"fmt"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/bass"
	files2 "github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/frame"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	video2 "github.com/wieku/danser-go/framework/graphics/video"
	"github.com/wieku/danser-go/framework/math/vector"
	"github.com/wieku/danser-go/framework/qpc"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Storyboard struct {
	textures map[string]*texture.TextureRegion
	atlas    *texture.TextureAtlas

	samples map[string]*bass.Sample

	background  *sprite.Manager
	pass        *sprite.Manager
	foreground  *sprite.Manager
	overlay     *sprite.Manager
	zIndex      int64
	bgFileUsed  bool
	widescreen  bool
	shouldRun   bool
	currentTime float64
	limiter     *frame.Limiter
	counter     *frame.Counter
	numSprites  int
	pathCache   *files2.FileMap
	hasVisuals  bool
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

	files := []string{
		filepath.Join(path, beatMap.File),
		filepath.Join(path, files2.FixName(fmt.Sprintf("%s - %s (%s).osb", beatMap.Artist, beatMap.Name, beatMap.Creator))),
	}

	storyboard := &Storyboard{
		textures:   make(map[string]*texture.TextureRegion),
		samples:    make(map[string]*bass.Sample),
		zIndex:     -1,
		background: sprite.NewManager(),
		pass:       sprite.NewManager(),
		foreground: sprite.NewManager(),
		overlay:    sprite.NewManager(),
		atlas:      nil,
	}

	storyboard.pathCache, _ = files2.NewFileMap(path)

	var currentSection string
	var currentSprite string
	var commands []string

	variables := make(map[string]string)
	hasVideo := false
	hasAudio := false

	for _, fS := range files {
		file, err := os.Open(fS)

		log.Println("Trying to load storyboard from: ", fS)

		if err != nil {
			log.Println(err)
			continue
		}

		scanner := files2.NewScannerBuf(file, 10*1024*1024)

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
			case "32", "Events":
				if strings.ContainsRune(line, '$') {
					for k, v := range variables {
						line = strings.Replace(line, k, v, -1)
					}
				}

				if settings.Playfield.Background.LoadVideos && (strings.HasPrefix(line, "Video") || strings.HasPrefix(line, "1")) {
					spl := strings.Split(line, ",")

					video := video2.NewVideo(filepath.Join(path, files2.FixName(spl[2])), -1, vector.NewVec2d(320, 240), vector.Centre)

					if video == nil {
						continue
					}

					video.SetScaleV(vector.NewVec2d(1, 1).Scl(480.0 / float64(video.Texture.Height)))

					offset, _ := strconv.ParseFloat(spl[1], 64)
					video.SetStartTime(offset)
					video.SetEndTime(math.MaxFloat64)
					video.ShowForever(false)

					storyboard.background.Add(video)

					hasVideo = true
				} else if settings.Playfield.Background.LoadStoryboards {
					if strings.HasPrefix(line, "Sample") || strings.HasPrefix(line, "5") {
						spl := strings.Split(line, ",")

						startTime, _ := strconv.ParseFloat(spl[1], 64)
						volume, _ := strconv.ParseFloat(spl[4], 64)

						sbSprite := sprite.NewAudioSprite(storyboard.getSample(files2.FixName(spl[3])), startTime, volume/100)

						storyboard.addSpriteToLayer(spl[2], sbSprite)

						hasAudio = true
					} else if strings.HasPrefix(line, "Sprite") || strings.HasPrefix(line, "4") || strings.HasPrefix(line, "Animation") || strings.HasPrefix(line, "6") {
						if currentSprite != "" {
							storyboard.loadSprite(currentSprite, commands)
						}

						currentSprite = line
						commands = make([]string, 0)
					} else if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "_") {
						commands = append(commands, line)
					}
				}
			}
		}

		if currentSprite != "" {
			storyboard.loadSprite(currentSprite, commands)
		}

		file.Close()
	}

	storyboard.hasVisuals = storyboard.numSprites > 0 || hasVideo

	if storyboard.numSprites == 0 {
		if storyboard.atlas != nil {
			storyboard.atlas.Dispose()
		}

		if !hasVideo && !hasAudio {
			return nil
		} else if !storyboard.widescreen {
			storyboard.widescreen = true
		}
	}

	for k := range storyboard.textures {
		if k == beatMap.Bg {
			storyboard.bgFileUsed = true
		}
	}

	log.Println("Storyboard loaded")

	storyboard.currentTime = -1000000
	storyboard.limiter = frame.NewLimiter(2000)
	storyboard.counter = frame.NewCounter()

	return storyboard
}

func (storyboard *Storyboard) loadSprite(currentSprite string, commands []string) {
	spl := strings.Split(currentSprite, ",")

	origin := parseOrigin(spl[2])

	x, _ := strconv.ParseFloat(spl[4], 64)
	y, _ := strconv.ParseFloat(spl[5], 64)

	pos := vector.NewVec2d(x, y)

	image := strings.Replace(spl[3], `"`, "", -1)

	if filepath.Ext(image) == "" {
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
			if tex := storyboard.getTexture(baseFile + strconv.Itoa(i) + extension); tex != nil {
				textures = append(textures, tex)
			}
		}
	} else {
		if tex := storyboard.getTexture(image); tex != nil {
			textures = append(textures, tex)
		}
	}

	storyboard.zIndex++

	if len(textures) != 0 {
		sbSprite := sprite.NewAnimation(textures, frameDelay, loopForever, float64(storyboard.zIndex), pos, origin)

		transforms := parseCommands(commands)

		sbSprite.ShowForever(false)
		sbSprite.AddTransforms(transforms)
		sbSprite.AdjustTimesToTransformations()
		sbSprite.ResetValuesToTransforms()

		storyboard.addSpriteToLayer(spl[1], sbSprite)

		storyboard.numSprites++
	}
}

func (storyboard *Storyboard) addSpriteToLayer(layer string, sbSprite sprite.ISprite) {
	switch layer {
	case "0", "Background":
		storyboard.background.Add(sbSprite)
	case "2", "Pass":
		storyboard.pass.Add(sbSprite)
	case "3", "Foreground":
		storyboard.foreground.Add(sbSprite)
	case "4", "Overlay":
		storyboard.overlay.Add(sbSprite)
	}
}

func (storyboard *Storyboard) getTexture(image string) *texture.TextureRegion {
	var texture1 *texture.TextureRegion

	if texture1 = storyboard.textures[image]; texture1 == nil {
		if texture1 = skin.GetTexture(strings.TrimSuffix(image, filepath.Ext(image))); texture1 != nil {
			storyboard.textures[image] = texture1
		} else {
			path, err := storyboard.pathCache.GetFile(image)
			if err != nil {
				log.Println("File:", image, "does not exist!")
				return texture1
			}

			img, err := texture.NewPixmapFileString(path)

			if err == nil {

				if img.Width > 512 || img.Height > 512 {
					tex := texture.NewTextureSingle(img.Width, img.Height, 0)
					tex.Bind(0)
					tex.SetData(0, 0, img.Width, img.Height, img.Data)
					rg := tex.GetRegion()
					texture1 = &rg
				} else {
					if storyboard.atlas == nil {
						storyboard.atlas = texture.NewTextureAtlas(4096, 0)
						storyboard.atlas.Bind(17)
					}

					texture1 = storyboard.atlas.AddTexture(image, img.Width, img.Height, img.Data)
				}

				img.Dispose()

				storyboard.textures[image] = texture1
			} else {
				log.Println(err)
			}
		}
	}

	return texture1
}

func (storyboard *Storyboard) getSample(sample string) (bassSample *bass.Sample) {
	if bassSample = storyboard.samples[sample]; bassSample == nil {
		path, err := storyboard.pathCache.GetFile(sample)
		if err != nil {
			log.Println("File:", sample, "does not exist!")
			return
		}

		bassSample = bass.NewSample(path)
	}

	return
}

func (storyboard *Storyboard) StartThread() {
	if storyboard.shouldRun {
		return
	}

	go func() {
		lastTime := qpc.GetMilliTimeF()

		for storyboard.shouldRun {
			time := qpc.GetMilliTimeF()
			storyboard.counter.PutSample(time - lastTime)
			lastTime = time

			storyboard.Update(storyboard.currentTime)

			storyboard.limiter.Sync()
		}
	}()

	storyboard.shouldRun = true
}

func (storyboard *Storyboard) StopThread() {
	storyboard.shouldRun = false
}

func (storyboard *Storyboard) IsThreadRunning() bool {
	return storyboard.shouldRun
}

func (storyboard *Storyboard) UpdateTime(time float64) {
	storyboard.currentTime = time
}

func (storyboard *Storyboard) GetFPS() float64 {
	return storyboard.counter.GetFPS()
}

func (storyboard *Storyboard) SetFPS(i int) {
	storyboard.limiter.FPS = i
}

func (storyboard *Storyboard) Update(time float64) {
	storyboard.background.Update(time)
	storyboard.pass.Update(time)
	storyboard.foreground.Update(time)
	storyboard.overlay.Update(time)
}

func (storyboard *Storyboard) Draw(time float64, batch *batch.QuadBatch) {
	batch.SetTranslation(vector.NewVec2d(-64, -48))
	storyboard.background.Draw(time, batch)
	storyboard.pass.Draw(time, batch)
	storyboard.foreground.Draw(time, batch)
	batch.SetTranslation(vector.NewVec2d(0, 0))
}

func (storyboard *Storyboard) DrawOverlay(time float64, batch *batch.QuadBatch) {
	batch.SetTranslation(vector.NewVec2d(-64, -48))
	storyboard.overlay.Draw(time, batch)
	batch.SetTranslation(vector.NewVec2d(0, 0))
}

func (storyboard *Storyboard) GetRenderedSprites() int {
	return storyboard.background.GetNumRendered() + storyboard.pass.GetNumRendered() + storyboard.foreground.GetNumRendered() + storyboard.overlay.GetNumRendered()
}

func (storyboard *Storyboard) GetProcessedSprites() int {
	return storyboard.background.GetNumProcessed() + storyboard.pass.GetNumProcessed() + storyboard.foreground.GetNumProcessed() + storyboard.overlay.GetNumProcessed()
}

func (storyboard *Storyboard) GetQueueSprites() int {
	return storyboard.background.GetNumInQueue() + storyboard.pass.GetNumInQueue() + storyboard.foreground.GetNumInQueue() + storyboard.overlay.GetNumInQueue()
}

func (storyboard *Storyboard) GetTotalSprites() int {
	return storyboard.numSprites
}

func (storyboard *Storyboard) BGFileUsed() bool {
	return storyboard.bgFileUsed
}

func (storyboard *Storyboard) HasVisuals() bool {
	return storyboard.hasVisuals
}

func (storyboard *Storyboard) IsWideScreen() bool {
	return storyboard.widescreen
}
