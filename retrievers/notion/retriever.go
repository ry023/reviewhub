package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dstotijn/go-notion"
	"github.com/ry023/reviewhub/reviewhub"
)

const MetaDataNotionId = "notion_id"

type NotionRetriever struct {
	Name                string
	ApiToken            string
	DatabaseId          string
	ReviewersProperty   string
	Filter              *notion.DatabaseQueryFilter
	StaticReviewerNames []string
}

type MetaData struct {
	ApiTokenEnv       string   `yaml:"api_token_env" validate:"required"`
	DatabaseId        string   `yaml:"database_id" validate:"required"`
	ReviewersProperty string   `yaml:"reviewers_property" validate:"required"`
	StaticReviewers   []string `yaml:"static_reviewers"`
	Filter            string   `yaml:"filter"`
}

func New(config *reviewhub.RetrieverConfig) (*NotionRetriever, error) {
	meta, err := reviewhub.ParseMetaData[MetaData](config.MetaData)
	if err != nil {
		return nil, err
	}

	token := os.Getenv(meta.ApiTokenEnv)

	var filter *notion.DatabaseQueryFilter
	if err = json.Unmarshal([]byte(meta.Filter), &filter); err != nil {
		return nil, err
	}

	return &NotionRetriever{
		Name:                config.Name,
		ApiToken:            token,
		DatabaseId:          meta.DatabaseId,
		ReviewersProperty:   meta.ReviewersProperty,
		Filter:              filter,
		StaticReviewerNames: meta.StaticReviewers,
	}, nil
}

func (p *NotionRetriever) Retrieve(allUsers []reviewhub.User) (*reviewhub.ReviewList, error) {
	cli := notion.NewClient(p.ApiToken)

	// Get all notion pages from the database
	hasmore := true
	nextCursor := ""
	results := []notion.Page{}
	ctx := context.Background()
	for hasmore {
		query := &notion.DatabaseQuery{
			Filter:      p.Filter,
			StartCursor: nextCursor,
		}

		res, err := cli.QueryDatabase(ctx, p.DatabaseId, query)
		if err != nil {
			return nil, err
		}
		results = append(results, res.Results...)

		hasmore = res.HasMore
		nextCursor = *res.NextCursor
	}

	// Convert to ReviewPage format
	var reviewPages []reviewhub.ReviewPage
	for _, page := range results {
		props, ok := page.Properties.(notion.DatabasePageProperties)
		if !ok {
			continue
		}

		title, err := getTitle(props)
		if err != nil {
			return nil, err
		}

		approved, err := getApprovedReviewers(p.ReviewersProperty, props, allUsers)
		if err != nil {
			return nil, err
		}

		owner, err := getOwner(props, allUsers)
		if err != nil {
			return nil, err
		}

		reviewPage := reviewhub.NewReviewPage(title, page.URL, *owner, approved, p.staticReviewers(allUsers, *owner))

		if len(reviewPage.ApprovedReviewers) < len(reviewPage.Reviewers) {
			reviewPages = append(reviewPages, reviewPage)
		}
	}

	return &reviewhub.ReviewList{
		Name:  p.Name,
		Pages: reviewPages,
	}, nil
}

func (r *NotionRetriever) staticReviewers(allUsers []reviewhub.User, owner reviewhub.User) []reviewhub.User {
	var reviewers []reviewhub.User
	for _, u := range allUsers {
		for _, n := range r.StaticReviewerNames {
			if u.Name == n && u.Name != owner.Name {
				reviewers = append(reviewers, u)
			}
		}
	}
	return reviewers
}

func getTitle(props notion.DatabasePageProperties) (string, error) {
	for _, p := range props {
		if p.Type == notion.DBPropTypeTitle && len(p.Title) > 0 {
			return p.Title[0].PlainText, nil
		}
	}
	return "", fmt.Errorf("Title not found")
}

func getApprovedReviewers(propertyName string, props notion.DatabasePageProperties, allUsers []reviewhub.User) ([]reviewhub.User, error) {
	prop, ok := props[propertyName]
	if !ok {
		return nil, fmt.Errorf("Property '%s' not found", propertyName)
	}

	if prop.Type != notion.DBPropTypePeople {
		return nil, fmt.Errorf("Property '%s' found but not People type", propertyName)
	}

	// Find approved reviewers from property
	var approved []reviewhub.User
	for _, u := range allUsers {
		notionId, ok := u.MetaData[MetaDataNotionId]
		if !ok {
			continue
		}
		for _, notionUser := range prop.People {
			if notionUser.ID == notionId {
				approved = append(approved, u)
			}
		}
	}
	return approved, nil
}

func getOwner(props notion.DatabasePageProperties, allUsers []reviewhub.User) (*reviewhub.User, error) {
	var got *notion.User
	for name, prop := range props {
		// TODO: let it specify prop name
		if prop.Type == notion.DBPropTypePeople && name == "Owner" && len(prop.People) > 0 {
			got = &prop.People[0]
			break
		}
	}

	if got == nil {
		return nil, fmt.Errorf("CreatedBy property not found")
	}

	for _, u := range allUsers {
		id, ok := u.MetaData[MetaDataNotionId]
		if !ok {
			continue
		}
		if got.ID == id {
			return &u, nil
		}
	}
	return reviewhub.NewUnknownUser(got.Name), nil
}
