package notion

import (
	"context"
	"fmt"

	"github.com/dstotijn/go-notion"
	"github.com/ry023/reviewhub/reviewhub"
)

const MetaDataNotionId = "notion_id"

type NotionRetriever struct {
	Name              string
	ApiToken          string
	DatabaseId        string
	ReviewersProperty string
	Filter            *notion.DatabaseQueryFilter
	StaticReviewers   []reviewhub.User
}

func (p *NotionRetriever) Retrieve(allUsers []reviewhub.User) (*reviewhub.ReviewList, error) {
	// Get all notion pages
	cli := notion.NewClient(p.ApiToken)
	ctx := context.Background()

	query := &notion.DatabaseQuery{
		Filter: p.Filter,
	}

	res, err := cli.QueryDatabase(ctx, p.DatabaseId, query)
	if err != nil {
		return nil, err
	}

	// Convert to ReviewPage format
	var reviewPages []reviewhub.ReviewPage
	for _, page := range res.Results {
		props, ok := page.Properties.(notion.DatabasePageProperties)
		if !ok {
			return nil, fmt.Errorf("Invalid page")
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

		reviewPage := reviewhub.NewReviewPage(title, page.URL, *owner, approved, subUser(p.StaticReviewers, *owner))

		if len(reviewPage.ApprovedReviewers) < len(reviewPage.Reviewers)-1 {
			reviewPages = append(reviewPages, reviewPage)
		}
	}

	return &reviewhub.ReviewList{
		Name:  p.Name,
		Pages: reviewPages,
	}, nil
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
		fmt.Println(id)
		if got.ID == id {
			return &u, nil
		}
	}
	return reviewhub.NewUnknownUser(got.Name), nil
}

func subUser(us []reviewhub.User, target reviewhub.User) []reviewhub.User {
	r := []reviewhub.User{}
	for _, u := range us {
		if u.Name == target.Name {
			continue
		}
		r = append(r, u)
	}
	return r
}
