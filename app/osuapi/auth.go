package osuapi

import (
	"context"
	"github.com/wieku/danser-go/app/settings"
	"github.com/wieku/danser-go/framework/goroutines"
	"github.com/wieku/danser-go/framework/platform"
	"github.com/wieku/danser-go/framework/util"
	"golang.org/x/oauth2"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AuthResult int

const (
	AuthSuccess AuthResult = iota
	AuthCancelled
	AuthError
)

const responseTemplate = `<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Auth</title>
		<link rel="icon" href="data:;base64,iVBORw0KGgo="> <!--hack to not request favicon.ico so server can be closed immediately-->
	</head>
	<body>
    	responseTemplate
	</body>
</html>
`

type AuthCallback func(result AuthResult, message string)

var clientConfig *oauth2.Config

func prepareConfig() {
	if clientConfig == nil {
		clientConfig = &oauth2.Config{
			Endpoint: oauth2.Endpoint{
				AuthURL:   "https://osu.ppy.sh/oauth/authorize",
				TokenURL:  "https://osu.ppy.sh/oauth/token",
				AuthStyle: oauth2.AuthStyleInParams,
			},
			Scopes: []string{"public", "identify"},
		}
	}

	clientConfig.RedirectURL = "http://localhost:" + strconv.Itoa(settings.Credentails.CallbackPort)
	clientConfig.ClientID = settings.Credentails.ClientId
	clientConfig.ClientSecret = settings.Credentails.ClientSecret
}

func Authorize(callback AuthCallback) {
	prepareConfig()
	settings.Credentails.AccessToken = ""
	settings.Credentails.RefreshToken = ""
	settings.Credentails.Expiry = time.Unix(0, 0)

	if settings.Credentails.ClientId == "" || settings.Credentails.ClientSecret == "" {
		callback(AuthError, "Please provide client id and client secret")
		return
	}

	if settings.Credentails.AuthType == "ClientCredentials" {
		authorizeClient(callback)
	} else {
		if settings.Credentails.CallbackPort < 0 || settings.Credentails.CallbackPort > 65535 {
			callback(AuthError, "Please provide a valid port number")
			return
		}

		authorizeCode(callback)
	}
}

func authorizeCode(callback AuthCallback) {
	securityCode := util.RandomHexString(64)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    "localhost:" + strconv.Itoa(settings.Credentails.CallbackPort),
		Handler: mux,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			goroutines.Run(func() {
				server.Shutdown(context.Background())
			})
		}()

		err := r.ParseForm()

		if err != nil {
			writeResponse(w, err.Error())

			callback(AuthError, err.Error())
			log.Println("ApiConnector: Failed to process auth:", err.Error())
			return
		}

		if r.Form.Get("state") != securityCode {
			writeResponse(w, "Invalid state")
			callback(AuthError, "Invalid state")
			log.Println("ApiConnector: Failed to process auth: Invalid state returned")
			return
		}

		if r.Form.Get("error") == "access_denied" {
			writeResponse(w, "User cancelled the auth request.")
			callback(AuthCancelled, "User cancelled the auth request.")
			log.Println("ApiConnector: Failed to process auth: Cancelled by user")
			return
		}

		if code := r.Form.Get("code"); code != "" {
			token, err2 := clientConfig.Exchange(context.Background(), code)

			if err2 != nil {
				writeResponse(w, err2.Error())
				callback(AuthError, err2.Error())
				log.Println("ApiConnector: Failed to process auth:", err.Error())
				return
			}

			tryUpdateToken(token)

			writeResponse(w, "Successful!")
			callback(AuthSuccess, "Successful!")
			log.Println("ApiConnector: Authorized!")
		}
	})

	platform.OpenURL(clientConfig.AuthCodeURL(securityCode))

	_ = server.ListenAndServe()
}

func authorizeClient(callback AuthCallback) {
	token, err := exchangeClientCredentials(clientConfig, context.Background())
	if err != nil {
		callback(AuthError, err.Error())
		return
	}

	tryUpdateToken(token)
	callback(AuthSuccess, "Successful!")
}

func writeResponse(w http.ResponseWriter, s string) {
	io.WriteString(w, strings.ReplaceAll(responseTemplate, "responseTemplate", s))
}

func exchangeClientCredentials(c *oauth2.Config, ctx context.Context) (token *oauth2.Token, err error) {
	// unlike code param, osu! client credentials grant fails if redirect_uri is provided, so we need to clear it temporarily
	tempURI := c.RedirectURL
	c.RedirectURL = ""

	token, err = c.Exchange(ctx, "1234", oauth2.SetAuthURLParam("grant_type", "client_credentials"), oauth2.SetAuthURLParam("scope", "public"))

	c.RedirectURL = tempURI

	return
}
