package stdout

import (
	"encoding/json"
	"fmt"

	"github.com/ry023/reviewhub/reviewhub"
)

const FormatJson = "json"

type StdoutNotifier struct {
	Format string
}

func New(config *reviewhub.NotifierConfig) (*StdoutNotifier, error) {
	return &StdoutNotifier{
		Format: FormatJson,
	}, nil
}

func (n *StdoutNotifier) Notify(user reviewhub.User, ls []reviewhub.ReviewList) error {
	var out string

	switch n.Format {
	case FormatJson:
		notif := notification{
			User:        user,
			ReviewLists: ls,
		}

		b, err := json.Marshal(notif)
		if err != nil {
			return err
		}

		out = string(b)
	default:
		return fmt.Errorf("Invalid format type: %s", n.Format)
	}

	fmt.Println(out)
	return nil
}

type notification struct {
	User        reviewhub.User
	ReviewLists []reviewhub.ReviewList
}
