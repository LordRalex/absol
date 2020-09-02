package log

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
)

func OnMessageEdit(ds *discordgo.Session, mc *discordgo.MessageUpdate) {
	if mc.Author != nil && mc.Author.ID == ds.State.User.ID {
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

	db, err := database.Get()
	if err != nil {
		logger.Err().Print(err.Error())
		return
	}

	stmt, _ := db.DB().Prepare("INSERT INTO edits (message_id, old_content) SELECT id, content FROM messages WHERE id =?")
	err = database.Execute(stmt, mc.Message.ID)
	if err != nil {
		logger.Err().Print(err.Error())
	}

	stmt, _ = db.DB().Prepare("UPDATE messages SET content = ? WHERE id = ?")
	err = database.Execute(stmt, message, mc.Message.ID)
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
