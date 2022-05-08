package pastes

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
)

type Module struct {
	api.Module
}

func (*Module) Load(ds *discordgo.Session) {
	api.RegisterIntentNeed(discordgo.IntentsGuildMessages)
	ds.AddHandler(HandleMessage)
	if viper.GetString("paste.url") == "" {
		logger.Err().Fatal("Pastebin root url required to use pastes module!")
	}
	if viper.GetString("paste.guilds") == "" {
		logger.Err().Fatal("At least one guild ID is required to use pastes module!")
	}
}

func HandleMessage(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if len(mc.Attachments) <= 0 {
		return
	}
	used := false
	for _, item := range strings.Split(viper.GetString("paste.guilds"), ";") {
		if item == mc.GuildID {
			used = true
		}
	}
	if !used {
		return
	}
	rows := []discordgo.MessageComponent{}
	row := []discordgo.MessageComponent{}
	for _, element := range mc.Attachments {
		if element.ContentType == "text/plain; charset=utf-8" || element.ContentType == "application/json; charset=utf-8" || element.ContentType == "text/html; charset=utf-8" {
			btn := discordgo.Button{
				Emoji: discordgo.ComponentEmoji{
					Name: "📜",
				},
				Label: "View " + element.Filename,
				Style: discordgo.LinkButton,
				URL:   fmt.Sprintf("%s/%s/%s/%s", viper.GetString("paste.url"), mc.ChannelID, mc.ID, element.Filename),
			}
			row = append(row, btn)
			if len(row) >= 5 {
				rows = append(rows, discordgo.ActionsRow{Components: row})
				row = []discordgo.MessageComponent{}
			}
		}
	}
	if len(row) > 0 {
		rows = append(rows, discordgo.ActionsRow{Components: row})
	}
	if len(rows) <= 0 {
		return
	}
	msg := &discordgo.MessageSend{
		Content:         "Web version of files from <@" + mc.Author.ID + ">",
		Components:      rows,
		AllowedMentions: &discordgo.MessageAllowedMentions{},
		Reference:       mc.Reference(),
	}
	_, err := ds.ChannelMessageSendComplex(mc.ChannelID, msg)
	if err != nil {
		logger.Err().Println(err.Error())
		return
	}
}
