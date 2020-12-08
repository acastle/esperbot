package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type HelpCommand struct {
}

func (h HelpCommand) Execute(ctx Context) error {
	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Esperbot help",
		},
		Title:       "",
		Description: "",
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://wow.zamimg.com/images/wow/icons/large/inv_misc_questionmark.jpg",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "!help",
				Value: "provide a list of available bot commands",
			},
			{
				Name:  "!events",
				Value: "list planned events for the upcomming week",
			},
			{
				Name:  "!setname [name]",
				Value: "sets what name the bot will use for your discord user (ex. !setname RingRingRingBananaPhone)",
			},
			{
				Name:  "!out [<date> to <date>]",
				Value: "mark yourself absent for all events over a period of time. (ex. !out Dec 10 to Dec 30)",
			},
			{
				Name:  "!in [<date> to <date>]",
				Value: "mark yourself in for all events over a period of time. (ex. !out Dec 10 to Dec 30)",
			},
			{
				Name:  "!late [<date> to <date>]",
				Value: "mark yourself late for all events over a period of time. (ex. !out Dec 10 to Dec 30)",
			},
			{
				Name:  "!ontime [<date> to <date>]",
				Value: "mark yourself on time for all events over a period of time. (ex. !out Dec 10 to Dec 30)",
			},
		},
	}

	_, err := ctx.Session.ChannelMessageSendEmbed(ctx.ChannelID, &embed)
	if err != nil {
		return fmt.Errorf("send help message: %w", err)
	}
	return nil
}
