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
				if strings.HasPrefix(id, "delete-") {
					deleteMessage(s, i.Interaction, id)
				} else if strings.HasPrefix(id, "mute-") {
					muteUser(s, i.Interaction, id)
				} else if strings.HasPrefix(id, "ban-") {
					banUser(s, i.Interaction, id)
				} else if strings.HasPrefix(id, "close-") {
					closeReport(s, i.Interaction, id)
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

	guild := api.GetGuild(s, i.GuildID)
	var channel *discordgo.Channel
	channels, err := s.GuildChannels(guild.ID)
	if err != nil {
		_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Submitting report failed"})
		return
	}

	for _, v := range channels {
		if v.ParentID == guilds[guild.ID] && v.Name == message.ID {
			channel = v
			break
		}
	}

	firstReport := true
	if channel == nil {
		channel, err = s.GuildChannelCreateComplex(i.GuildID, discordgo.GuildChannelCreateData{
			Name:     message.ID,
			Type:     discordgo.ChannelTypeGuildText,
			Topic:    "Reported Message - " + message.ID,
			ParentID: guilds[guild.ID],
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
			URL:         fmt.Sprintf("https://discord.com/channels/%s/%s/%s", guild.ID, message.ChannelID, message.ID),
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
						CustomID: "delete-" + message.ID + "-" + message.ChannelID,
						Label:    "Delete Message",
						Style:    discordgo.PrimaryButton,
					}, discordgo.Button{
						CustomID: "mute-" + message.ID + "-" + i.Member.User.ID,
						Label:    "Mute User",
						Style:    discordgo.SecondaryButton,
					}, discordgo.Button{
						CustomID: "ban-" + message.ID + "-" + i.Member.User.ID,
						Label:    "Ban User",
						Style:    discordgo.DangerButton,
					}}},
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{discordgo.Button{
						CustomID: "close-" + message.ID,
						Label:    "Close Report",
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

func deleteMessage(s *discordgo.Session, i *discordgo.Interaction, customId string) {
	data := strings.Split(customId, "-")
	channelId := data[2]
	messageId := data[1]

	err := s.ChannelMessageDelete(channelId, messageId)
	if err != nil {
		_, _ = s.InteractionResponseEdit(appId, i, &discordgo.WebhookEdit{Content: "Failed to delete message"})
	} else {
		_, _ = s.InteractionResponseEdit(appId, i, &discordgo.WebhookEdit{Content: "Message deleted"})
	}
}

func muteUser(s *discordgo.Session, i *discordgo.Interaction, customId string) {
	_, _ = s.InteractionResponseEdit(appId, i, &discordgo.WebhookEdit{Content: "Muting is not supported"})
}

func banUser(s *discordgo.Session, i *discordgo.Interaction, customId string) {
	_, _ = s.InteractionResponseEdit(appId, i, &discordgo.WebhookEdit{Content: "Banning is not supported"})
}

func closeReport(s *discordgo.Session, i *discordgo.Interaction, customId string) {
	messageId := strings.Split(customId, "-")[1]

	guild := api.GetGuild(s, i.GuildID)
	channels, err := s.GuildChannels(guild.ID)
	if err != nil {
		_, _ = s.InteractionResponseEdit(appId, i, &discordgo.WebhookEdit{Content: "Closing report failed"})
		return
	}

	for _, v := range channels {
		if v.ParentID == guilds[guild.ID] && v.Name == messageId {
			_, _ = s.ChannelDelete(v.ID)
			break
		}
	}
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
