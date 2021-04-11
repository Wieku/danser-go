package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20201117 struct {}

func (m *M20201117) RequiredSections() []string {
	return nil
}

func (m *M20201117) FieldsToMigrate() []string {
	return nil
}

func (m *M20201117) GetValues(_ *beatmap.BeatMap) []interface{} {
	return nil
}

func (m *M20201117) Date() int {
	return 20201117
}

func (m *M20201117) GetMigrationStmts() string {
	return "ALTER TABLE beatmaps ADD COLUMN stars REAL DEFAULT -1;"
}