package log

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lordralex/absol/database"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"log"
	"strings"
	"sync"
)

var db *sql.DB

var lastAuditIds = make(map[string]string)
var auditLastCheck sync.Mutex
var loggedServers []string

func RegisterCore(session *discordgo.Session) {
	var err error

	loggedServers = strings.Split(viper.GetString("LOGGED_SERVERS"), ";")

	db, err = database.Get()
	if err != nil {
		log.Fatalf("Database connection failed: %s", err.Error())
		return
	}

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

func OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	g := getGuild(ds, mc.GuildID)
	if g == nil {
		return
	}

	logged := false
	for _, v := range loggedServers {
		if v == g.Name {
			logged = true
		}
	}

	if !logged {
		return
	}

	c := getChannel(ds, mc.ChannelID)

	if c == nil || c.Type == discordgo.ChannelTypeDM {
		return
	}

	message, err := mc.ContentWithMoreMentionsReplaced(ds)
	if err != nil {
		message = mc.ContentWithMentionsReplaced()
	}

	logger.Debug().Printf("[%s] [%s] [%s#%s] [%s]", c.Name, mc.ID, mc.Author.Username, mc.Author.Discriminator, message)

	stmt, err := db.Prepare("INSERT INTO messages (id, channel, sender, content) VALUES (?, ?, ?, ?);")
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}
	_, err = stmt.Exec(mc.ID, c.Name, mc.Author.Username+"#"+mc.Author.Discriminator, message)
	if err != nil {
		logger.Err().Print(err.Error())
	}
}

func OnMessageEdit(ds *discordgo.Session, mc *discordgo.MessageUpdate) {
	if mc.Author != nil && mc.Author.ID == ds.State.User.ID {
		return
	}

	g := getGuild(ds, mc.GuildID)
	if g == nil {
		return
	}

	logged := false
	for _, v := range loggedServers {
		if v == g.Name {
			logged = true
		}
	}

	if !logged {
		return
	}

	c := getChannel(ds, mc.ChannelID)

	if c == nil || c.Type == discordgo.ChannelTypeDM {
		return
	}

	message, err := mc.ContentWithMoreMentionsReplaced(ds)
	if err != nil {
		message = mc.ContentWithMentionsReplaced()
	}

	for _, v := range mc.Embeds {
		if v.Author != nil {
			if message != "" {
				message += "\r\n"
			}
			message += v.Author.Name
		}
		if v.Description != "" {
			if message != "" {
				message += "\r\n"
			}
			message += v.Description
		}
	}

	logger.Debug().Printf("[EDIT] [%s] [%s] [%s]", c.Name, mc.ID, message)

	stmt, err := db.Prepare("INSERT INTO edits (message_id, old_content) SELECT id, content FROM messages WHERE id =?")
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}
	_, err = stmt.Exec(mc.Message.ID)
	if err != nil {
		logger.Err().Print(err.Error())
	}

	stmt, err = db.Prepare("UPDATE messages SET content = ? WHERE id = ?")
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}
	_, err = stmt.Exec(message, mc.Message.ID)
	if err != nil {
		logger.Err().Print(err.Error())
	}

}

func OnMessageDelete(ds *discordgo.Session, mc *discordgo.MessageDelete) {
	g := getGuild(ds, mc.GuildID)
	if g == nil {
		return
	}

	logged := false
	for _, v := range loggedServers {
		if v == g.Name {
			logged = true
		}
	}

	if !logged {
		return
	}

	if mc.Author != nil && mc.Author.Username != "" {
		logger.Debug().Printf("[DELETE] [%s] [%s]", mc.Author.Username, mc.ID)
	} else {
		logger.Debug().Printf("[DELETE] [%s]", mc.ID)
	}

	go func() {
		auditLastCheck.Lock()
		defer auditLastCheck.Unlock()
		auditLog, err := ds.GuildAuditLog(mc.GuildID, "", "", discordgo.AuditLogActionMessageDelete, 1)
		if err != nil {
			logger.Err().Printf("Failed to check audit log: %s", err.Error())
		} else {
			for _, v := range auditLog.AuditLogEntries {
				if lastAuditIds[mc.GuildID] == v.ID {
					//we have already processed this ID, which means the delete was a self-delete
				} else {
					lastAuditIds[mc.GuildID] = v.ID

					var deleter, messenger string
					for _, u := range auditLog.Users {
						if u.ID == v.UserID {
							deleter = u.Username + "#" + u.Discriminator
						} else if u.ID == v.TargetID {
							messenger = u.Username + "#" + u.Discriminator
						}
					}

					logger.Debug().Printf("[AUDIT] [%s] deleted messages by [%s]", deleter, messenger)
					for _, v := range ds.State.Guilds[0].Channels {
						if v.Name == "bot" {
							_, err = ds.ChannelMessageSend(v.ID, fmt.Sprintf("LAST AUDIT ACTION - %s deleted messages by %s", deleter, messenger))
							if err != nil {
								logger.Err().Printf("Error sending audit message: %s", err.Error())
							}
							break
						}
					}
				}
			}
		}
	}()

	stmt, err := db.Prepare("UPDATE messages SET deleted = 1 WHERE id = ?;")
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}
	_, err = stmt.Exec(mc.ID)
	if err != nil {
		logger.Err().Print(err.Error())
	}
}

func OnMessageDeleteBulk(ds *discordgo.Session, mc *discordgo.MessageDeleteBulk) {
	g := getGuild(ds, mc.GuildID)
	if g == nil {
		return
	}

	logged := false
	for _, v := range loggedServers {
		if v == g.Name {
			logged = true
		}
	}

	if !logged {
		return
	}

	logger.Debug().Printf("[DELETE-BULK] [%s]", mc.Messages)

	for _, v := range mc.Messages {
		go func(id string) {
			stmt, err := db.Prepare("UPDATE messages SET deleted = 1 WHERE id = ?;")
			if err != nil {
				logger.Err().Print(err.Error())
				return
			}
			_, err = stmt.Exec(id)
			if err != nil {
				logger.Err().Print(err.Error())
			}
		}(v)
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
