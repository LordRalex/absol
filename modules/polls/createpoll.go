package polls

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/database"
	"strings"
)

var createPollOperation = &discordgo.ApplicationCommand{
	Name:        "createpoll",
	Description: "Create a poll with a bunch of options",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
		{
			Name:        "title",
			Description: "Title for the poll",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    true,
		},
		{
			Name:        "choices",
			Description: "Allowed choices",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
		{
			Name:        "choices-file",
			Description: "File containing allowed choices",
			Type:        discordgo.ApplicationCommandOptionAttachment,
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

func runCreateCommand(ds *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: uint64(discordgo.MessageFlagsEphemeral)},
	})

	commandData := i.ApplicationCommandData()

	var title string
	var description string
	var choices []string
	var err error

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
		case "choices":
			{
				choices = strings.Split(v.StringValue(), " ")
			}
		case "choices-file":
			{
				fileId := v.Value.(string)
				data, err := downloadFile(commandData.Resolved.Attachments[fileId].URL)
				if err != nil {
					_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Error downloading file: " + err.Error()})
					return
				}
				choices = strings.Split(data, "\r\n")
			}
		}
	}

	if len(choices) < 2 {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "You need at least 2 choices"})
	}

	if len(choices) > 15 {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Limit of 15 choices"})
		return
	}

	if hasDupes(choices) {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Choices cannot repeat"})
		return
	}

	for _, v := range choices {
		if len(v) > 50 {
			_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Choices can be at most 50 characters"})
			return
		}
	}

	embeds := []*discordgo.MessageEmbed{{
		Title:       title,
		Description: description,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    i.Member.Nick,
			IconURL: i.Member.AvatarURL(""),
		}},
	}

	m := &discordgo.MessageSend{
		Embeds:     embeds,
		Components: splitToRows(choices),
	}

	db, err := database.Get()
	if err != nil {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Error connecting to database: " + err.Error()})
		return
	}

	message, err := ds.ChannelMessageSendComplex(i.ChannelID, m)
	if err != nil {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Error sending poll: " + err.Error()})
		return
	}

	err = db.Create(&Poll{Title: title, MessageId: message.ID}).Error
	if err != nil {
		_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Error saving poll to database: " + err.Error()})
		_ = ds.ChannelMessageDelete(message.ChannelID, message.ID)
		return
	}

	_, _ = ds.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Poll created"})
}
