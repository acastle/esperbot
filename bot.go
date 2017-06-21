package main

import (
  "github.com/bwmarrin/discordgo"
  "log"
  "os"
  "os/signal"
  "syscall"
  "fmt"
  "github.com/acastle/apiai-go"
  "github.com/go-redis/redis"
  "strings"
)

var GuildID = "256295245816397824"
var AI *apiaigo.APIAI
var Redis *redis.Client

func main() {
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

  AI = &apiaigo.APIAI{
    AuthToken: apiKey,
    Language:  "en-US",
    SessionID: "64f16405-5b58-4209-9fd1-c3e327267861",
    Version:   "20150910",
  }

  Redis = redis.NewClient(&redis.Options{
    Addr:     "redis:6379",
    DB:       0,
  })

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
    log.Printf("Query for '%v'", m.Content)
    resp, err := AI.SendText(m.Content)
    if err != nil {
      log.Println("err: " + err.Error())
    }

    if resp.Result.Action == "attendance.missraid" {
      results, err := MissRaid(m.Author, resp)
      if err != nil {
        log.Println(err.Error())
        s.ChannelMessageSend(m.ChannelID, err.Error())
      } else {
        for _,r := range results {
          s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Marked %v out on %v", r.Name, r.Dates))
        }
      }
    }

    if resp.Result.Action == "attendance.query" {
      results, err := Query(resp)
      if err != nil {
        log.Println(err.Error())
        s.ChannelMessageSend(m.ChannelID, err.Error())
      } else {

        for _,r := range results {
          year,month,day := r.Date.Date()
          members := strings.Join(r.Members, "\n  ")
          msg := fmt.Sprintf("**Raiders out for %d/%d/%d**\n  %s", month, day, year, members)
          s.ChannelMessageSend(m.ChannelID, msg)
        }
      }

    }

    if resp.Result.Action == "input.runsim" {

    }
  }
}