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
	"strings"
)

var Credentails = &credentials{}

type credentials struct {
	ApiV1Key string `label:"API V1 Key" long:"true" password:"true" tooltip:"Valid API V1 Key has to be provided to have access to global leaderboards\n\nDON'T SHARE IT WITH OTHERS!!!" hidePath:"true"`

	//Future stuff
	//ClientId     string
	//ClientSecret string
	//
	//CallbackPort uint
	//
	//AccessToken  string
	//ExpireDate   time.Time
	//RefreshToken string
}

var srcDataCred []byte

func LoadCredentials() {
	migrateKey()

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

func migrateKey() {
	_, err := os.Stat(filepath.Join(env.DataDir(), "api.txt"))
	if err == nil {
		err = os.Rename(filepath.Join(env.DataDir(), "api.txt"), filepath.Join(env.ConfigDir(), "api.txt"))
		if err != nil {
			panic(err)
		}
	}

	file, err := os.Open(filepath.Join(env.ConfigDir(), "api.txt"))
	if err == nil {
		var data []byte

		data, err = io.ReadAll(files.NewUnicodeReader(file))

		file.Close()

		if err == nil {
			Credentails.ApiV1Key = strings.TrimSpace(string(data))
			SaveCredentials(false)

			os.Remove(filepath.Join(env.ConfigDir(), "api.txt"))
		}
	}
}
