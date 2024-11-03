package database

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/rulesets/osu/performance/pp241007"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/files"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/util"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

var dbFile *sql.DB

const databaseVersion = 20220622

var currentPreVersion = databaseVersion
var currentSchemaPreVersion = databaseVersion

type mapLocation struct {
	dir  string
	file string
}

type modMap struct {
	location mapLocation
	modTime  time.Time
}

var migrations []Migration

var songsDir string

var difficultyCalc = pp241007.NewDifficultyCalculator()

func Init() error {
	log.Println("DatabaseManager: Initializing database...")

	var err error

	songsDir, err = filepath.Abs(settings.General.GetSongsDir())
	if err != nil {
		return fmt.Errorf("invalid song path given: %s", settings.General.GetSongsDir())
	}

	_, err = os.Open(songsDir)
	if os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", songsDir)
	}

	migrations = []Migration{
		&M20181111{},
		&M20201027{},
		&M20201112{},
		&M20201117{},
		&M20201118{},
		&M20210104{},
		&M20210326{},
		&M20210423{},
		&M20220605{},
		&M20220622{},
	}

	dbFile, err = sql.Open("sqlite3", filepath.Join(env.DataDir(), "danser.db"))
	if err != nil {
		return err
	}

	_, err = dbFile.Exec(`
		CREATE TABLE IF NOT EXISTS beatmaps (dir TEXT, file TEXT, lastModified INTEGER, title TEXT, titleUnicode TEXT, artist TEXT, artistUnicode TEXT, creator TEXT, version TEXT, source TEXT, tags TEXT, cs REAL, ar REAL, sliderMultiplier REAL, sliderTickRate REAL, audioFile TEXT, previewTime INTEGER, sampleSet INTEGER, stackLeniency REAL, mode INTEGER, bg TEXT, md5 TEXT, dateAdded INTEGER, playCount INTEGER, lastPlayed INTEGER, hpdrain REAL, od REAL, stars REAL DEFAULT -1, bpmMin REAL, bpmMax REAL, circles INTEGER, sliders INTEGER, spinners INTEGER, endTime INTEGER, setID INTEGER, mapID INTEGER, starsVersion INTEGER DEFAULT 0, localOffset INTEGER DEFAULT 0);
		CREATE INDEX IF NOT EXISTS idx ON beatmaps (dir, file);
		CREATE TABLE IF NOT EXISTS info (key TEXT NOT NULL UNIQUE, value TEXT);
	`)

	if err != nil {
		return err
	}

	schemaVersionExists := false

	res, err := dbFile.Query("SELECT key, value FROM info")
	if err != nil {
		return err
	}

	for res.Next() {
		var key, value string

		err = res.Scan(&key, &value)
		if err != nil {
			return err
		}

		if key == "version" {
			currentPreVersion, _ = strconv.Atoi(value)
		}

		if key == "schema_version" {
			schemaVersionExists = true
			currentSchemaPreVersion, _ = strconv.Atoi(value)
		}
	}

	if !schemaVersionExists {
		currentSchemaPreVersion = currentPreVersion
	}

	log.Println("DatabaseManager: Database schema version:", currentSchemaPreVersion)
	log.Println("DatabaseManager: Database data version:", currentPreVersion)

	if currentSchemaPreVersion != databaseVersion {
		log.Println("DatabaseManager: Database schema is too old! Updating...")

		statement := ""

		for _, m := range migrations {
			if currentPreVersion < m.Date() {
				statement += m.GetMigrationStmts()
			}
		}

		_, err = dbFile.Exec(statement)
		if err != nil {
			panic(err)
		}

		log.Println("DatabaseManager: Schema has been updated!")
	}

	_, err = dbFile.Exec("REPLACE INTO info (key, value) VALUES ('schema_version', ?)", strconv.FormatInt(databaseVersion, 10))
	if err != nil {
		return err
	}

	if currentPreVersion != databaseVersion {
		migrateBeatmaps()
	}

	_, err = dbFile.Exec("REPLACE INTO info (key, value) VALUES ('version', ?)", strconv.FormatInt(databaseVersion, 10))
	if err != nil {
		return err
	}

	return nil
}

