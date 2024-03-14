package notion

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const apiEndpoint = "https://api.notion.com/v1/"

type queryParam struct {
	Filter      any    `json:"filter,omitempty"`
	StartCursor string `json:"start_cursor,omitempty"`
}

type Page any

type response struct {
	Results    []Page  `json:"results"`
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
}

func queryDatabase(databaseId, filterJSON, token string) ([]Page, error) {
	var pages []Page

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

		pages = append(pages, res.Results...)

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
	if err := json.Unmarshal(resBody, res); err != nil {
		return nil, err
	}
	return res, nil
}
