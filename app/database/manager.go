package database

import (
	"database/sql"
	oppai "github.com/flesnuk/oppai5"
	"github.com/karrick/godirwalk"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"log"
	"os"
	"path/filepath"
	"strings"

	"crypto/md5"
	"encoding/hex"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wieku/danser-go/app/beatmap"
	"io"
	"strconv"
	"time"
)

var dbFile *sql.DB

const databaseVersion = 20201118

var currentPreVersion = databaseVersion

type toRemove struct {
	dir  string
	file string
}

func Init() {
	var err error
	dbFile, err = sql.Open("sqlite3", "danser.db")
	if err != nil {
		panic(err)
	}

	_, err = dbFile.Exec(`
		CREATE TABLE IF NOT EXISTS beatmaps (dir TEXT, file TEXT, lastModified INTEGER, title TEXT, titleUnicode TEXT, artist TEXT, artistUnicode TEXT, creator TEXT, version TEXT, source TEXT, tags TEXT, cs REAL, ar REAL, sliderMultiplier REAL, sliderTickRate REAL, audioFile TEXT, previewTime INTEGER, sampleSet INTEGER, stackLeniency REAL, mode INTEGER, bg TEXT, md5 TEXT, dateAdded INTEGER, playCount INTEGER, lastPlayed INTEGER, hpdrain REAL, od REAL, stars REAL DEFAULT -1, bpmMin REAL, bpmMax REAL, circles INTEGER, sliders INTEGER, spinners INTEGER, endTime INTEGER);
		CREATE INDEX IF NOT EXISTS idx ON beatmaps (dir, file);
		CREATE TABLE IF NOT EXISTS info (key TEXT NOT NULL UNIQUE, value TEXT);
	`)

	if err != nil {
		panic(err)
	}

	res, _ := dbFile.Query("SELECT key, value FROM info")

	for res.Next() {
		var key, value string

		res.Scan(&key, &value)
		if key == "version" {
			currentPreVersion, _ = strconv.Atoi(value)
		}
	}

	log.Println("Database version: ", currentPreVersion)

	if currentPreVersion == databaseVersion {
		return
	}

	log.Println("Database is too old! Updating...")

	if currentPreVersion < 20181111 {
		_, err = dbFile.Exec(`ALTER TABLE beatmaps ADD COLUMN hpdrain REAL;
							 ALTER TABLE beatmaps ADD COLUMN od REAL;`)

		if err != nil {
			panic(err)
		}
	}

	if currentPreVersion < 20201027 {
		_, err = dbFile.Exec(`
			BEGIN TRANSACTION;
			CREATE TEMPORARY TABLE beatmaps_backup(dir TEXT, file TEXT, lastModified INTEGER, title TEXT, titleUnicode TEXT, artist TEXT, artistUnicode TEXT, creator TEXT, version TEXT, source TEXT, tags TEXT, cs REAL, ar REAL, sliderMultiplier REAL, sliderTickRate REAL, audioFile TEXT, previewTime INTEGER, sampleSet INTEGER, stackLeniency REAL, mode INTEGER, bg TEXT, md5 TEXT, dateAdded INTEGER, playCount INTEGER, lastPlayed INTEGER, hpdrain REAL, od REAL);
			INSERT INTO beatmaps_backup SELECT dir, file, lastModified, title, titleUnicode, artist, artistUnicode, creator, version, source, tags, cs, ar, sliderMultiplier, sliderTickRate, audioFile, previewTime, sampleSet, stackLeniency, mode, bg, md5, dateAdded, playCount, lastPlayed, hpdrain, od FROM beatmaps;
			DROP TABLE beatmaps;
			CREATE TABLE beatmaps(dir TEXT, file TEXT, lastModified INTEGER, title TEXT, titleUnicode TEXT, artist TEXT, artistUnicode TEXT, creator TEXT, version TEXT, source TEXT, tags TEXT, cs REAL, ar REAL, sliderMultiplier REAL, sliderTickRate REAL, audioFile TEXT, previewTime INTEGER, sampleSet INTEGER, stackLeniency REAL, mode INTEGER, bg TEXT, md5 TEXT, dateAdded INTEGER, playCount INTEGER, lastPlayed INTEGER, hpdrain REAL, od REAL);
			INSERT INTO beatmaps SELECT * FROM beatmaps_backup;
			DROP TABLE beatmaps_backup;
			CREATE INDEX IF NOT EXISTS idx ON beatmaps (dir, file);
			COMMIT;
			vacuum;
		`)

		if err != nil {
			panic(err)
		}
	}

	if currentPreVersion < 20201117 {
		_, err = dbFile.Exec(`ALTER TABLE beatmaps ADD COLUMN stars REAL DEFAULT -1;`)
		if err != nil {
			panic(err)
		}
	}

	if currentPreVersion < 20201118 {
		_, err = dbFile.Exec(`
			ALTER TABLE beatmaps ADD COLUMN bpmMin REAL DEFAULT 0;
 			ALTER TABLE beatmaps ADD COLUMN bpmMax REAL DEFAULT 0;
  			ALTER TABLE beatmaps ADD COLUMN circles INTEGER DEFAULT 0;
   			ALTER TABLE beatmaps ADD COLUMN sliders INTEGER DEFAULT 0;
    		ALTER TABLE beatmaps ADD COLUMN spinners INTEGER DEFAULT 0;
     		ALTER TABLE beatmaps ADD COLUMN endTime INTEGER DEFAULT 0;
     	`)

		if err != nil {
			panic(err)
		}
	}

	_, err = dbFile.Exec("REPLACE INTO info (key, value) VALUES ('version', ?)", strconv.FormatInt(databaseVersion, 10))
	if err != nil {
		log.Println(err)
	}
}

