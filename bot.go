package main

import (
	"crypto/rand"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
	ai "github.com/mlabouardy/dialogflow-go-client"
	"github.com/mlabouardy/dialogflow-go-client/models"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var GuildID = "256295245816397824"
var AdminID = "93921947854835712"
var apiKey string
var Redis *redis.Client
var Dialog *ai.DialogFlowClient
var DevDialog *ai.DialogFlowClient

func main() {
	apiKey = os.Getenv("API_KEY")
	var botToken = os.Getenv("BOT_TOKEN")
	var devApiKey = os.Getenv("DEV_API_KEY")

	bot, err := discordgo.New("Bot " + botToken)
	if err != nil {
		log.Fatal(err)
	}

	bot.State.User, err = bot.User("@me")
	if err != nil {
		log.Fatal(err)
	}

	err = bot.Open()
	if err != nil {
		log.Fatal(err)
	}

	bot.AddHandler(onMessageCreate)

	Redis = redis.NewClient(&redis.Options{
		Addr: "redis:6379",
		DB:   0,
	})

	err, Dialog = ai.NewDialogFlowClient(models.Options{
		AccessToken: apiKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	err, DevDialog = ai.NewDialogFlowClient(models.Options{
		AccessToken: devApiKey,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf(`Now running. Press CTRL-C to exit.`)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Clean up
	bot.Close()
}

func onMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		return
	}

	if c.Type == discordgo.ChannelTypeDM && m.Author.ID == "93921947854835712" {
		if m.Content == "refresh" {
			err = RefreshAliases()
			if err == nil {
				s.ChannelMessageSend(m.ChannelID, "Refresh complete")
			} else {
				log.Println(err.Error())
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
		}
	}

	if c.Name == "attendance" {
		log.Printf("Query for '%v'", m.Content)

		uuid, err := newUUID()
		if err != nil {
			return
		}

		query := models.Query{
			Query:     m.Content,
			SessionID: uuid,
		}

		resp, err := Dialog.QueryFindRequest(query)
		if err != nil {
			log.Println("err: " + err.Error())
		}

		user := GetUserByAlias(m.Author.Username)
		results, err := DispatchActions(user, resp)

		if err != nil {
			log.Println(err.Error())
		} else {
			for _, r := range results {
				s.ChannelMessageSend(m.ChannelID, r.String())
			}
		}
	}
}

func newUUID() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8]&^0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6]&^0xf0 | 0x40
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}
