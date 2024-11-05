package osuapi

type ScoresResult struct {
	Scores []Score `json:"scores"`
}

type User struct {
	AvatarURL   string `json:"avatar_url"`
	CountryCode string `json:"country_code"`
	ID          int64  `json:"id"`
	Username    string `json:"username"`
}

type Score struct {
	ClassicTotalScore     int64   `json:"classic_total_score"`
	BeatmapID             int64   `json:"beatmap_id"`
	ID                    int64   `json:"id"`
	UserID                int64   `json:"user_id"`
	Accuracy              float64 `json:"accuracy"`
	LegacyScoreID         int64   `json:"legacy_score_id"`
	LegacyTotalScore      int64   `json:"legacy_total_score"`
	MaxCombo              int64   `json:"max_combo"`
	Score                 int64   `json:"score"`
	TotalScore            int64   `json:"total_score"`
	User                  User    `json:"user"`
	TotalScoreWithoutMods int64   `json:"total_score_without_mods,omitempty"`
}

type LookupResult struct {
	BeatmapsetID int64      `json:"beatmapset_id"`
	ID           int64      `json:"id"`
	Mode         string     `json:"mode"`
	URL          string     `json:"url"`
	Checksum     string     `json:"checksum"`
	Beatmapset   Beatmapset `json:"beatmapset"`
}

type Beatmapset struct {
	ID     int64   `json:"id"`
	Offset float64 `json:"offset"`
}
