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

		timer := time.NewTicker(10 * time.Minute)

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

	for _, g := range ds.State.Guilds {
		guild, _ := ds.Guild(g.ID)
		if guild.Name == server {
			for _, channel := range channels {
				c, err := ds.State.Channel(channel)
				if err != nil {
					// Try fetching via REST API
					c, err = ds.Channel(channel)
					if err != nil {
						logger.Err().Printf("unable to fetch Channel for Message, %s", err)
						return
					} else {
						// Attempt to add this channel into our State
						err = ds.State.ChannelAdd(c)
						if err != nil {
							logger.Err().Printf("error updating State with Channel, %s", err)
							return
						}
					}
				}

				pinned, err := ds.ChannelMessagesPinned(c.ID)
				if err != nil {
					logger.Err().Printf("Error cleaning channel: %v", err.Error())
					continue
				}

				messages := make([]string, 0)
				//we use 1 here so that we go back to the beginning of time
				chanMessages, err := ds.ChannelMessages(c.ID, 100, "", "1", "")
				if err != nil {
					logger.Err().Printf("Error cleaning channel: %v", err.Error())
					continue
				}

				for _, m := range chanMessages {
					skip := false
					for _, p := range pinned {
						if p.ID == m.ID {
							skip = true
						}
					}

					if skip {
						continue
					}

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
