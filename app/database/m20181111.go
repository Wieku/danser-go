package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20181111 struct{}

func (m *M20181111) RequiredSections() []string {
	return []string{
		"Difficulty",
		"HitObjectStats",
	}
}

func (m *M20181111) FieldsToMigrate() []string {
	return []string{
		"hpdrain",
		"od",
	}
}

func (m *M20181111) GetValues(beatMap *beatmap.BeatMap) []interface{} {
	return []interface{}{
		beatMap.Diff.GetHPDrain(),
		beatMap.Diff.GetOD(),
	}
}

func (m *M20181111) Date() int {
	return 20181111
}

func (m *M20181111) GetMigrationStmts() string {
	return `
		ALTER TABLE beatmaps ADD COLUMN hpdrain REAL;
		ALTER TABLE beatmaps ADD COLUMN od REAL;`
}
