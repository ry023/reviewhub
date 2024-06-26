package stdout

import (
	"encoding/json"
	"fmt"

	"github.com/ry023/reviewhub/reviewhub"
)

const (
	FormatJson      = "json"
	FormatPlainText = "plaintext"
)

type StdoutNotifier struct {
}

type MetaData struct {
	Format string `yaml:"format"`
}

func (m *MetaData) Validate() error {
	if m.Format == "" {
		m.Format = FormatPlainText
	}

	valid := []string{
		FormatJson,
		FormatPlainText,
	}
	for _, v := range valid {
		if v == m.Format {
			return nil
		}
	}
	return fmt.Errorf("Invalid format type: %s", m.Format)
}

func (n *StdoutNotifier) Notify(config reviewhub.NotifierConfig, user reviewhub.User, ls []reviewhub.ReviewList) error {
	meta, err := reviewhub.ParseMetaData[MetaData](config.MetaData)
	if err != nil {
		return err
	}

	if err := meta.Validate(); err != nil {
		return err
	}

	switch meta.Format {
	case FormatPlainText:
		fmt.Printf("User: %s\n", user.Name)
		for _, l := range ls {
			fmt.Printf("ReviewName: %s\n", l.Name)
			for _, p := range l.Pages {
				fmt.Printf("- %s\n", p.Title)
			}
		}
		fmt.Println("")

	case FormatJson:
		notif := notification{
			User:        user,
			ReviewLists: ls,
		}
		b, err := json.Marshal(notif)
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	default:
		return fmt.Errorf("Invalid format type: %s", meta.Format)
	}

	return nil
}

type notification struct {
	User        reviewhub.User
	ReviewLists []reviewhub.ReviewList
}
