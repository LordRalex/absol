package mcf

import (
	"encoding/xml"
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"net/http"
	"net/url"
	"time"
)

var silent = false
var client = &http.Client{}

const ErrorUrl = "https://www.minecraftforum.net/cp/elmah/rss"

func Schedule(d *discordgo.Session) {
	go func(ds *discordgo.Session) {
		timer := time.NewTicker(time.Minute)

		select {
		case <-timer.C:
			{
				runTick(ds)
			}}
	}(d)
}

func RunTick() {
	runTick(nil)
}

func runTick(ds *discordgo.Session) {
	if silent {
		return
	}

	var err error

	req := &http.Request{}
	req.URL, err = url.Parse(ErrorUrl)

	req.Header = http.Header{}
	req.Method = "GET"

	req.AddCookie(&http.Cookie{
		Name:    "AWSELB",
		Value:   viper.GetString("cookies_awselb"),
		Path:    "/",
		Domain:  "www.minecraftforum.net",
		Expires: time.Now().Add(time.Hour * 24 * 365),
		Secure:  true,
	})
	req.AddCookie(&http.Cookie{
		Name:    "__cfduid",
		Value:   viper.GetString("cookies___cfduid"),
		Path:    "/",
		Domain:  ".minecraftforum.net",
		Expires: time.Now().Add(time.Hour * 24 * 365),
		Secure:  true,
	})
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
		sendMessage(ds, "Error pinging main site: "+err.Error())
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
}

func sendMessage(ds *discordgo.Session, msg string) {
	channel := viper.GetString("alertChannel")
	server := viper.GetString("alertServer")

	for _, guild := range ds.State.Guilds {
		if guild.ID == server {
			for _, c := range guild.Channels {
				if c.Name == channel {
					_, _ = ds.ChannelMessageSend(c.ID, msg)
					time.AfterFunc(time.Minute*5, func() {
						silent = false
					})
				}
			}
		}
	}
}
