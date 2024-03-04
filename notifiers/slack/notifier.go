package slack

import (
	"fmt"
	"log"
	"os"

	"github.com/ry023/reviewhub/reviewhub"
	"github.com/slack-go/slack"
)

const MetaDataSlackId = "slack_id"

type SlackNotifier struct {
	// Bot Token or User Token
	ApiToken string
	// Destination
	Channel string
}

func New(config *reviewhub.NotifierConfig) (*SlackNotifier, error) {
	validKeys := []string{
		"api_token_env",
		"channel",
	}

	for _, k := range validKeys {
		_, ok := config.MetaData[k]
		if !ok {
			return nil, fmt.Errorf("MetaData '%s' not found", k)
		}
	}

	token := os.Getenv(config.MetaData["api_token_env"])

	return &SlackNotifier{
		ApiToken: token,
		Channel:  config.MetaData["channel"],
	}, nil
}

func (n *SlackNotifier) Notify(user reviewhub.User, ls []reviewhub.ReviewList) error {
	cli := slack.New(n.ApiToken)

	slackId, ok := user.MetaData[MetaDataSlackId]
	if !ok {
		// This user not slack information
		return nil
	}

	b := []slack.Block{
		// Header Block
		slack.NewHeaderBlock(
			slack.NewTextBlockObject(
				slack.PlainTextType,
				fmt.Sprintf("Notification For You (%s)", user.Name),
				false,
				false,
			),
		),
	}

	for _, c := range ls {
		// Page List
		b = append(b, buildPageListBlock(c))
		// Divider
		b = append(b, slack.NewDividerBlock())
	}

	_, err := cli.PostEphemeral(n.Channel, slackId, slack.MsgOptionBlocks(b...))
	if err != nil {
		log.Printf("Failed to send to %s: %v", user.Name, err)
	}

	return nil
}

func buildPageListBlock(r reviewhub.ReviewList) slack.Block {
	// List Name
	name := slack.NewRichTextSection(
		slack.NewRichTextSectionTextElement(
			r.Name,
			&slack.RichTextSectionTextStyle{
				Bold: true,
			},
		),
	)

	// Early return if no review page!
	if len(r.Pages) == 0 {
		s := slack.NewRichTextSection(
			// emoji
			slack.NewRichTextSectionEmojiElement("tada", 2, nil),
			// message
			slack.NewRichTextSectionTextElement(
				"There are no pages you need to review! Thank you for your cooperation!",
				&slack.RichTextSectionTextStyle{Italic: true},
			),
		)
		return slack.NewRichTextBlock("", name, s)
	}

	// Page List
	var els []slack.RichTextElement
	for _, page := range r.Pages {
		// Building List Element...
		s := slack.NewRichTextSection(
			// Page URL
			slack.NewRichTextSectionLinkElement(
				page.Url,
				page.Title,
				&slack.RichTextSectionTextStyle{
					Bold: true,
				},
			),
			// Owner
			slack.NewRichTextSectionTextElement(
				fmt.Sprintf("(by %s)", page.Owner.Name),
				&slack.RichTextSectionTextStyle{},
			),
		)
		els = append(els, s)
	}
	list := slack.NewRichTextList(slack.RTEListBullet, 0, els...) // Sum up to RichTextList block

	return slack.NewRichTextBlock("", name, list)
}