package ghdiscussions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
  "log"

	"github.com/buger/jsonparser"
)

type page struct {
	isAnswered  bool
	closed      bool
	title       string
	url         string
	authorLogin string
}

func request(owner, repository, token, apiendpoint string) ([]page, error) {
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
		owner, repository,
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
	req, err := http.NewRequest(http.MethodPost, "https://api.github.com/graphql", bytes.NewReader(bodybytes))
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

	_, err = jsonparser.ArrayEach(
		respBody,
		// handler
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			if err != nil {
				log.Printf("Failed to parse nodes: %v", err)
				return
			}
		},
		// array path
		"data", "repository", "discussions", "nodes",
	)
}
