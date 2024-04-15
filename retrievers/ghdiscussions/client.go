package ghdiscussions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/buger/jsonparser"
)

type page struct {
	isAnswered  bool
	closed      bool
	title       string
	url         string
	authorLogin string
}

func request(repositoryOwner, repository, token, apiEndpoint string) ([]page, error) {
	// GraphQL Query
	query := fmt.Sprintf(`
query {
  repository(owner: "%s", name: "%s") {
    discussions(last: 100) {
      nodes {
        isAnswered
        closed
        title
        url
        author {
          login
        }
      }
    }
  }
}
`,
		repositoryOwner, repository,
	)

	// build json request body
	body := struct {
		Query string `json:"query"`
	}{Query: query}
	bodybytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	// build http request
	req, err := http.NewRequest(http.MethodPost, apiEndpoint, bytes.NewReader(bodybytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// do request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pages := []page{}
	_, err = jsonparser.ArrayEach(
		respBody,
		// handler
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				log.Printf("Failed to parse nodes: %v", err)
				return
			}

			isAnswered, err := jsonparser.GetBoolean(value, "isAnswered")
			if err != nil {
				return
			}
			closed, err := jsonparser.GetBoolean(value, "closed")
			if err != nil {
				return
			}
			title, err := jsonparser.GetString(value, "title")
			if err != nil {
				return
			}
			url, err := jsonparser.GetString(value, "url")
			if err != nil {
				return
			}
			authorLogin, err := jsonparser.GetString(value, "author", "login")
			if err != nil {
				return
			}
			pages = append(pages, page{
				isAnswered:  isAnswered,
				closed:      closed,
				title:       title,
				url:         url,
				authorLogin: authorLogin,
			})
		},
		// array path
		"data", "repository", "discussions", "nodes",
	)
	if err != nil {
		return nil, err
	}
	return pages, nil
}
