package runner

import (
	"fmt"

	"github.com/ry023/reviewhub/notifiers/slack"
	"github.com/ry023/reviewhub/reviewhub"
)

type ReviewHubRunner struct {
	config reviewhub.Config
	users  []reviewhub.User
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
