package alert

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/database"
	"github.com/lordralex/absol/logger"
	"net/http"
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
			}

			break
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
