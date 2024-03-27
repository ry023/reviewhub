package runners

import (
	"fmt"
	"log"

	"github.com/ry023/reviewhub/notifiers/slack"
	"github.com/ry023/reviewhub/notifiers/stdout"
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

func New(config *reviewhub.Config) (*ReviewHubRunner, error) {
	var notifiers []reviewhub.Notifier
	for _, c := range config.Notifiers {
		n, err := parseBuiltinNotifier(&c)
		if err != nil {
			return nil, err
		}
		notifiers = append(notifiers, n)
	}

	var retrievers []reviewhub.Retriever
	for _, c := range config.Retrievers {
		r, err := parseBuiltinRetriever(&c)
		if err != nil {
			return nil, err
		}
		retrievers = append(retrievers, r)
	}

	return &ReviewHubRunner{
		config:     *config,
		notifiers:  notifiers,
		users:      config.Users,
		retrievers: retrievers,
	}, nil
}

func (r *ReviewHubRunner) Run() error {
	var ls []reviewhub.ReviewList
	for _, retriever := range r.retrievers {
		l, err := retriever.Retrieve(r.users)
		if err != nil {
			log.Printf("Failed to retrieve: %s", err)
			break
		}
		ls = append(ls, *l)
	}

	for _, notifier := range r.notifiers {
		for _, u := range r.users {
			filtered := reviewhub.FilterReviewList(ls, u, false)
			if err := notifier.Notify(u, filtered); err != nil {
				log.Printf("Failed to notify to user: %s", err)
				break
			}
		}
	}

	return nil
}

func parseBuiltinNotifier(config *reviewhub.NotifierConfig) (reviewhub.Notifier, error) {
	if config.Type == "" {
		return nil, fmt.Errorf("'type' field empty")
	}

	switch config.Type {
	case "slack":
		return slack.New(config)
	case "stdout":
		return stdout.New(config)
	}

	return nil, ErrNotBuiltIn
}

func parseBuiltinRetriever(config *reviewhub.RetrieverConfig) (reviewhub.Retriever, error) {
	if config.Type == "" {
		return nil, fmt.Errorf("'type' field empty")
	}

	switch config.Type {
	case "notion":
		return notion.New(config)
	}

	return nil, ErrNotBuiltIn
}
