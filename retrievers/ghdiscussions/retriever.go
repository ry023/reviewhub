package ghdiscussions

import (
	"os"

	"github.com/ry023/reviewhub/reviewhub"
)

type GitHubDiscussionsRetriever struct {
}

type MetaData struct {
	ApiTokenEnv string `yaml:"api_token_env" validate:"required"`
  ApiEndpoint string `yaml:"api_endpoint"`

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

  return nil, err
}