func LoadBeatmaps(skipDatabaseCheck bool, importListener ImportListener) []*beatmap.BeatMap {
	var unpackedMaps []string
	if settings.General.UnpackOszFiles {
		unpackedMaps = unpackMaps()
	}

	importMaps(skipDatabaseCheck, unpackedMaps, importListener)

	log.Println("DatabaseManager: Loading beatmaps from database...")

	allMaps := loadBeatmapsFromDatabase()

	stdMaps := make([]*beatmap.BeatMap, 0, len(allMaps)/2)

	for _, b := range allMaps {
		if b.Mode == 0 {
			stdMaps = append(stdMaps, b)
		}
	}

	log.Println("DatabaseManager: Loaded", len(stdMaps), "total.")

	return stdMaps
}

func unpackMaps() (dirs []string) {
	oszs, err := files.SearchFiles(songsDir, "*.osz", 0)

	if err == nil && len(oszs) > 0 {
		for _, osz := range oszs {
			dirName := strings.TrimSuffix(filepath.Base(osz), ".osz")

			destination := filepath.Join(filepath.Dir(osz), dirName)

			log.Println("DatabaseManager: Unpacking", osz, "->", destination)

			utils.Unzip(osz, destination)
			os.Remove(osz)

			dirs = append(dirs, dirName)
		}
	}

	return
}

type ImportListener func(stage ImportStage, progress, target int)

type ImportStage int

const (
	Discovery = ImportStage(iota)
	Comparison
	Cleanup
	Import
	Finished
)

