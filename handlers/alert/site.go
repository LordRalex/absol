package alert

import (
	"encoding/xml"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/logger"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type site struct {
	SiteName     string   `gorm:"name"`
	RSSUrl       string   `gorm:"rss"`
	AlertChannel []string `gorm:"-"`
	Channels     string
	AlertServer  []string `gorm:"-"`
	Servers      string
	Cookie       string `gorm:"cookie_cobaltsession"`
	Domain       string
	MaxErrors    int
	Period       int

	lastPingFailed bool
	silent         bool
}

type sites []site

func (s *site) runTick(ds *discordgo.Session) {
	defer func() {
		if err := recover(); err != nil {
			logger.Err().Printf("Error running site checker tick for %s: %v", s.SiteName, err)
		}
	}()

	if s.silent {
		return
	}

	logger.Debug().Printf("Pinging %s", s.SiteName)

	var err error

	req := &http.Request{}
	req.URL, err = url.Parse(s.RSSUrl)

	req.Header = http.Header{}
	req.Method = "GET"

	req.AddCookie(&http.Cookie{
		Name:    "CobaltSession",
		Value:   s.Cookie,
		Path:    "/",
		Domain:  s.Domain,
		Expires: time.Now().Add(time.Hour * 24 * 365),
		Secure:  true,
	})

	response, err := client.Do(req)
	if err != nil {
		if s.lastPingFailed {
			s.sendMessage(ds, fmt.Sprintf("Error pinging: %s", err.Error()))
			s.lastPingFailed = false
		} else {
			s.lastPingFailed = true
		}
		return
	}
	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if response.StatusCode != 200 {
		s.sendMessage(ds, fmt.Sprintf("Error pinging: Status Code %s", response.Status))
	}

	data := &RootXML{}
	err = xml.NewDecoder(response.Body).Decode(data)

	if err != nil {
		s.sendMessage(ds, fmt.Sprintf("Error pinging: %s", err.Error()))
	}
	counter := 0

	if len(data.Channel.Item) == 0 {
		s.sendMessage(ds, fmt.Sprintf("Error pinging: RSS Log is empty"))
	}

	for _, e := range data.Channel.Item {
		if s.isReportable(e) {
			counter++
		}
	}

	if counter >= s.MaxErrors {
		s.sendMessage(ds, fmt.Sprintf("%d errors detected in report log in last %d minutes, please investigate", counter, s.Period))
	}
}

func (s *site) sendMessage(ds *discordgo.Session, msg string) {
	logger.Out().Printf(msg)

	logger.Debug().Printf("Sending message to server '%s' and channel '%s'", s.AlertServer, s.AlertChannel)

	for k, v := range s.AlertServer {
		for _, guild := range ds.State.Guilds {
			if guild.Name == v {
				for _, c := range guild.Channels {
					if c.Name == s.AlertChannel[k] {
						_, _ = ds.ChannelMessageSend(c.ID, fmt.Sprintf("[%s] %s", s.SiteName, msg))
						s.silent = true
						time.AfterFunc(time.Minute*5, func() {
							s.silent = false
						})
					}
				}
			}
		}
	}
}

func (s *site) isReportable(data Item) bool {
	cutoffTime := time.Now().Add(time.Duration(-1*s.Period) * time.Minute)

	return data.PublishDate.After(cutoffTime)
}

func (s *site) AfterFind() (err error) {
	s.AlertChannel = strings.Split(s.Channels, ";")
	s.AlertServer = strings.Split(s.Servers, ";")
	return
}
