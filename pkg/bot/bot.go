package bot

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session *discordgo.Session
}

func NewBot(session *discordgo.Session) (*Bot, error) {
	return &Bot{
		session: session,
	}, nil
}

func (b *Bot) Run() error {
	err := b.session.Open()
	if err != nil {
		log.Fatal(err)
	}

	defer b.session.Close()
	b.session.AddHandler(b.handleMessage)

	log.Printf(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	return nil
}

func (b *Bot) handleMessage(s *discordgo.Session, m *discordgo.MessageCreate) {
	cmd, err := Parse(m.Content)
	if err != nil {
		log.Printf("parse command: %s", err.Error())
	}

	err = cmd.Execute()
	if err != nil {
		log.Printf("execute command: %s", err.Error())
	}

	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return
	}

	if c.Type == discordgo.ChannelTypeDM && m.Author.ID == "93921947854835712" {
		log.Printf("Responding to DM from user %s", m.Author.ID)
		//s.ChannelMessageSend(m.ChannelID, "Refresh complete")
	}
}
