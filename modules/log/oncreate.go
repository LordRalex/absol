package log

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
)

func OnMessageCreate(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if mc.Author.ID == ds.State.User.ID {
		return
	}

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

	logger.Debug().Printf("[%s] [%s] [%s#%s] [%s]", c.Name, mc.ID, mc.Author.Username, mc.Author.Discriminator, message)

	db, err := database.Get()
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}

	guild := api.GetGuild(ds, mc.GuildID)

	stmt, _ := db.DB().Prepare("INSERT INTO guilds (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?;")
	err = database.Execute(stmt, guild.ID, guild.Name, guild.Name)
	if err != nil {
		logger.Err().Print(err.Error())
	}

	stmt, _ = db.DB().Prepare("INSERT INTO channels (id, name) VALUES (?, ?) ON DUPLICATE KEY UPDATE name = ?;")
	err = database.Execute(stmt, c.ID, c.Name, c.Name)
	if err != nil {
		logger.Err().Print(err.Error())
	}

	stmt, _ = db.DB().Prepare("INSERT INTO messages (id, sender, content, guild_id, channel_id) VALUES (?, ?, ?, ?, ?);")
	err = database.Execute(stmt, mc.ID, mc.Author.Username+"#"+mc.Author.Discriminator, message, mc.GuildID, mc.ChannelID)
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
