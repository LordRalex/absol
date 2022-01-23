package messagereport

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"log"
	"os"
	"strings"
	"time"
)

type Module struct {
	api.Module
}

var guilds = make(map[string]string)

func (*Module) Load(ds *discordgo.Session) {
	maps := strings.Split(os.Getenv("MESSAGEREPORT_GUILDS"), ";")
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
			fmt.Printf("Registering %s for guild %s\n", reportOperation.Name, k)
			_, err := s.ApplicationCommandCreate(appId, k, reportOperation)
			if err != nil {
				log.Printf("Cannot create slash command %q: %v", reportOperation.Name, err)
			}
		}
	})

	ds.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.ApplicationCommandData().Name == reportOperation.Name {
			submitReport(s, i)
		}
	})
}

var reportOperation = &discordgo.ApplicationCommand{
	Name: "report-message",
	Type: discordgo.MessageApplicationCommand,
}

var appId = os.Getenv("APP_ID")

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
		if v.Name == message.ID {
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
			log.Printf("Error submitting report for %s: %s\n", message.ID, err.Error())
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

		_, err = s.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
			Content: "Report submitted by " + i.Member.Mention(),
			Embeds:  embeds,
			Components: []discordgo.MessageComponent{discordgo.Button{
				Emoji: discordgo.ComponentEmoji{
					Name: "",
				},
				Label: "Delete Message",
				Style: discordgo.LinkButton,
			}, discordgo.Button{
				Emoji: discordgo.ComponentEmoji{
					Name: "",
				},
				Label: "Mute User",
				Style: discordgo.LinkButton,
			}, discordgo.Button{
				Emoji: discordgo.ComponentEmoji{
					Name: "",
				},
				Label: "Ban User",
				Style: discordgo.LinkButton,
			}},
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Parse: []discordgo.AllowedMentionType{},
			},
		})

		if err != nil {
			log.Printf("Error submitting report for %s: %s\n", message.ID, err.Error())
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
			log.Printf("Error submitting report for %s: %s\n", message.ID, err.Error())
			_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Submitting report failed"})
			return
		}
	}

	_, _ = s.InteractionResponseEdit(appId, i.Interaction, &discordgo.WebhookEdit{Content: "Report submitted"})
}
