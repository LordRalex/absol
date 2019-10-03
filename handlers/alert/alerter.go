package alert

import (
	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"time"
)

var client = &http.Client{
	Timeout: time.Second * 30,
}

var sites []*site

func Schedule(d *discordgo.Session) {

	viper.SetDefault("MCF_COUNT", 5)
	viper.SetDefault("MCF_PERIOD", 2)

	siteKeys := strings.Split(viper.GetString("SITES"), ";")

	for _, v := range siteKeys {
		sites = append(sites, &site{
			SiteName:       v,
			RSSUrl:         viper.GetString("SITES_" + v + "_RSS"),
			AlertChannel:   strings.Split(viper.GetString("SITES_"+v+"_CHANNELS"), ";"),
			AlertServer:    strings.Split(viper.GetString("SITES_"+v+"_SERVERS"), ";"),
			Cookie:         viper.GetString("SITES_" + v + "_COOKIES_COBALTSESSION"),
			Domain:         viper.GetString("SITES_" + v + "_DOMAIN"),
			MaxErrors:      viper.GetInt("SITES_" + v + "_MAXERRORS"),
			Period:         viper.GetInt("SITES_" + v + "_PERIOD"),
			lastPingFailed: false,
			silent:         false,
		})
	}

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
