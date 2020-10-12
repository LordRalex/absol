package alert

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type site struct {
	SiteName     string   `gorm:"column:name"`
	RSSUrl       string   `gorm:"column:rss"`
	ElmahUrl     string   `gorm:"column:elmahUrl"`
	AlertChannel []string `gorm:"-"`
	Channels     string
	AlertServer  []string `gorm:"-"`
	Servers      string
	Cookie       string   `gorm:"column:cookie_cobaltsession"`
	Domain       string
	MaxErrors    int
	Period       int

	lastPingFailed bool
	silent         bool
	fullyIgnore    bool
}

type sites []*site

func (s *site) runTick(ds *discordgo.Session) {
	if s.fullyIgnore {
		return
	}

	defer func() {
		if err := recover(); err != nil {
			logger.Err().Printf("Error running site checker tick for %s: %v", s.SiteName, err)
		}
	}()

	logger.Debug().Printf("Pinging %s", s.SiteName)
	req, err := s.createRequest(s.RSSUrl)
	if err != nil {
		s.sendMessage(ds, fmt.Sprintf("Error pinging: %s", err.Error()))
		return
	}

	response, err := client.Do(req)
	if err != nil {
		if s.lastPingFailed && !s.silent {
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

	if response.StatusCode != 200 && !s.silent {
		s.sendMessage(ds, fmt.Sprintf("Error pinging: Status Code %s", response.Status))
		return
	}

	data := &RootXML{}
	err = xml.NewDecoder(response.Body).Decode(data)

	if err != nil && !s.silent {
		s.sendMessage(ds, fmt.Sprintf("Error pinging: %s", err.Error()))
		return
	}
	counter := 0

	if len(data.Channel.Item) == 0 && !s.silent {
		s.sendMessage(ds, fmt.Sprintf("Error pinging: RSS Log is empty"))
		return
	}

	var importantErrors []string
	for _, e := range data.Channel.Item {
		if s.isReportable(e) {
			counter++
		}

		if s.isImportantError(e) {
			if importantErrors == nil {
				importantErrors = []string{e.Title}
			} else {
				importantErrors = append(importantErrors, e.Title)
			}
		}

		s.isLoggable(e)
	}

	if importantErrors != nil && len(importantErrors) >= 0 {
		s.sendMessage(ds, fmt.Sprintf("Important Errors: \n%s", strings.Join(importantErrors, "\n")))
	}

	if counter >= s.MaxErrors && !s.silent {
		s.sendMessage(ds, fmt.Sprintf("%d errors detected in report log in last %d minutes, please investigate", counter, s.Period))
	}
}

func (s *site) sendMessage(ds *discordgo.Session, msg string) {
	logger.Out().Printf(msg)

	logger.Debug().Printf("Sending message to server '%s' and channel '%s'", s.AlertServer, s.AlertChannel)

	for _, v := range s.AlertChannel {
		_, _ = ds.ChannelMessageSend(v, fmt.Sprintf("[%s] [%s]\n%s", s.SiteName, s.ElmahUrl, msg))
		s.silent = true
		time.AfterFunc(time.Minute*5, func() {
			s.silent = false
		})
	}
}

func (s *site) isReportable(data Item) bool {
	if s.silent {
		return false
	}

	cutoffTime := time.Now().Add(time.Duration(-1*s.Period) * time.Minute)
	if !data.PublishDate.After(cutoffTime) {
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

func (s *site) isImportantError(data Item) bool {
	cutoffTime := time.Now().Add(time.Duration(-1*s.Period) * time.Minute)
	if !data.PublishDate.After(cutoffTime) {
		return false
	}

	db, err := database.Get()
	if err != nil {
		logger.Err().Printf("Error connecting to database: %s\n", err.Error())
		return false
	}

	stmt, err := db.DB().Prepare("SELECT COUNT(1) AS Matches FROM sites_important_errors WHERE site = ? AND ( ? LIKE title OR ? LIKE description)")
	if err != nil {
		logger.Err().Printf("Error checking if record is ignorable: %s\n", err.Error())
		return false
	}
	defer stmt.Close()

	results, err := stmt.Query(s.SiteName, data.Title, data.Description)
	if err != nil {
		logger.Err().Printf("Error checking if record is ignorable: %s\n", err.Error())
		return false
	}
	defer results.Close()

	results.Next()
	var count int
	err = results.Scan(&count)
	if err != nil {
		logger.Err().Printf("Error checking if record is ignorable: %s\n", err.Error())
		return false
	}

	return count != 0
}

func (s *site) isLoggable(item Item) bool {
	cutoffTime := time.Now().Add(time.Duration(-1*s.Period) * time.Minute)
	if !item.PublishDate.After(cutoffTime) {
		return false
	}

	return true
}

func (s *site) AfterFind() (err error) {
	s.AlertChannel = strings.Split(s.Channels, ";")
	s.AlertServer = strings.Split(s.Servers, ";")
	return
}

func (s *site) createRequest(requestUrl string) (req *http.Request, err error) {
	req = &http.Request{}
	req.URL, err = url.Parse(requestUrl)
	if err != nil {
		return
	}

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

	return
}

func submitToElastic(id string, data map[string]interface{}) error {
	delete(data, "cookies")
	delete(data, "host")
	delete(data, "form")
	if data["serverVariables"] != nil {
		serverVars := data["serverVariables"].(map[string]interface{})
		delete(serverVars, "ALL_HTTP")
		delete(serverVars, "ALL_RAW")
		delete(serverVars, "HTTP_COOKIE")
		delete(serverVars, "AUTH_PASSWORD")

		forwards := serverVars["HTTP_X_FORWARDED_FOR"]
		if ips, ok := forwards.(string); ok && ips != "" {
			ipList := strings.Split(ips, ",")
			for k, v := range ipList {
				ipList[k] = strings.TrimSpace(v)
			}
			serverVars["HTTP_X_FORWARDED_FOR"] = ipList
		} else {
			delete(serverVars, "HTTP_X_FORWARDED_FOR")
		}
	}
	delete(data, "webHostHtmlMessage")
	delete(data, "HTTP_COOKIE")

	//convert query string to a standard string instead
	qs := data["queryString"]
	if qs != nil {
		d, err := json.Marshal(qs)
		if err != nil {
			logger.Err().Printf("Failed to convert query string to json; %s\n", err.Error())
			delete(data, "queryString")
		} else {
			queryString := bytes.NewBuffer(d).String()
			data["queryString"] = queryString
		}
	} else {
		delete(data, "queryString")
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	es, err := http.NewRequest("PUT", viper.GetString("elastic.url")+"/_doc/"+id, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}
	es.SetBasicAuth(viper.GetString("elastic.user"), viper.GetString("elastic.pass"))
	es.Header.Set("Content-Type", "application/json")

	response, err := client.Do(es)

	defer func() {
		if response != nil && response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if err == nil && !(response.StatusCode == 201 || response.StatusCode == 200) {
		body, _ := ioutil.ReadAll(response.Body)
		err = errors.New(fmt.Sprintf("Failed to save log (%s): %s", response.Status, body))
	}

	return err
}
