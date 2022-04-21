package messagereport

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api/logger"
	"strings"
)

type InteractionAction struct {
	Function             func(action *InteractionAction, session *discordgo.Session, interaction *discordgo.Interaction, id *CustomId)
	ButtonText           string
	RequiresConfirmation bool
	Action               string
}

var confirmationAction *InteractionAction
var cancelConfirmationAction *InteractionAction
var deleteAction *InteractionAction
var muteAction *InteractionAction
var banAction *InteractionAction
var closeAction *InteractionAction

func init() {
	confirmationAction = &InteractionAction{
		Function:             sendConfirmation,
		ButtonText:           "Yes",
		RequiresConfirmation: false,
		Action:               "confirm",
	}
	cancelConfirmationAction = &InteractionAction{
		Function:             cancelConfirmation,
		ButtonText:           "No",
		RequiresConfirmation: false,
		Action:               "cancel",
	}
	deleteAction = &InteractionAction{
		Function:             deleteMessage,
		ButtonText:           "Delete Message",
		RequiresConfirmation: true,
		Action:               "delete",
	}
	muteAction = &InteractionAction{
		Function:             muteUser,
		ButtonText:           "Mute User",
		RequiresConfirmation: true,
		Action:               "mute",
	}
	banAction = &InteractionAction{
		Function:             banUser,
		ButtonText:           "Ban User",
		RequiresConfirmation: true,
		Action:               "ban",
	}
	closeAction = &InteractionAction{
		Function:             closeReport,
		ButtonText:           "Close Report",
		RequiresConfirmation: true,
		Action:               "close",
	}

	actions = []*InteractionAction{deleteAction, muteAction, banAction, closeAction}
}

var actions []*InteractionAction

func getAction(name string) *InteractionAction {
	for _, v := range actions {
		if v.RequiresConfirmation {
			if v.Action == name {
				return confirmationAction
			}
			if cancelConfirmationAction.Action+v.Action == name {
				return cancelConfirmationAction
			}
			if confirmationAction.Action+v.Action == name {
				return v
			}
		} else if v.Action == name {
			return v
		}
	}

	return nil
}

func getRawAction(name string) *InteractionAction {
	for _, v := range actions {
		if v.Action == name {
			return v
		}
	}

	return nil
}

func deleteMessage(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *CustomId) {
	err := s.ChannelMessageDelete(id.ChannelId, id.MessageId)
	if err != nil {
		_, _ = s.ChannelMessageSend(i.Message.ChannelID, "Failed to delete message, it may already be deleted")
		return
	} else {
		_, _ = s.ChannelMessageSend(i.Message.ChannelID, "Message deleted by "+i.Member.User.Username)
	}
}

func sendConfirmation(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *CustomId) {
	id.BaseMessageId = i.Message.ID

	confirm := id.Clone()
	confirm.Action = confirmationAction.Action + id.Action

	cancel := id.Clone()
	cancel.Action = cancelConfirmationAction.Action + id.Action

	m := &discordgo.MessageSend{
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

	_, _ = s.ChannelMessageSendComplex(i.ChannelID, m)
}

func muteUser(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *CustomId) {
}

func banUser(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *CustomId) {
}

func cancelConfirmation(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *CustomId) {
	err := s.ChannelMessageDelete(i.Message.ChannelID, i.Message.ID)
	if err != nil {
		logger.Err().Println(err.Error())
	}

	actionName := strings.TrimPrefix(id.Action, cancelConfirmationAction.Action)
	raw := getRawAction(actionName)
	if raw != nil {
		toggleButton(s, i.Message.ChannelID, id.BaseMessageId, raw.ButtonText)
	}
}

func closeReport(action *InteractionAction, s *discordgo.Session, i *discordgo.Interaction, id *CustomId) {
	_, _ = s.ChannelDelete(i.ChannelID)
}
