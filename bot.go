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
  "io"
  "crypto/rand"
)

var GuildID = "256295245816397824"
var apiKey string;
var Redis *redis.Client

func main() {
  apiKey = os.Getenv("API_KEY")
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

  uuid, err := newUUID()
  if err != nil {
    return
  }

  ai := &apiaigo.APIAI{
    AuthToken: apiKey,
    Language:  "en-US",
    SessionID: uuid,
    Version:   "20150910",
  }

  if c.Name == "bottraining" {
    log.Printf("Query for '%v'", m.Content)
    resp, err := ai.SendText(m.Content)
    if err != nil {
      log.Println("err: " + err.Error())
    }

    if resp.Result.Action == "attendance.missraid" {
      results, err := MissRaid(m.Author, resp)
      if err != nil {
        log.Println(err.Error())
        //s.ChannelMessageSend(m.ChannelID, err.Error())
      } else {
        for _,r := range results {
          s.ChannelMessageSend(m.ChannelID, r.String())
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
