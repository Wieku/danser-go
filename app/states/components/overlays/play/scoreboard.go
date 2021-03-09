package play

import (
	"github.com/thehowl/go-osuapi"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/bmath"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/skin"
	"github.com/wieku/danser-go/framework/graphics/batch"
	"github.com/wieku/danser-go/framework/graphics/sprite"
	"github.com/wieku/danser-go/framework/math/animation"
	"github.com/wieku/danser-go/framework/math/animation/easing"
	"github.com/wieku/danser-go/framework/math/vector"
	"io/ioutil"
	"log"
	"sort"
)

const spacing = 57.6
const padding = 142.0
const start = 348.0
const visible = 6

type ScoreBoard struct {
	scores        []*ScoreboardEntry
	displayScores []*ScoreboardEntry

	time float64

	playerIndex     int
	lastPlayerIndex int
	playerEntry     *ScoreboardEntry

	explosionManager *sprite.SpriteManager
	first            bool
}

func NewScoreboard(beatMap *beatmap.BeatMap, omitID int64) *ScoreBoard {
	board := &ScoreBoard{
		first:            true,
		explosionManager: sprite.NewSpriteManager(),
	}

	skin.GetTextureSource("scoreboard-explosion-1", skin.LOCAL)
	skin.GetTextureSource("scoreboard-explosion-2", skin.LOCAL)

	if settings.Gameplay.ScoreBoard.HideOthers {
		return board
	}

	data, err := ioutil.ReadFile("api.txt")
	if err != nil {
		log.Println("Please put your osu!api v1 key into 'api.txt' file")
	} else {
		client := osuapi.NewClient(string(data))
		err := client.Test()

		if err != nil {
			log.Println("Can't connect to osu!api:", err)
		} else {
			beatMaps, err := client.GetBeatmaps(osuapi.GetBeatmapsOpts{BeatmapHash: beatMap.MD5})
			if len(beatMaps) == 0 || err != nil {
				log.Println("Online beatmap not found!")
				if err != nil {
					log.Println(err)
				}
			} else {
				scores, err := client.GetScores(osuapi.GetScoresOpts{BeatmapID: beatMaps[0].BeatmapID, Limit: 51})
				if len(scores) == 0 || err != nil {
					log.Println("Can't find online scores!")
					if err != nil {
						log.Println(err)
					}
				} else {
					for i := 0; i < len(scores); i++ {
						if scores[i].ScoreID == omitID {
							scores = append(scores[:i], scores[i+1:]...)
							i--
						}
					}

					sort.SliceStable(scores, func(i, j int) bool {
						return scores[i].Score.Score > scores[j].Score.Score
					})

					for i := 0; i < bmath.MinI(len(scores), 50); i++ {
						s := scores[i]

						entry := NewScoreboardEntry(s.Username, s.Score.Score, int64(s.MaxCombo), i+1, false)

						board.scores = append(board.scores, entry)
						board.displayScores = append(board.displayScores, entry)
					}
				}

				log.Println("SCORES", len(scores))
			}
		}
	}

	return board
}

func (board *ScoreBoard) AddPlayer(name string) {
	board.playerEntry = NewScoreboardEntry(name, 0, 0, len(board.scores)+1, true)
	board.playerIndex = len(board.scores)
	board.lastPlayerIndex = board.playerIndex

	board.scores = append(board.scores, board.playerEntry)
	board.displayScores = append(board.displayScores, board.playerEntry)

	board.UpdatePlayer(0, 0)
}

func (board *ScoreBoard) UpdatePlayer(score, combo int64) {
	board.playerEntry.score = score
	board.playerEntry.combo = combo

	sort.SliceStable(board.scores, func(i, j int) bool {
		return board.scores[i].score > board.scores[j].score
	})

	for i := 0; i < len(board.scores); i++ {
		entry := board.scores[i]

		entry.rank = i + 1

		entry.UpdateData()

		if entry == board.playerEntry {
			board.playerIndex = i
			if board.playerIndex < board.lastPlayerIndex {
				playerPos := board.playerEntry.GetPosition()
				playerPos.X = 0

				sprite2 := sprite.NewSpriteSingle(skin.GetTexture("scoreboard-explosion-2"), 0.5, playerPos, bmath.Origin.CentreLeft)
				sprite2.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, board.time, board.time+400, 1, 0))
				sprite2.AddTransform(animation.NewVectorTransform(animation.ScaleVector, easing.OutQuad, board.time, board.time+200, 1, 1, 16, 1.2))
				sprite2.ResetValuesToTransforms()
				sprite2.AdjustTimesToTransformations()
				sprite2.ShowForever(false)

				sprite1 := sprite.NewSpriteSingle(skin.GetTexture("scoreboard-explosion-1"), 1, playerPos, bmath.Origin.CentreLeft)
				sprite1.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, board.time, board.time+700, 1, 0))
				sprite1.AddTransform(animation.NewVectorTransform(animation.ScaleVector, easing.OutQuad, board.time, board.time+700, 1, 1, 1, 1.3))
				sprite1.ResetValuesToTransforms()
				sprite1.AdjustTimesToTransformations()
				sprite1.ShowForever(false)

				board.explosionManager.Add(sprite2)
				board.explosionManager.Add(sprite1)
			}

			board.lastPlayerIndex = board.playerIndex
		}
	}

	startI := 0
	if board.playerIndex > visible-1 {
		startI = board.playerIndex - (visible - 1)
	}

	shiftI := 0

	for i := 0; i < len(board.scores); i++ {
		entry := board.scores[i]

		display := i == 0
		if i > startI && shiftI < visible {
			display = true
		}

		target := vector.NewVec2d(0, start+float64(shiftI)*spacing)

		if board.first {
			entry.SetPosition(target)
		} else {
			entry.ClearTransformationsOfType(animation.Move)
			entry.AddTransform(animation.NewVectorTransformV(animation.Move, easing.OutQuad, board.time, board.time+600, entry.GetPosition(), target))
		}

		alpha := 0.8
		if i != board.playerIndex {
			alpha -= 0.3 * float64(i) / float64(len(board.scores))
		}

		if display {
			entry.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, board.time, board.time+600, entry.GetAlpha(), alpha))

			shiftI++
		} else if entry.visible {
			entry.AddTransform(animation.NewSingleTransform(animation.Fade, easing.Linear, board.time, board.time+600, entry.GetAlpha(), 0))
		}

		entry.visible = display
	}

	board.first = false
}

func (board *ScoreBoard) Update(time float64) {
	board.time = time

	board.explosionManager.Update(time)

	for _, e := range board.scores {
		e.Update(time)
	}
}

func (board *ScoreBoard) Draw(batch *batch.QuadBatch, alpha float64) {
	if !settings.Gameplay.ScoreBoard.Show {
		return
	}

	alpha *= settings.Gameplay.ScoreBoard.Opacity

	for _, e := range board.displayScores {
		e.Draw(board.time, batch, alpha)
	}

	board.explosionManager.Draw(board.time, batch)
}
