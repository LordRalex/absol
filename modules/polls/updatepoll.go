package polls

import "github.com/bwmarrin/discordgo"

var updatePollOperation = &discordgo.ApplicationCommand{
	Name:        "updatepoll",
	Description: "Update a poll with a bunch of options",
	Type:        discordgo.ChatApplicationCommand,
	Options: []*discordgo.ApplicationCommandOption{
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

	_ = ds.InteractionResponseDelete(appId, i.Interaction)
}
