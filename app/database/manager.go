package database

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/karrick/godirwalk"
	_ "github.com/mattn/go-sqlite3"
	"github.com/wieku/danser-go/app/beatmap"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/app/utils"
	"github.com/wieku/danser-go/framework/math/mutils"
	"github.com/wieku/danser-go/framework/util"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

func SetLogOutput(w io.Writer) {
	logger.SetOutput(w)
}

var dbFile *sql.DB

const databaseVersion = 20210423

var currentPreVersion = databaseVersion
var currentSchemaPreVersion = databaseVersion

type mapLocation struct {
	dir  string
	file string
}

var migrations []Migration

var songsDir string

func Init() error {
	logger.Println("DatabaseManager: Initializing database...")

	var err error

	songsDir, err = filepath.Abs(settings.General.OsuSongsDir)
	if err != nil {
		return fmt.Errorf("invalid song path given: %s", settings.General.OsuSongsDir)
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
	}

	dbFile, err = sql.Open("sqlite3", "danser.db")
	if err != nil {
		return err
	}

	_, err = dbFile.Exec(`
		CREATE TABLE IF NOT EXISTS beatmaps (dir TEXT, file TEXT, lastModified INTEGER, title TEXT, titleUnicode TEXT, artist TEXT, artistUnicode TEXT, creator TEXT, version TEXT, source TEXT, tags TEXT, cs REAL, ar REAL, sliderMultiplier REAL, sliderTickRate REAL, audioFile TEXT, previewTime INTEGER, sampleSet INTEGER, stackLeniency REAL, mode INTEGER, bg TEXT, md5 TEXT, dateAdded INTEGER, playCount INTEGER, lastPlayed INTEGER, hpdrain REAL, od REAL, stars REAL DEFAULT -1, bpmMin REAL, bpmMax REAL, circles INTEGER, sliders INTEGER, spinners INTEGER, endTime INTEGER, setID INTEGER, mapID INTEGER);
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

	logger.Println("DatabaseManager: Database schema version:", currentSchemaPreVersion)
	logger.Println("DatabaseManager: Database data version:", currentPreVersion)

	if currentSchemaPreVersion != databaseVersion {
		logger.Println("DatabaseManager: Database schema is too old! Updating...")

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

		logger.Println("DatabaseManager: Schema has been updated!")
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

func LoadBeatmaps(skipDatabaseCheck bool) []*beatmap.BeatMap {
	if settings.General.UnpackOszFiles {
		unpackMaps()
	}

	if !skipDatabaseCheck {
		importMaps()
	}

	logger.Println("DatabaseManager: Loading beatmaps from database...")

	allMaps := loadBeatmapsFromDatabase()

	stdMaps := make([]*beatmap.BeatMap, 0, len(allMaps)/2)

	for _, b := range allMaps {
		if b.Mode == 0 {
			stdMaps = append(stdMaps, b)
		}
	}

	logger.Println("DatabaseManager: Loaded", len(stdMaps), "total.")

	return stdMaps
}

func unpackMaps() {
	_ = godirwalk.Walk(songsDir, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() && osPathname != songsDir {
				return godirwalk.SkipThis
			}

			if strings.HasSuffix(de.Name(), ".osz") {
				destination := filepath.Join(filepath.Dir(osPathname), strings.TrimSuffix(de.Name(), ".osz"))

				logger.Println("DatabaseManager: Unpacking", osPathname, "->", destination)

				utils.Unzip(osPathname, destination)
				os.Remove(osPathname)
			}

			return nil
		},
		Unsorted: true,
	})
}

func importMaps() {
	mapsInDB := getLastModified()
	candidates := make([]mapLocation, 0)

	logger.Println(fmt.Sprintf("DatabaseManager: Scanning \"%s\" for .osu files...", songsDir))

	err := godirwalk.Walk(songsDir, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			if de.IsDir() && osPathname != songsDir && filepath.Dir(osPathname) != songsDir {
				return godirwalk.SkipThis
			}

			if strings.HasSuffix(de.Name(), ".osu") {
				candidates = append(candidates, mapLocation{
					dir:  filepath.Base(filepath.Dir(osPathname)),
					file: de.Name(),
				})
			}

			return nil
		},
		Unsorted: true,
	})

	if err != nil {
		panic(err)
	}

	logger.Println("DatabaseManager: Scan complete. Found", len(candidates), "files.")
	logger.Println("DatabaseManager: Comparing files with database...")

	mapsToImport := make([]interface{}, 0)

	for _, candidate := range candidates {
		partialPath := filepath.Join(candidate.dir, candidate.file)
		mapPath := filepath.Join(songsDir, partialPath)

		stat, err := os.Stat(mapPath)
		if err != nil {
			logger.Println("DatabaseManager: Failed to read file stats, skipping:", partialPath)
			logger.Println("DatabaseManager: Error:", err)

			// If file does exist we assume it's a permission error, don't remove it from database in that case
			if !os.IsNotExist(err) {
				delete(mapsInDB, candidate)
			}

			continue
		}

		if lastModified, ok := mapsInDB[candidate]; ok {
			if lastModified == stat.ModTime().UnixNano()/1000000 {
				// Map is up to date, so remove it from mapsInDB because values left in that map are later removed from database.
				delete(mapsInDB, candidate)

				continue
			}

			logger.Println("DatabaseManager: New beatmap version found:", candidate.file)
		} else {
			logger.Println("DatabaseManager: New beatmap found:", candidate.file)
		}

		mapsToImport = append(mapsToImport, candidate)
	}

	logger.Println("DatabaseManager: Compare complete.")

	if len(mapsInDB) > 0 {
		logger.Println("DatabaseManager: Removing leftover maps from database...")

		mapsToRemove := make([]mapLocation, 0, len(mapsInDB))

		for k := range mapsInDB {
			mapsToRemove = append(mapsToRemove, k)
		}

		removeBeatmaps(mapsToRemove)

		logger.Println("DatabaseManager: Removal complete.")
	}

	if len(mapsToImport) > 0 {
		logger.Println("DatabaseManager: Starting import of", len(mapsToImport), "maps...")

		loaded := util.Balance(4, mapsToImport, func(a interface{}) interface{} {
			candidate := a.(mapLocation)

			partialPath := filepath.Join(candidate.dir, candidate.file)
			mapPath := filepath.Join(songsDir, partialPath)

			file, err := os.Open(mapPath)
			if err != nil {
				logger.Println(fmt.Sprintf("\"DatabaseManager: Failed to read \"%s\", skipping. Error: %s", partialPath, err))
				return nil
			}

			defer file.Close()

			logger.Println("DatabaseManager: Importing:", partialPath)

			if bMap := beatmap.ParseBeatMapFile(file); bMap != nil {
				stat, _ := file.Stat()
				bMap.LastModified = stat.ModTime().UnixNano() / 1000000
				bMap.TimeAdded = time.Now().UnixNano() / 1000000

				hash := md5.New()
				if _, err := io.Copy(hash, file); err == nil {
					bMap.MD5 = hex.EncodeToString(hash.Sum(nil))
				}

				logger.Println("DatabaseManager: Imported:", partialPath)
				return bMap
			} else {
				logger.Println("DatabaseManager: Failed to import:", partialPath)
			}

			return nil
		})

		newBeatmaps := make([]*beatmap.BeatMap, len(loaded))
		for i, o := range loaded {
			newBeatmaps[i] = o.(*beatmap.BeatMap)
		}

		logger.Println("DatabaseManager: Imported", len(newBeatmaps), "new/updated beatmaps. Inserting to database...")

		insertBeatmaps(newBeatmaps)

		logger.Println("DatabaseManager: Insert complete.")
	}
}

func UpdatePlayStats(beatmap *beatmap.BeatMap) {
	_, err := dbFile.Exec("UPDATE beatmaps SET playCount = ?, lastPlayed = ? WHERE dir = ? AND file = ?", beatmap.PlayCount, beatmap.LastPlayed, beatmap.Dir, beatmap.File)
	if err != nil {
		logger.Println(err)
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
					logger.Println(err1)
				}
			}
		} else {
			panic(err)
		}

		st.Close()
		tx.Commit()
	}

	if err != nil {
		logger.Println(err)
	}
}

func migrateBeatmaps() {
	lastModified := getLastModified()

	var removeList []mapLocation

	if currentPreVersion < databaseVersion {
		updateBeatmaps := false

		for _, m := range migrations {
			if currentPreVersion < m.Date() {
				updateBeatmaps = updateBeatmaps || m.FieldsToMigrate() != nil
			}
		}

		if updateBeatmaps {
			logger.Println("Updating cached beatmaps...")

			logger.Println("Loading cached beatmaps from disk...")

			toUpdate := make([]*beatmap.BeatMap, 0)

			for location := range lastModified {
				file, err := os.Open(filepath.Join(songsDir, location.dir, location.file))
				if err != nil {
					logger.Println("Failed to open file, removing from database:", location.file)
					logger.Println("Error:", err)

					removeList = append(removeList, location)

					continue
				}

				bMap := beatmap.ParseBeatMapFile(file)
				if bMap == nil {
					logger.Println("Corrupted cached beatmap found. Removing from database:", location.file)

					removeList = append(removeList, location)

					continue
				}

				toUpdate = append(toUpdate, bMap)
			}

			logger.Println("Cached beatmaps loaded! Performing migrations...")

			tx, err := dbFile.Begin()
			if err != nil {
				panic(err)
			}

			for _, m := range migrations {
				if currentPreVersion < m.Date() {
					logger.Println("Performing", m.Date(), "migration...")

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

			logger.Println("Committing migrations to database...")

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
		st, err = tx.Prepare("INSERT INTO beatmaps VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")

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
					bMap.SetID,
					bMap.ID,
				)

				if err1 != nil {
					logger.Println(err1)
				}
			}
		} else {
			panic(err)
		}

		st.Close()
		tx.Commit()
	}

	if err != nil {
		logger.Println(err)
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
		)

		beatMap.Diff.SetCS(mutils.ClampF64(cs, 0, 10))
		beatMap.Diff.SetAR(mutils.ClampF64(ar, 0, 10))
		beatMap.Diff.SetHPDrain(mutils.ClampF64(hp, 0, 10))
		beatMap.Diff.SetOD(mutils.ClampF64(od, 0, 10))

		beatmaps = append(beatmaps, beatMap)
	}

	return beatmaps
}

func getLastModified() map[mapLocation]int64 {
	res, _ := dbFile.Query("SELECT dir, file, lastModified FROM beatmaps")

	mod := make(map[mapLocation]int64)

	for res.Next() {
		var dir, file string
		var lastModified int64

		res.Scan(&dir, &file, &lastModified)

		mod[mapLocation{
			dir:  dir,
			file: file,
		}] = lastModified
	}

	return mod
}

func Close() {
	if dbFile != nil {
		err := dbFile.Close()
		if err != nil {
			logger.Println("Failed to close database:", err)
		}
	}
}
