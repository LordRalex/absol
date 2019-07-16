package handlers

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
	_ "github.com/go-sql-driver/mysql"
	"github.com/lordralex/absol/logger"
	"github.com/spf13/viper"
	"os"
	"sync"
)

var db *sql.DB

var lastAuditId string
var auditLastCheck sync.Mutex

func RegisterCore(session *discordgo.Session) {
	var err error

	connString := viper.GetString("database")
	if connString == "" {
		connString = "discord:discord@/discord"
	}

	db, err = sql.Open("mysql", connString)
	if err != nil {
		logger.Err().Print(err.Error())
		os.Exit(1)
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
	auditLog, err := ds.GuildAuditLog(ds.State.Guilds[0].ID, "", "", discordgo.AuditLogActionMessageDelete, 1)
	if err != nil {
		logger.Err().Printf("Failed to check audit log: %s", err.Error())
	} else {
		for _, v := range auditLog.AuditLogEntries {
			lastAuditId = v.ID
		}
	}
}

func OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == ds.State.User.ID {
		return
	}

	c, err := ds.State.Channel(mc.ChannelID)
	if err != nil {
		// Try fetching via REST API
		c, err = ds.Channel(mc.ChannelID)
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

	c, err := ds.State.Channel(mc.ChannelID)
	if err != nil {
		// Try fetching via REST API
		c, err = ds.Channel(mc.ChannelID)
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
				if lastAuditId == v.ID {
					//we have already processed this ID, which means the delete was a self-delete
				} else {

					lastAuditId = v.ID

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
