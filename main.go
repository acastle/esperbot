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

func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	var botToken = os.Getenv("BOT_TOKEN")
	var redisAddr = os.Getenv("REDIS_ADDR")
	session, err := discordgo.New("Bot " + botToken)

	rd := redis.NewClient(&redis.Options{
		Addr: redisAddr,
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
