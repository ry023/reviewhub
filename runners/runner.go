package runner

import (
	"fmt"

	"github.com/ry023/reviewhub/notifiers/slack"
	"github.com/ry023/reviewhub/retrievers/notion"
	"github.com/ry023/reviewhub/reviewhub"
)

type ReviewHubRunner struct {
	config reviewhub.Config

	users      []reviewhub.User
	notifiers  []reviewhub.Notifier
	retrievers []reviewhub.Retriever
}

var ErrNotBuiltIn error = fmt.Errorf("This type is not built-in")

func (r *ReviewHubRunner) parseBuiltinNotifier(config *reviewhub.NotifierConfig) (reviewhub.Notifier, error) {
	if config.Type == "" {
		return nil, fmt.Errorf("'type' field empty")
	}

	switch config.Type {
	case "slack":
		return slack.New(config)
	}

	return nil, ErrNotBuiltIn
}

func (r *ReviewHubRunner) parseBuiltinRetriever(config *reviewhub.RetrieverConfig) (reviewhub.Retriever, error) {
	if config.Type == "" {
		return nil, fmt.Errorf("'type' field empty")
	}

	switch config.Type {
	case "notion":
		return notion.New(config)
	}

	return nil, ErrNotBuiltIn
}
