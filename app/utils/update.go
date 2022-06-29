package utils

import (
	"encoding/json"
	"fmt"
	"github.com/wieku/danser-go/build"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// GetLatestVersionFromGitHub makes a request to GitHub and returns url and tag of the latest version found
func GetLatestVersionFromGitHub() (url string, tag string, err error) {
	request, err := http.NewRequest(http.MethodGet, "https://api.github.com/repos/Wieku/danser-go/releases/latest", nil)
	if err != nil {
		return "", "", err
	}

	client := new(http.Client)
	response, err := client.Do(request)

	if err != nil || response.StatusCode != 200 {
		return "", "", err
	}

	defer func() {
		err := response.Body.Close()
		if err != nil {
			panic(err)
		}
	}()

	var data struct {
		URL string `json:"html_url"`
		Tag string `json:"tag_name"`
	}

	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return "", "", err
	}

	return data.URL, data.Tag, nil
}

// TransformVersion transfers version to a more comparable format.
//   - 0.6.7 becomes 600079999
//   - 0.6.7-s(napshot)12 becomes 600070012
//   - 1.0.0 becomes 1000000009999
func TransformVersion(version string) uint64 {
	currentSplit := strings.Split(version, "-")
	splitDots := strings.Split(strings.TrimSuffix(currentSplit[0], "b"), ".")

	for i, s := range splitDots {
		splitDots[i] = fmt.Sprintf("%04s", s)
	}

	snapshot := "9999"
	if len(currentSplit) > 1 && !strings.HasPrefix(currentSplit[1], "dev") {
		snapshot = fmt.Sprintf("%04s", strings.TrimPrefix(strings.TrimPrefix(currentSplit[1], "s"), "napshot"))
	}

	versionInt, err := strconv.ParseUint(strings.Join(splitDots, "")+snapshot, 10, 64)
	if err != nil {
		panic(err)
	}

	return versionInt
}

type UpdateStatus int

const (
	Failed = UpdateStatus(iota)
	Ignored
	UpToDate
	Snapshot
	UpdateAvailable
)

func CheckForUpdate() (UpdateStatus, string, error) {
	if build.Stream != "Release" || strings.Contains(build.VERSION, "dev") { // false positive, those are changed during compile
		return Ignored, "", nil
	}

	log.Println("Checking Github for a new version of danser...")

	url, tag, err := GetLatestVersionFromGitHub()
	if err != nil {
		return Failed, "", err
	}

	githubVersion := TransformVersion(tag)
	exeVersion := TransformVersion(build.VERSION)

	if exeVersion >= githubVersion {
		if strings.Contains(build.VERSION, "snapshot") {
			return Snapshot, "https://wieku.me/lair", nil
		} else {
			return UpToDate, "", nil
		}
	} else {
		return UpdateAvailable, url, nil
	}
}
