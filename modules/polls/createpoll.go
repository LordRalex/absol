package polls

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/database"
	"strconv"
	"strings"
	"time"
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
		{
			Name:        "timeout",
			Description: "How long the poll should be valid for (default is 1 day)",
			Type:        discordgo.ApplicationCommandOptionString,
			Required:    false,
		},
	},
}

func runCreateCommand(ds *discordgo.Session, i *discordgo.InteractionCreate) {
	_ = ds.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	commandData := i.ApplicationCommandData()

	var title string
	var description string
	var choices []string
	var timeout string
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
					msg := "Error downloading file: " + err.Error()
					_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
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
					msg := "Error downloading file: " + err.Error()
					_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
					return
				}
				choices = strings.Split(data, "\r\n")
			}
		case "timeout":
			{
				timeout = v.StringValue()
			}
		}
	}

	if len(choices) < 2 {
		msg := "You need at least 2 choices"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	if len(choices) > 15 {
		msg := "Limit of 15 choices"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	if hasDupes(choices) {
		msg := "Choices cannot repeat"
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	endDate := time.Now().AddDate(0, 0, 1)
	if timeout != "" {
		msg := "Timeout is invalid"
		if strings.HasSuffix(timeout, "d") {
			//parse as days
			part := strings.TrimSuffix(timeout, "d")
			numDays, err := strconv.Atoi(part)
			if err != nil {
				_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
				return
			}
			endDate = time.Now().AddDate(0, 0, numDays)
		} else {
			timer, err := time.ParseDuration(timeout)
			if err != nil {
				_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
				return
			}
			endDate = time.Now().Add(timer)
		}
	}

	endDate = endDate.UTC()

	for _, v := range choices {
		if len(v) > 50 {
			msg := "Choices can be at most 50 characters"
			_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
			return
		}
	}

	if description == "" {
		description = fmt.Sprintf("Poll ends <t:%d:R>", endDate.Unix())
	} else {
		description = fmt.Sprintf("%s\n\nPoll ends <t:%d:R>", description, endDate.Unix())
	}

	embeds := []*discordgo.MessageEmbed{{
		Title:       title,
		Description: description,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    i.Member.Nick,
			IconURL: i.Member.AvatarURL(""),
		},
	}}

	m := &discordgo.MessageSend{
		Embeds:     embeds,
		Components: splitToRows(choices),
	}

	db, err := database.Get()
	if err != nil {
		msg := "Error connecting to database: " + err.Error()
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	message, err := ds.ChannelMessageSendComplex(i.ChannelID, m)
	if err != nil {
		msg := "Error sending poll: " + err.Error()
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		return
	}

	err = db.Create(&Poll{Title: title, MessageId: message.ID, ChannelId: i.ChannelID, EndAt: endDate, Started: time.Now()}).Error
	if err != nil {
		msg := "Error saving poll to database: " + err.Error()
		_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
		_ = ds.ChannelMessageDelete(message.ChannelID, message.ID)
		return
	}

	msg := "Poll created"
	_, _ = ds.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &msg})
}
