package messagereport

import (
	"bytes"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"strings"
	"time"
)

type Module struct {
	api.Module
}

var guilds = make(map[string]string)
var client = &http.Client{}
var appId string

func (*Module) Load(ds *discordgo.Session) {
	appId = viper.GetString("app.id")

	maps := strings.Split(viper.GetString("MESSAGEREPORT_GUILDS"), ";")
	for _, v := range maps {
		if v == "" {
			continue
		}
		parts := strings.Split(v, ":")
		if len(parts) != 2 {
			continue
		}
		guilds[parts[0]] = parts[1]
	}

	ds.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		for k, _ := range guilds {
			logger.Out().Printf("Registering %s for guild %s\n", reportOperation.Name, k)
			_, err := s.ApplicationCommandCreate(appId, k, reportOperation)
			if err != nil {
				logger.Err().Printf("Cannot create slash command %q: %v", reportOperation.Name, err)
			}
		}
	})

	ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			{
				if i.ApplicationCommandData().Name == reportOperation.Name {
					submitReport(s, i)
				}
			}
		case discordgo.InteractionMessageComponent:
			{
				_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "Processing",
					},
				})
				id := i.Interaction.MessageComponentData().CustomID

				customId := &InteractionId{}
				customId.FromString(id)

				action := getAction(customId.Action)

				if action != nil {
					action.Function(action, s, i.Interaction, customId)
				} else {
					_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Unknown action"})
				}
			}
		}
	})
}

var reportOperation = &discordgo.ApplicationCommand{
	Name: "report-message",
	Type: discordgo.MessageApplicationCommand,
}

func getFile(url string) (io.Reader, error) {
	response, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(response.Body)
	return &buf, err
}

func getChannelForReport(s *discordgo.Session, guildId string, messageId string) (*discordgo.Channel, error) {
	guild := api.GetGuild(s, guildId)
	var channel *discordgo.Channel
	channels, err := s.GuildChannels(guild.ID)
	if err != nil {
		return nil, err
	}

	for _, v := range channels {
		if v.ParentID == guilds[guild.ID] && v.Name == messageId {
			channel = v
			break
		}
	}

	return channel, nil
}

func toggleButton(s *discordgo.Session, i *discordgo.Interaction, buttonText string) {
	originalMessage, err := s.ChannelMessage(i.Message.ChannelID, i.Message.ID)
	if err != nil {
		logger.Err().Println(err.Error())
		return
	}

	edit := discordgo.NewMessageEdit(originalMessage.ChannelID, originalMessage.ID)
	edit.Content = &originalMessage.Content
	edit.Components = originalMessage.Components
	edit.Embeds = originalMessage.Embeds

	for _, v := range edit.Components {
		if v.Type() == discordgo.ActionsRowComponent {
			row := v.(*discordgo.ActionsRow)
			for _, b := range row.Components {
				if b.Type() == discordgo.ButtonComponent {
					button := b.(*discordgo.Button)
					if button.Label == buttonText {
						button.Disabled = !button.Disabled
					}
				}
			}
		}
	}

	_, _ = s.ChannelMessageEditComplex(edit)
}

func submitReport(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error

	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Submitting report",
			Flags:   1 << 6,
		},
	})

	messageId := i.ApplicationCommandData().TargetID
	message, err := s.ChannelMessage(i.ChannelID, messageId)
	if err != nil {
		_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Submitting report failed"})
		return
	}

	channel, err := getChannelForReport(s, i.GuildID, messageId)
	if err != nil {
		_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Submitting report failed"})
		return
	}

	firstReport := true
	if channel == nil {
		channel, err = s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
			Name:     message.ID,
			Type:     discordgo.ChannelTypeGuildText,
			Topic:    "Reported Message - " + message.ID,
			ParentID: guilds[i.GuildID],
			NSFW:     false,
		})

		if err != nil {
			logger.Err().Printf("Error submitting report for %s: %s\n", message.ID, err.Error())
			_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Submitting report failed"})
			return
		}
	} else {
		firstReport = false
	}

	if firstReport {
		embeds := []*discordgo.MessageEmbed{{
			URL:         fmt.Sprintf("https://discord.com/channels/%s/%s/%s", i.GuildID, message.ChannelID, message.ID),
			Title:       fmt.Sprintf("Message Link"),
			Description: message.Content,
			Timestamp:   message.Timestamp.Format(time.RFC3339),
			Thumbnail:   &discordgo.MessageEmbedThumbnail{URL: message.Author.AvatarURL("")},
			Author: &discordgo.MessageEmbedAuthor{
				Name:    message.Author.Username,
				IconURL: message.Author.AvatarURL(""),
			}},
		}

		embeds = append(embeds, message.Embeds...)

		files := make([]*discordgo.File, 0)
		for _, v := range message.Attachments {
			file, err := getFile(v.URL)
			if err != nil {
				logger.Err().Printf("Error downloading %s: %s\n", v.URL, err.Error())
				continue
			}
			files = append(files, &discordgo.File{
				Name:        v.Filename,
				ContentType: "application/octet-stream",
				Reader:      file,
			})
		}

		m := &discordgo.MessageSend{
			Content: "Report submitted by " + i.Member.Mention(),
			Files:   files,
			Embeds:  embeds,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{discordgo.Button{
						CustomID: (&InteractionId{Action: deleteAction.Action, ChannelId: message.ChannelID, MessageId: message.ID, UserId: i.Member.User.ID}).ToString(),
						Label:    deleteAction.ButtonText,
						Style:    discordgo.PrimaryButton,
					}, discordgo.Button{
						CustomID: (&InteractionId{Action: muteAction.Action, MessageId: message.ID, UserId: i.Member.User.ID}).ToString(),
						Label:    muteAction.ButtonText,
						Style:    discordgo.SecondaryButton,
					}, discordgo.Button{
						CustomID: (&InteractionId{Action: banAction.Action, MessageId: message.ID, UserId: i.Member.User.ID}).ToString(),
						Label:    banAction.ButtonText,
						Style:    discordgo.DangerButton,
					}}},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{discordgo.Button{
						CustomID: (&InteractionId{Action: closeAction.Action, MessageId: message.ID}).ToString(),
						Label:    closeAction.ButtonText,
						Style:    discordgo.PrimaryButton,
					}}},
			},
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{},
			},
		}

		_, err = s.ChannelMessageSendComplex(channel.ID, m)

		if err != nil {
			logger.Err().Printf("Error submitting report for %s: %s\n", message.ID, err.Error())
			_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Submitting report failed"})
			return
		}
	} else {
		_, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
			Content: "Additional report submitted by " + i.Member.Mention(),
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{},
			},
		})
		if err != nil {
			logger.Err().Printf("Error submitting report for %s: %s\n", message.ID, err.Error())
			_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Submitting report failed"})
			return
		}
	}

	_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Report submitted"})
}
