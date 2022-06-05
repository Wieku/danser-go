package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20220605 struct{}

func (m *M20220605) RequiredSections() []string {
	return nil
}

func (m *M20220605) FieldsToMigrate() []string {
	return nil
}

func (m *M20220605) GetValues(_ *beatmap.BeatMap) []interface{} {
	return nil
}

func (m *M20220605) Date() int {
	return 20220605
}

func (m *M20220605) GetMigrationStmts() string {
	return "ALTER TABLE beatmaps ADD COLUMN starsVersion INTEGER DEFAULT 0;"
}
