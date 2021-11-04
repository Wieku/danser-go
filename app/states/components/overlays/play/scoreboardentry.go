package play

import (
	"fmt"
	"github.com/thehowl/go-osuapi"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/assets"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/font"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/graphics/texture"
	color2 "github.com/wieku/danser-go/framework/math/color"
	"github.com/wieku/danser-go/framework/math/vector"
	"log"
	"net/http"
	"strconv"
)

type ScoreboardEntry struct {
	*sprite.Sprite

	name    string
	score   int64
	combo   int64
	rank    int
	visible bool

	scoreHumanized string
	comboHumanized string
	rankHumanized  string
	avatar         *sprite.Sprite
	showAvatar     bool
}

func NewScoreboardEntry(name string, score int64, combo int64, rank int, isPlayer bool) *ScoreboardEntry {
	bg := skin.GetTexture("menu-button-background")
	entry := &ScoreboardEntry{
		Sprite: sprite.NewSpriteSingle(bg, 0, vector.NewVec2d(0, 0), vector.CentreRight),
		name:   name,
		score:  score,
		combo:  combo,
		rank:   rank,
	}

	entry.Sprite.SetScale(0.625)
	entry.Sprite.SetCutOrigin(vector.CentreRight)
	entry.SetAlpha(0)

	if isPlayer {
		entry.SetColor(color2.NewL(0.9))
	} else {
		entry.SetColor(color2.NewIRGB(31, 115, 153))
	}

	fnt := font.GetFont("Ubuntu Regular")
	fnt.Overlap = 2.5

	testName := entry.name

	for fnt.GetWidth(20, testName) > 135 {
		entry.name = entry.name[:len(entry.name)-1]
		testName = entry.name + "..."
	}

	entry.name = testName

	fnt.Overlap = 0

	entry.UpdateData()

	return entry
}

func (entry *ScoreboardEntry) UpdateData() {
	entry.scoreHumanized = utils.Humanize(entry.score)
	entry.comboHumanized = utils.Humanize(entry.combo) + "x"
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
		entry.Sprite.SetCutX(1.0 - (230+offset/0.625)/float64(entry.Sprite.Texture.Width))
	} else {
		entry.Sprite.SetCutOrigin(vector.CentreRight)
		entry.Sprite.SetOrigin(vector.CentreRight)
		batch.SetTranslation(vector.NewVec2d((float64(entry.Sprite.Texture.Width-470)*0.625+offset)*scale, 0))
		entry.Sprite.SetCutX(1.0 - (float64(entry.Sprite.Texture.Width-470)+offset/0.625)/float64(entry.Sprite.Texture.Width))
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

	ubu := font.GetFont("Ubuntu Regular")
	ubu.Overlap = 2.5

	batch.SetColor(0.1, 0.1, 0.1, a*0.8)
	ubu.DrawOrigin(batch, entryPos.X+posScale*(3.5*scale), entryPos.Y-18.5*scale, topLeft, 20*scale, false, entry.name)

	batch.SetColor(1, 1, 1, a)
	ubu.DrawOrigin(batch, entryPos.X+posScale*(3*scale), entryPos.Y-19*scale, topLeft, 20*scale, false, entry.name)

	ubu.Overlap = 0

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
	url := "https://a.ppy.sh/" + strconv.Itoa(id)

	log.Println("Trying to fetch avatar from:", url)

	request, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		log.Println("Can't create request")
		return
	}

	client := new(http.Client)
	response, err := client.Do(request)

	if err != nil {
		log.Println(fmt.Sprintf("Failed to create request to: \"%s\": %s", url, err))
		return
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Println("a.ppy.sh responded with:", response.StatusCode)

		if response.StatusCode == 404 {
			log.Println("Avatar for user", id, "not found!")
		}

		return
	}

	pixmap, err := texture.NewPixmapReader(response.Body, response.ContentLength)
	if err != nil {
		log.Println("Can't load avatar! Error:", err)
		return
	}

	entry.loadAvatar(pixmap)

	pixmap.Dispose()
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
	key, err := utils.GetApiKey()
	if err != nil {
		log.Println("Please put your osu!api v1 key into 'api.txt' file")
	} else {
		client := osuapi.NewClient(key)
		err := client.Test()

		if err != nil {
			log.Println("Can't connect to osu!api:", err)
		} else {
			sUser, err := client.GetUser(osuapi.GetUserOpts{Username: user})
			if err != nil {
				log.Println("Can't find user:", user)
				log.Println(err)
			} else {
				entry.LoadAvatarID(sUser.UserID)
			}
		}
	}
}

func (entry *ScoreboardEntry) IsAvatarLoaded() bool {
	return entry.avatar != nil
}

func (entry *ScoreboardEntry) ShowAvatar(value bool) {
	entry.showAvatar = value
}
