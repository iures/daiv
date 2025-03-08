package github

import (
	"github.com/google/go-github/v68/github"
)

type GithubClientSettings struct {
	Username string
	Token string
	Org string
	Repos []string
}

type GithubClient struct {
	client *github.Client
	settings GithubClientSettings
}

func (gc *GithubClient) Init(settings GithubClientSettings) {
	authToken := github.BasicAuthTransport{
		Username: settings.Username,
		Password: settings.Token,
	}

	gc.client = github.NewClient(authToken.Client())
	gc.settings = settings
}
