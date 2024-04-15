package runners

import (
	"fmt"
	"log"

	"github.com/ry023/reviewhub/notifiers/slack"
	"github.com/ry023/reviewhub/notifiers/stdout"
	"github.com/ry023/reviewhub/retrievers/ghdiscussions"
	"github.com/ry023/reviewhub/retrievers/notion"
	"github.com/ry023/reviewhub/reviewhub"
)

type ReviewHubRunner struct {
	config reviewhub.Config

	users      []reviewhub.User
	notifiers  []notifier
	retrievers []retriever
}

type notifier struct {
	config   reviewhub.NotifierConfig
	notifier reviewhub.Notifier
}

type retriever struct {
	config    reviewhub.RetrieverConfig
	retriever reviewhub.Retriever
}

var ErrNotBuiltIn error = fmt.Errorf("This type is not built-in")

func New(config *reviewhub.Config) (*ReviewHubRunner, error) {
	var notifiers []notifier
	for _, c := range config.Notifiers {
		n, err := parseBuiltinNotifier(&c)
		if err != nil {
			return nil, err
		}
		notifiers = append(notifiers, notifier{
			notifier: n,
			config:   c,
		})
	}

	var retrievers []retriever
	for _, c := range config.Retrievers {
		r, err := parseBuiltinRetriever(&c)
		if err != nil {
			return nil, err
		}
		retrievers = append(retrievers, retriever{
			retriever: r,
			config:    c,
		})
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
	for _, v := range r.retrievers {
		l, err := v.retriever.Retrieve(v.config, r.users)
		if err != nil {
			return err
		}
		ls = append(ls, *l)
	}

	for _, v := range r.notifiers {
		for _, u := range r.users {
			filtered := reviewhub.FilterReviewList(ls, u, false)
			if err := v.notifier.Notify(v.config, u, filtered); err != nil {
				log.Printf("Failed to notify to user by %T: %s", v.notifier, err)
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
		return new(slack.SlackNotifier), nil
	case "stdout":
		return new(stdout.StdoutNotifier), nil
	}

	return nil, ErrNotBuiltIn
}

func parseBuiltinRetriever(config *reviewhub.RetrieverConfig) (reviewhub.Retriever, error) {
	if config.Type == "" {
		return nil, fmt.Errorf("'type' field empty")
	}

	switch config.Type {
	case "notion":
		return new(notion.NotionRetriever), nil
	case "github-discussions":
		return new(ghdiscussions.GitHubDiscussionsRetriever), nil
	}

	return nil, ErrNotBuiltIn
}
