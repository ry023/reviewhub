package notion

import (
	"fmt"
	"os"

	"github.com/ry023/reviewhub/reviewhub"
)

type NotionRetriever struct {
	conf  reviewhub.RetrieverConfig
	token string
	meta  MetaData
}

type MetaData struct {
	ApiTokenEnv           string   `yaml:"api_token_env" validate:"required"`
	DatabaseId            string   `yaml:"database_id" validate:"required"`
	OwnerProperty         string   `yaml:"owner_property" validate:"required"`
	ApprovedUsersProperty string   `yaml:"approved_users_property" validate:"required"`
	ReviewersProperty     string   `yaml:"reviewers_property"`
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

	if meta.ReviewersProperty == "" && (meta.StaticReviewers == nil || len(meta.StaticReviewers) == 0) {
		return nil, fmt.Errorf("Either static_reviewers or reviewers_property required")
	}

	token := os.Getenv(meta.ApiTokenEnv)

	return &NotionRetriever{
		conf:  *config,
		token: token,
		meta:  *meta,
	}, nil
}

func (p *NotionRetriever) Retrieve(knownUsers []reviewhub.User) (*reviewhub.ReviewList, error) {
	pages, err := queryDatabase(p.meta.DatabaseId, p.meta.Filter, p.token)
	if err != nil {
		return nil, fmt.Errorf("Failed to query database: %w", err)
	}

	// Convert to ReviewPage format
	var reviewPages []reviewhub.ReviewPage
	for _, page := range pages {
		title, err := page.title(p.meta.TitleProperty)
		if err != nil {
			return nil, err
		}

		owners, err := page.peopleProp(p.meta.OwnerProperty, knownUsers)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse owner_property (%s): %w", p.meta.OwnerProperty, err)
		} else if len(owners) != 1 {
			return nil, fmt.Errorf("owner must be single (owners=%v)", owners)
		}
		owner := owners[0]

		approvedUsers, err := page.peopleProp(p.meta.ApprovedUsersProperty, knownUsers)
		if err != nil {
			return nil, fmt.Errorf("Failed to parse approved_users_property (%s): %w", p.meta.ApprovedUsersProperty, err)
		}

		var reviewers []reviewhub.User
		if p.meta.StaticReviewers != nil && len(p.meta.StaticReviewers) > 0 {
			reviewers = p.staticReviewers(knownUsers, owner)
		} else {
			us, err := page.peopleProp(p.meta.ReviewersProperty, knownUsers)
			if err != nil {
				return nil, fmt.Errorf("Failed to parse approved_users_property (%s): %w", p.meta.ApprovedUsersProperty, err)
			}
			reviewers = us
		}

		url, err := page.url()
		if err != nil {
			return nil, err
		}

		reviewPage := reviewhub.NewReviewPage(title, url, owner, approvedUsers, reviewers)

		if len(reviewPage.ApprovedReviewers) < len(reviewPage.Reviewers) {
			reviewPages = append(reviewPages, reviewPage)
		}
	}

	return &reviewhub.ReviewList{
		Name:  p.conf.Name,
		Pages: reviewPages,
	}, nil
}

func (r *NotionRetriever) staticReviewers(knownUsers []reviewhub.User, owner reviewhub.User) []reviewhub.User {
	var reviewers []reviewhub.User
	for _, u := range knownUsers {
		for _, n := range r.meta.StaticReviewers {
			if u.Name == n && u.Name != owner.Name {
				reviewers = append(reviewers, u)
			}
		}
	}
	return reviewers
}
