package messagereport

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
	"github.com/lordralex/absol/api/logger"
)

type InteractionAction struct {
	Function             func(action *InteractionAction, session *discordgo.Session, interaction *discordgo.Interaction, id *InteractionId)
	ButtonText           string
	RequiresConfirmation bool
	Action               string
}

var confirmationAction = &InteractionAction{
	Function:             sendConfirmation,
	ButtonText:           "Yes",
	RequiresConfirmation: false,
	Action:               "confirm",
}
var cancelConfirmationAction = &InteractionAction{
	Function:             cancelConfirmation,
	ButtonText:           "No",
	RequiresConfirmation: false,
	Action:               "cancel",
}
var deleteAction = &InteractionAction{
	Function:             deleteMessage,
	ButtonText:           "Delete Message",
	RequiresConfirmation: true,
	Action:               "delete",
}
var muteAction = &InteractionAction{
	Function:             confirmMuteUser,
	ButtonText:           "Mute User",
	RequiresConfirmation: true,
	Action:               "mute",
}
var banAction = &InteractionAction{
	Function:             confirmBanUser,
	ButtonText:           "Ban User",
	RequiresConfirmation: true,
	Action:               "ban",
}
var closeAction = &InteractionAction{
	Function:             closeReport,
	ButtonText:           "Close Report",
	RequiresConfirmation: true,
	Action:               "close",
}

var actions = []*InteractionAction{deleteAction, muteAction, banAction, closeAction}

func getAction(name string) *InteractionAction {
	for _, v := range actions {
		if v.RequiresConfirmation {
			if v.Action == name {
				return confirmationAction
			}
			if "cancel"+v.Action == name {
				return cancelConfirmationAction
			}
			if "confirm"+v.Action == name {
				return v
			}
		} else if v.Action == name {
			return v
		}
	}

	return nil
}

func deleteMessage(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *InteractionId) {
	err := s.ChannelMessageDelete(id.ChannelId, id.MessageId)
	if err != nil {
		_, _ = s.ChannelMessageSend(i.Message.ChannelID, "Failed to delete message, it may already be deleted")
		return
	} else {
		_, _ = s.ChannelMessageSend(i.Message.ChannelID, "Message deleted by "+i.Member.User.Username)
	}
	_ = s.InteractionResponseDelete(appId, i)

	toggleButton(s, i, "Delete Message")
}

func sendConfirmation(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *InteractionId) {
	confirm := id.Clone()
	confirm.Action = "confirm" + id.Action

	cancel := id.Clone()
	cancel.Action = "cancel" + id.Action

	m := &discordgo.WebhookEdit{
		Content: "Confirm " + id.Action,
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{discordgo.Button{
					CustomID: confirm.ToString(),
					Label:    confirmationAction.ButtonText,
					Style:    discordgo.DangerButton,
				}, discordgo.Button{
					CustomID: cancel.ToString(),
					Label:    cancelConfirmationAction.ButtonText,
					Style:    discordgo.SecondaryButton,
				}}},
		},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Parse: []discordgo.AllowedMentionType{},
		},
	}

	_, _ = s.InteractionResponseEdit(appId, i, m)
}

func confirmMuteUser(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *InteractionId) {
}

func confirmBanUser(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *InteractionId) {
}

func cancelConfirmation(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *InteractionId) {
	err := s.ChannelMessageDelete(i.Message.ChannelID, i.Message.ID)
	if err != nil {
		logger.Err().Println(err.Error())
	}
	_ = s.InteractionResponseDelete(appId, i)
}

func closeReport(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *InteractionId) {
	guild := api.GetGuild(s, i.GuildID)
	channels, err := s.GuildChannels(guild.ID)
	if err != nil {
		_, _ = s.InteractionResponseEdit(appId, i, &discordgo.WebhookEdit{Content: "Closing report failed"})
		return
	}

	for _, v := range channels {
		if v.ParentID == guilds[guild.ID] && v.Name == id.MessageId {
			_, _ = s.ChannelDelete(v.ID)
			break
		}
	}
}
