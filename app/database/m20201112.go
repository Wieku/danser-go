package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20201112 struct{}

func (m *M20201112) RequiredSections() []string {
	return []string{
		"General",
	}
}

func (m *M20201112) FieldsToMigrate() []string {
	return []string{
		"previewTime",
	}
}

func (m *M20201112) GetValues(beatMap *beatmap.BeatMap) []interface{} {
	return []interface{}{
		beatMap.PreviewTime,
	}
}

func (m *M20201112) Date() int {
	return 20201112
}

func (m *M20201112) GetMigrationStmts() string {
	return ""
}
