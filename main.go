package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/drone/drone-go/drone"
	"golang.org/x/oauth2"
)

func main() {
	token := os.Getenv("PLUGIN_TOKEN")
	host := "https://cloud.drone.io"
	if hostPathArr := strings.Split(os.Getenv("DRONE_BUILD_LINK"), "/"); len(hostPathArr) > 2 {
		host = hostPathArr[0] + "//" + hostPathArr[2]
	}
	repoName := os.Getenv("DRONE_REPO_NAME")
	repoNamespace := os.Getenv("DRONE_REPO_NAMESPACE")
	buildNumber, _ := strconv.Atoi(os.Getenv("DRONE_BUILD_NUMBER"))
	currentStageNumber, _ := strconv.Atoi(os.Getenv("DRONE_STAGE_NUMBER"))
	currentBranch := os.Getenv("DRONE_COMMIT_BRANCH")
	onSuccess := os.Getenv("PLUGIN_ON_SUCCESS")
	if onSuccess == "" {
		onSuccess = "change"
	}
	onFailure := os.Getenv("PLUGIN_ON_FAILURE")
	if onFailure == "" {
		onFailure = "always"
	}
	currentBuildStatus := os.Getenv("DRONE_BUILD_STATUS")

	// create an http client with oauth authentication.
	config := new(oauth2.Config)
	auther := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	// create the drone client with authenticator
	client := drone.NewClient(host, auther)

	if len(repoNamespace) > 0 && len(repoName) > 0 {
		var showNotify int

		if onSuccess == "change" || onFailure == "change" {
			var lastBuild int
			// Get last build information
			for page, foundLastBuild := 1, 0; page <= 4 && foundLastBuild == 0; page++ {
				if gotBuildlist, err := client.BuildList(repoNamespace, repoName, drone.ListOptions{page}); err == nil {
					for _, element := range gotBuildlist {
						if currentBranch == element.Source && int64(buildNumber) > element.Number {
							fmt.Println(element.Source + "[" + strconv.FormatInt(element.Number, 10) + "]:" + element.Status)
							if element.Status == "success" {
								lastBuild = 1
							} else {
								lastBuild = 2
							}
							foundLastBuild = 1
							break
						}
					}
				}
			}

			if onSuccess == "change" && currentBuildStatus == "success" && lastBuild != 1 {
				showNotify = 1
			}
			if onFailure == "change" && currentBuildStatus != "success" && lastBuild == 1 {
				showNotify = 1
			}
		}
		if onSuccess == "always" && currentBuildStatus == "success" {
			showNotify = 1
		}
		if onFailure == "always" && currentBuildStatus != "success" {
			showNotify = 1
		}

		if showNotify > 0 {
			if gotBuild, err := client.Build(repoNamespace, repoName, buildNumber); err == nil {
				for index, element := range gotBuild.Stages {
					if index != currentStageNumber-1 {
						fmt.Println(element.Name + ":" + element.Status)
					}
				}
			}
		}
	} else {
		fmt.Println("No Repo or build.")
		return
	}

}