func importMaps(skipDatabaseCheck bool, mustCheckDirs []string, importListener ImportListener) {
	const workers = 4

	cachedFolders, mapsInDB := getLastModified()

	candidates := make([]modMap, 0)

	log.Println(fmt.Sprintf("DatabaseManager: Scanning \"%s\" for .osu files...", songsDir))

	if skipDatabaseCheck {
		log.Println("DatabaseManager: '-nodbcheck' is active so only new directories will be imported.")
	}

	trySendStatus(importListener, Discovery, 0, 0)

	err := files.WalkDir(songsDir, func(path string, level int, de files.DirEntry) error {
		if de.IsDir() {
			if skipDatabaseCheck && level > 0 {
				dirName := filepath.Base(path)

				if _, ok := cachedFolders[dirName]; ok && (mustCheckDirs == nil || !slices.Contains(mustCheckDirs, dirName)) {
					return files.SkipDir
				}
			}

			return nil
		}

		// Don't read .osu files in main directory
		if level > 0 && strings.HasSuffix(de.Name(), ".osu") {
			relDir, err1 := filepath.Rel(songsDir, filepath.Dir(path))
			info, err2 := de.Info()
			if err1 != nil || err2 != nil {
				return nil
			}

			candidates = append(candidates, modMap{
				location: mapLocation{
					dir:  filepath.ToSlash(relDir),
					file: de.Name(),
				},
				modTime: info.ModTime(),
			})

			return files.SkipChildDirs
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	log.Println("DatabaseManager: Scan complete. Found", len(candidates), "files.")

	if len(candidates) == 0 {
		return
	}

	log.Println("DatabaseManager: Comparing files with database...")

	mapsToImport := make([]mapLocation, 0)

	trySendStatus(importListener, Comparison, 0, 0)

	for _, candidate := range candidates {
		if lastModified, ok := mapsInDB[candidate.location]; ok {
			if lastModified == candidate.modTime.UnixNano()/1000000 {
				// Map is up-to-date, so remove it from mapsInDB because values left in that map are later removed from database.
				delete(mapsInDB, candidate.location)

				continue
			}

			if settings.General.VerboseImportLogs {
				log.Println("DatabaseManager: New beatmap version found:", candidate.location.file)
			}
		} else if settings.General.VerboseImportLogs {
			log.Println("DatabaseManager: New beatmap found:", candidate.location.file)
		}

		mapsToImport = append(mapsToImport, candidate.location)
	}

	log.Println("DatabaseManager: Compare complete.")

	if len(mapsInDB) > 0 && !skipDatabaseCheck {
		trySendStatus(importListener, Cleanup, 100, 100)

		log.Println("DatabaseManager: Removing leftover maps from database...")

		mapsToRemove := make([]mapLocation, 0, len(mapsInDB))

		for k := range mapsInDB {
			mapsToRemove = append(mapsToRemove, k)
		}

		removeBeatmaps(mapsToRemove)

		log.Println("DatabaseManager: Removal complete.")
	}

	if len(mapsToImport) == 0 {
		return
	}

	log.Println("DatabaseManager: Starting import of", len(mapsToImport), "maps. It may take up to several minutes...")

	trySendStatus(importListener, Import, 0, len(mapsToImport))

	receive := make(chan *beatmap.BeatMap, workers)

	goroutines.Run(func() {
		util.BalanceChan(workers, mapsToImport, receive, func(candidate mapLocation) (*beatmap.BeatMap, bool) {
			defer func() {
				if err := recover(); err != nil { //TODO: Technically should be fixed but unexpected parsing problem won't crash whole process
					log.Println("DatabaseManager: Failed to load \"", candidate.dir+"/"+candidate.file, "\":", err)
				}
			}()

			partialPath := filepath.Join(candidate.dir, candidate.file)
			mapPath := filepath.Join(songsDir, partialPath)

			file, err := os.Open(mapPath)
			if err != nil {
				log.Println(fmt.Sprintf("\"DatabaseManager: Failed to read \"%s\", skipping. Error: %s", partialPath, err))
				return nil, false
			}

			defer file.Close()

			if settings.General.VerboseImportLogs {
				log.Println("DatabaseManager: Importing:", partialPath)
			}

			if bMap := beatmap.ParseBeatMapFile(file); bMap != nil {
				stat, _ := file.Stat()
				bMap.LastModified = stat.ModTime().UnixNano() / 1000000
				bMap.TimeAdded = time.Now().UnixNano() / 1000000

				hash := md5.New()
				if _, err := io.Copy(hash, file); err == nil {
					bMap.MD5 = hex.EncodeToString(hash.Sum(nil))
				}

				if settings.General.VerboseImportLogs {
					log.Println("DatabaseManager: Imported:", partialPath)
				}

				return bMap, true
			} else {
				log.Println("DatabaseManager: Failed to import:", partialPath)
			}

			return nil, false
		})

		close(receive)
	})

	var numImported int
	var imported []*beatmap.BeatMap

	for bMap := range receive {
		numImported++
		trySendStatus(importListener, Import, numImported, len(mapsToImport))

		imported = append(imported, bMap)

		if len(imported) >= 1000 { // Commit to database every 1k beatmaps to not lose progress in case of crash/close
			insertBeatmaps(imported)

			imported = imported[:0]
		}
	}

	if len(imported) > 0 {
		insertBeatmaps(imported)
	}

	trySendStatus(importListener, Finished, 100, 100)

	if numImported > 0 {
		log.Println("DatabaseManager: Imported", numImported, "new/updated beatmaps.")
	} else {
		log.Println("DatabaseManager: No new/updated beatmaps imported.")
	}
}

func trySendStatus(listener ImportListener, stage ImportStage, progress, target int) {
	if listener != nil {
		listener(stage, progress, target)
	}
}

func UpdateStarRating(maps []*beatmap.BeatMap, progressListener func(processed, target int)) {
	const workers = 1 // For now using only one thread because calculating 4 aspire maps at once can OOM since (de)allocation can't keep up with many complex sliders

	var toCalculate []*beatmap.BeatMap

	for _, b := range maps {
		if b.Mode == 0 && (b.Stars < 0 || b.StarsVersion < difficultyCalc.GetVersion()) {
			toCalculate = append(toCalculate, b)
		}
	}

	if len(toCalculate) == 0 {
		return
	}

	if progressListener != nil {
		progressListener(0, len(toCalculate))
	}

	receive := make(chan *beatmap.BeatMap, workers)

	goroutines.Run(func() {
		util.BalanceChan(workers, toCalculate, receive, func(bMap *beatmap.BeatMap) (ret *beatmap.BeatMap, ret2 bool) {
			ret = bMap // HACK: still return the beatmap even if execution panics: https://golangbyexample.com/return-value-function-panic-recover-go/
			ret2 = true

			defer func() {
				bMap.StarsVersion = difficultyCalc.GetVersion()
				bMap.Clear() //Clear objects and timing to avoid OOM

				if err := recover(); err != nil { //TODO: Technically should be fixed but unexpected parsing problem won't crash whole process
					bMap.Stars = 0
					log.Println("DatabaseManager: Failed to load \"", bMap.Dir+"/"+bMap.File, "\":", err)
				}
			}()

			beatmap.ParseTimingPointsAndPauses(bMap)
			beatmap.ParseObjects(bMap, true, false)

			if len(bMap.HitObjects) < 2 {
				log.Println("DatabaseManager:", bMap.Dir+"/"+bMap.File, "doesn't have enough hitobjects")
				bMap.Stars = 0
			} else {
				attr := difficultyCalc.CalculateSingle(bMap.HitObjects, bMap.Diff)
				bMap.Stars = attr.Total
			}

			return
		})

		close(receive)
	})

	var calculated []*beatmap.BeatMap
	var progress int

	for bMap := range receive {
		if progressListener != nil {
			progress++
			progressListener(progress, len(toCalculate))
		}

		calculated = append(calculated, bMap)

		if len(calculated) >= 1000 { // Commit to database every 1k beatmaps to not lose progress in case of crash/close
			pushSRToDB(calculated)

			calculated = calculated[:0]
		}
	}

	if len(calculated) > 0 {
		pushSRToDB(calculated)
	}

	log.Println("DatabaseManager: Star rating updated!")
}

func pushSRToDB(maps []*beatmap.BeatMap) {
	tx, err := dbFile.Begin()
	if err != nil {
		panic(err)
	}

	st, err := tx.Prepare("UPDATE beatmaps SET stars = ?, starsVersion = ? WHERE dir = ? AND file = ?")
	if err != nil {
		panic(err)
	}

	for _, bMap := range maps {
		_, err1 := st.Exec(
			bMap.Stars,
			bMap.StarsVersion,
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
}

func UpdatePlayStats(beatmap *beatmap.BeatMap) {
	_, err := dbFile.Exec("UPDATE beatmaps SET playCount = ?, lastPlayed = ? WHERE dir = ? AND file = ?", beatmap.PlayCount, beatmap.LastPlayed, beatmap.Dir, beatmap.File)
	if err != nil {
		log.Println(err)
	}
}

func UpdateLocalOffset(beatmap *beatmap.BeatMap) {
	_, err := dbFile.Exec("UPDATE beatmaps SET localOffset = ? WHERE dir = ? AND file = ?", beatmap.LocalOffset, beatmap.Dir, beatmap.File)
	if err != nil {
		log.Println(err)
	}
}

func removeBeatmaps(toRemove []mapLocation) {
	if len(toRemove) == 0 {
		return
	}

	tx, err := dbFile.Begin()

	if err == nil {
		st, err := tx.Prepare("DELETE FROM beatmaps WHERE dir = ? AND file = ?")

		if err == nil {
			for _, bMap := range toRemove {
				_, err1 := st.Exec(bMap.dir, bMap.file)

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

func migrateBeatmaps() {
	_, lastModified := getLastModified()

	var removeList []mapLocation

	if currentPreVersion < databaseVersion {
		updateBeatmaps := false

		for _, m := range migrations {
			if currentPreVersion < m.Date() {
				updateBeatmaps = updateBeatmaps || m.FieldsToMigrate() != nil
			}
		}

		if updateBeatmaps {
			log.Println("Updating cached beatmaps...")

			log.Println("Loading cached beatmaps from disk...")

			toUpdate := make([]*beatmap.BeatMap, 0)

			for location := range lastModified {
				file, err := os.Open(filepath.Join(songsDir, location.dir, location.file))
				if err != nil {
					log.Println("Failed to open file, removing from database:", location.file)
					log.Println("Error:", err)

					removeList = append(removeList, location)

					continue
				}

				bMap := beatmap.ParseBeatMapFile(file)
				if bMap == nil {
					log.Println("Corrupted cached beatmap found. Removing from database:", location.file)

					removeList = append(removeList, location)

					continue
				}

				toUpdate = append(toUpdate, bMap)
			}

			log.Println("Cached beatmaps loaded! Performing migrations...")

			tx, err := dbFile.Begin()
			if err != nil {
				panic(err)
			}

			for _, m := range migrations {
				if currentPreVersion < m.Date() {
					log.Println("Performing", m.Date(), "migration...")

					if m.FieldsToMigrate() == nil {
						continue
					}

					fieldsArray := m.FieldsToMigrate()
					for i := range fieldsArray {
						fieldsArray[i] += " = ?"
					}

					st, err := tx.Prepare(fmt.Sprintf("UPDATE beatmaps SET %s WHERE dir = ? AND file = ?", strings.Join(fieldsArray, ", ")))
					if err != nil {
						panic(err)
					}

					for _, bMap := range toUpdate {
						values := append(m.GetValues(bMap), bMap.Dir, bMap.File)

						_, err = st.Exec(values...)

						if err != nil {
							panic(err)
						}
					}

					if err = st.Close(); err != nil {
						panic(err)
					}
				}
			}

			log.Println("Committing migrations to database...")

			err = tx.Commit()
			if err != nil {
				panic(err)
			}
		}
	}

	removeBeatmaps(removeList)
}

func insertBeatmaps(bMaps []*beatmap.BeatMap) {
	tx, err := dbFile.Begin()

	if err == nil {
		var st *sql.Stmt
		st, err = tx.Prepare("INSERT INTO beatmaps VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

		if err == nil {
			for _, bMap := range bMaps {
				_, err1 := st.Exec(
					bMap.Dir,
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
					bMap.Diff.GetHP(),
					bMap.Diff.GetOD(),
					bMap.Stars,
					bMap.MinBPM,
					bMap.MaxBPM,
					bMap.Circles,
					bMap.Sliders,
					bMap.Spinners,
					bMap.Length,
					bMap.SetID,
					bMap.ID,
					bMap.StarsVersion,
					bMap.LocalOffset,
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

func loadBeatmapsFromDatabase() []*beatmap.BeatMap {
	beatmaps := make([]*beatmap.BeatMap, 0)

	res, _ := dbFile.Query("SELECT * FROM beatmaps")

	for res.Next() {
		beatMap := beatmap.NewBeatMap()

		var cs, ar, hp, od float64

		res.Scan(
			&beatMap.Dir,
			&beatMap.File,
			&beatMap.LastModified,
			&beatMap.Name,
			&beatMap.NameUnicode,
			&beatMap.Artist,
			&beatMap.ArtistUnicode,
			&beatMap.Creator,
			&beatMap.Difficulty,
			&beatMap.Source,
			&beatMap.Tags,
			&cs,
			&ar,
			&beatMap.Timings.SliderMult,
			&beatMap.Timings.TickRate,
			&beatMap.Audio,
			&beatMap.PreviewTime,
			&beatMap.Timings.BaseSet,
			&beatMap.StackLeniency,
			&beatMap.Mode,
			&beatMap.Bg,
			&beatMap.MD5,
			&beatMap.TimeAdded,
			&beatMap.PlayCount,
			&beatMap.LastPlayed,
			&hp,
			&od,
			&beatMap.Stars,
			&beatMap.MinBPM,
			&beatMap.MaxBPM,
			&beatMap.Circles,
			&beatMap.Sliders,
			&beatMap.Spinners,
			&beatMap.Length,
			&beatMap.SetID,
			&beatMap.ID,
			&beatMap.StarsVersion,
			&beatMap.LocalOffset,
		)

		beatMap.Diff.SetCS(mutils.Clamp(cs, 0, 10))
		beatMap.Diff.SetAR(mutils.Clamp(ar, 0, 10))
		beatMap.Diff.SetHP(mutils.Clamp(hp, 0, 10))
		beatMap.Diff.SetOD(mutils.Clamp(od, 0, 10))

		beatmaps = append(beatmaps, beatMap)
	}

	return beatmaps
}

func getLastModified() (map[string]uint8, map[mapLocation]int64) {
	res, _ := dbFile.Query("SELECT dir, file, lastModified FROM beatmaps")

	dirs := make(map[string]uint8)

	mod := make(map[mapLocation]int64)

	for res.Next() {
		var dir, file string
		var lastModified int64

		res.Scan(&dir, &file, &lastModified)

		dirs[dir] = 1

		mod[mapLocation{
			dir:  dir,
			file: file,
		}] = lastModified
	}

	return dirs, mod
}

func Close() {
	if dbFile != nil {
		err := dbFile.Close()
		if err != nil {
			log.Println("Failed to close database:", err)
		}

		dbFile = nil
	}
}
