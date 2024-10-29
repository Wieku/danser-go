package settings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wieku/danser-go/framework/env"
	"github.com/wieku/danser-go/framework/files"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var Credentails = &credentials{
	AuthType:     "ClientCredentials",
	CallbackPort: 8294,
}

type credentials struct {
	ClientId     string
	ClientSecret string `long:"true" password:"true"`

	AuthType string `combo:"ClientCredentials|Client credentials (Anonymous),AuthorizationCode|Authorization code (User authenticated)"`
	//
	CallbackPort int `string:"true" min:"0" max:"65535" showif:"AuthType=AuthorizationCode"`

	AccessToken  string    `skip:"true" long:"true" password:"true"`
	Expiry       time.Time `skip:"true"`
	RefreshToken string    `skip:"true" long:"true" password:"true" showif:"AuthType=AuthorizationCode"`
}

var srcDataCred []byte

func LoadCredentials() {
	if err := os.MkdirAll(env.ConfigDir(), 0755); err != nil {
		panic(err)
	}

	file, err := os.Open(filepath.Join(env.ConfigDir(), "credentials.json"))

	if os.IsNotExist(err) {
		SaveCredentials(true)
	} else if err != nil {
		panic(err)
	} else {
		defer file.Close()

		loadCredentials(file)

		SaveCredentials(false) // this is done to save additions from the current format
	}
}

func loadCredentials(file *os.File) {
	log.Println(fmt.Sprintf(`ApiConnector: Loading "%s"`, file.Name()))

	data, err := io.ReadAll(files.NewUnicodeReader(file))
	if err != nil {
		panic(err)
	}

	srcDataCred = data

	if err = json.Unmarshal(data, Credentails); err != nil {
		panic(fmt.Sprintf("ApiConnector: Failed to parse %s! Please re-check the file for mistakes. Error: %s", file.Name(), err))
	}
}

func SaveCredentials(forceSave bool) {
	data, err := json.MarshalIndent(Credentails, "", "\t")
	if err != nil {
		panic(err)
	}

	fPath := filepath.Join(env.ConfigDir(), "credentials.json")

	if forceSave || !bytes.Equal(data, srcDataCred) { // Don't rewrite the file unless necessary
		log.Println(fmt.Sprintf(`ApiConnector: Saving current settings to "%s"`, fPath))

		srcDataCred = data

		if err = os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
			panic(err)
		}

		if err = os.WriteFile(fPath, data, 0644); err != nil {
			panic(err)
		}
	}
}
