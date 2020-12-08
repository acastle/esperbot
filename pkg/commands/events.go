package commands

import (
	"fmt"
	"time"

	"github.com/acastle/esperbot/pkg/events"
	"github.com/acastle/esperbot/pkg/util"
	"github.com/bwmarrin/discordgo"
	log "github.com/sirupsen/logrus"
)

type EventsCommand struct {
}

func (c EventsCommand) Execute(ctx Context) error {
	now := time.Now().UTC()
	begin := util.BeginningOfWeek(now)
	end := util.EndOfWeek(now)

	evts, err := events.GetEventsForWeek(ctx.Redis, now)
	if err != nil {
		log.Error(err)
	}

	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name: "Upcoming events",
		},
		Title:       "",
		Description: fmt.Sprintf("for the week of %s to %s", begin.Format("Monday Jan _2 2006"), end.Format("Monday Jan _2 2006")),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://wow.zamimg.com/images/Icon/large/hilite/default.png",
		},
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, evt := range evts {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  evt.Name,
			Value: evt.Time.Format("Monday Jan _2 2006"),
		})
	}

	ctx.Session.ChannelMessageSendEmbed(ctx.ChannelID, &embed)
	return nil
}
