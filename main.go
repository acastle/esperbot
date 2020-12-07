package main

import (
	"os"
	"time"

	"github.com/go-co-op/gocron"
	log "github.com/sirupsen/logrus"

	"github.com/acastle/esperbot/pkg/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
)

var GuildID = "256295245816397824"
var AdminID = "93921947854835712"
var Redis *redis.Client

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	var botToken = os.Getenv("BOT_TOKEN")
	session, err := discordgo.New("Bot " + botToken)

	rd := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	instance, err := bot.NewBot(session, rd, gocron.NewScheduler(time.UTC))
	if err != nil {
		log.Fatal(err)
	}

	err = instance.Run()
	if err != nil {
		log.Fatal(err)
	}
}
