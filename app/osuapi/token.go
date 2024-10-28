package osuapi

import (
	"context"
	"github.com/wieku/danser-go/app/settings"
	"golang.org/x/oauth2"
	"time"
)

func getToken() *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  settings.Credentails.AccessToken,
		TokenType:    "Bearer",
		RefreshToken: settings.Credentails.RefreshToken,
		Expiry:       settings.Credentails.Expiry,
		ExpiresIn:    int64(settings.Credentails.Expiry.Sub(time.Now()).Seconds()),
	}
}

func getTokenSource() (oauth2.TokenSource, error) {
	prepareConfig()

	token := getToken()

	if settings.Credentails.AuthType == "ClientCredentials" {
		if !token.Valid() {
			var err error

			token, err = exchangeClientCredentials(clientConfig, context.Background())

			if err != nil {
				return nil, err
			}
		}

		return oauth2.StaticTokenSource(token), nil
	}

	return clientConfig.TokenSource(context.Background(), token), nil
}

func TryRefreshToken() error {
	if settings.Credentails.AccessToken == "" {
		return nil
	}

	tSource, err := getTokenSource()

	if err != nil {
		return err
	}

	tk, err := tSource.Token()

	if err != nil {
		return err
	}

	tryUpdateToken(tk)

	return nil
}

func tryUpdateToken(token *oauth2.Token) {
	if settings.Credentails.AccessToken == token.AccessToken &&
		settings.Credentails.RefreshToken == token.RefreshToken &&
		settings.Credentails.Expiry == token.Expiry {
		return
	}

	settings.Credentails.AccessToken = token.AccessToken
	settings.Credentails.RefreshToken = token.RefreshToken
	settings.Credentails.Expiry = token.Expiry

	settings.SaveCredentials(false)
}
