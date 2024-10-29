package osuapi

import (
	"encoding/json"
	"io"
	"net/url"
	"strconv"
)

type ScoreType int

const (
	NormalMode ScoreType = iota
	FriendsMode
	CountryMode
)

func LookupBeatmap(checksum string) (int, error) {
	resp, err := makeRequest("beatmaps/lookup?checksum=" + checksum)

	if err != nil {
		return -1, err
	}

	buf, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		return -1, err
	}

	lRes := &LookupResult{}
	if err = json.Unmarshal(buf, &lRes); err != nil {
		return -1, err
	}

	return lRes.ID, nil
}

func GetScoresCheksum(checksum string, legacyOnly bool, mode ScoreType, limit int, mods ...string) ([]Score, error) {
	id, err := LookupBeatmap(checksum)

	if err != nil {
		return nil, err
	}

	return GetScores(id, legacyOnly, mode, limit, mods...)
}

func GetScores(beatmapId int, legacyOnly bool, mode ScoreType, limit int, mods ...string) ([]Score, error) {
	vls := url.Values{}

	prefix := "solo-"
	if legacyOnly {
		prefix = ""
		vls.Set("legacy_only", "1")
	}

	switch mode {
	case CountryMode:
		vls.Set("type", "country")
	case FriendsMode:
		vls.Set("type", "friend")
	}

	if limit > -1 {
		vls.Add("limit", strconv.Itoa(limit))
	}

	if len(mods) > 0 {
		for _, m := range mods {
			vls.Add("mods[]", m)
		}
	}

	resp, err := makeRequest("beatmaps/" + strconv.Itoa(beatmapId) + "/" + prefix + "scores?" + vls.Encode())

	if err != nil {
		return nil, err
	}

	buf, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		return nil, err
	}

	sRes := &ScoresResult{}
	if err = json.Unmarshal(buf, &sRes); err != nil {
		return nil, err
	}

	return sRes.Scores, nil
}

func LookupUser(nickname string) (*User, error) {
	resp, err := makeRequest("users/@" + nickname + "/osu")

	if err != nil {
		return nil, err
	}

	buf, err2 := io.ReadAll(resp.Body)
	if err2 != nil {
		return nil, err
	}

	user := &User{}
	if err = json.Unmarshal(buf, &user); err != nil {
		return nil, err
	}

	return user, nil
}
