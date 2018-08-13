package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"
	//"github.com/wieku/danser/beatmap"
	"github.com/wieku/danser/settings"

	"github.com/wieku/danser/beatmap"
	"crypto/md5"
	"io"
	"encoding/hex"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var dbFile *sql.DB

func Init() {
	var err error
	dbFile, err = sql.Open("sqlite3", "danser.db")

	if err != nil {
		panic(err)
	}

	_, err = dbFile.Exec("CREATE TABLE IF NOT EXISTS beatmaps (dir TEXT, file TEXT, lastModified INTEGER, title TEXT, titleUnicode TEXT, artist TEXT, artistUnicode TEXT, creator TEXT, version TEXT, source TEXT, tags TEXT, cs REAL, ar REAL, sliderMultiplier REAL, sliderTickRate REAL, audioFile TEXT, previewTime INTEGER, sampleSet INTEGER, stackLeniency REAL, mode INTEGER, bg TEXT, pauses TEXT, timingPoints TEXT, md5 TEXT, dateAdded INTEGER, playCount INTEGER, lastPlayed INTEGER)")

	if err != nil {
		panic(err)
	}
}

func LoadBeatmaps() []*beatmap.BeatMap {
	log.Println("Checking database...")

	searchDir := settings.General.OsuSongsDir

	_, err := os.Open(searchDir)
	if os.IsNotExist(err) {
		log.Println(searchDir + " does not exist!")
		return nil
	}

	mod := getLastModified()

	newBeatmaps := make([]*beatmap.BeatMap, 0)
	cachedBeatmaps := make([]*beatmap.BeatMap, 0)

	filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(f.Name(), ".osu") {
			cachedTime := mod[filepath.Base(filepath.Dir(path))+"/"+f.Name()]
			if cachedTime != f.ModTime().UnixNano()/1000000 {
				removeBeatmap(filepath.Base(filepath.Dir(path)), f.Name())

				file, err := os.Open(path)

				if err == nil {
					defer file.Close()

					if bMap := beatmap.ParseBeatMap(file); bMap != nil {

						bMap.Dir = filepath.Base(filepath.Dir(path))
						bMap.File = f.Name()
						bMap.LastModified = f.ModTime().UnixNano() / 1000000
						bMap.TimeAdded = time.Now().UnixNano() / 1000000
						log.Println("New beatmap:", bMap.File)

						hash := md5.New()

						if _, err := io.Copy(hash, file); err == nil {
							bMap.MD5 = hex.EncodeToString(hash.Sum(nil))
							log.Println("Checksum:", bMap.MD5)

							newBeatmaps = append(newBeatmaps, bMap)
						}
					}
				}
			} else {
				bMap := beatmap.NewBeatMap()
				bMap.Dir = filepath.Base(filepath.Dir(path))
				bMap.File = f.Name()
				cachedBeatmaps = append(cachedBeatmaps, bMap)
			}
		}
		return nil
	})

	log.Println("Found", len(newBeatmaps), "new beatmaps.")

	updateBeatmaps(newBeatmaps)

	log.Println("Found", len(cachedBeatmaps), "cached beatmaps. Loading...")

	loadBeatmaps(cachedBeatmaps)

	allMaps := append(newBeatmaps, cachedBeatmaps...)

	log.Println("Loaded", len(allMaps), "total.")

	return allMaps
}

func UpdatePlayStats(beatmap *beatmap.BeatMap) {
	_, err := dbFile.Exec("UPDATE beatmaps SET playCount = ?, lastPlayed = ? WHERE dir = ? AND file = ?", beatmap.PlayCount, beatmap.LastPlayed, beatmap.Dir, beatmap.File)
	if err != nil {
		log.Println(err)
	}
}

func removeBeatmap(dir, file string) {
	dbFile.Exec("DELETE FROM beatmaps WHERE dir = ? AND file = ?", dir, file)
}

func loadBeatmaps(bMaps []*beatmap.BeatMap) {

	beatmaps := make(map[string]int)

	for i, bMap := range bMaps  {
		beatmaps[bMap.Dir+"/"+bMap.File] = i+1
	}

	res, _ := dbFile.Query("SELECT * FROM beatmaps")

	for res.Next() {
		beatmap := beatmap.NewBeatMap()
		var mode int
		res.Scan(
				&beatmap.Dir,
				&beatmap.File,
				&beatmap.LastModified,
				&beatmap.Name,
				&beatmap.NameUnicode,
				&beatmap.Artist,
				&beatmap.ArtistUnicode,
				&beatmap.Creator,
				&beatmap.Difficulty,
				&beatmap.Source,
				&beatmap.Tags,
				&beatmap.CircleSize,
				&beatmap.AR,
				&beatmap.Timings.SliderMult,
				&beatmap.Timings.TickRate,
				&beatmap.Audio,
				&beatmap.PreviewTime,
				&beatmap.Timings.BaseSet,
				&beatmap.StackLeniency,
				&mode,
				&beatmap.Bg,
				&beatmap.PausesText,
				&beatmap.TimingPoints,
				&beatmap.MD5,
				&beatmap.TimeAdded,
				&beatmap.PlayCount,
				&beatmap.LastPlayed)

		key := beatmap.Dir + "/" + beatmap.File

		if beatmaps[key] > 0 {
			bMaps[beatmaps[key] - 1] = beatmap
			beatmap.LoadPauses()
			beatmap.LoadTimingPoints()
		}

	}

}

func updateBeatmaps(bMaps []*beatmap.BeatMap) {
	tx, err := dbFile.Begin()

	if err == nil {
		var st *sql.Stmt
		st, err = tx.Prepare("INSERT INTO beatmaps VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		if err == nil {
			for _, bMap := range bMaps {
				_, err1 := st.Exec(bMap.Dir,
					bMap.File,
					bMap.LastModified,
					bMap.Name,
					bMap.NameUnicode,
					bMap.Artist,
					bMap.ArtistUnicode,
					bMap.Creator,
					bMap.Difficulty,
					bMap.Source,
					bMap.Tags,
					bMap.CircleSize,
					bMap.AR,
					bMap.SliderMultiplier,
					bMap.Timings.TickRate,
					bMap.Audio,
					bMap.PreviewTime,
					bMap.Timings.BaseSet,
					bMap.StackLeniency,
					0,
					bMap.Bg,
					bMap.PausesText,
					bMap.TimingPoints,
					bMap.MD5,
					bMap.TimeAdded,
					bMap.PlayCount,
					bMap.LastPlayed)

				if err1 != nil {
					log.Println(err1)
				}
			}
		}

		st.Close()
		tx.Commit()
	}

	if err != nil {
		log.Println(err)
	}

}

func getLastModified() map[string]int64 {
	res, _ := dbFile.Query("SELECT dir, file, lastModified FROM beatmaps")

	mod := make(map[string]int64)

	for res.Next() {
		var dir string
		var file string
		var lastModified int64

		res.Scan(&dir, &file, &lastModified)
		mod[dir+"/"+file] = lastModified
	}

	return mod
}