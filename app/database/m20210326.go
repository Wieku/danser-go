package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20210326 struct {}

func (m *M20210326) RequiredSections() []string {
	return []string{
		"Difficulty",
	}
}

func (m *M20210326) FieldsToMigrate() []string {
	return []string{
		"ar",
	}
}

func (m *M20210326) GetValues(beatMap *beatmap.BeatMap) []interface{} {
	return []interface{}{
		beatMap.Diff.GetAR(),
	}
}

func (m *M20210326) Date() int {
	return 20210326
}

func (m *M20210326) GetMigrationStmts() string {
	return ""
}