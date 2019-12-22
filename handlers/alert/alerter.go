package alert

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/database"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"net/http"
	"strings"
	"time"
)

var knownSites []*site

var client = &http.Client{
	Timeout: time.Second * 30,
}

func Schedule(d *discordgo.Session) {
	go func(ds *discordgo.Session) {
		timer := time.NewTicker(time.Minute)

		for {
			select {
			case <-timer.C:
				{
					syncSites()
					for _, v := range knownSites {
						go func(s *site) {
							s.runTick(ds)
						}(v)
					}
				}}
		}
	}(d)
}

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, c *discordgo.Channel, cmd string, args []string) {
	//do not run when in DMs
	if c == nil || c.Type == discordgo.ChannelTypeDM {
		return
	}

	//only permit usage of this command in certain servers
	allowed := viper.GetString("alerter.server")
	guild, err := ds.State.Guild(mc.GuildID)

	if err != nil {
		logger.Err().Printf("Error getting guild information")
		return
	}
	if guild == nil {
		return
	}

	if guild.Name != allowed {
		return
	}

	if len(args) == 0 {
		if cmd == "silent" {
			_, _ = ds.ChannelMessageSend(c.ID, "Usage: silent <sitename> [duration]")
		} else if cmd == "resume" {
			_, _ = ds.ChannelMessageSend(c.ID, "Usage: resume <sitename>")
		}
		return
	}

	siteName := strings.ToLower(args[0])
	var targetSite *site
	for _, v := range knownSites {
		if strings.ToLower(v.SiteName) == siteName {
			targetSite = v
			break
		}
	}

	if targetSite == nil {
		_, _ = ds.ChannelMessageSend(c.ID, "Usage: No site with given name")
		return
	}

	switch cmd {
	case "silent":
		{
			silentTime := time.Hour
			if len(args) == 2 {
				silentTime, err = time.ParseDuration(args[1])
				if err != nil {
					_, _ = ds.ChannelMessageSend(c.ID, "Failed to parse time duration: "+err.Error())
				}
			}
			targetSite.fullyIgnore = true
			go func(ts *site, duration time.Duration, chanId string) {
				<-time.After(silentTime)
				if targetSite.fullyIgnore {
					targetSite.fullyIgnore = false
					_, _ = ds.ChannelMessageSend(chanId, fmt.Sprintf("Reporting re-enabled for %s", targetSite.SiteName))
				}
			}(targetSite, silentTime, c.ID)

			_, _ = ds.ChannelMessageSend(c.ID, fmt.Sprintf("Will not report errors from %s for %s", targetSite.SiteName, silentTime.String()))
			break
		}
	case "resume":
		{
			targetSite.silent = false
			_, _ = ds.ChannelMessageSend(c.ID, fmt.Sprintf("Reporting re-enabled for %s", targetSite.SiteName))
			break
		}
	}
}

func syncSites() {
	db, err := database.Get()
	if err != nil {
		logger.Err().Printf("Error getting DB connection: %s\n", err.Error())
		return
	}

	dbSites := &sites{}
	err = db.Find(dbSites).Error
	if err != nil {
		logger.Err().Printf("Error looking for new db sites: %s\n", err.Error())
		return
	}

	for _, v := range *dbSites {
		exists := false
		for _, e := range knownSites {
			if v.SiteName == e.SiteName {
				e.AlertServer = v.AlertServer
				e.RSSUrl = v.RSSUrl
				e.AlertChannel = v.AlertChannel
				e.Channels = v.Channels
				e.AlertServer = v.AlertServer
				e.Servers = v.Servers
				e.Cookie = v.Cookie
				e.Domain = v.Domain
				e.MaxErrors = v.MaxErrors
				e.Period = v.Period
				exists = true
				break
			}
		}

		if !exists {
			knownSites = append(knownSites, v)
		}
	}

	for i := 0; i < len(knownSites); i++ {
		exists := false
		for _, v := range *dbSites {
			if v.SiteName == knownSites[i].SiteName {
				exists = true
				break
			}
		}
		if !exists {
			copy(knownSites[i:], knownSites[i+1:])
			knownSites[len(knownSites)-1] = nil
			knownSites = knownSites[:len(knownSites)-1]
			i--
		}
	}
}
