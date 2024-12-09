package log

import (
	"database/sql"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
)

func OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.GuildID == "" {
		return
	}

	logged := false
	for _, v := range loggedServers {
		if v == mc.GuildID {
			logged = true
		}
	}

	if !logged {
		return
	}

	c := api.GetChannel(ds, mc.ChannelID)

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

	logger.Debug().Printf("[%s] [%s] [%s] [%s]", c.Name, mc.ID, mc.Author.Username, message)

	gorm, err := database.Get()
	if err != nil {
		logger.Err().Printf("Error connecting to database: %s\n", err.Error())
		return
	}

	db, err := gorm.DB()
	if err != nil {
		logger.Err().Printf("Error connecting to database: %s\n", err.Error())
		return
	}

	guild := api.GetGuild(ds, mc.GuildID)

	stmt, _ := db.Prepare("INSERT INTO guilds (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?;")
	err = database.Execute(stmt, guild.ID, guild.Name, guild.Name)
	if err != nil {
		logger.Err().Print(err.Error())
	}

	stmt, _ = db.Prepare("INSERT INTO channels (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?;")
	err = database.Execute(stmt, c.ID, c.Name, c.Name)
	if err != nil {
		logger.Err().Print(err.Error())
	}

	stmt, _ = db.Prepare("INSERT INTO users (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?;")
	err = database.Execute(stmt, mc.Author.ID, mc.Author.Username, mc.Author.Username)
	if err != nil {
		logger.Err().Print(err.Error())
	}

	replyId := sql.NullString{}
	if mc.MessageReference != nil {
		replyId.String = mc.MessageReference.MessageID
		replyId.Valid = true
	}

	stmt, _ = db.Prepare("INSERT INTO messages (id, user_id, content, guild_id, channel_id, reply_id) VALUES (?, ?, ?, ?, ?, ?);")
	err = database.Execute(stmt, mc.ID, mc.Author.ID, message, mc.GuildID, mc.ChannelID, replyId)
	if err != nil {
		logger.Err().Print(err.Error())
	}

	if mc.Attachments != nil && len(mc.Attachments) > 0 {
		for _, attachment := range mc.Attachments {
			//trigger a download of the file
			go downloadAttachment(db, mc.ID, attachment.URL, attachment.Filename)
		}
	}
}
