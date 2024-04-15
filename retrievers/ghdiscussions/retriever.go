package ghdiscussions

import (
	"os"

	"github.com/ry023/reviewhub/reviewhub"
)

type GitHubDiscussionsRetriever struct {
}

type MetaData struct {
	RepositoryOwner string `yaml:"repository_owner"`
	RepositoryName  string `yaml:"repository_name"`
	ApiTokenEnv     string `yaml:"api_token_env" validate:"required"`
	ApiEndpoint     string `yaml:"api_endpoint"`
}

type UserMetaData struct {
	GitHubId string `yaml:"github_id"`
}

const defaultApiEndpoint = "https://api.github.com/graphql"

func (p *GitHubDiscussionsRetriever) Retrieve(config reviewhub.RetrieverConfig, knownUsers []reviewhub.User) (*reviewhub.ReviewList, error) {
	meta, err := reviewhub.ParseMetaData[MetaData](config.MetaData)
	if err != nil {
		return nil, err
	}

	token := os.Getenv(meta.ApiTokenEnv)
	apiEndpoint := defaultApiEndpoint
	if meta.ApiEndpoint != "" {
		apiEndpoint = meta.ApiEndpoint
	}

	l := []reviewhub.ReviewPage{}
	pages, err := request(meta.RepositoryOwner, meta.RepositoryName, token, apiEndpoint)
	if err != nil {
		return nil, err
	}

	for _, u := range knownUsers {
		umeta, err := reviewhub.ParseMetaData[UserMetaData](u.MetaData)
		if err != nil {
			continue // skip this user
		}

		for _, page := range pages {
			if page.authorLogin == umeta.GitHubId || page.authorLogin == u.Name {
				if !page.closed && !page.isAnswered {
					l = append(l,
						// all member as reviewers and 0 approved members
						reviewhub.NewReviewPage(page.title, page.url, u, []reviewhub.User{}, knownUsers),
					)
				}
			}
		}
	}

	return &reviewhub.ReviewList{
		Name:  config.Name,
		Pages: l,
	}, nil
}
