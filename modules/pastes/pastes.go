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
	ds.AddHandler(onDelete)
	if viper.GetString("pastes.url") == "" {
		logger.Err().Fatal("Paste root url required to use pastes module!")
	}
}

func HandleMessage(ds *discordgo.Session, mc *discordgo.MessageCreate) {
	if len(mc.Attachments) <= 0 {
		return
	}
	used := false
	for _, item := range strings.Split(viper.GetString("pastes.guilds"), ";") {
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
		if isAcceptedFile(element) {
			btn := discordgo.Button{
				Emoji: discordgo.ComponentEmoji{
					Name: "📜",
				},
				Label: "View " + element.Filename,
				Style: discordgo.LinkButton,
				URL:   fmt.Sprintf("%s/%s/%s/%s", viper.GetString("pastes.url"), mc.ChannelID, element.ID, element.Filename),
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

func onDelete(ds *discordgo.Session, md *discordgo.MessageDelete) {
	deleteIfReferenced(ds, md.ChannelID, md.ID)
}

func deleteIfReferenced(ds *discordgo.Session, channel string, messageId string) {
	messages, err := ds.ChannelMessages(channel, 5, "", messageId, "")
	if err != nil {
		return
	}
	for _, message := range messages {
		if message.Author.ID == ds.State.User.ID {
			if message.MessageReference != nil {
				if message.MessageReference.MessageID == messageId {
					_ = ds.ChannelMessageDelete(channel, message.ID)
				}
			}
		}
	}
}

func isAcceptedFile(attachment *discordgo.MessageAttachment) bool {
	fileType := strings.Split(attachment.ContentType, ";")[0]
	fileType = strings.TrimSpace(fileType)

	if fileType == "text/plain" || fileType == "application/json" || fileType == "text/html" {
		return true
	}

	return false
}

func (Module) Name() string {
	return "pastes"
}
