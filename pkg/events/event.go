package events

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/acastle/esperbot/pkg/util"
	"github.com/bwmarrin/discordgo"
	"github.com/go-redis/redis"

	log "github.com/sirupsen/logrus"
)

func EventKeyForID(id string) string {
	return fmt.Sprintf("event:%s", id)
}

func EventIndexKeyForDate(date time.Time) string {
	start := util.BeginningOfWeek(date.UTC())
	return fmt.Sprintf("index:events_by_week:%d", start.Unix())
}

func EventIndexKeyForMessageId(channelID string, messageID string) string {
	return fmt.Sprintf("index:event_by_message:%s:%s", channelID, messageID)
}

type EventStatus string

const (
	Unscheduled EventStatus = "Unscheduled"
	Scheduled               = "Scheduled"
	Canceled                = "Canceled"
)

type Event struct {
	ID     string
	Name   string
	Time   time.Time
	Status EventStatus

	RecurringEventID  string
	AnnounceMessageID string
	AnnounceChannelID string
}

func ScheduleEvent(redis *redis.Client, event Event) error {
	pipe := redis.Pipeline()
	key := EventKeyForID(event.ID)
	pipe.HSet(key, "id", event.ID)
	pipe.HSet(key, "name", event.Name)
	pipe.HSet(key, "time", event.Time.Unix())
	pipe.HSet(key, "status", Scheduled)
	pipe.HSet(key, "recurring_event_id", event.RecurringEventID)
	pipe.HSet(key, "announce_message_id", event.AnnounceMessageID)
	pipe.HSet(key, "announce_channel_id", event.AnnounceChannelID)

	if event.AnnounceMessageID != "" && event.AnnounceChannelID != "" {
		messageIndexKey := EventIndexKeyForMessageId(event.AnnounceChannelID, event.AnnounceMessageID)
		pipe.Set(messageIndexKey, event.ID, time.Duration(365*24*time.Hour))
	}

	indexKey := EventIndexKeyForDate(event.Time)
	pipe.SAdd(indexKey, event.ID)
	pipe.Expire(key, time.Duration(365*24*time.Hour))
	pipe.Expire(indexKey, time.Duration(365*24*time.Hour))
	_, err := pipe.Exec()
	if err != nil {
		return fmt.Errorf("upsert event: %w", err)
	}

	return nil
}

func GetEventById(redis *redis.Client, id string) (Event, error) {
	key := EventKeyForID(id)
	result := redis.HGetAll(key)
	if result.Err() != nil {
		return Event{}, fmt.Errorf("get properties for event: %w", result.Err())
	}

	data := result.Val()
	timestamp, err := strconv.ParseInt(data["time"], 10, 64)
	if err != nil {
		return Event{}, fmt.Errorf("parse date from time: %w", err)
	}

	return Event{
		ID:                data["id"],
		Name:              data["name"],
		Time:              time.Unix(timestamp, 0).UTC(),
		Status:            EventStatus(data["status"]),
		RecurringEventID:  data["recurring_event_id"],
		AnnounceMessageID: data["announce_message_id"],
		AnnounceChannelID: data["announce_channel_id"],
	}, nil
}

var ErrEventNotFound = errors.New("event could not be found")

func GetEventByMessage(r *redis.Client, channelID string, messageID string) (Event, error) {
	key := EventIndexKeyForMessageId(channelID, messageID)
	result := r.Get(key)
	if result.Err() == redis.Nil {
		return Event{}, ErrEventNotFound
	} else if result.Err() != nil {
		return Event{}, fmt.Errorf("lookup event by message: %w", result.Err())
	}

	id := result.Val()
	if key == "" {
		return Event{}, ErrEventNotFound
	}

	return GetEventById(r, id)
}

func GetEventsForWeek(redis *redis.Client, date time.Time) ([]Event, error) {
	result := redis.SMembers(EventIndexKeyForDate(date))
	if result.Err() != nil {
		return nil, fmt.Errorf("get events for week: %w", result.Err())
	}

	evts := []Event{}
	for _, id := range result.Val() {
		evt, err := GetEventById(redis, id)
		if err != nil {
			return nil, fmt.Errorf("get event by id: %w", err)
		}

		evts = append(evts, evt)
	}

	return evts, nil
}

