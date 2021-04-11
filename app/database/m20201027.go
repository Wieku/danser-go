package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20201027 struct {}

func (m *M20201027) RequiredSections() []string {
	return nil
}

func (m *M20201027) FieldsToMigrate() []string {
	return nil
}

func (m *M20201027) GetValues(_ *beatmap.BeatMap) []interface{} {
	return nil
}

func (m *M20201027) Date() int {
	return 20201027
}

func (m *M20201027) GetMigrationStmts() string {
	return `
		BEGIN TRANSACTION;
		CREATE TEMPORARY TABLE beatmaps_backup(dir TEXT, file TEXT, lastModified INTEGER, title TEXT, titleUnicode TEXT, artist TEXT, artistUnicode TEXT, creator TEXT, version TEXT, source TEXT, tags TEXT, cs REAL, ar REAL, sliderMultiplier REAL, sliderTickRate REAL, audioFile TEXT, previewTime INTEGER, sampleSet INTEGER, stackLeniency REAL, mode INTEGER, bg TEXT, md5 TEXT, dateAdded INTEGER, playCount INTEGER, lastPlayed INTEGER, hpdrain REAL, od REAL);
		INSERT INTO beatmaps_backup SELECT dir, file, lastModified, title, titleUnicode, artist, artistUnicode, creator, version, source, tags, cs, ar, sliderMultiplier, sliderTickRate, audioFile, previewTime, sampleSet, stackLeniency, mode, bg, md5, dateAdded, playCount, lastPlayed, hpdrain, od FROM beatmaps;
		DROP TABLE beatmaps;
		CREATE TABLE beatmaps(dir TEXT, file TEXT, lastModified INTEGER, title TEXT, titleUnicode TEXT, artist TEXT, artistUnicode TEXT, creator TEXT, version TEXT, source TEXT, tags TEXT, cs REAL, ar REAL, sliderMultiplier REAL, sliderTickRate REAL, audioFile TEXT, previewTime INTEGER, sampleSet INTEGER, stackLeniency REAL, mode INTEGER, bg TEXT, md5 TEXT, dateAdded INTEGER, playCount INTEGER, lastPlayed INTEGER, hpdrain REAL, od REAL);
		INSERT INTO beatmaps SELECT * FROM beatmaps_backup;
		DROP TABLE beatmaps_backup;
		CREATE INDEX IF NOT EXISTS idx ON beatmaps (dir, file);
		COMMIT;
		vacuum;`
}