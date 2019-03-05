package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-template-lib/template"
	"golang.org/x/oauth2"
)

type Pipeline struct {
	Name  string `json:"pipename"`
	Status  string `json:"pipestatus`
}

type (
	Repo struct {
		Owner string `json:"owner"`
		Name  string `json:"name"`
	}

	Build struct {
		Tag     string `json:"tag"`
		Event   string `json:"event"`
		Number  int    `json:"number"`
		Commit  string `json:"commit"`
		Ref     string `json:"ref"`
		Branch  string `json:"branch"`
		Author  string `json:"author"`
		Message string `json:"message"`
		Status  string `json:"status"`
		Link    string `json:"link"`
		Started int64  `json:"started"`
		Created int64  `json:"created"`
	}

	Config struct {
		Method        string
		Username      string
		Password      string
		ContentType   string
		Template      string
		Headers       []string
		URLs          []string
		ValidCodes    []int
		Debug         bool
		SkipVerify    bool
		Token         string
		OnSuccess     string
		OnFailure     string
		PipelineName  string
	}

	Job struct {
		Started int64
		Status []Pipeline `json:"jobstatus"`
	}

	Plugin struct {
		Repo   Repo
		Build  Build
		Config Config
		Job    Job
	}
)

func (p Plugin) Exec() error {
	var (
		buf bytes.Buffer
		b   []byte
	)

	host := "https://cloud.drone.io"
	if hostPathArr := strings.Split(p.Build.Link, "/"); len(hostPathArr) > 2 {
		host = hostPathArr[0] + "//" + hostPathArr[2]
	}

	// create an http client with oauth authentication.
	config := new(oauth2.Config)
	auther := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: p.Config.Token,
		},
	)
	// create the drone client with authenticator
	client := drone.NewClient(host, auther)

	if len(p.Repo.Owner) > 0 && len(p.Repo.Name) > 0 {
		var showNotify int

		if p.Config.OnSuccess == "change" || p.Config.OnFailure == "change" {
			var lastBuild int
			// Get last build information
			for page, foundLastBuild := 1, 0; page <= 4 && foundLastBuild == 0; page++ {
				if gotBuildlist, err := client.BuildList(p.Repo.Owner, p.Repo.Name, drone.ListOptions{page}); err == nil {
					for _, element := range gotBuildlist {
						if p.Build.Branch == element.Source && int64(p.Build.Number) > element.Number {
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

			if p.Config.OnSuccess == "change" && p.Build.Status == "success" && lastBuild != 1 {
				showNotify = 1
			}
			if p.Config.OnFailure == "change" && p.Build.Status != "success" && lastBuild == 1 {
				showNotify = 1
			}
		}

		if p.Config.OnSuccess == "always" && p.Build.Status == "success" {
			showNotify = 1
		}
		if p.Config.OnFailure == "always" && p.Build.Status != "success" {
			showNotify = 1
		}

		if showNotify > 0 {
			fmt.Println(p)
			if gotBuild, err := client.Build(p.Repo.Owner, p.Repo.Name, p.Build.Number); err == nil {
				for _, element := range gotBuild.Stages {
					if p.Config.PipelineName != element.Name {
						var pipe Pipeline
						pipe.Name = element.Name
						pipe.Status = element.Status
						p.Job.Status = append(p.Job.Status, pipe)
						if element.Status != "success" {
							p.Build.Status = element.Status
						}
					}
				}
			}

			if p.Config.Template == "" {
				data := struct {
					Repo  Repo  `json:"repo"`
					Build Build `json:"build"`
				}{p.Repo, p.Build}

				if err := json.NewEncoder(&buf).Encode(&data); err != nil {
					fmt.Printf("Error: Failed to encode JSON payload. %s\n", err)
					return err
				}

				b = buf.Bytes()
			} else {
				txt, err := template.RenderTrim(p.Config.Template, p)

				if err != nil {
					return err
				}

				text := txt
				b = []byte(text)
			}

			// build and execute a request for each url.
			// all auth, headers, method, template (payload),
			// and content_type values will be applied to
			// every webhook request.

			for i, rawurl := range p.Config.URLs {
				uri, err := url.Parse(rawurl)

				if err != nil {
					fmt.Printf("Error: Failed to parse the hook URL. %s\n", err)
					os.Exit(1)
				}

				r := bytes.NewReader(b)
				req, err := http.NewRequest(p.Config.Method, uri.String(), r)

				if err != nil {
					fmt.Printf("Error: Failed to create the HTTP request. %s\n", err)
					return err
				}

				req.Header.Set("Content-Type", p.Config.ContentType)

				for _, value := range p.Config.Headers {
					header := strings.Split(value, "=")
					req.Header.Set(header[0], header[1])
				}

				if p.Config.Username != "" && p.Config.Password != "" {
					req.SetBasicAuth(p.Config.Username, p.Config.Password)
				}

				client := http.DefaultClient

				if p.Config.SkipVerify {
					client = &http.Client{
						Transport: &http.Transport{
							TLSClientConfig: &tls.Config{
								InsecureSkipVerify: true,
							},
						},
					}
				}

				resp, err := client.Do(req)

				if err != nil {
					fmt.Printf("Error: Failed to execute the HTTP request. %s\n", err)
					return err
				}

				defer resp.Body.Close()

				if p.Config.Debug || resp.StatusCode >= http.StatusBadRequest {
					body, err := ioutil.ReadAll(resp.Body)

					if err != nil {
						fmt.Printf("Error: Failed to read the HTTP response body. %s\n", err)
					}

					if p.Config.Debug {
						fmt.Printf(
							debugFormat,
							i+1,
							req.URL,
							req.Method,
							req.Header,
							string(b),
							resp.Status,
							string(body),
						)
					} else {
						fmt.Printf(
							respFormat,
							i+1,
							req.URL,
							resp.Status,
							string(body),
						)
					}
				}

				if len(p.Config.ValidCodes) > 0 && !intInSlice(p.Config.ValidCodes, resp.StatusCode) {
					return fmt.Errorf("Error: Response code %d not found among valid response codes", resp.StatusCode)
				}
			}
		}
	}

	return nil
}

// Function checks if int is in slice of ints
func intInSlice(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
