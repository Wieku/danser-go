package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20201118 struct {}

func (m *M20201118) RequiredSections() []string {
	return []string{
		"TimingPoints",
		"HitObjectStats",
	}
}

func (m *M20201118) FieldsToMigrate() []string {
	return []string{
		"bpmMin",
		"bpmMax",
		"circles",
		"sliders",
		"spinners",
		"endTime",
	}
}

func (m *M20201118) GetValues(beatMap *beatmap.BeatMap) []interface{} {
	return []interface{}{
		beatMap.MinBPM,
		beatMap.MaxBPM,
		beatMap.Circles,
		beatMap.Sliders,
		beatMap.Spinners,
		beatMap.Length,
	}
}

func (m *M20201118) Date() int {
	return 20201118
}

func (m *M20201118) GetMigrationStmts() string {
	return `
		ALTER TABLE beatmaps ADD COLUMN bpmMin REAL DEFAULT 0;
		ALTER TABLE beatmaps ADD COLUMN bpmMax REAL DEFAULT 0;
		ALTER TABLE beatmaps ADD COLUMN circles INTEGER DEFAULT 0;
		ALTER TABLE beatmaps ADD COLUMN sliders INTEGER DEFAULT 0;
		ALTER TABLE beatmaps ADD COLUMN spinners INTEGER DEFAULT 0;
		ALTER TABLE beatmaps ADD COLUMN endTime INTEGER DEFAULT 0;`
}