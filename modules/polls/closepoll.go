package polls

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/database"
	"github.com/lordralex/absol/api/logger"
	"gorm.io/gorm"
)

var closePollOperation = &discordgo.ApplicationCommand{
	Name:        "closepoll",
	Description: "Closes a poll",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "id",
			Description: "ID for the poll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
	},
}

func runCloseCommand(ds *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: uint64(discordgo.MessageFlagsEphemeral)},
	})

	messageId := i.ApplicationCommandData().Options[0].StringValue()

	originalMessage, err := ds.ChannelMessage(i.ChannelID, messageId)
	if err != nil {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Unable to get poll"})
		return
	}

	if originalMessage.Author.ID != ds.State.User.ID {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "This does not appear to be a poll"})
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

	for _, v := range edit.Components {
		if v.Type() == discordgo.ActionsRowComponent {
			row := v.(*discordgo.ActionsRow)
			for _, b := range row.Components {
				if b.Type() == discordgo.ButtonComponent {
					button := b.(*discordgo.Button)
					key := button.Label
					votes := &Vote{MessageId: poll.MessageId, Vote: key}
					var count int64
					db.Model(votes).Where(votes).Count(&count)
					button.Label = fmt.Sprintf("%s (%d)", button.Label, count)
					button.Disabled = true
				}
			}
		}
	}

	poll.Closed = true

	_ = db.Save(poll).Error

	_, _ = ds.ChannelMessageEditComplex(edit)
	_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Poll closed"})
}