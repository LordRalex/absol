package servers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"strings"
	"time"
)

const PostDelay = -1 * time.Hour * 24

func Schedule(d *discordgo.Session) {
	go func(ds *discordgo.Session) {
		runTick(ds)

		timer := time.NewTicker(30 * time.Minute)

		select {
		case <-timer.C:
			{
				runTick(ds)
			}}
	}(d)
}

func runTick(ds *discordgo.Session) {
	logger.Debug().Print("Running channel cleanup")
	envChan := viper.GetString("cleanerChannel")
	server := viper.GetString("cleanerServer")

	channels := strings.Split(envChan, "|")

	cutOff := time.Now().Add(PostDelay)

	for _, guild := range ds.State.Guilds {
		if guild.Name == server {
			for _, ch := range guild.Channels {
				c, err := ds.State.Channel(ch.ID)
				if err != nil {
					// Try fetching via REST API
					c, err = ds.Channel(ch.ID)
					if err != nil {
						logger.Err().Printf("unable to fetch Channel for Message, %s", err)
					} else {
						// Attempt to add this channel into our State
						err = ds.State.ChannelAdd(c)
						if err != nil {
							logger.Err().Printf("error updating State with Channel, %s", err)
						}
					}
				}
				for _, channel := range channels {
					if c.Name == channel {
						messages := make([]string, 0)
						chanMessages, err := ds.ChannelMessages(c.ID, 100, "", "", "")
						if err != nil {
							logger.Err().Printf("Error cleaning channel: %v", err.Error())
							continue
						}

						for _, m := range chanMessages {
							creationDate, err := m.Timestamp.Parse()
							if err != nil {
								continue
							}
							if creationDate.Before(cutOff) {
								messages = append(messages, m.ID)
								if len(messages) > 20 {
									break
								}
							}
						}

						logger.Debug().Printf("Deleting %d messages", len(messages))

						err = ds.ChannelMessagesBulkDelete(c.ID, messages)
						if err != nil {
							logger.Err().Printf("Error cleaning channel: %v", err.Error())
						}
					}
				}
			}
		}
	}
}
