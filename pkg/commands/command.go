package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"
)

type Command interface {
	Execute(Context) error
}

type Context struct {
	Session   *discordgo.Session
	ChannelID string
	Sender    *discordgo.User
	Redis     *redis.Client
}
