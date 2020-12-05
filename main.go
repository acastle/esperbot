package main

import (
	"log"
	"os"

	"github.com/acastle/esperbot/pkg/bot"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
)

var GuildID = "256295245816397824"
var AdminID = "93921947854835712"
var Redis *redis.Client

func main() {
	var botToken = os.Getenv("BOT_TOKEN")
	session, err := discordgo.New("Bot " + botToken)

	instance, err := bot.NewBot(session)
	if err != nil {
		log.Fatal(err)
	}

	err = instance.Run()
	if err != nil {
		log.Fatal(err)
	}
}
