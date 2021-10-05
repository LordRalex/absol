package hyperlinkscanner

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"net/http"
	"os"
	"strings"
	"time"
)

const Api = "https://api.hyperphish.com/gimme-domains"

var badLinks = make([]string, 0)
var client = &http.Client{}
var watchedChannels []string
var loggerChannels = make(map[string]string)

type Module struct {
	api.Module
}

func (*Module) Load(session *discordgo.Session) {
	session.AddHandler(OnMessageCreate)
	session.AddHandler(OnMessageEdit)

	watchedChannels = strings.Split(os.Getenv("HYPERSCANNER_CHANNELS"), ",")
	mapping := strings.Split(os.Getenv("HYPERSCANNER_LOGTO"), ",")
	for _, v := range mapping {
		d := strings.Split(v, ":")
		loggerChannels[d[0]] = d[1]
	}

	api.RegisterIntentNeed(discordgo.IntentsGuildMessages)

	go func() {
		runTimer()
		timer := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-timer.C:
				{
					runTimer()
				}
			}
		}
	}()
}

func OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	skip := true
	for _, v := range watchedChannels {
		if v == mc.ChannelID {
			skip = false
		}
	}
	if skip {
		return
	}

	for _, v := range badLinks {
		if strings.Contains(mc.Content, v) {
			//THIS IS BAD!!!!
			_ = ds.ChannelMessageDelete(mc.ChannelID, mc.ID)
			_, _ = ds.GuildBan(mc.GuildID, mc.Author.ID)

			targetChan := loggerChannels[mc.GuildID]
			if targetChan != "" {
				_, _ = ds.ChannelMessageSend(targetChan, "User "+mc.Author.ID+" banned for posting a malicious link")
			}
		}
	}
}

func OnMessageEdit(ds *discordgo.Session, mc *discordgo.MessageUpdate) {}

func runTimer() {
	err := refreshList()
	if err != nil {
		logger.Err().Printf("Error refreshing bad link API: %s\n", err.Error())
	} else {
		logger.Debug().Printf("Refreshed the link database\n")
	}
}

func refreshList() error {
	response, err := client.Get(Api)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	var data []string
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		return err
	}
	if len(data) > 0 {
		badLinks = data
	}
	return nil
}
