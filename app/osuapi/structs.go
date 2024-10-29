package osuapi

import (
	"time"
)

type ScoresResult struct {
	Scores []Score `json:"scores"`
}

type Mods struct {
	Acronym string `json:"acronym"`
}

type CurrentUserAttributes struct {
	Pin any `json:"pin"`
}

type Country struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Cover struct {
	CustomURL string `json:"custom_url"`
	URL       string `json:"url"`
	ID        any    `json:"id"`
}

type User struct {
	AvatarURL     string  `json:"avatar_url"`
	CountryCode   string  `json:"country_code"`
	DefaultGroup  string  `json:"default_group"`
	ID            int     `json:"id"`
	IsActive      bool    `json:"is_active"`
	IsBot         bool    `json:"is_bot"`
	IsDeleted     bool    `json:"is_deleted"`
	IsOnline      bool    `json:"is_online"`
	IsSupporter   bool    `json:"is_supporter"`
	LastVisit     any     `json:"last_visit"`
	PmFriendsOnly bool    `json:"pm_friends_only"`
	ProfileColour any     `json:"profile_colour"`
	Username      string  `json:"username"`
	Country       Country `json:"country"`
	Cover         Cover   `json:"cover"`
}

type Score struct {
	ClassicTotalScore     int64                 `json:"classic_total_score"`
	Preserve              bool                  `json:"preserve"`
	Processed             bool                  `json:"processed"`
	Ranked                bool                  `json:"ranked"`
	MaximumStatistics     map[string]int64      `json:"maximum_statistics,omitempty"`
	Mods                  []any                 `json:"mods"`
	Statistics            map[string]int64      `json:"statistics,omitempty"`
	BeatmapID             int                   `json:"beatmap_id"`
	BestID                any                   `json:"best_id"`
	ID                    int64                 `json:"id"`
	Rank                  string                `json:"rank"`
	Type                  string                `json:"type"`
	UserID                int                   `json:"user_id"`
	Accuracy              float64               `json:"accuracy"`
	BuildID               any                   `json:"build_id"`
	EndedAt               time.Time             `json:"ended_at"`
	HasReplay             bool                  `json:"has_replay"`
	IsPerfectCombo        bool                  `json:"is_perfect_combo"`
	LegacyPerfect         bool                  `json:"legacy_perfect"`
	LegacyScoreID         int64                 `json:"legacy_score_id"`
	LegacyTotalScore      int64                 `json:"legacy_total_score"`
	MaxCombo              int64                 `json:"max_combo"`
	Passed                bool                  `json:"passed"`
	Pp                    float64               `json:"pp"`
	RulesetID             int                   `json:"ruleset_id"`
	StartedAt             any                   `json:"started_at"`
	Score                 int64                 `json:"score"`
	TotalScore            int64                 `json:"total_score"`
	Replay                bool                  `json:"replay"`
	CurrentUserAttributes CurrentUserAttributes `json:"current_user_attributes"`
	User                  User                  `json:"user"`
	TotalScoreWithoutMods int64                 `json:"total_score_without_mods,omitempty"`
}

type LookupResult struct {
	BeatmapsetID     int        `json:"beatmapset_id"`
	DifficultyRating float64    `json:"difficulty_rating"`
	ID               int        `json:"id"`
	Mode             string     `json:"mode"`
	Status           string     `json:"status"`
	TotalLength      int        `json:"total_length"`
	UserID           int        `json:"user_id"`
	Version          string     `json:"version"`
	Accuracy         float64    `json:"accuracy"`
	Ar               float64    `json:"ar"`
	Bpm              float64    `json:"bpm"`
	Convert          bool       `json:"convert"`
	CountCircles     int        `json:"count_circles"`
	CountSliders     int        `json:"count_sliders"`
	CountSpinners    int        `json:"count_spinners"`
	Cs               float64    `json:"cs"`
	DeletedAt        any        `json:"deleted_at"`
	Drain            float64    `json:"drain"`
	HitLength        int        `json:"hit_length"`
	IsScoreable      bool       `json:"is_scoreable"`
	LastUpdated      time.Time  `json:"last_updated"`
	ModeInt          int        `json:"mode_int"`
	Passcount        int        `json:"passcount"`
	Playcount        int        `json:"playcount"`
	Ranked           int        `json:"ranked"`
	URL              string     `json:"url"`
	Checksum         string     `json:"checksum"`
	Beatmapset       Beatmapset `json:"beatmapset"`
	Failtimes        Failtimes  `json:"failtimes"`
	MaxCombo         int        `json:"max_combo"`
}

type Covers struct {
	Cover       string `json:"cover"`
	Cover2X     string `json:"cover@2x"`
	Card        string `json:"card"`
	Card2X      string `json:"card@2x"`
	List        string `json:"list"`
	List2X      string `json:"list@2x"`
	Slimcover   string `json:"slimcover"`
	Slimcover2X string `json:"slimcover@2x"`
}

type RequiredMeta struct {
	MainRuleset    int `json:"main_ruleset"`
	NonMainRuleset int `json:"non_main_ruleset"`
}

type NominationsSummary struct {
	Current              int          `json:"current"`
	EligibleMainRulesets []string     `json:"eligible_main_rulesets"`
	RequiredMeta         RequiredMeta `json:"required_meta"`
}

type Availability struct {
	DownloadDisabled bool `json:"download_disabled"`
	MoreInformation  any  `json:"more_information"`
}

type Beatmapset struct {
	Artist             string             `json:"artist"`
	ArtistUnicode      string             `json:"artist_unicode"`
	Covers             Covers             `json:"covers"`
	Creator            string             `json:"creator"`
	FavouriteCount     int                `json:"favourite_count"`
	Hype               any                `json:"hype"`
	ID                 int                `json:"id"`
	Nsfw               bool               `json:"nsfw"`
	Offset             int                `json:"offset"`
	PlayCount          int                `json:"play_count"`
	PreviewURL         string             `json:"preview_url"`
	Source             string             `json:"source"`
	Spotlight          bool               `json:"spotlight"`
	Status             string             `json:"status"`
	Title              string             `json:"title"`
	TitleUnicode       string             `json:"title_unicode"`
	TrackID            any                `json:"track_id"`
	UserID             int                `json:"user_id"`
	Video              bool               `json:"video"`
	Bpm                int                `json:"bpm"`
	CanBeHyped         bool               `json:"can_be_hyped"`
	DeletedAt          any                `json:"deleted_at"`
	DiscussionEnabled  bool               `json:"discussion_enabled"`
	DiscussionLocked   bool               `json:"discussion_locked"`
	IsScoreable        bool               `json:"is_scoreable"`
	LastUpdated        time.Time          `json:"last_updated"`
	LegacyThreadURL    string             `json:"legacy_thread_url"`
	NominationsSummary NominationsSummary `json:"nominations_summary"`
	Ranked             int                `json:"ranked"`
	RankedDate         time.Time          `json:"ranked_date"`
	Storyboard         bool               `json:"storyboard"`
	SubmittedDate      time.Time          `json:"submitted_date"`
	Tags               string             `json:"tags"`
	Availability       Availability       `json:"availability"`
	Ratings            []int              `json:"ratings"`
}

type Failtimes struct {
	Fail []int `json:"fail"`
	Exit []int `json:"exit"`
}
