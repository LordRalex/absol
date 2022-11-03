package alert

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"io"
	"net/http"
	"time"
)

var knownSites []*Site

var client = &http.Client{
	Timeout: time.Second * 30,
}

type Module struct {
	api.Module
}

func (m *Module) Load(d *discordgo.Session) {
	d.AddHandlerOnce(func(ds *discordgo.Session, e *discordgo.Connect) {
		db, err := database.Get()
		if err != nil {
			logger.Err().Printf("Error getting DB connection: %s\n", err.Error())
			return
		}

		_ = db.AutoMigrate(&Site{})

		var hooks []string
		err = db.Model(&Site{}).Select("webhook").Where("webhook <> ''").Distinct().Find(&hooks).Error
		if err != nil {
			logger.Err().Printf("Error looking for new db sites: %s\n", err.Error())
			return
		}

		if len(hooks) == 0 {
			return
		}

		//update the webhooks with our avatar, just so it's nice to see
		url := ds.State.User.AvatarURL("")

		body, err := client.Get(url)
		if err != nil {
			return
		}
		defer body.Body.Close()
		data, err := io.ReadAll(body.Body)
		body.Body.Close()
		if err != nil {
			logger.Err().Printf("Error updating avatar: %s\n", err.Error())
			return
		}

		mimeType := http.DetectContentType(data)
		base64Encoding := "data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data)
		data = nil

		type WebhookUpdate struct {
			Avatar string `json:"avatar"`
		}
		st := &WebhookUpdate{Avatar: base64Encoding}
		base64Encoding = ""

		jsonData, err := json.Marshal(st)
		if err != nil {
			logger.Err().Printf("Error updating avatar: %s\n", err.Error())
			return
		}

		for _, v := range hooks {
			logger.Debug().Printf("Updating webhook with our avatar: %s\n", v)
			request, _ := http.NewRequest("PATCH", v, bytes.NewBuffer(jsonData))
			request.Header.Set("Content-Type", "application/json")
			res, err := client.Do(request)
			if res != nil && res.Body != nil {
				response, _ := io.ReadAll(res.Body)
				logger.Debug().Printf("%s\n", response)
				_ = res.Body.Close()
			}
			if err != nil {
				logger.Err().Printf("Error updating avatar: %s\n", err.Error())
			}
		}
	})

	d.AddHandlerOnce(func(d *discordgo.Session, e *discordgo.Connect) {
		go func(ds *discordgo.Session) {
			timer := time.NewTicker(time.Minute)
			for {
				syncSites()
				for _, v := range knownSites {
					go func(s *Site) {
						s.runTick(ds)
					}(v)
				}
				<-timer.C
			}
		}(d)
	})
}

func syncSites() {
	db, err := database.Get()
	if err != nil {
		logger.Err().Printf("Error getting DB connection: %s\n", err.Error())
		return
	}

	var dbSites []*Site
	err = db.Find(&dbSites).Error
	if err != nil {
		logger.Err().Printf("Error looking for new db sites: %s\n", err.Error())
		return
	}

	for _, v := range dbSites {
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
		for _, v := range dbSites {
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

func (*Module) Name() string {
	return "alert"
}
