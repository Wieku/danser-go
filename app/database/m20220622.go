package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20220622 struct{}

func (m *M20220622) RequiredSections() []string {
	return nil
}

func (m *M20220622) FieldsToMigrate() []string {
	return nil
}

func (m *M20220622) GetValues(_ *beatmap.BeatMap) []interface{} {
	return nil
}

func (m *M20220622) Date() int {
	return 20220622
}

func (m *M20220622) GetMigrationStmts() string {
	return "ALTER TABLE beatmaps ADD COLUMN localOffset INTEGER DEFAULT 0;"
}
