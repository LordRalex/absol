package log

import (
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

var lastAuditIds = make(map[string]string)
var auditLastCheck sync.Mutex
var loggedServers []string
var client = &http.Client{}

type Module struct {
	api.Module
}

func (*Module) Load(session *discordgo.Session) {
	loggedServers = strings.Split(viper.GetString("LOGGED_SERVERS"), ";")

	session.AddHandler(OnMessageCreate)
	session.AddHandler(OnMessageDelete)
	session.AddHandler(OnMessageDeleteBulk)
	session.AddHandler(OnMessageEdit)
	session.AddHandlerOnce(OnConnect)
}

func OnConnect(ds *discordgo.Session, mc *discordgo.Connect) {
	auditLastCheck.Lock()
	defer auditLastCheck.Unlock()

	for _, guild := range ds.State.Guilds {
		auditLog, err := ds.GuildAuditLog(guild.ID, "", "", discordgo.AuditLogActionMessageDelete, 1)
		if err != nil {
			logger.Err().Printf("Failed to check audit log: %s", err.Error())
		} else {
			for _, v := range auditLog.AuditLogEntries {
				lastAuditIds[guild.ID] = v.ID
			}
		}
	}
}

func getGuild(ds *discordgo.Session, guildId string) *discordgo.Guild {
	g, err := ds.State.Guild(guildId)
	if err != nil {
		// Try fetching via REST API
		g, err = ds.Guild(guildId)
		if err != nil {
			logger.Err().Printf("unable to fetch Guild for Message, %s", err)
		} else {
			// Attempt to add this channel into our State
			err = ds.State.GuildAdd(g)
			if err != nil {
				logger.Err().Printf("error updating Guild with Channel, %s", err)
			}
		}
	}

	return g
}
func getChannel(ds *discordgo.Session, channelId string) *discordgo.Channel {
	c, err := ds.State.Channel(channelId)
	if err != nil {
		// Try fetching via REST API
		c, err = ds.Channel(channelId)
		if err != nil {
			logger.Err().Printf("unable to fetch Channel for Message, %s", err)
		} else {
			// Attempt to add this channel into our State
			err = ds.State.ChannelAdd(c)
			if err != nil {
				logger.Err().Printf("error updating State with Channel, %s", err)
			}
		}
	}

	return c
}

func downloadAttachment(db *gorm.DB, id, url, filename string) {
	//check to see if URL already exists, if so, skip

	stmt, _ := db.DB().Prepare("SELECT id from attachments WHERE url = ?")
	rows, err := stmt.Query(url)
	_ = stmt.Close()
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}
	hasRows := rows.Next()
	_ = rows.Close()
	if hasRows {
		return
	}

	var data []byte
	response, err := client.Get(url)
	if err == nil {
		defer response.Body.Close()
		data, _ = ioutil.ReadAll(response.Body)
	}

	stmt, _ = db.DB().Prepare("INSERT INTO attachments (message_id, url, name, contents) VALUES (?, ?, ?, ?);")
	err = database.Execute(stmt, id, url, filename, data)
	if err != nil {
		logger.Err().Print(err.Error())
	}
}