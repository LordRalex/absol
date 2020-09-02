package log

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
)

func OnMessageDelete(ds *discordgo.Session, mc *discordgo.MessageDelete) {
	logged := false
	for _, v := range loggedServers {
		if v == mc.GuildID {
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

	go func(guildId string) {
		auditLastCheck.Lock()
		defer auditLastCheck.Unlock()
		auditLog, err := ds.GuildAuditLog(guildId, "", "", discordgo.AuditLogActionMessageDelete, 1)
		if err != nil {
			logger.Err().Printf("Failed to check audit log: %s", err.Error())
		} else {
			for _, v := range auditLog.AuditLogEntries {
				if lastAuditIds[guildId] == v.ID {
					//we have already processed this ID, which means the delete was a self-delete
				} else {
					lastAuditIds[guildId] = v.ID

					var deleter, messenger string
					for _, u := range auditLog.Users {
						if u.ID == v.UserID {
							deleter = u.Username + "#" + u.Discriminator
						} else if u.ID == v.TargetID {
							messenger = u.Username + "#" + u.Discriminator
						}
					}

					logger.Debug().Printf("[AUDIT] [%s] deleted messages by [%s]", deleter, messenger)
					guild := getGuild(ds, guildId)
					for _, v := range guild.Channels {
						if v.Name == "bot" || v.Name == "log" {
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
	}(mc.GuildID)

	db, err := database.Get()
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}

	stmt, err := db.DB().Prepare("UPDATE messages SET deleted = 1 WHERE id = ?;")
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}
	err = database.Execute(stmt, mc.ID)
	if err != nil {
		logger.Err().Print(err.Error())
	}
}

func OnMessageDeleteBulk(ds *discordgo.Session, mc *discordgo.MessageDeleteBulk) {
	logged := false
	for _, v := range loggedServers {
		if v == mc.GuildID {
			logged = true
		}
	}

	if !logged {
		return
	}

	logger.Debug().Printf("[DELETE-BULK] [%s]", mc.Messages)

	db, err := database.Get()
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}

	stmt, err := db.DB().Prepare("UPDATE messages SET deleted = 1 WHERE id = ?;")
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}
	for _, v := range mc.Messages {
		err = database.Execute(stmt, v)
		if err != nil {
			logger.Err().Print(err.Error())
		}
	}
}
