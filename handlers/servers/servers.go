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
		timer := time.NewTicker(30 * time.Minute)

		select {
		case <-timer.C:
			{
				runTick(ds)
			}}
	}(d)
}

func runTick(ds *discordgo.Session) {
	envChan := viper.GetString("cleanerChannel")
	server := viper.GetString("cleanerServer")

	channels := strings.Split(envChan, "|")

	cutOff := time.Now().Add(PostDelay)

	for _, guild := range ds.State.Guilds {
		if guild.ID == server {
			for _, c := range guild.Channels {
				for _, channel := range channels {
					if c.Name == channel {
						messages := make([]string, 0)
						for _, m := range c.Messages {
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

						err := ds.ChannelMessagesBulkDelete(c.ID, messages)
						if err != nil {
							logger.Err().Printf("Error cleaning channel: %v", err.Error())
						}
					}
				}
			}
		}
	}
}
