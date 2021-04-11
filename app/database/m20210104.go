package database

import (
	"github.com/wieku/danser-go/app/beatmap"
)

type M20210104 struct {}

func (m *M20210104) RequiredSections() []string {
	return []string{
		"Metadata",
	}
}

func (m *M20210104) FieldsToMigrate() []string {
	return []string{
		"title",
		"titleUnicode",
		"artist",
		"artistUnicode",
		"creator",
		"version",
		"`source`",
		"tags",
	}
}

func (m *M20210104) GetValues(beatMap *beatmap.BeatMap) []interface{} {
	return []interface{}{
		beatMap.Name,
		beatMap.NameUnicode,
		beatMap.Artist,
		beatMap.ArtistUnicode,
		beatMap.Creator,
		beatMap.Difficulty,
		beatMap.Source,
		beatMap.Tags,
	}
}

func (m *M20210104) Date() int {
	return 20210104
}

func (m *M20210104) GetMigrationStmts() string {
	return ""
}