package notion

import (
	"fmt"
	"os"

	"github.com/ry023/reviewhub/reviewhub"
)

type NotionRetriever struct {
	Name     string
	ApiToken string
	MetaData MetaData
}

type MetaData struct {
	ApiTokenEnv           string   `yaml:"api_token_env" validate:"required"`
	DatabaseId            string   `yaml:"database_id" validate:"required"`
	OwnerProperty         string   `yaml:"owner_property" validate:"required"`
	ApprovedUsersProperty string   `yaml:"approved_users_property" validate:"required"`
	TitleProperty         string   `yaml:"title_property" validate:"required"`
	StaticReviewers       []string `yaml:"static_reviewers"`
	Filter                string   `yaml:"filter"`
}

type UserMetaData struct {
	NotionId string `yaml:"notion_id"`
}

func New(config *reviewhub.RetrieverConfig) (*NotionRetriever, error) {
	meta, err := reviewhub.ParseMetaData[MetaData](config.MetaData)
	if err != nil {
		return nil, err
	}

	token := os.Getenv(meta.ApiTokenEnv)

	return &NotionRetriever{
		Name:     config.Name,
		ApiToken: token,

		MetaData: *meta,
	}, nil
}

func (p *NotionRetriever) Retrieve(knownUsers []reviewhub.User) (*reviewhub.ReviewList, error) {
	pages, err := queryDatabase(p.MetaData.DatabaseId, p.MetaData.Filter, p.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("Failed to query database: %w", err)
	}

	// Convert to ReviewPage format
	var reviewPages []reviewhub.ReviewPage
	for _, page := range pages {
		title, err := page.title(p.MetaData.TitleProperty)
		if err != nil {
			return nil, err
		}

		owners, err := page.peopleProp(p.MetaData.OwnerProperty, knownUsers)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse owner_property (%s): %w", p.MetaData.OwnerProperty, err)
		} else if len(owners) != 1 {
			return nil, fmt.Errorf("owner must be one (owners=%v)", owners)
		}
		owner := owners[0]

		approvedUsers, err := page.peopleProp(p.MetaData.ApprovedUsersProperty, knownUsers)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse approved_users_property (%s): %w", p.MetaData.ApprovedUsersProperty, err)
		}

		url, err := page.url()
		if err != nil {
			return nil, err
		}

		reviewPage := reviewhub.NewReviewPage(title, url, owner, approvedUsers, p.staticReviewers(knownUsers, owner))

		if len(reviewPage.ApprovedReviewers) < len(reviewPage.Reviewers) {
			reviewPages = append(reviewPages, reviewPage)
		}
	}

	return &reviewhub.ReviewList{
		Name:  p.Name,
		Pages: reviewPages,
	}, nil
}

func (r *NotionRetriever) staticReviewers(knownUsers []reviewhub.User, owner reviewhub.User) []reviewhub.User {
	var reviewers []reviewhub.User
	for _, u := range knownUsers {
		for _, n := range r.MetaData.StaticReviewers {
			if u.Name == n && u.Name != owner.Name {
				reviewers = append(reviewers, u)
			}
		}
	}
	return reviewers
}