func LoadBeatmaps() []*beatmap.BeatMap {
	log.Println("Checking database...")

	searchDir, err := filepath.Abs(settings.General.OsuSongsDir)
	if err != nil {
		log.Println("Invalid song path given:", settings.General.OsuSongsDir)
		return nil
	}

	_, err = os.Open(searchDir)
	if os.IsNotExist(err) {
		log.Println(searchDir + " does not exist!")
		return nil
	}

	mod := getLastModified()

	newBeatmaps := make([]*beatmap.BeatMap, 0)
	cachedBeatmaps := make([]*beatmap.BeatMap, 0)

	_ = godirwalk.Walk(searchDir, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() && osPathname != searchDir {
				return godirwalk.SkipThis
			}

			if strings.HasSuffix(de.Name(), ".osz") {
				log.Println("Unpacking", osPathname, "to", filepath.Dir(osPathname)+"/"+strings.TrimSuffix(de.Name(), ".osz"))
				utils.Unzip(osPathname, filepath.Dir(osPathname)+"/"+strings.TrimSuffix(de.Name(), ".osz"))
				os.Remove(osPathname)
			}

			return nil
		},
		Unsorted: true,
	})

	err = godirwalk.Walk(searchDir, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() && osPathname != searchDir && filepath.Dir(osPathname) != searchDir {
				return godirwalk.SkipThis
			}

			if strings.HasSuffix(de.Name(), ".osu") {
				cachedTime := mod[filepath.Base(filepath.Dir(osPathname))+"/"+de.Name()]

				stat, err := os.Stat(osPathname)
				if err != nil {
					log.Println("Failed to read file stats, skipping:", osPathname)
					return nil
				}

				if cachedTime != stat.ModTime().UnixNano()/1000000 {
					if cachedTime > 0 {
						log.Println(cachedTime, stat.ModTime().UnixNano()/1000000, osPathname)
						removeBeatmap(filepath.Base(filepath.Dir(osPathname)), de.Name())
						log.Println("Found new beatmap version:", de.Name())
					} else {
						log.Println("New beatmap found:", de.Name())
					}

					file, err := os.Open(osPathname)

					if err == nil {
						defer file.Close()

						if bMap := beatmap.ParseBeatMapFile(file); bMap != nil {
							bMap.LastModified = stat.ModTime().UnixNano() / 1000000
							bMap.TimeAdded = time.Now().UnixNano() / 1000000
							log.Println("Importing:", bMap.File)

							hash := md5.New()
							if _, err := io.Copy(hash, file); err == nil {
								bMap.MD5 = hex.EncodeToString(hash.Sum(nil))
								newBeatmaps = append(newBeatmaps, bMap)
							}
						}
					}
				} else {
					bMap := beatmap.NewBeatMap()
					bMap.Dir = filepath.Base(filepath.Dir(osPathname))
					bMap.File = de.Name()
					cachedBeatmaps = append(cachedBeatmaps, bMap)
				}
			}

			return nil
		},
		Unsorted: true,
	})

	if err != nil {
		panic(err)
	}

	log.Println("Imported", len(newBeatmaps), "new beatmaps.")

	updateBeatmaps(newBeatmaps)

	log.Println("Found", len(cachedBeatmaps), "cached beatmaps. Loading...")

	loadBeatmaps(cachedBeatmaps)

	allMaps := append(newBeatmaps, cachedBeatmaps...)

	log.Println("Loaded", len(allMaps), "total.")

	result := make([]*beatmap.BeatMap, 0)
	stars := make([]interface{}, 0)

	for _, b := range allMaps {
		if b.Mode == 0 {
			if b.Stars < 0 {
				stars = append(stars, b)
			}

			result = append(result, b)
		}
	}

	if len(stars) > 0 {
		log.Println("Updating star rating...")

		utils.Balance(4, stars, func(a interface{}) interface{} {
			b := a.(*beatmap.BeatMap)

			f, err := os.Open(filepath.Join(settings.General.OsuSongsDir, b.Dir, b.File))
			if err == nil {
				mp := oppai.Parse(f)

				if len(mp.Objects) > 0 {
					calc := &oppai.DiffCalc{Beatmap: *mp}
					calc.Calc(0, oppai.DefaultSingletapThreshold)
					b.Stars = calc.Total
				} else {
					b.Stars = 0
				}
			}

			return a
		})

		tx, err := dbFile.Begin()
		if err != nil {
			panic(err)
		}

		st, err := tx.Prepare("UPDATE beatmaps SET stars = ? WHERE dir = ? AND file = ?")
		if err != nil {
			panic(err)
		}

		for _, b := range stars {
			bMap := b.(*beatmap.BeatMap)
			_, err1 := st.Exec(
				bMap.Stars,
				bMap.Dir,
				bMap.File)

			if err1 != nil {
				log.Println(err1)
			}
		}

		if err = st.Close(); err != nil {
			panic(err)
		}

		if err = tx.Commit(); err != nil {
			panic(err)
		}

		log.Println("Calculations finished")
	}

	return result
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
	var removeList []toRemove

	for i, bMap := range bMaps {
		beatmaps[bMap.Dir+"/"+bMap.File] = i + 1
	}

	if currentPreVersion < 20201118 {
		log.Println("Updating cached beatmaps")

		toUpdate := make([]*beatmap.BeatMap, 0)

		for _, bMap := range bMaps {
			err2 := beatmap.ParseBeatMap(bMap)
			if err2 != nil {
				log.Println("Corrupted cached beatmap found. Removing from database:", bMap.File)
				removeList = append(removeList, toRemove{bMap.Dir, bMap.File})
			} else {
				toUpdate = append(toUpdate, bMap)
			}
		}

		tx, err := dbFile.Begin()
		if err != nil {
			panic(err)
		}

		if currentPreVersion < 20181111 {
			st, err := tx.Prepare("UPDATE beatmaps SET hpdrain = ?, od = ? WHERE dir = ? AND file = ?")
			if err != nil {
				panic(err)
			}

			for _, bMap := range toUpdate {
				_, err1 := st.Exec(
					bMap.Diff.GetHPDrain(),
					bMap.Diff.GetOD(),
					bMap.Dir,
					bMap.File)

				if err1 != nil {
					log.Println(err1)
				}
			}

			if err = st.Close(); err != nil {
				panic(err)
			}
		}

		if currentPreVersion < 20201112 {
			st, err := tx.Prepare("UPDATE beatmaps SET previewTime = ? WHERE dir = ? AND file = ?")
			if err != nil {
				panic(err)
			}

			for _, bMap := range toUpdate {
				_, err1 := st.Exec(
					bMap.PreviewTime,
					bMap.Dir,
					bMap.File)

				if err1 != nil {
					log.Println(err1)
				}
			}

			if err = st.Close(); err != nil {
				panic(err)
			}
		}

		if currentPreVersion < 20201118 {
			st, err := tx.Prepare("UPDATE beatmaps SET bpmMin = ?, bpmMax = ?, circles = ?, sliders = ?, spinners = ?, endTime = ? WHERE dir = ? AND file = ?")
			if err != nil {
				panic(err)
			}

			for _, bMap := range toUpdate {
				_, err1 := st.Exec(
					bMap.MinBPM,
					bMap.MaxBPM,
					bMap.Circles,
					bMap.Sliders,
					bMap.Spinners,
					bMap.Length,
					bMap.Dir,
					bMap.File)

				if err1 != nil {
					log.Println(err1)
				}
			}

			if err = st.Close(); err != nil {
				panic(err)
			}
		}

		err = tx.Commit()
		if err != nil {
			panic(err)
		}
	} else {
		res, _ := dbFile.Query("SELECT * FROM beatmaps")

		for res.Next() {
			beatmap := beatmap.NewBeatMap()

			var cs float64
			var ar float64
			var hp float64
			var od float64

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
				&cs,
				&ar,
				&beatmap.Timings.SliderMult,
				&beatmap.Timings.TickRate,
				&beatmap.Audio,
				&beatmap.PreviewTime,
				&beatmap.Timings.BaseSet,
				&beatmap.StackLeniency,
				&beatmap.Mode,
				&beatmap.Bg,
				&beatmap.MD5,
				&beatmap.TimeAdded,
				&beatmap.PlayCount,
				&beatmap.LastPlayed,
				&hp,
				&od,
				&beatmap.Stars,
				&beatmap.MinBPM,
				&beatmap.MaxBPM,
				&beatmap.Circles,
				&beatmap.Sliders,
				&beatmap.Spinners,
				&beatmap.Length,
			)

			beatmap.Diff.SetCS(cs)
			beatmap.Diff.SetAR(ar)
			beatmap.Diff.SetHPDrain(hp)
			beatmap.Diff.SetOD(od)

			if beatmap.Name+beatmap.Artist+beatmap.Creator == "" {
				log.Println("Corrupted cached beatmap found. Removing from database:", beatmap.File)
				removeList = append(removeList, toRemove{beatmap.Dir, beatmap.File})
				continue
			}

			key := beatmap.Dir + "/" + beatmap.File

			if beatmaps[key] > 0 {
				bMaps[beatmaps[key]-1] = beatmap
			}

		}
	}

	for _, b := range removeList {
		removeBeatmap(b.dir, b.file)
	}
}

func updateBeatmaps(bMaps []*beatmap.BeatMap) {
	tx, err := dbFile.Begin()

	if err == nil {
		var st *sql.Stmt
		st, err = tx.Prepare("INSERT INTO beatmaps VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

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
					bMap.Diff.GetCS(),
					bMap.Diff.GetAR(),
					bMap.SliderMultiplier,
					bMap.Timings.TickRate,
					bMap.Audio,
					bMap.PreviewTime,
					bMap.Timings.BaseSet,
					bMap.StackLeniency,
					bMap.Mode,
					bMap.Bg,
					bMap.MD5,
					bMap.TimeAdded,
					bMap.PlayCount,
					bMap.LastPlayed,
					bMap.Diff.GetHPDrain(),
					bMap.Diff.GetOD(),
					bMap.Stars,
					bMap.MinBPM,
					bMap.MaxBPM,
					bMap.Circles,
					bMap.Sliders,
					bMap.Spinners,
					bMap.Length,
				)

				if err1 != nil {
					log.Println(err1)
				}
			}
		} else {
			panic(err)
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