func AnnounceEvent(session *discordgo.Session, redis *redis.Client, evt Event) error {
	embed, err := GetEmbedForEvent(session, redis, evt)
	if err != nil {
		log.Error(err)
	}

	if evt.AnnounceChannelID != "" && evt.AnnounceMessageID != "" {
		log.WithFields(log.Fields{
			"id":         evt.ID,
			"message_id": evt.AnnounceMessageID,
			"channel_id": evt.AnnounceChannelID,
		}).Debug("update announce message for event")

		_, err = session.ChannelMessageEditEmbed(evt.AnnounceChannelID, evt.AnnounceMessageID, embed)
		if err != nil {
			return fmt.Errorf("update announce message embed")
		}
	} else {
		log.WithFields(log.Fields{
			"id": evt.ID,
		}).Info("create new announcement for event")
		msg, err := session.ChannelMessageSendEmbed(evt.AnnounceChannelID, embed)
		if err != nil {
			log.Error(err)
		}

		evt.AnnounceMessageID = msg.ID
		err = ScheduleEvent(redis, evt)
		if err != nil {
			return fmt.Errorf("update event: %w", err)
		}
	}

	err = session.MessageReactionAdd(evt.AnnounceChannelID, evt.AnnounceMessageID, "❌")
	if err != nil {
		return fmt.Errorf("add reaction: %w", err)
	}

	err = session.MessageReactionAdd(evt.AnnounceChannelID, evt.AnnounceMessageID, "🕘")
	if err != nil {
		return fmt.Errorf("add reaction: %w", err)
	}

	return nil
}

func FormattedUserList(session *discordgo.Session, redis *redis.Client, ids []string) (string, error) {
	if len(ids) == 0 {
		return "No one 👍", nil
	}

	result := ""
	for _, id := range ids {
		alias, err := GetUserAlias(redis, session, id)
		if err != nil {
			return "", fmt.Errorf("get user alias: %w", err)
		}
		result = result + alias + "\n"
	}

	return result, nil
}

func GetEmbedForEvent(session *discordgo.Session, redis *redis.Client, evt Event) (*discordgo.MessageEmbed, error) {
	attendance, err := GetAttendanceForEvent(redis, evt)
	if err != nil {
		return nil, fmt.Errorf("get attendance for event: %w", err)
	}

	out, err := FormattedUserList(session, redis, attendance.Absent)
	if err != nil {
		return nil, fmt.Errorf("format absent user list: %w", err)
	}

	late, err := FormattedUserList(session, redis, attendance.Late)
	if err != nil {
		return nil, fmt.Errorf("format late user list: %w", err)
	}

	embed := discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			Name:    evt.Name,
			IconURL: session.State.User.AvatarURL(""),
		},
		Title:       "Castle Nathria",
		Description: evt.Time.Format("Monday Jan _2 2006"),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://wow.zamimg.com/images/wow/icons/large/achievement_raid_revendrethraid_castlenathria.jpg",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "❌ Out",
				Value:  out,
				Inline: true,
			},
			{
				Name:   "🕘 Late",
				Value:  late,
				Inline: true,
			},
			{
				Name:  "Instructions",
				Value: "React with 🕘 to mark yourself late or ❌ for out",
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("id: %s", evt.ID),
		},
	}

	return &embed, nil
}

func UserAliasKey(userID string) string {
	return fmt.Sprintf("alias:%s", userID)
}

func SetUserAlias(redis *redis.Client, userID string, alias string) error {
	key := UserAliasKey(userID)
	result := redis.Set(key, alias, 0)
	if result.Err() != nil {
		return fmt.Errorf("redis write: %w", result.Err())
	}

	return nil
}

func GetUserAlias(r *redis.Client, session *discordgo.Session, userID string) (string, error) {
	key := UserAliasKey(userID)
	result := r.Get(key)
	if result.Err() == nil {
		return result.Val(), nil
	}

	user, err := session.User(userID)
	if err != nil {
		return "", fmt.Errorf("query user from discord: %w", err)
	}

	err = SetUserAlias(r, userID, user.Username)
	if err != nil {
		return "", fmt.Errorf("set user alias: %w", err)
	}

	return user.Username, nil
}
