package twitch

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
)

type Module struct {
	api.Module
}

func (tm *Module) Load(session *discordgo.Session) {
	api.RegisterIntentNeed(discordgo.IntentsGuildMessages)

	api.RegisterCommand("twitchid", RunCommand)
	api.RegisterCommand("twitchname", RunCommand)
}

func (*Module) Name() string {
	return "twitch"
}
