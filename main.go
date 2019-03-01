package main

import (
	"github.com/drone/drone-go/drone"
	"golang.org/x/oauth2"
	"fmt"
	"os"
	"strings"
	"strconv"
)

func main() {
 	token := os.Getenv("PLUGIN_TOKEN")
	host := "https://cloud.drone.io"
	if host_path_arr := strings.Split(os.Getenv("DRONE_BUILD_LINK"), "/"); len(host_path_arr) > 2 {
		host = host_path_arr[0]+"//"+host_path_arr[2]
	}
	repo_name := os.Getenv("DRONE_REPO_NAME")
	repo_namespace := os.Getenv("DRONE_REPO_NAMESPACE")
	build_number, _ := strconv.Atoi(os.Getenv("DRONE_BUILD_NUMBER"))
//	current_stage_number, _ := strconv.Atoi(os.Getenv("DRONE_STAGE_NUMBER"))
	current_branch := os.Getenv("DRONE_COMMIT_BRANCH")

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

	if len(repo_namespace) > 0 && len(repo_name) > 0 {
		got, err := client.BuildLast(repo_namespace, repo_name, current_branch)
		fmt.Println(got.Status, err)

		gotBuild, err := client.Build(repo_namespace, repo_name, build_number)
		//for index, element := range gotBuild.Stages {
		for _, element := range gotBuild.Stages {
			//if index != current_stage_number {
				fmt.Println(element.Name+":"+element.Status, err)
			//}
		}
	} else {
		fmt.Println("No Repo or build.")
		return
	}

}