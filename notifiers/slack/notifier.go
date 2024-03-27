package slack

import (
	"fmt"
	"log"
	"os"

	"github.com/ry023/reviewhub/reviewhub"
	"github.com/slack-go/slack"
)

type SlackNotifier struct {
	// Bot Token or User Token
	ApiToken string
	// Destination
	Channel string
}

type MetaData struct {
	ApiTokenEnv string `yaml:"api_token_env" validate:"required"`
	Channel     string `yaml:"channel" validate:"required"`
}

type UserMetaData struct {
	SlackId string `yaml:"slack_id" validate:"required"`
}

func (n *SlackNotifier) Notify(config reviewhub.NotifierConfig, user reviewhub.User, ls []reviewhub.ReviewList) error {
  meta, err := reviewhub.ParseMetaData[MetaData](config.MetaData)
	if err != nil {
		return err
	}

	cli := slack.New(os.Getenv(meta.ApiTokenEnv))

	usermeta, err := reviewhub.ParseMetaData[UserMetaData](config.MetaData)
	if err != nil {
		// user metadata not satisfied
		return nil
	}
	slackId := usermeta.SlackId

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

	if _, err := cli.PostEphemeral(n.Channel, slackId, slack.MsgOptionBlocks(b...)); err != nil {
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
