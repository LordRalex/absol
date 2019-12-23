package factoids

import (
	"github.com/bwmarrin/discordgo"
	"github.com/jinzhu/gorm"
	"github.com/lordralex/absol/database"
	"github.com/lordralex/absol/logger"
	"strings"
)

func RunCommand(ds *discordgo.Session, mc *discordgo.MessageCreate, c *discordgo.Channel, cmd string, args []string) {
	if len(args) == 0 {
		return
	}

	factoidName := args[0]

	db, err := database.Get()
	if err != nil {
		_, err = ds.ChannelMessageSend(c.ID, "Failed to connect to database")
		logger.Err().Printf("Failed to connect to database\n%s", err)
		return
	}

	var data factoid

	err = db.Where("name = ?", factoidName).First(&data).Error

	if err != nil && gorm.IsRecordNotFoundError(err) {
		_, err = ds.ChannelMessageSend(c.ID, "No factoid with the given name was found")
		return
	} else if err != nil {
		logger.Err().Printf("Failed to pull data from database\n%s", err)
		return
	}

	msg := data.Content
	msg = strings.Replace(msg, "[b]", "**", -1)
	msg = strings.Replace(msg, "[/b]", "**", -1)
	msg = strings.Replace(msg, "[u]", "__", -1)
	msg = strings.Replace(msg, "[/u]", "__", -1)
	msg = strings.Replace(msg, "[i]", "*", -1)
	msg = strings.Replace(msg, "[/i]", "*", -1)
	msg = strings.Replace(msg, ";;", "\n", -1)

	if strings.Contains(msg, "http") {
		msgsplit := strings.Split(msg, " ")
		for k, v := range msgsplit {
			if strings.HasPrefix(v, "http") {
				msgsplit[k] = "<" + v + ">"
			}
		}
		msg = strings.Join(msgsplit, " ")
	}

	if len(mc.Mentions) > 0 {
		//if we have an @, we'll add it to the message
		header := ""
		for _, v := range mc.Mentions {
			header += "<@" + v.ID + "> "
		}
		msg = header + "Please refer to the below information:\n" + msg
	}

	_ = ds.ChannelMessageDelete(c.ID, mc.ID)
	_, err = ds.ChannelMessageSend(c.ID, ">>> " + msg)
}

type factoid struct {
	Name    string `gorm:"name"`
	Content string `gorm:"content"`
}
