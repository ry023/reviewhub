package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/ry023/reviewhub/reviewhub"
)

const apiEndpoint = "https://api.notion.com/v1"

type queryParam struct {
	Filter      any    `json:"filter,omitempty"`
	StartCursor string `json:"start_cursor,omitempty"`
}

type response struct {
	Results    []any   `json:"results"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
}

type jsonPage []byte

func (p jsonPage) title(prop string) (string, error) {
	// TODO: parse richtext strictly
	return jsonparser.GetString(p, "properties", prop, "title", "[0]", "text", "content")
}

func (p jsonPage) url() (string, error) {
	// TODO: parse richtext strictly
	return jsonparser.GetString(p, "url")
}

func (p jsonPage) owner(prop string, knownUsers []reviewhub.User) (*reviewhub.User, error) {
	propid, err := jsonparser.GetString(p, "properties", prop, "people", "[0]", "id")
	if err != nil {
		return nil, err
	}

	// search in known user
	for _, u := range knownUsers {
		meta, err := reviewhub.ParseMetaData[UserMetaData](u.MetaData)
		if err != nil {
			log.Printf("Skip user %s because it may not have notion metadata: %v", u.Name, err)
			continue
		}

		if meta.NotionId == propid {
			return &u, nil
		}
	}

	// return as unknown user
	name, err := jsonparser.GetString(p, "properties", prop, "people", "[0]", "name")
	if err != nil {
		return nil, err
	}
	return reviewhub.NewUnknownUser(name), nil
}

func (p jsonPage) peopleProp(prop string, knownUsers []reviewhub.User) ([]reviewhub.User, error) {
	var peopleIds []string
	// parse properties
	_, err := jsonparser.ArrayEach(
		// json raw bytes
		p,

		// callback
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				log.Printf("Failed to parse reviewer array: %v", err)
				return
			}

			id, err := jsonparser.GetString(value, "id")
			if err != nil {
				log.Printf("Failed to parse reviewer's id: %v", err)
				return
			}
			peopleIds = append(peopleIds, id)
		},

		// array path
		"properties", prop, "people",
	)
	if err != nil {
		return nil, err
	}

	// search in known user
	var people []reviewhub.User
	for _, propid := range peopleIds {
		for _, u := range knownUsers {
			meta, err := reviewhub.ParseMetaData[UserMetaData](u.MetaData)
			if err != nil {
				log.Printf("Skip user %s because it may not have notion metadata: %v", u.Name, err)
				continue
			}

			if meta.NotionId == propid {
				people = append(people, u)
			}
		}
	}

	return people, nil
}

func queryDatabase(databaseId, filterJSON, token string) ([]jsonPage, error) {
	var pages []jsonPage

	var filter any
	if err := json.Unmarshal([]byte(filterJSON), &filter); err != nil {
		return nil, fmt.Errorf("json of filter invalid: %w", err)
	}

	var cur string
	more := true
	for more {
		p := queryParam{
			Filter:      filter,
			StartCursor: cur,
		}
		res, err := request(databaseId, token, p)
		if err != nil {
			return nil, err
		}

		for _, r := range res.Results {
			b, err := json.Marshal(r)
			if err != nil {
				return nil, err
			}
			pages = append(pages, b)
		}

		more = res.HasMore
		if res.NextCursor != nil {
			cur = *res.NextCursor
		}
	}

	return pages, nil
}

func request(databaseId, token string, param queryParam) (*response, error) {
	// build request body bytes
	p, err := json.Marshal(param)
	if err != nil {
		return nil, err
	}
	reqBody := bytes.NewBuffer(p)

	// build request
	url := fmt.Sprintf("%s/databases/%s/query", apiEndpoint, databaseId)
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("Failed to query to api: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	// request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read body
	resBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// parse response body
	var res *response
	if err := json.Unmarshal(resBody, &res); err != nil {
		return nil, err
	}
	return res, nil
}
