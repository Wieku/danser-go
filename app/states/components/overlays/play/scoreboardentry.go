package play

import (
	"errors"
	"fmt"
	"github.com/wieku/danser-go/app/osuapi"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type ScoreboardEntry struct {
	*sprite.Sprite

	name       string
	score      osuapi.Score
	lazerScore bool

	rank    int
	visible bool

	scoreHumanized string
	comboHumanized string
	rankHumanized  string
	avatar         *sprite.Sprite
	showAvatar     bool
}

func NewScoreboardEntry(name string, score osuapi.Score, lazerScore bool, rank int, isPlayer bool) *ScoreboardEntry {
	bg := skin.GetTexture("menu-button-background")
	entry := &ScoreboardEntry{
		Sprite:     sprite.NewSpriteSingle(bg, 0, vector.NewVec2d(0, 0), vector.CentreRight),
		name:       name,
		score:      score,
		lazerScore: lazerScore,
		rank:       rank,
	}

	entry.Sprite.SetScale(0.625)
	entry.Sprite.SetCutOrigin(vector.CentreRight)
	entry.SetAlpha(0)

	if isPlayer {
		entry.SetColor(color2.NewL(0.9))
	} else {
		entry.SetColor(color2.NewIRGB(31, 115, 153))
	}

	fnt := font.GetFont("SBFont")
	if fnt == font.GetFont("Ubuntu Regular") {
		fnt.Overlap = 2.5
	} else {
		fnt.Overlap = 0
	}

	testName := entry.name

	for fnt.GetWidth(19, testName) > 135 {
		entry.name = entry.name[:len(entry.name)-1]
		testName = entry.name + "..."
	}

	entry.name = testName

	fnt.Overlap = 0

	entry.UpdateData()

	return entry
}

func (entry *ScoreboardEntry) UpdateData() {
	entry.scoreHumanized = utils.Humanize(entry.getScore())
	entry.comboHumanized = utils.Humanize(entry.score.MaxCombo) + "x"
	entry.rankHumanized = fmt.Sprintf("%d", entry.rank)
}

func (entry *ScoreboardEntry) Draw(time float64, batch *batch.QuadBatch, alpha float64) {
	batch.ResetTransform()

	a := entry.Sprite.GetAlpha() * alpha

	scale := settings.Gameplay.ScoreBoard.Scale

	if a < 0.01 {
		return
	}

	offset := 0.0
	if entry.showAvatar {
		offset = 52
	}

	topLeft := vector.TopLeft
	topRight := vector.TopRight
	posScale := 1.0

	if settings.Gameplay.ScoreBoard.AlignRight {
		topLeft = vector.TopRight
		topRight = vector.TopLeft
		posScale = -1.0

		entry.Sprite.SetCutOrigin(vector.CentreLeft)
		entry.Sprite.SetOrigin(vector.CentreLeft)
		batch.SetTranslation(vector.NewVec2d(-(230*0.625+offset)*scale, 0))
		entry.Sprite.SetCutX(0, 1.0-(230+offset/0.625)/float64(entry.Sprite.Texture.Width))
	} else {
		entry.Sprite.SetCutOrigin(vector.CentreRight)
		entry.Sprite.SetOrigin(vector.CentreRight)
		batch.SetTranslation(vector.NewVec2d((float64(entry.Sprite.Texture.Width-470)*0.625+offset)*scale, 0))
		entry.Sprite.SetCutX(1.0-(float64(entry.Sprite.Texture.Width-470)+offset/0.625)/float64(entry.Sprite.Texture.Width), 0)
	}

	batch.SetColor(1, 1, 1, 0.6*alpha)
	batch.SetScale(scale, scale)

	entry.Sprite.Draw(time, batch)

	batch.ResetTransform()
	batch.SetTranslation(vector.NewVec2d(posScale*offset*scale, 0))

	batch.SetColor(1, 1, 1, a)

	entryPos := entry.GetPosition()

	if entry.showAvatar && entry.avatar != nil {
		batch.SetSubScale(scale, scale)
		entry.avatar.SetPosition(entryPos.SubS(26*scale*posScale, 0))
		entry.avatar.Draw(time, batch)
		batch.SetSubScale(1, 1)
	}

	fnt := skin.GetFont("scoreentry")

	fnt.Overlap = 2.5
	fnt.DrawOrigin(batch, entryPos.X+posScale*(3.2*scale), entryPos.Y+8.8*scale, topLeft, fnt.GetSize()*scale, true, entry.scoreHumanized)

	if entry.rank <= 50 {
		batch.SetColor(1, 1, 1, a*0.32)

		fnt.Overlap = 3
		fnt.DrawOrigin(batch, entryPos.X+posScale*(padding-10)*scale, entryPos.Y-22*scale, topRight, fnt.GetSize()*2.2*scale, true, entry.rankHumanized)
	}

	batch.SetColor(0.6, 0.98, 1, a)

	fnt.Overlap = 2.5
	fnt.DrawOrigin(batch, entryPos.X+posScale*(padding-10)*scale, entryPos.Y+8.8*scale, topRight, fnt.GetSize()*scale, true, entry.comboHumanized)

	sbFnt := font.GetFont("SBFont")

	if sbFnt == font.GetFont("Ubuntu Regular") {
		sbFnt.Overlap = 2.5
	} else {
		sbFnt.Overlap = 0
	}

	batch.SetColor(0.1, 0.1, 0.1, a*0.8)
	sbFnt.DrawOrigin(batch, entryPos.X+posScale*(3.5*scale), entryPos.Y-18.5*scale, topLeft, 19*scale, false, entry.name)

	batch.SetColor(1, 1, 1, a)
	sbFnt.DrawOrigin(batch, entryPos.X+posScale*(3*scale), entryPos.Y-19*scale, topLeft, 19*scale, false, entry.name)

	sbFnt.Overlap = 0

	batch.ResetTransform()
	batch.SetColor(1, 1, 1, 1)
}

func (entry *ScoreboardEntry) loadAvatar(pixmap *texture.Pixmap) {
	tex := texture.LoadTextureSingle(pixmap.RGBA(), 4)
	region := tex.GetRegion()

	entry.avatar = sprite.NewSpriteSingle(&region, 0, vector.NewVec2d(26, 0), vector.Centre)
	entry.avatar.SetScale(float64(52 / region.Height))
}

func (entry *ScoreboardEntry) LoadAvatarID(id int) {
	entry.LoadAvatarURL("https://a.ppy.sh/" + strconv.Itoa(id))
}

func (entry *ScoreboardEntry) LoadAvatarURL(url string) {
	if url == "" { // just in case
		return
	}

	fileName := strings.ReplaceAll(url[strings.LastIndex(url, "/")+1:], "?", "-")
	filePath := filepath.Join(env.DataDir(), "cache", "avatars", fileName)

	if s, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) || s.Size() == 0 { // Avatar does not exist or is empty, try to download
		log.Println("Trying to fetch avatar from:", url)

		err2 := downloadAvatar(url, filePath)
		if err2 != nil {
			log.Println(err2)
			return
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Println(fmt.Sprintf("Failed to open avatar \"%s\": %s", fileName, err.Error()))
		return
	}

	defer file.Close()

	fStat, err2 := file.Stat()
	if err2 != nil {
		log.Println(fmt.Sprintf("Failed to open file stats \"%s\": %s", fileName, err2.Error()))
		return
	}

	pixmap, err := texture.NewPixmapReader(file, fStat.Size())
	if err != nil {
		log.Println("Can't load avatar! Error:", err)
		return
	}

	entry.loadAvatar(pixmap)

	pixmap.Dispose()
}

func downloadAvatar(url, path string) error {
	response, err := http.Get(url)

	if err != nil {
		return fmt.Errorf("failed to create request to: \"%s\": %s", url, err)
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return fmt.Errorf("failed to create request to: \"%s\": %s", url, response.StatusCode)
	}

	out, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file: \"%s\": %s", path, err)
	}

	defer out.Close()

	_, err = io.Copy(out, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write to file: \"%s\": %s", path, err)
	}

	return nil
}

func (entry *ScoreboardEntry) LoadDefaultAvatar() {
	pixmap, err := assets.GetPixmap("assets/textures/dansercoin256.png")
	if err != nil {
		log.Println("Can't load avatar! Error:", err)
		return
	}

	entry.loadAvatar(pixmap)

	pixmap.Dispose()
}

func (entry *ScoreboardEntry) LoadAvatarUser(user string) {
	sUser, err := osuapi.LookupUser(user)

	if err != nil {
		log.Println("Error connecting to osu!api:", err)
	} else {
		entry.LoadAvatarURL(sUser.AvatarURL)
	}
}

func (entry *ScoreboardEntry) IsAvatarLoaded() bool {
	return entry.avatar != nil
}

func (entry *ScoreboardEntry) ShowAvatar(value bool) {
	entry.showAvatar = value
}

func (entry *ScoreboardEntry) getScore() int64 {
	if entry.lazerScore {
		if settings.Gameplay.LazerClassicScore {
			return entry.score.ClassicTotalScore
		}

		return entry.score.TotalScore
	}

	return entry.score.Score
}
