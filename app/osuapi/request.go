package osuapi

import (
	"context"
	"errors"
	"golang.org/x/oauth2"
	"net/http"
)

const osuEndpoint = "https://osu.ppy.sh/api/v2/"

func makeRequest(reqString string) (res *http.Response, err error) {
	src, err := getTokenSource()

	if err != nil {
		return nil, err
	}

	res, err = oauth2.NewClient(context.Background(), src).Get(osuEndpoint + reqString)

	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New(res.Status)
	}

	tk, _ := src.Token()

	tryUpdateToken(tk)

	return
}
