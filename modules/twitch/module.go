package twitch

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lordralex/absol/api"
)

type Module struct {
	api.Module
}

func (tm *Module) Load(session *discordgo.Session) {
	api.RegisterCommand("twitchid", RunCommand)
}
