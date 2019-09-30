package alert

import (
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

var client = &http.Client{}

var sites []*site

func Schedule(d *discordgo.Session) {

	viper.SetDefault("MCF_COUNT", 5)
	viper.SetDefault("MCF_PERIOD", 2)

	sites = append(sites, &site{
		SiteName:       "MinecraftForum",
		RSSUrl:         "https://www.minecraftforum.net/cp/elmah/rss",
		AlertChannel:   viper.GetString("ALERTCHANNEL"),
		AlertServer:    viper.GetString("ALERTSERVER"),
		Cookie:         viper.GetString("COOKIES_COBALTSESSION"),
		Domain:         ".minecraftforum.net",
		MaxErrors:      viper.GetInt("MCF_COUNT"),
		Period:         viper.GetInt("MCF_PERIOD"),
		lastPingFailed: false,
		silent:         false,
	})

	go func(ds *discordgo.Session) {
		timer := time.NewTicker(time.Minute)

		for {
			select {
			case <-timer.C:
				{
					for _, v := range sites {
						go func(s *site) {
							s.runTick(ds)
						}(v)
					}
				}}
		}
	}(d)
}
