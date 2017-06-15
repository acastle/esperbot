package main

import (
  "github.com/bwmarrin/discordgo"
  "log"
  "os"
  "os/signal"
  "syscall"
  "github.com/kkdai/luis"
  "fmt"
)

var GuildID = "256295245816397824"
var cog *luis.Luis

func main() {
  var appId = os.Getenv("APP_ID")
  var apiKey = os.Getenv("API_KEY")
  var botToken = os.Getenv("BOT_TOKEN")

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
  cog = luis.NewLuis(apiKey, appId)

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

  if c.Name == "bottraining" {
    res, err := cog.Predict(m.Content)
    if err != nil {
      log.Println("err: " + err.Err.Error())
    }

    pr := luis.NewPredictResponse(res)
    intent := luis.GetBestScoreIntent(pr)

    if intent.Name != "None" {
      s.ChannelMessageSend(m.ChannelID, "Intent: " + intent.Name)
      for _,e := range (*pr)[0].EntitiesResults {
        fmt.Println(e)
      }
    }
  }
}