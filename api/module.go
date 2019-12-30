package api

import "github.com/bwmarrin/discordgo"

type Module interface {
	Load(session *discordgo.Session)
}
