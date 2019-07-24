package mcf

import (
	"encoding/xml"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"time"
)

var silent = false
var client = &http.Client{}
var lastPingFailed = false

const ErrorUrl = "https://www.minecraftforum.net/cp/elmah/rss"

func init() {
	viper.SetDefault("MCF_PERIOD", 2)
	viper.SetDefault("MCF_COUNT", 5)
}

func Schedule(d *discordgo.Session) {
	go func(ds *discordgo.Session) {
		timer := time.NewTicker(time.Minute)

		for {
			select {
			case <-timer.C:
				{
					runTick(ds)
				}}
		}
	}(d)
}

func RunTick() {
	runTick(nil)
}

func runTick(ds *discordgo.Session) {
	defer func() {
		if err := recover(); err != nil {
			logger.Err().Printf("Error running MCF tick: %v", err)
		}
	}()

	if silent {
		return
	}

	logger.Debug().Printf("Pinging MCF")

	var err error

	req := &http.Request{}
	req.URL, err = url.Parse(ErrorUrl)

	req.Header = http.Header{}
	req.Method = "GET"

	req.AddCookie(&http.Cookie{
		Name:    "CobaltSession",
		Value:   viper.GetString("cookies_cobaltsession"),
		Path:    "/",
		Domain:  ".minecraftforum.net",
		Expires: time.Now().Add(time.Hour * 24 * 365),
		Secure:  true,
	})

	response, err := client.Do(req)
	if err != nil {
		if lastPingFailed {
			sendMessage(ds, "Error pinging main site: "+err.Error())
			lastPingFailed = false
		} else {
			lastPingFailed = true
		}
		return
	}
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if response.StatusCode != 200 {
		sendMessage(ds, "I did not get to the site correctly")
	}

	data := &RootXML{}
	err = xml.NewDecoder(response.Body).Decode(data)

	if err != nil {
		sendMessage(ds, "RSS feed failed: "+err.Error())
	}

	period := viper.GetInt("MCF_PERIOD")
	cutoffTime := time.Now().Add(time.Duration(-1*period) * time.Minute)
	counter := 0

	if len(data.Channel.Item) == 0 {
		sendMessage(ds, "RSS feed failed: No items in log")
	}

	for _, e := range data.Channel.Item {
		if e.PublishDate.After(cutoffTime) {
			counter++
		}
	}

	if counter >= viper.GetInt("MCF_COUNT") {
		sendMessage(ds, fmt.Sprintf("%d errors detected in report log in last %d minutes, please investigate", counter, period))
	}
}

func sendMessage(ds *discordgo.Session, msg string) {
	logger.Out().Printf(msg)
	channel := viper.GetString("alertChannel")
	server := viper.GetString("alertServer")

	logger.Debug().Printf("Sending message to server '%s' and channel '%s'", server, channel)

	for _, guild := range ds.State.Guilds {
		if guild.Name == server {
			for _, c := range guild.Channels {
				if c.Name == channel {
					_, _ = ds.ChannelMessageSend(c.ID, msg)
					silent = true
					time.AfterFunc(time.Minute*5, func() {
						silent = false
					})
				}
			}
		}
	}
}
