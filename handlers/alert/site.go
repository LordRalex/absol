package alert

import (
	"encoding/xml"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/database"
	"github.com/lordralex/absol/logger"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type site struct {
	SiteName     string   `gorm:"column:name"`
	RSSUrl       string   `gorm:"column:rss"`
	AlertChannel []string `gorm:"-"`
	Channels     string
	AlertServer  []string `gorm:"-"`
	Servers      string
	Cookie       string `gorm:"column:cookie_cobaltsession"`
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
	if ! data.PublishDate.After(cutoffTime) {
		return false
	}

	//if database connection fails, assume we can report this since it's in the right time-range
	db, err := database.Get()
	if err != nil {
		logger.Err().Printf("Error connecting to database: %s\n", err.Error())
		return true
	}

	stmt, err := db.DB().Prepare("SELECT COUNT(1) AS Matches FROM sites_ignored_errors WHERE site = ? AND ( ? LIKE title OR ? LIKE description)")
	if err != nil {
		logger.Err().Printf("Error checking if record is ignorable: %s\n", err.Error())
		return true
	}
	defer stmt.Close()

	results, err := stmt.Query(s.SiteName, data.Title, data.Description)
	if err != nil {
		logger.Err().Printf("Error checking if record is ignorable: %s\n", err.Error())
		return true
	}
	defer results.Close()

	results.Next()
	var count int
	err = results.Scan(&count)
	if err != nil {
		logger.Err().Printf("Error checking if record is ignorable: %s\n", err.Error())
		return true
	}

	return count == 0
}

func (s *site) AfterFind() (err error) {
	s.AlertChannel = strings.Split(s.Channels, ";")
	s.AlertServer = strings.Split(s.Servers, ";")
	return
}
