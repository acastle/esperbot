package bot

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"

	"github.com/acastle/esperbot/pkg/commands"
	"github.com/acastle/esperbot/pkg/events"
	"github.com/acastle/esperbot/pkg/parser"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
)

type Bot struct {
	session   *discordgo.Session
	redis     *redis.Client
	scheduler *gocron.Scheduler
}

func NewBot(session *discordgo.Session, redis *redis.Client, scheduler *gocron.Scheduler) (*Bot, error) {
	return &Bot{
		session:   session,
		redis:     redis,
		scheduler: scheduler,
	}, nil
}

func (b *Bot) Run() error {
	defer b.session.Close()
	b.session.AddHandler(b.handleMessage)
	b.session.AddHandler(b.handleReactionAdd)
	b.session.AddHandler(b.handleReactionRemove)
	b.session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages | discordgo.IntentsDirectMessageReactions | discordgo.IntentsGuildMessageReactions)

	b.scheduler.StartAsync()

	err := b.session.Open()
	if err != nil {
		log.Fatal(err)
	}

	b.session.State.User, err = b.session.User("@me")
	if err != nil {
		log.Fatal(err)
	}

	events.UpsertRecurringEvent(b.redis, events.RecurringEvent{
		ID:       "MainRaid",
		Name:     "Main Raid",
		Weekdays: []time.Weekday{time.Wednesday},
	})

	b.scheduler.Every(1).Day().Do(b.scheduleEvents)
	b.scheduleEvents()

	log.Printf(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return nil
}

func (b *Bot) scheduleEvents() {
	log.Info("scheduling events")
	err := events.ScheduleEventsForWeek(b.redis, time.Now())
	if err != nil {
		log.Error(err)
	}

	evts, err := events.GetEventsForWeek(b.redis, time.Now().UTC())
	if err != nil {
		log.Error(err)
	}

	for _, evt := range evts {
		evt.AnnounceChannelID = "256297257052274688"
		err := events.AnnounceEvent(b.session, b.redis, evt)
		if err != nil {
			log.Error(err)
		}
	}
}

func (b *Bot) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	cmd, err := parser.Parse(m.Content)
	if err != nil {
		log.WithField("input", m.Content).Error(err)
		return
	}

	ctx := commands.Context{
		Session:   s,
		Sender:    m.Author,
		ChannelID: m.ChannelID,
		Redis:     b.redis,
	}
	err = cmd.Execute(ctx)
	if err != nil {
		log.Error(err)
		return
	}
}

func (b *Bot) handleReactionAdd(s *discordgo.Session, m *discordgo.MessageReactionAdd) {
	if m.UserID == s.State.User.ID {
		return
	}

	var t events.UserListType
	switch m.Emoji.Name {
	case "❌":
		t = events.Absent
	case "🕘":
		t = events.Late
	default:
		return
	}

	evt, err := events.GetEventByMessage(b.redis, m.ChannelID, m.MessageID)
	if errors.Is(err, events.ErrEventNotFound) {
		return
	} else if err != nil {
		log.Error(err)
		return
	}

	err = events.EventUserListAdd(b.redis, evt, m.UserID, t)
	if err != nil {
		log.Error(err)
		return
	}

	err = events.AnnounceEvent(s, b.redis, evt)
	if err != nil {
		log.Error(err)
		return
	}
}

func (b *Bot) handleReactionRemove(s *discordgo.Session, m *discordgo.MessageReactionRemove) {
	if m.UserID == s.State.User.ID {
		return
	}

	var t events.UserListType
	switch m.Emoji.Name {
	case "❌":
		t = events.Absent
	case "🕘":
		t = events.Late
	default:
		return
	}

	evt, err := events.GetEventByMessage(b.redis, m.ChannelID, m.MessageID)
	if errors.Is(err, events.ErrEventNotFound) {
		return
	} else if err != nil {
		log.Error(err)
		return
	}

	err = events.EventUserListRemove(b.redis, evt, m.UserID, t)
	if err != nil {
		log.Error(err)
		return
	}

	err = events.AnnounceEvent(s, b.redis, evt)
	if err != nil {
		log.Error(err)
		return
	}
}
