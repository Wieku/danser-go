package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20210423 struct {}

func (m *M20210423) RequiredSections() []string {
	return []string{
		"Metadata",
	}
}

func (m *M20210423) FieldsToMigrate() []string {
	return []string{
		"setID",
		"mapID",
	}
}

func (m *M20210423) GetValues(beatMap *beatmap.BeatMap) []interface{} {
	return []interface{}{
		beatMap.SetID,
		beatMap.ID,
	}
}

func (m *M20210423) Date() int {
	return 20210423
}

func (m *M20210423) GetMigrationStmts() string {
	return `
		ALTER TABLE beatmaps ADD COLUMN setID INTEGER DEFAULT 0;
		ALTER TABLE beatmaps ADD COLUMN mapID INTEGER DEFAULT 0;`
}