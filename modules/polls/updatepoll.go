package polls

import (
	"errors"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"gorm.io/gorm"
)

var updatePollOperation = &discordgo.ApplicationCommand{
	Name:        "updatepoll",
	Description: "Update a poll with a bunch of options",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "id",
			Description: "ID for the poll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        "title",
			Description: "Title for the poll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
		{
			Name:        "description",
			Description: "Full information about the poll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
		{
			Name:        "description-file",
			Description: "Full information about the poll (based from a file)",
			Type:        discordgo.ApplicationCommandOptionAttachment,
			Required:    false,
		},
	},
}

func runUpdateCommand(ds *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: uint64(discordgo.MessageFlagsEphemeral)},
	})

	commandData := i.ApplicationCommandData()

	var messageId string
	for _, v := range commandData.Options {
		if v.Name == "id" {
			messageId = v.StringValue()
			break
		}
	}

	originalMessage, err := ds.ChannelMessage(i.ChannelID, messageId)
	if err != nil {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Unable to get poll"})
		return
	}

	db, err := database.Get()
	if err != nil {
		logger.Err().Println(err.Error())
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Unable to get poll"})
		return
	}

	poll := &Poll{MessageId: messageId}
	err = db.Where(poll).First(poll).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Err().Println(err.Error())
		}
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Unable to get poll"})
		return
	}
	if poll.Closed {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Poll is already closed"})
		return
	}

	edit := discordgo.NewMessageEdit(originalMessage.ChannelID, originalMessage.ID)
	edit.Components = originalMessage.Components
	edit.Embeds = originalMessage.Embeds

	title := edit.Embeds[0].Title
	description := edit.Embeds[0].Description

	for _, v := range commandData.Options {
		switch v.Name {
		case "title":
			{
				title = v.StringValue()
			}
		case "description":
			{
				description = v.StringValue()
			}
		case "description-file":
			{
				fileId := v.Value.(string)
				description, err = downloadFile(commandData.Resolved.Attachments[fileId].URL)
				if err != nil {
					_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Error downloading file: " + err.Error()})
					return
				}
			}
		}
	}
	edit.Embeds[0].Title = title
	edit.Embeds[0].Description = description

	poll.Title = title
	_ = db.Save(poll).Error

	_, _ = ds.ChannelMessageEditComplex(edit)

	_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Poll updated"})
}
