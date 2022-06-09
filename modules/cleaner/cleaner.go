package cleaner

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"strings"
	"time"
)

type Module struct {
	api.Module
}

func (c *Module) Load(d *discordgo.Session) {
	go func(ds *discordgo.Session) {
		runTick(ds)

		timer := time.NewTicker(10 * time.Minute)

		for {
			select {
			case <-timer.C:
				{
					runTick(ds)
				}
			}
		}
	}(d)
}

func runTick(ds *discordgo.Session) {
	logger.Debug().Print("Running channel cleanup")
	envChan := viper.GetString("cleanerChannel")

	postDelay := -1 * time.Hour * 24
	delay := viper.GetInt("cleanerTime")
	if delay != 0 {
		postDelay = -1 * time.Hour * time.Duration(delay)
	}

	channels := strings.Split(envChan, ";")

	cutOff := time.Now().Add(postDelay)

	for _, channel := range channels {
		c := api.GetChannel(ds, channel)

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

			if m.Timestamp.Before(cutOff) {
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

func (Module) Name() string {
	return "cleaner"
}
